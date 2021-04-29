package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type ReportPermissionChecker struct {
	Operator        *entity.Operator
	AllowTeacherIDs []string
}

func NewReportPermissionChecker(operator *entity.Operator) *ReportPermissionChecker {
	return &ReportPermissionChecker{Operator: operator}
}

func (c *ReportPermissionChecker) CheckTeachers(ctx context.Context, teacherIDs []string) (bool, error) {
	permissions := []external.PermissionName{
		external.ReportTeacherReports603,
		external.ReportViewReports610,
		external.ReportViewMySchoolReports611,
		external.ReportViewMyOrganizationsReports612,
		external.ReportViewMyReports614,
	}
	permissionsMap, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, c.Operator, permissions)
	if err != nil {
		return false, err
	}
	if !permissionsMap[external.ReportTeacherReports603] {
		return false, nil
	}
	if permissionsMap[external.ReportViewReports610] || permissionsMap[external.ReportViewMyOrganizationsReports612] {
		if err := c.SearchOrg(ctx); err != nil {
			log.Error(ctx, "CheckTeachers: c.SearchOrg: search failed",
				log.Err(err),
				log.Any("operator", c.Operator),
			)
			return false, err
		}
	}
	if permissionsMap[external.ReportViewReports610] || permissionsMap[external.ReportViewMySchoolReports611] {
		if err := c.SearchSchool(ctx); err != nil {
			log.Error(ctx, "CheckTeachers: c.SearchOrg: search failed",
				log.Err(err),
				log.Any("operator", c.Operator),
			)
			return false, err
		}
	}
	if permissionsMap[external.ReportViewMyReports614] {
		c.SearchMe(ctx)
	}
	return utils.HasSubset(c.AllowTeacherIDs, teacherIDs), nil
}

func (c *ReportPermissionChecker) SearchOrg(ctx context.Context) error {
	teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, c.Operator, c.Operator.OrgID)
	if err != nil {
		log.Error(ctx, "SearchOrg: external.GetTeacherServiceProvider().GetByOrganization: get failed",
			log.Err(err),
			log.Any("operator", c.Operator),
		)
		return err
	}
	for _, t := range teachers {
		c.AllowTeacherIDs = append(c.AllowTeacherIDs, t.ID)
	}
	return nil
}

func (c *ReportPermissionChecker) SearchSchool(ctx context.Context) error {
	schools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, c.Operator)
	if err != nil {
		log.Error(ctx, "SearchSchool: external.GetSchoolServiceProvider().GetByOperator: get failed",
			log.Err(err),
			log.Any("operator", c.Operator),
		)
		return err
	}
	var schoolIDs []string
	for _, school := range schools {
		schoolIDs = append(schoolIDs, school.ID)
	}
	schoolTeachersMap, err := external.GetTeacherServiceProvider().GetBySchools(ctx, c.Operator, schoolIDs)
	if err != nil {
		log.Error(ctx, "SearchSchool: external.GetTeacherServiceProvider().GetBySchools: get failed",
			log.Err(err),
			log.Any("operator", c.Operator),
		)
		return err
	}
	for _, teachers := range schoolTeachersMap {
		for _, t := range teachers {
			c.AllowTeacherIDs = append(c.AllowTeacherIDs, t.ID)
		}
	}
	return nil
}

func (c *ReportPermissionChecker) SearchMe(ctx context.Context) {
	c.AllowTeacherIDs = append(c.AllowTeacherIDs, c.Operator.UserID)
}
