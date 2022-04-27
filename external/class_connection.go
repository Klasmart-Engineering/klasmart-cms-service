package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
	condition := NewCondition(options...)
	filter := ClassFilter{
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		},
	}

	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	filter.OR = []ClassFilter{
		{StudentID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(userID)}},
		{TeacherID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(userID)}},
	}

	var pages []ClassesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get age by program failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	var classes []*Class
	for _, page := range pages {
		for _, v := range page.Edges {
			class := Class{
				ID:       v.Node.ID,
				Name:     v.Node.Name,
				Status:   APStatus(v.Node.Status),
				JoinType: IsTeaching, // IsStudy
			}
			classes = append(classes, &class)
		}
	}
	return classes, nil
}
func (accs AmsClassConnectionService) GetByUserIDs(ctx context.Context, operator *entity.Operator, userIDs []string, options ...APOption) (map[string][]*Class, error) {
	panic("implement me")
}
func (accs AmsClassConnectionService) GetByOrganizationIDs(ctx context.Context, operator *entity.Operator, orgIDs []string, options ...APOption) (map[string][]*Class, error) {
	condition := NewCondition(options...)
	filter := ClassFilter{
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		},
	}

	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	filterOrgs := make([]ClassFilter, 0, len(orgIDs))
	for _, id := range orgIDs {
		fo := ClassFilter{OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(id)}}
		filterOrgs = append(filterOrgs, fo)
	}
	filter.OR = filterOrgs
	var pages []ClassesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get class by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	var classes []*Class
	for _, page := range pages {
		for _, v := range page.Edges {
			class := Class{
				ID:       v.Node.ID,
				Name:     v.Node.Name,
				Status:   APStatus(v.Node.Status),
				JoinType: IsTeaching, // IsStudy
			}
			classes = append(classes, &class)
		}
	}

	panic("implement me")
	//return classes, nil
}
func (accs AmsClassConnectionService) GetBySchoolIDs(ctx context.Context, operator *entity.Operator, schoolIDs []string, options ...APOption) (map[string][]*Class, error) {
	condition := NewCondition(options...)
	filter := ClassFilter{
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		},
	}

	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	filterSchs := make([]ClassFilter, 0, len(schoolIDs))
	for _, id := range schoolIDs {
		vuuid := UUID(id)
		fo := ClassFilter{SchoolID: &UUIDExclusiveFilter{Operator: UUIDExclusiveOperator(OperatorTypeEq), Value: &vuuid}}
		filterSchs = append(filterSchs, fo)
	}
	filter.OR = filterSchs
	var pages []ClassesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get class by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	var classes []*Class
	for _, page := range pages {
		for _, v := range page.Edges {
			class := Class{
				ID:       v.Node.ID,
				Name:     v.Node.Name,
				Status:   APStatus(v.Node.Status),
				JoinType: IsTeaching, // IsStudy
			}
			classes = append(classes, &class)
		}
	}

	panic("implement me")
}

func (accs AmsClassConnectionService) GetOnlyUnderOrgClasses(ctx context.Context, operator *entity.Operator, orgID string) ([]*NullableClass, error) {

	filter := ClassFilter{
		//Status:         &StringFilter{Operator: StringOperator(OperatorTypeEq), Value: Active.String()},
		OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(orgID)},
		SchoolID:       &UUIDExclusiveFilter{Operator: UUIDExclusiveOperator(OperatorTypeIsNull)},
	}

	var pages []ClassesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get class by organization failed",
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
