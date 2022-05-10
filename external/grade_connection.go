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

func (gcs AmsGradeConnectionService) pageNodes(ctx context.Context, operator *entity.Operator, pages []GradesConnectionResponse) []*Grade {
	if len(pages) == 0 {
		log.Warn(ctx, "pageNodes is empty",
			log.Any("operator", operator))
		return []*Grade{}
	}
	grades := make([]*Grade, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, edge := range page.Edges {
			if _, ok := exists[edge.Node.ID]; ok {
				log.Warn(ctx, "pageNodes: grade exist",
					log.Any("grade", edge.Node),
					log.Any("operator", operator))
				continue
			}
			exists[edge.Node.ID] = true
			obj := &Grade{
				ID:     edge.Node.ID,
				Name:   edge.Node.Name,
				Status: APStatus(edge.Node.Status),
				System: edge.Node.System,
			}
			grades = append(grades, obj)
		}
	}
	return grades
}

func (gcs AmsGradeConnectionService) NewGradeFilter(ctx context.Context, operator *entity.Operator, options ...APOption) *GradeFilter {
	condition := NewCondition(options...)
	var filter GradeFilter
	if condition.Status.Valid && condition.Status.Status != Ignore {
		filter.Status = &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    condition.Status.Status.String(),
		}
	} else if !condition.Status.Valid {
		filter.Status = &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		}
	}
	if condition.System.Valid {
		filter.System = &BooleanFilter{
			Operator: OperatorTypeEq,
			Value:    condition.System.Bool,
		}
	}
	return &filter
}

func (gcs AmsGradeConnectionService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Grade, error) {
	filter := gcs.NewGradeFilter(ctx, operator, options...)
	filter.ProgramID = &UUIDFilter{
		Operator: UUIDOperator(OperatorTypeEq),
		Value:    UUID(programID),
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

	grades := gcs.pageNodes(ctx, operator, pages)
	return grades, nil
}
func (gcs AmsGradeConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Grade, error) {
	filter := gcs.NewGradeFilter(ctx, operator, options...)
	filter.OR = []GradeFilter{
		{OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(operator.OrgID)}},
		{System: &BooleanFilter{Operator: OperatorTypeEq, Value: true}},
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

	grades := gcs.pageNodes(ctx, operator, pages)
	return grades, nil
}
