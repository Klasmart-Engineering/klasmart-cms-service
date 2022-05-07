package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type AmsGradeConnectionService struct {
	AmsGradeService
}

type GradeFilter struct {
	ID             *UUIDFilter    `json:"id,omitempty" gqls:"id,omitempty"`
	Name           *StringFilter  `json:"name,omitempty" gqls:"name,omitempty"`
	Status         *StringFilter  `json:"status,omitempty" gqls:"status,omitempty"`
	System         *BooleanFilter `json:"system,omitempty" gqls:"system,omitempty"`
	OrganizationID *UUIDFilter    `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	CategoryID     *UUIDFilter    `json:"categoryId,omitempty" gqls:"categoryId,omitempty"`
	ClassID        *UUIDFilter    `json:"classId,omitempty" gqls:"classId,omitempty"`
	ProgramID      *UUIDFilter    `json:"programId,omitempty" gqls:"programId,omitempty"`
	FromGradeID    *UUIDFilter    `json:"fromGradeId,omitempty" gqls:"fromGradeId,omitempty"`
	ToGradeID      *UUIDFilter    `json:"toGradeId,omitempty" gqls:"toGradeId,omitempty"`
	AND            []GradeFilter  `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR             []GradeFilter  `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (GradeFilter) FilterName() FilterType {
	return GradeFilterType
}

func (GradeFilter) ConnectionName() ConnectionType {
	return GradesConnectionType
}

type GradeSummaryNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}
type GradeConnectionNode struct {
	ID        string           `json:"id" gqls:"id"`
	Name      string           `json:"name" gqls:"name"`
	Status    string           `json:"status" gqls:"status"`
	System    bool             `json:"system" gqls:"system"`
	FromGrade GradeSummaryNode `json:"fromGrade" gqls:"fromGrade"`
	ToGrade   GradeSummaryNode `json:"toGrade" gqls:"toGrade"`
}

type GradesConnectionEdge struct {
	Cursor string              `json:"cursor" gqls:"cursor"`
	Node   GradeConnectionNode `json:"node" gqls:"node"`
}

type GradesConnectionResponse struct {
	TotalCount int                    `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo     `json:"pageInfo" gqls:"pageInfo"`
	Edges      []GradesConnectionEdge `json:"edges" gqls:"edges"`
}

func (scs GradesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &scs.PageInfo
}
func (gcs AmsGradeConnectionService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Grade, error) {
	condition := NewCondition(options...)

	filter := GradeFilter{
		ProgramID: &UUIDFilter{
			Operator: UUIDOperator(OperatorTypeEq),
			Value:    UUID(programID),
		},
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		},
	}
	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	if condition.System.Valid {
		filter.System = &BooleanFilter{
			Operator: OperatorTypeEq,
			Value:    condition.System.Valid,
		}
	}

	var pages []GradesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get grade by program failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	if len(pages) == 0 {
		log.Warn(ctx, "grade is empty",
			log.Any("operator", operator),
			log.Any("filter", filter))
		return []*Grade{}, nil
	}

	grades := make([]*Grade, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, p := range pages {
		for _, v := range p.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Warn(ctx, "grade exists",
					log.Any("grade", v.Node),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
			obj := &Grade{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			grades = append(grades, obj)
		}
	}
	return grades, nil
}
func (gcs AmsGradeConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Grade, error) {
	condition := NewCondition(options...)
	filter := GradeFilter{
		Status: &StringFilter{Operator: StringOperator(OperatorTypeEq), Value: Active.String()},
		OR: []GradeFilter{
			{OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(operator.OrgID)}},
			{System: &BooleanFilter{Operator: OperatorTypeEq, Value: true}},
		},
	}
	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	if condition.System.Valid {
		filter.System = &BooleanFilter{Operator: OperatorTypeEq, Value: condition.System.Valid}
	}

	var pages []GradesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get grade by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	if len(pages) == 0 {
		log.Warn(ctx, "grade is empty",
			log.Any("operator", operator),
			log.Any("filter", filter))
		return []*Grade{}, nil
	}

	grades := make([]*Grade, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, p := range pages {
		for _, v := range p.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Warn(ctx, "grade exists",
					log.Any("grade", v.Node),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
			obj := &Grade{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			grades = append(grades, obj)
		}
	}
	return grades, nil
}
