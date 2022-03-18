package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

type AssessmentOrgPermission struct {
	OrgID  string
	Status entity.NullStrings
}
type AssessmentSchoolPermission struct {
	UserIDs []string
	Status  entity.NullStrings
}
type AssessmentMyPermission struct {
	UserID string
	Status entity.NullStrings
}

type AssessmentPermission struct {
	OrgPermission    AssessmentOrgPermission
	SchoolPermission AssessmentSchoolPermission
	MyPermission     AssessmentMyPermission

	allowStatusComplete   bool
	allowStatusInProgress bool
}

func (c *AssessmentPermission) SearchAllPermissions(ctx context.Context, op *entity.Operator) error {
	permissionNames := []external.PermissionName{
		external.AssessmentViewCompletedAssessments414,
		external.AssessmentViewInProgressAssessments415,

		external.AssessmentViewOrgCompletedAssessments424,
		external.AssessmentViewOrgInProgressAssessments425,

		external.AssessmentViewSchoolCompletedAssessments426,
		external.AssessmentViewSchoolInProgressAssessments427,
	}
	permissionMap, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, permissionNames)
	if err != nil {
		return err
	}

	completeStatus := v2.GetAssessmentStatusByReq()[v2.AssessmentStatusCompliantCompleted]
	notCompleteStatus := v2.GetAssessmentStatusByReq()[v2.AssessmentStatusCompliantNotCompleted]

	c.OrgPermission = AssessmentOrgPermission{
		OrgID: op.OrgID,
	}
	if permissionMap[external.AssessmentViewOrgCompletedAssessments424] {
		c.allowStatusComplete = true
		c.OrgPermission.Status.Strings = append(c.OrgPermission.Status.Strings, completeStatus...)
	}
	if permissionMap[external.AssessmentViewOrgInProgressAssessments425] {
		c.allowStatusInProgress = true
		c.OrgPermission.Status.Strings = append(c.OrgPermission.Status.Strings, notCompleteStatus...)
	}
	c.OrgPermission.Status.Valid = len(c.OrgPermission.Status.Strings) > 0
	if c.OrgPermission.Status.Valid {
		log.Debug(ctx, "user has org permission", log.Any("OrgPermission", c.OrgPermission))
		return nil
	}

	// school permission
	c.SchoolPermission = AssessmentSchoolPermission{}

	if permissionMap[external.AssessmentViewSchoolCompletedAssessments426] {
		c.allowStatusComplete = true
		c.SchoolPermission.Status.Strings = append(c.SchoolPermission.Status.Strings, completeStatus...)
	}
	if permissionMap[external.AssessmentViewSchoolInProgressAssessments427] {
		c.allowStatusInProgress = true
		c.SchoolPermission.Status.Strings = append(c.SchoolPermission.Status.Strings, notCompleteStatus...)
	}
	c.SchoolPermission.Status.Valid = len(c.SchoolPermission.Status.Strings) > 0

	if c.SchoolPermission.Status.Valid {
		schools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, op)
		if err != nil {
			return err
		}
		var schoolIDs = make([]string, len(schools))
		for i, school := range schools {
			schoolIDs[i] = school.ID
		}
		schoolID2TeachersMap, err := external.GetTeacherServiceProvider().GetBySchools(ctx, op, schoolIDs)
		if err != nil {
			return err
		}
		var teacherMap = make(map[string]struct{})
		for _, teachers := range schoolID2TeachersMap {
			for _, teacher := range teachers {
				if _, ok := teacherMap[teacher.ID]; ok {
					continue
				}
				c.SchoolPermission.UserIDs = append(c.SchoolPermission.UserIDs, teacher.ID)
				teacherMap[teacher.ID] = struct{}{}
			}
		}

		log.Debug(ctx, "user has school permission", log.Any("SchoolPermission", c.SchoolPermission))

		//return nil
	}

	// self permission
	c.MyPermission = AssessmentMyPermission{
		UserID: op.UserID,
	}
	if permissionMap[external.AssessmentViewCompletedAssessments414] {
		c.allowStatusComplete = true
		c.MyPermission.Status.Strings = append(c.MyPermission.Status.Strings, completeStatus...)
	}
	if permissionMap[external.AssessmentViewInProgressAssessments415] {
		c.allowStatusInProgress = true
		c.MyPermission.Status.Strings = append(c.MyPermission.Status.Strings, notCompleteStatus...)
	}
	c.MyPermission.Status.Valid = len(c.MyPermission.Status.Strings) > 0

	if !c.allowStatusComplete && !c.allowStatusInProgress {
		log.Warn(ctx, "user has no permission", log.Any("permissionMap", permissionMap), log.Any("permissionResult", c))
		return constant.ErrForbidden
	}

	log.Debug(ctx, "user has my permission", log.Any("MyPermission", c.MyPermission))

	return nil
}

func (c *AssessmentPermission) IsAllowEdit(ctx context.Context, op *entity.Operator) error {
	isAllow, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.AssessmentEditInProgressAssessment439)
	if err != nil {
		log.Error(ctx, "HasP439: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 439 failed",
			log.Err(err),
			log.Any("operator", op),
		)
		return err
	}
	if !isAllow {
		log.Warn(ctx, "user not has edit permission", log.Any("operator", op), log.String("permission", external.AssessmentEditInProgressAssessment439.String()))
		return constant.ErrForbidden
	}

	return nil
}
