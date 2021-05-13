package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type ReportPermissionChecker struct {
	operator              *entity.Operator
	allowTeacherIDs       []string
	hasMyOrgPermission    bool
	hasMySchoolPermission bool
	hasMyselfPermission   bool
	searched              bool
}

func NewReportPermissionChecker(operator *entity.Operator) *ReportPermissionChecker {
	return &ReportPermissionChecker{operator: operator}
}

func (c *ReportPermissionChecker) SearchAndCheckTeachers(ctx context.Context, teacherIDs []string) (bool, error) {
	if !c.searched {
		if err := c.Search(ctx); err != nil {
			log.Error(ctx, "SearchAndCheckTeachers: c.Search: search failed",
				log.Strings("teacher_ids", teacherIDs),
				log.Any("c", c),
			)
			return false, err
		}
	}
	return c.CheckTeachers(teacherIDs), nil
}

func (c *ReportPermissionChecker) Search(ctx context.Context) error {
	permissions := []external.PermissionName{
		external.ReportTeacherReports603,
		external.ReportViewReports610,
		external.ReportViewMySchoolReports611,
		external.ReportViewMyOrganizationsReports612,
		external.ReportViewMyReports614,
	}
	permissionsMap, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, c.operator, permissions)
	if err != nil {
		return err
	}
	if !permissionsMap[external.ReportTeacherReports603] {
		return nil
	}
	if permissionsMap[external.ReportViewReports610] || permissionsMap[external.ReportViewMyOrganizationsReports612] {
		c.hasMyOrgPermission = true
		if err := c.SearchOrg(ctx); err != nil {
			log.Error(ctx, "SearchAndCheckTeachers: c.SearchOrg: search failed",
				log.Err(err),
				log.Any("operator", c.operator),
			)
			return err
		}
	}
	if permissionsMap[external.ReportViewReports610] || permissionsMap[external.ReportViewMySchoolReports611] {
		c.hasMySchoolPermission = true
		if err := c.SearchSchool(ctx); err != nil {
			log.Error(ctx, "SearchAndCheckTeachers: c.SearchOrg: search failed",
				log.Err(err),
				log.Any("operator", c.operator),
			)
			return err
		}
	}
	if permissionsMap[external.ReportViewMyReports614] {
		c.hasMyselfPermission = true
		c.SearchMe(ctx)
	}
	c.searched = true
	return nil
}

func (c *ReportPermissionChecker) CheckTeachers(teacherIDs []string) bool {
	return utils.HasSubset(c.allowTeacherIDs, teacherIDs)
}

func (c *ReportPermissionChecker) SearchOrg(ctx context.Context) error {
	teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, c.operator, c.operator.OrgID)
	if err != nil {
		log.Error(ctx, "SearchOrg: external.GetTeacherServiceProvider().GetByOrganization: get failed",
			log.Err(err),
			log.Any("operator", c.operator),
		)
		return err
	}
	for _, t := range teachers {
		c.allowTeacherIDs = append(c.allowTeacherIDs, t.ID)
	}
	return nil
}

func (c *ReportPermissionChecker) SearchSchool(ctx context.Context) error {
	schools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, c.operator)
	if err != nil {
		log.Error(ctx, "SearchSchool: external.GetSchoolServiceProvider().GetByOperator: get failed",
			log.Err(err),
			log.Any("operator", c.operator),
		)
		return err
	}
	var schoolIDs []string
	for _, school := range schools {
		schoolIDs = append(schoolIDs, school.ID)
	}
	schoolTeachersMap, err := external.GetTeacherServiceProvider().GetBySchools(ctx, c.operator, schoolIDs)
	if err != nil {
		log.Error(ctx, "SearchSchool: external.GetTeacherServiceProvider().GetBySchools: get failed",
			log.Err(err),
			log.Any("operator", c.operator),
		)
		return err
	}
	for _, teachers := range schoolTeachersMap {
		for _, t := range teachers {
			c.allowTeacherIDs = append(c.allowTeacherIDs, t.ID)
		}
	}
	return nil
}

func (c *ReportPermissionChecker) SearchMe(ctx context.Context) {
	c.allowTeacherIDs = append(c.allowTeacherIDs, c.operator.UserID)
}

func (c *ReportPermissionChecker) HasMyOrgPermission() bool {
	return c.hasMyOrgPermission
}

func (c *ReportPermissionChecker) HasMySchoolPermission() bool {
	return c.hasMySchoolPermission
}

func (c *ReportPermissionChecker) HasMyselfPermission() bool {
	return c.hasMyselfPermission
}
