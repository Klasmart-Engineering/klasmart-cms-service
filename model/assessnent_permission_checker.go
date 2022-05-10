package model

import (
	"context"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
)

type AssessmentPermissionChecker struct {
	operator              *entity.Operator
	allowStatusComplete   bool
	allowStatusInProgress bool
	allowTeacherIDs       []string
	allowPairs            []*entity.AssessmentAllowTeacherIDAndStatusPair
}

func NewAssessmentPermissionChecker(operator *entity.Operator) *AssessmentPermissionChecker {
	return &AssessmentPermissionChecker{operator: operator}
}

func (c *AssessmentPermissionChecker) SearchAllPermissions(ctx context.Context) error {
	permissionNames := []external.PermissionName{
		external.AssessmentViewCompletedAssessments414,
		external.AssessmentViewInProgressAssessments415,

		external.AssessmentViewOrgCompletedAssessments424,
		external.AssessmentViewOrgInProgressAssessments425,

		external.AssessmentViewSchoolCompletedAssessments426,
		external.AssessmentViewSchoolInProgressAssessments427,
	}
	permissionMap, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, c.operator, permissionNames)
	if err != nil {
		return err
	}

	if err := c.SearchOrgPermissions(ctx, permissionMap); err != nil {
		return err
	}

	if err := c.SearchSchoolPermissions(ctx, permissionMap); err != nil {
		return err
	}

	if err := c.SearchSelfPermissions(ctx, permissionMap); err != nil {
		return err
	}

	return nil
}

func (c *AssessmentPermissionChecker) SearchOrgPermissions(ctx context.Context, permission map[external.PermissionName]bool) error {
	c.allowStatusComplete = permission[external.AssessmentViewOrgCompletedAssessments424]
	c.allowStatusInProgress = permission[external.AssessmentViewOrgInProgressAssessments425]

	if c.allowStatusComplete || c.allowStatusInProgress {
		teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, c.operator, c.operator.OrgID)
		if err != nil {
			return err
		}

		for _, teacher := range teachers {
			c.allowTeacherIDs = append(c.allowTeacherIDs, teacher.ID)

			paris := &entity.AssessmentAllowTeacherIDAndStatusPair{
				TeacherID: teacher.ID,
			}

			if c.allowStatusComplete {
				paris.Status = append(paris.Status, entity.AssessmentStatusComplete)
			}
			if c.allowStatusInProgress {
				paris.Status = append(paris.Status, entity.AssessmentStatusInProgress)
			}

			c.allowPairs = append(c.allowPairs, paris)
		}
	}

	return nil
}

func (c *AssessmentPermissionChecker) SearchSchoolPermissions(ctx context.Context, permission map[external.PermissionName]bool) error {
	c.allowStatusComplete = permission[external.AssessmentViewSchoolCompletedAssessments426]
	c.allowStatusInProgress = permission[external.AssessmentViewSchoolInProgressAssessments427]

	if c.allowStatusComplete || c.allowStatusInProgress {
		schools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, c.operator)
		if err != nil {
			return err
		}
		var schoolIDs []string
		for _, school := range schools {
			schoolIDs = append(schoolIDs, school.ID)
		}
		schoolID2TeachersMap, err := external.GetTeacherServiceProvider().GetBySchools(ctx, c.operator, schoolIDs)
		if err != nil {
			return err
		}

		var teacherMap = make(map[string]struct{})
		for _, teachers := range schoolID2TeachersMap {
			for _, teacher := range teachers {
				if _, ok := teacherMap[teacher.ID]; ok {
					continue
				}

				teacherMap[teacher.ID] = struct{}{}
				c.allowTeacherIDs = append(c.allowTeacherIDs, teacher.ID)

				paris := &entity.AssessmentAllowTeacherIDAndStatusPair{
					TeacherID: teacher.ID,
				}

				if c.allowStatusComplete {
					paris.Status = append(paris.Status, entity.AssessmentStatusComplete)
				}
				if c.allowStatusInProgress {
					paris.Status = append(paris.Status, entity.AssessmentStatusInProgress)
				}

				c.allowPairs = append(c.allowPairs, paris)
			}
		}
	}

	return nil
}

func (c *AssessmentPermissionChecker) SearchSelfPermissions(ctx context.Context, permission map[external.PermissionName]bool) error {
	c.allowStatusComplete = permission[external.AssessmentViewCompletedAssessments414]
	c.allowStatusInProgress = permission[external.AssessmentViewInProgressAssessments415]

	if c.allowStatusComplete || c.allowStatusInProgress {
		paris := &entity.AssessmentAllowTeacherIDAndStatusPair{
			TeacherID: c.operator.UserID,
		}

		if c.allowStatusComplete {
			paris.Status = append(paris.Status, entity.AssessmentStatusComplete)
		}
		if c.allowStatusInProgress {
			paris.Status = append(paris.Status, entity.AssessmentStatusInProgress)
		}

		c.allowPairs = append(c.allowPairs, paris)
		c.allowTeacherIDs = append(c.allowTeacherIDs, c.operator.UserID)
	}

	return nil
}

func (c *AssessmentPermissionChecker) AllowTeacherIDs() []string {
	return c.allowTeacherIDs
}

func (c *AssessmentPermissionChecker) AllowPairs() []*entity.AssessmentAllowTeacherIDAndStatusPair {
	var result []*entity.AssessmentAllowTeacherIDAndStatusPair

	for _, item := range c.allowPairs {
		if len(item.Status) != 1 {
			continue
		}
		result = append(result, item)
	}

	return result
}

func (c *AssessmentPermissionChecker) AllowStatuses() []entity.AssessmentStatus {
	var result []entity.AssessmentStatus
	if c.allowStatusInProgress {
		result = append(result, entity.AssessmentStatusInProgress)
	}
	if c.allowStatusComplete {
		result = append(result, entity.AssessmentStatusComplete)
	}
	return result
}

func (c *AssessmentPermissionChecker) CheckTeacherIDs(ids []string) bool {
	allow := c.AllowTeacherIDs()
	for _, id := range ids {
		for _, a := range allow {
			if a == id {
				return true
			}
		}
	}
	return false
}

func (c *AssessmentPermissionChecker) CheckStatus(s entity.AssessmentStatus) bool {
	if !s.Valid() {
		return true
	}
	allow := c.AllowStatuses()
	for _, a := range allow {
		if a == s {
			return true
		}
	}
	return false
}

func (c *AssessmentPermissionChecker) HasP439(ctx context.Context) (bool, error) {
	hasP439, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.operator, external.AssessmentEditInProgressAssessment439)
	if err != nil {
		log.Error(ctx, "HasP439: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 439 failed",
			log.Err(err),
			log.Any("operator", c.operator),
			log.Any("c", c),
		)
		return false, err
	}
	return hasP439, nil
}
