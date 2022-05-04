package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type ClassFilter struct {
	ID                *UUIDFilter          `json:"id,omitempty" gqls:"id,omitempty"`
	Name              *StringFilter        `json:"name,omitempty" gqls:"name,omitempty"`
	Status            *StringFilter        `json:"status,omitempty" gqls:"status,omitempty"`
	OrganizationID    *UUIDFilter          `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	AgeRangeValueFrom *AgeRangeValueFilter `json:"ageRangeValueFrom,omitempty" gqls:"ageRangeValueFrom,omitempty"`
	AgeRangeUnitFrom  *AgeRangeUnitFilter  `json:"ageRangeUnitFrom,omitempty" gqls:"ageRangeUnitFrom,omitempty"`
	AgeRangeValueTo   *AgeRangeValueFilter `json:"ageRangeValueTo,omitempty" gqls:"ageRangeValueTo,omitempty"`
	AgeRangeUnitTo    *AgeRangeUnitFilter  `json:"ageRangeUnitTo,omitempty" gqls:"ageRangeUnitTo,omitempty"`
	SchoolID          *UUIDExclusiveFilter `json:"schoolId,omitempty" gqls:"schoolId,omitempty"`
	GradeID           *UUIDFilter          `json:"gradeId,omitempty" gqls:"gradeId,omitempty"`
	SubjectID         *UUIDFilter          `json:"subjectId,omitempty" gqls:"subjectId,omitempty"`
	AcademicTermID    *UUIDFilter          `json:"academicTermId,omitempty" gqls:"academicTermId,omitempty"`
	StudentID         *UUIDFilter          `json:"studentId,omitempty" gqls:"studentId,omitempty"`
	TeacherID         *UUIDFilter          `json:"teacherId,omitempty" gqls:"teacherId,omitempty"`
	ProgramID         *UUIDFilter          `json:"programId,omitempty" gqls:"programId,omitempty"`
	AND               []ClassFilter        `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR                []ClassFilter        `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (ClassFilter) FilterName() FilterType {
	return ClassFilterType
}

func (ClassFilter) ConnectionName() ConnectionType {
	return ClassesConnectionType
}

type ClassConnectionNode struct {
	ID        string `json:"id" gqls:"id"`
	Name      string `json:"name" gqls:"name"`
	Status    string `json:"status" gqls:"status"`
	ShortCode string `json:"shortCode" gqls:"shortCode"`
}

type ClassConnectionEdge struct {
	Cursor string              `json:"cursor" gqls:"cursor"`
	Node   ClassConnectionNode `json:"node" gqls:"node"`
}

type ClassesConnectionResponse struct {
	TotalCount int                   `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo    `json:"pageInfo" gqls:"pageInfo"`
	Edges      []ClassConnectionEdge `json:"edges" gqls:"edges"`
}

func (ccs ClassesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &ccs.PageInfo
}

type AmsClassConnectionService struct {
	AmsClassService
}

func (accs AmsClassConnectionService) GetByUserID(ctx context.Context, operator *entity.Operator, userID string, options ...APOption) ([]*Class, error) {
	classesMap, err := accs.GetByUserIDs(ctx, operator, []string{userID}, options...)
	if err != nil {
		log.Error(ctx, "GetByUserID: GetByUserIDs failed",
			log.Err(err),
			log.String("user", userID))
		return nil, err
	}
	return classesMap[userID], nil
}

func (accs AmsClassConnectionService) getByTeacherIDs(ctx context.Context, operator *entity.Operator, IDs []string, options ...APOption) (map[string][]*Class, error) {
	result := make(map[string][]ClassesConnectionResponse)
	err := subPageQuery(ctx, operator, "userNode", "classesTeachingConnection", IDs, result)
	if err != nil {
		log.Error(ctx, "getByTeacherIDs: subPageQuery failed",
			log.Err(err),
			log.Strings("user_ids", IDs))
		return nil, err
	}

	condition := NewCondition(options...)
	classesMap := make(map[string][]*Class)
	for k, pages := range result {
		for _, page := range pages {
			for _, edge := range page.Edges {
				class := Class{
					ID:       edge.Node.ID,
					Name:     edge.Node.Name,
					Status:   APStatus(edge.Node.Status),
					JoinType: IsTeaching,
				}
				if condition.Status.Valid && condition.Status.Status != class.Status {
					continue
				} else if condition.Status.Valid && class.Status != Active {
					// only status = "Active" data is returned by default
					continue
				}
				classesMap[k] = append(classesMap[k], &class)
			}
		}
	}
	return classesMap, nil
}

func (accs AmsClassConnectionService) getByStudentIDs(ctx context.Context, operator *entity.Operator, IDs []string, options ...APOption) (map[string][]*Class, error) {
	result := make(map[string][]ClassesConnectionResponse)
	err := subPageQuery(ctx, operator, "userNode", "classesStudyingConnection", IDs, result)
	if err != nil {
		log.Error(ctx, "getByStudentIDs: subPageQuery failed",
			log.Err(err),
			log.Strings("user_ids", IDs))
		return nil, err
	}

	condition := NewCondition(options...)
	classesMap := make(map[string][]*Class)
	for k, pages := range result {
		for _, page := range pages {
			for _, edge := range page.Edges {
				class := Class{
					ID:       edge.Node.ID,
					Name:     edge.Node.Name,
					Status:   APStatus(edge.Node.Status),
					JoinType: IsTeaching,
				}
				if condition.Status.Valid && condition.Status.Status != class.Status {
					continue
				} else if condition.Status.Valid && class.Status != Active {
					// only status = "Active" data is returned by default
					continue
				}
				classesMap[k] = append(classesMap[k], &class)
			}
		}
	}
	return classesMap, nil
}

func (accs AmsClassConnectionService) GetByUserIDs(ctx context.Context, operator *entity.Operator, userIDs []string, options ...APOption) (map[string][]*Class, error) {
	IDs := utils.SliceDeduplicationExcludeEmpty(userIDs)
	var wg sync.WaitGroup
	wg.Add(2)
	var tErr, sErr error
	var teachClasses, studyClasses map[string][]*Class
	go func() {
		defer wg.Done()
		teachClasses, tErr = accs.getByTeacherIDs(ctx, operator, IDs, options...)
	}()
	go func() {
		defer wg.Done()
		studyClasses, sErr = accs.getByStudentIDs(ctx, operator, IDs, options...)
	}()
	wg.Wait()
	if tErr != nil {
		log.Error(ctx, "GetByUserIDs: subPageQuery failed",
			log.Err(tErr),
			log.Strings("user_ids", userIDs))
		return nil, tErr
	}
	if sErr != nil {
		log.Error(ctx, "GetByUserIDs: subPageQuery failed",
			log.Err(sErr),
			log.Strings("user_ids", userIDs))
		return nil, sErr
	}
	var classesMap map[string][]*Class
	for _, k := range IDs {
		if teachClasses[k] == nil && studyClasses[k] != nil {
			classesMap[k] = studyClasses[k]
		}
		if teachClasses[k] != nil && studyClasses[k] == nil {
			classesMap[k] = teachClasses[k]
		}
		if teachClasses[k] != nil && studyClasses[k] != nil {
			classesMap[k] = teachClasses[k]
			classesMap[k] = append(classesMap[k], studyClasses[k]...)
		}
	}
	return classesMap, nil
}
func (accs AmsClassConnectionService) GetByOrganizationIDs(ctx context.Context, operator *entity.Operator, orgIDs []string, options ...APOption) (map[string][]*Class, error) {
	result := make(map[string][]ClassesConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(orgIDs)
	err := subPageQuery(ctx, operator, "organizationNode", "classesConnection", IDs, result)
	if err != nil {
		log.Error(ctx, "GetByOrganizationIDs: subPageQuery failed",
			log.Err(err),
			log.Strings("organization_ids", orgIDs))
		return nil, err
	}

	condition := NewCondition(options...)
	classesMap := make(map[string][]*Class)
	for k, pages := range result {
		for _, page := range pages {
			for _, edge := range page.Edges {
				class := Class{
					ID:     edge.Node.ID,
					Name:   edge.Node.Name,
					Status: APStatus(edge.Node.Status),
				}
				if condition.Status.Valid && condition.Status.Status != class.Status {
					continue
				} else if condition.Status.Valid && class.Status != Active {
					// only status = "Active" data is returned by default
					continue
				}
				classesMap[k] = append(classesMap[k], &class)
			}
		}
	}
	return classesMap, nil
}

func (accs AmsClassConnectionService) GetBySchoolIDs(ctx context.Context, operator *entity.Operator, schoolIDs []string, options ...APOption) (map[string][]*Class, error) {
	result := make(map[string][]ClassesConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(schoolIDs)
	err := subPageQuery(ctx, operator, "schoolNode", "classesConnection", IDs, result)
	if err != nil {
		log.Error(ctx, "GetBySchoolIDs: subPageQuery failed",
			log.Err(err),
			log.Strings("school_ids", schoolIDs))
		return nil, err
	}

	condition := NewCondition(options...)
	classesMap := make(map[string][]*Class)
	for k, pages := range result {
		for _, page := range pages {
			for _, edge := range page.Edges {
				class := Class{
					ID:     edge.Node.ID,
					Name:   edge.Node.Name,
					Status: APStatus(edge.Node.Status),
				}
				if condition.Status.Valid && condition.Status.Status != class.Status {
					continue
				} else if condition.Status.Valid && class.Status != Active {
					// only status = "Active" data is returned by default
					continue
				}
				classesMap[k] = append(classesMap[k], &class)
			}
		}
	}
	return classesMap, nil
}

func (accs AmsClassConnectionService) GetOnlyUnderOrgClasses(ctx context.Context, operator *entity.Operator, orgID string) ([]*NullableClass, error) {

	filter := ClassFilter{
		OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(orgID)},
		SchoolID:       &UUIDExclusiveFilter{Operator: UUIDExclusiveOperator(OperatorTypeIsNull)},
	}

	var pages []ClassesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get only under org class failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	var classes []*NullableClass
	for _, page := range pages {
		for _, v := range page.Edges {
			class := Class{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
			}
			nullableClass := NullableClass{Class: class, StrID: class.ID, Valid: true}
			classes = append(classes, &nullableClass)
		}
	}

	return classes, nil
}
