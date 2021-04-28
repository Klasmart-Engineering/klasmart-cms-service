package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

type ReportPermissionChecker struct {
	Operator        *entity.Operator
	AllowTeacherIDs []string
}

func NewReportPermissionChecker(operator *entity.Operator) *ReportPermissionChecker {
	return &ReportPermissionChecker{Operator: operator}
}

func (c *ReportPermissionChecker) CheckTeachersPermission(ctx context.Context, teacherIDs []string) (bool, error) {
	ok, err := c.Check603(ctx)
	if err != nil {
		log.Error(ctx, "CheckTeachersPermission: c.Check603: check failed",
			log.Err(err),
			log.Any("operator", c.Operator),
		)
		return false, err
	}
	if !ok {
		log.Error(ctx, "CheckTeachersPermission: c.Check603: check failed",
			log.Err(err),
			log.Any("operator", c.Operator),
		)
		return false, nil
	}

	if err := c.SearchAll(ctx); err != nil {
		log.Error(ctx, "CheckTeachersPermission: c.SearchAll: check failed",
			log.Err(err),
			log.Any("operator", c.Operator),
		)
		return false, err
	}

	return c.CheckTeacherIDs(ctx, teacherIDs), nil
}

func (c *ReportPermissionChecker) SearchAll(ctx context.Context) error {
	searchFns := []func(context.Context) error{c.Search610, c.Search614, c.Search611, c.Search612}
	for _, fn := range searchFns {
		if err := fn(ctx); err != nil {
			log.Error(ctx, "SearchAll: search failed",
				log.Err(err),
				log.Any("operator", c.Operator),
			)
			return err
		}
	}
	return nil
}

func (c *ReportPermissionChecker) CheckTeacherIDs(ctx context.Context, tids []string) bool {
	for _, tid := range tids {
		for _, tid2 := range c.AllowTeacherIDs {
			if tid == tid2 {
				return true
			}
		}
	}
	return false
}

func (c *ReportPermissionChecker) Check603(ctx context.Context) (bool, error) {
	ok, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.Operator, external.ReportTeacherReports603)
	if err != nil {
		log.Error(ctx, "Check603: external.GetPermissionServiceProvider().HasOrganizationPermission: check failed",
			log.Err(err),
			log.Any("operator", c.Operator),
		)
		return false, err
	}
	return ok, nil
}

func (c *ReportPermissionChecker) Search610(ctx context.Context) error {
	ok, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.Operator, external.ReportViewReports610)
	if err != nil {
		log.Error(ctx, "Search610: external.GetPermissionServiceProvider().HasOrganizationPermission: check failed",
			log.Err(err),
			log.Any("operator", c.Operator),
		)
		return err
	}

	if ok {
		teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, c.Operator, c.Operator.OrgID)
		if err != nil {
			log.Error(ctx, "Search610: external.GetTeacherServiceProvider().GetByOrganization: get failed",
				log.Err(err),
				log.Any("operator", c.Operator),
			)
			return err
		}
		for _, t := range teachers {
			c.AllowTeacherIDs = append(c.AllowTeacherIDs, t.ID)
		}
		schools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, c.Operator)
		var schoolIDs []string
		for _, school := range schools {
			schoolIDs = append(schoolIDs, school.ID)
		}
		schoolTeachersMap, err := external.GetTeacherServiceProvider().GetBySchools(ctx, c.Operator, schoolIDs)
		if err != nil {
			log.Error(ctx, "Search610: external.GetTeacherServiceProvider().GetBySchools: get failed",
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
	}

	return nil
}

func (c *ReportPermissionChecker) Search614(ctx context.Context) error {
	ok, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.Operator, external.ReportViewMyReports614)
	if err != nil {
		log.Error(ctx, "Search614: external.GetPermissionServiceProvider().HasOrganizationPermission:  search failed",
			log.Err(err),
			log.Any("operator", c.Operator),
		)
		return err
	}
	if ok {
		c.AllowTeacherIDs = append(c.AllowTeacherIDs, c.Operator.UserID)
	}
	return nil
}

func (c *ReportPermissionChecker) Search611(ctx context.Context) error {
	ok, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.Operator, external.ReportViewMySchoolReports611)
	if err != nil {
		log.Error(ctx, "Search611: external.GetPermissionServiceProvider().HasOrganizationPermission: check failed",
			log.Err(err),
			log.Any("operator", c.Operator),
		)
		return err
	}

	if ok {
		schools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, c.Operator)
		if err != nil {
			log.Error(ctx, "Search611: external.GetSchoolServiceProvider().GetByOperator: get failed",
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
			log.Error(ctx, "Search611: external.GetTeacherServiceProvider().GetBySchools: get failed",
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
	}

	return nil
}

func (c *ReportPermissionChecker) Search612(ctx context.Context) error {
	ok, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.Operator, external.ReportViewMyOrganizationsReports612)
	if err != nil {
		log.Error(ctx, "Search612: external.GetPermissionServiceProvider().HasOrganizationPermission: check failed",
			log.Err(err),
			log.Any("operator", c.Operator),
		)
		return err
	}

	if ok {
		teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, c.Operator, c.Operator.OrgID)
		if err != nil {
			log.Error(ctx, "Search612: external.GetTeacherServiceProvider().GetByOrganization: get failed",
				log.Err(err),
				log.Any("operator", c.Operator),
			)
			return err
		}
		for _, t := range teachers {
			c.AllowTeacherIDs = append(c.AllowTeacherIDs, t.ID)
		}
	}

	return nil
}
