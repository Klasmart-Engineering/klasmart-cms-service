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
	ID             *UUIDFilter    `gqls:"id,omitempty"`
	Name           *StringFilter  `gqls:"name,omitempty"`
	Status         *StringFilter  `gqls:"status,omitempty"`
	System         *BooleanFilter `gqls:"system,omitempty"`
	OrganizationID *UUIDFilter    `gqls:"organizationId,omitempty"`
	CategoryID     *UUIDFilter    `gqls:"categoryId,omitempty"`
	ClassID        *UUIDFilter    `gqls:"classId,omitempty"`
	ProgramID      *UUIDFilter    `gqls:"programId,omitempty"`
	FromGradeID    *UUIDFilter    `gqls:"fromGradeId,omitempty"`
	ToGradeID      *UUIDFilter    `gqls:"toGradeId,omitempty"`
	AND            []GradeFilter  `gqls:"AND,omitempty"`
	OR             []GradeFilter  `gqls:"OR,omitempty"`
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
	Name      string           `json:"name" gqls:"id"`
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

	var grades []*Grade
	var pages []GradesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get grade by program failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	for _, p := range pages {
		for _, v := range p.Edges {
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
		OrganizationID: &UUIDFilter{
			Operator: UUIDOperator(OperatorTypeEq),
			Value:    UUID(operator.OrgID),
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

	var grades []*Grade
	var pages []GradesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get grade by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	for _, p := range pages {
		for _, v := range p.Edges {
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
