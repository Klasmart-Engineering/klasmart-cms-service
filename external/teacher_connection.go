package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type AmsTeacherConnectionService struct {
	AmsTeacherService
}

func (ascs AmsTeacherConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, organizationID string) ([]*Teacher, error) {
	teacherMap, err := ascs.GetByOrganizations(ctx, operator, []string{organizationID})
	if err != nil {
		log.Error(ctx, "GetByOrganization: GetByOrganizations failed",
			log.Err(err),
			log.String("organization", organizationID))
		return nil, err
	}
	return teacherMap[organizationID], nil
}
func (ascs AmsTeacherConnectionService) GetByOrganizations(ctx context.Context, operator *entity.Operator, organizationIDs []string) (map[string][]*Teacher, error) {
	classes, err := GetClassServiceProvider().GetByOrganizationIDs(ctx, operator, organizationIDs)
	if err != nil {
		log.Error(ctx, "GetByOrganizations: get classes failed",
			log.Err(err),
			log.Strings("organization_ids", organizationIDs))
		return nil, err
	}
	clsOrgMap := make(map[string]string)
	var clsIDs []string
	for k, v := range classes {
		for _, cls := range v {
			clsOrgMap[cls.ID] = k
			clsIDs = append(clsIDs, cls.ID)
		}
	}
	if len(clsIDs) == 0 {
		log.Debug(ctx, "GetByOrganizations: class is empty",
			log.Strings("organization_ids", organizationIDs))
		return map[string][]*Teacher{}, nil
	}
	classTeachers, err := ascs.GetByClasses(ctx, operator, clsIDs)
	if err != nil {
		log.Error(ctx, "GetByOrganizations: GetByClasses failed",
			log.Err(err),
			log.Strings("organization_ids", organizationIDs))
		return nil, err
	}
	organizationTeachers := make(map[string][]*Teacher)
	for k, v := range classTeachers {
		if org, ok := clsOrgMap[k]; ok {
			organizationTeachers[org] = append(organizationTeachers[org], v...)
		}
	}
	return organizationTeachers, nil
}
func (ascs AmsTeacherConnectionService) GetByClasses(ctx context.Context, operator *entity.Operator, classIDs []string) (map[string][]*Teacher, error) {
	result := make(map[string][]UsersConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(classIDs)
	err := subPageQuery(ctx, operator, "classNode", "teachersConnection", IDs, result)
	if err != nil {
		log.Error(ctx, "GetByClassIDs: subPageQuery failed",
			log.Err(err),
			log.Strings("class_ids", classIDs))
		return nil, err
	}

	teachersMap := make(map[string][]*Teacher)
	for k, pages := range result {
		for _, page := range pages {
			for _, edge := range page.Edges {
				teacher := &Teacher{
					ID:   edge.Node.ID,
					Name: edge.Node.GivenName,
				}
				teachersMap[k] = append(teachersMap[k], teacher)
			}
		}
	}
	return teachersMap, nil
}
func (ascs AmsTeacherConnectionService) GetBySchool(ctx context.Context, operator *entity.Operator, schoolID string) ([]*Teacher, error) {
	result, err := ascs.GetBySchools(ctx, operator, []string{schoolID})
	if err != nil {
		log.Error(ctx, "GetBySchool: subPageQuery failed",
			log.Err(err),
			log.String("school_id", schoolID))
		return nil, err
	}
	return result[schoolID], nil
}
func (ascs AmsTeacherConnectionService) GetBySchools(ctx context.Context, operator *entity.Operator, schoolIDs []string) (map[string][]*Teacher, error) {
	result := make(map[string][]ClassesConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(schoolIDs)
	err := subPageQuery(ctx, operator, "schoolNode", "classesConnection", IDs, result)
	if err != nil {
		log.Error(ctx, "GetBySchools: subPageQuery failed",
			log.Err(err),
			log.Strings("school_ids", schoolIDs))
		return nil, err
	}
	classSchoolMap := make(map[string]string)
	var classIDs []string
	for k, pages := range result {
		for _, page := range pages {
			for _, edge := range page.Edges {
				classSchoolMap[edge.Node.ID] = k
				classIDs = append(classIDs, edge.Node.ID)
			}
		}
	}

	classTeacherMap, err := ascs.GetByClasses(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "GetBySchools: GetByClasses failed",
			log.Err(err),
			log.Strings("class_ids", classIDs),
			log.Strings("school_ids", schoolIDs))
		return nil, err
	}

	teachersMap := make(map[string][]*Teacher)
	for k, v := range classTeacherMap {
		school := classSchoolMap[k]
		existMap := make(map[string]bool)
		for _, teacher := range v {
			if _, ok := existMap[teacher.ID]; !ok {
				existMap[teacher.ID] = true
				teachersMap[school] = append(teachersMap[school], teacher)
			}
		}
	}

	return teachersMap, nil
}
