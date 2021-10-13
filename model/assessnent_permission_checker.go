package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
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
	hasP424 := permission[external.AssessmentViewOrgCompletedAssessments424]
	hasP425 := permission[external.AssessmentViewOrgInProgressAssessments425]

	if hasP424 || hasP425 {
		teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, c.operator, c.operator.OrgID)
		if err != nil {
			log.Error(ctx, "SearchOrgPermissions: external.GetTeacherServiceProvider().GetByOrganization: get teachers failed",
				log.Err(err),
				log.Any("c", c),
			)
			return err
		}
		var teacherIDs []string
		for _, teacher := range teachers {
			teacherIDs = append(teacherIDs, teacher.ID)
		}
		c.allowTeacherIDs = append(c.allowTeacherIDs, teacherIDs...)

		if hasP424 {
			c.allowStatusComplete = true
			for _, teacherID := range teacherIDs {
				c.allowPairs = append(c.allowPairs, &entity.AssessmentAllowTeacherIDAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusComplete,
				})
			}
		}

		if hasP425 {
			c.allowStatusInProgress = true
			for _, teacherID := range teacherIDs {
				c.allowPairs = append(c.allowPairs, &entity.AssessmentAllowTeacherIDAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusInProgress,
				})
			}
		}
	}

	return nil
}

func (c *AssessmentPermissionChecker) SearchSchoolPermissions(ctx context.Context, permission map[external.PermissionName]bool) error {
	hasP426 := permission[external.AssessmentViewSchoolCompletedAssessments426]
	hasP427 := permission[external.AssessmentViewSchoolInProgressAssessments427]

	if hasP426 || hasP427 {
		schools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, c.operator)
		if err != nil {
			log.Error(ctx, "SearchSchoolPermissions: external.GetSchoolServiceProvider().GetByOperator: get schools failed",
				log.Err(err),
				log.Any("c", c),
			)
			return err
		}
		var schoolIDs []string
		for _, school := range schools {
			schoolIDs = append(schoolIDs, school.ID)
		}
		schoolID2TeachersMap, err := external.GetTeacherServiceProvider().GetBySchools(ctx, c.operator, schoolIDs)
		if err != nil {
			log.Error(ctx, "SearchSchoolPermissions: external.GetTeacherServiceProvider().GetBySchools: get teachers failed",
				log.Err(err),
				log.Any("c", c),
				log.Any("school_ids", schoolIDs),
			)
			return err
		}

		var teacherIDs []string
		for _, teachers := range schoolID2TeachersMap {
			for _, teacher := range teachers {
				teacherIDs = append(teacherIDs, teacher.ID)
			}
		}
		c.allowTeacherIDs = append(c.allowTeacherIDs, teacherIDs...)

		if hasP426 {
			c.allowStatusComplete = true
			for _, teacherID := range c.allowTeacherIDs {
				c.allowPairs = append(c.allowPairs, &entity.AssessmentAllowTeacherIDAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusComplete,
				})
			}
		}

		if hasP427 {
			c.allowStatusInProgress = true
			for _, teacherID := range teacherIDs {
				c.allowPairs = append(c.allowPairs, &entity.AssessmentAllowTeacherIDAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusInProgress,
				})
			}
		}
	}

	return nil
}

func (c *AssessmentPermissionChecker) SearchSelfPermissions(ctx context.Context, permission map[external.PermissionName]bool) error {
	hasP414 := permission[external.AssessmentViewCompletedAssessments414]
	hasP415 := permission[external.AssessmentViewInProgressAssessments415]

	if hasP414 || hasP415 {
		if hasP414 {
			c.allowStatusComplete = true
			c.allowPairs = append(c.allowPairs, &entity.AssessmentAllowTeacherIDAndStatusPair{
				TeacherID: c.operator.UserID,
				Status:    entity.AssessmentStatusComplete,
			})
		}
		if hasP415 {
			c.allowStatusInProgress = true
			c.allowPairs = append(c.allowPairs, &entity.AssessmentAllowTeacherIDAndStatusPair{
				TeacherID: c.operator.UserID,
				Status:    entity.AssessmentStatusInProgress,
			})
		}
		c.allowTeacherIDs = append(c.allowTeacherIDs, c.operator.UserID)
	}

	return nil
}

func (c *AssessmentPermissionChecker) AllowTeacherIDs() []string {
	return c.allowTeacherIDs
}

func (c *AssessmentPermissionChecker) AllowPairs() []*entity.AssessmentAllowTeacherIDAndStatusPair {
	var result []*entity.AssessmentAllowTeacherIDAndStatusPair

	m := map[string]*struct {
		allowStatusInProgress bool
		allowStatusComplete   bool
	}{}
	for _, pair := range c.allowPairs {
		if m[pair.TeacherID] == nil {
			m[pair.TeacherID] = &struct {
				allowStatusInProgress bool
				allowStatusComplete   bool
			}{}
		}
		switch pair.Status {
		case entity.AssessmentStatusComplete:
			m[pair.TeacherID].allowStatusComplete = true
		case entity.AssessmentStatusInProgress:
			m[pair.TeacherID].allowStatusInProgress = true
		}
	}

	for teacherID, statuses := range m {
		if statuses.allowStatusInProgress && statuses.allowStatusComplete {
			continue
		}
		if !statuses.allowStatusInProgress && !statuses.allowStatusComplete {
			continue
		}
		if statuses.allowStatusComplete {
			result = append(result, &entity.AssessmentAllowTeacherIDAndStatusPair{
				TeacherID: teacherID,
				Status:    entity.AssessmentStatusComplete,
			})
		}
		if statuses.allowStatusInProgress {
			result = append(result, &entity.AssessmentAllowTeacherIDAndStatusPair{
				TeacherID: teacherID,
				Status:    entity.AssessmentStatusInProgress,
			})
		}
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
