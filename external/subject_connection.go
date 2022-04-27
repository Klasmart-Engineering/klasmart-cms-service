package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type SubjectFilter struct {
	ID             *UUIDFilter     `json:"id,omitempty" gqls:"id,omitempty"`
	Name           *StringFilter   `json:"name,omitempty" gqls:"name,omitempty"`
	Status         *StringFilter   `json:"status,omitempty" gqls:"status,omitempty"`
	System         *BooleanFilter  `json:"system,omitempty" gqls:"system,omitempty"`
	OrganizationID *UUIDFilter     `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	CategoryId     *UUIDFilter     `json:"categoryId,omitempty" gqls:"categoryId,omitempty"`
	ClassID        *UUIDFilter     `json:"classId,omitempty" gqls:"classId,omitempty"`
	ProgramID      *UUIDFilter     `json:"programId,omitempty" gqls:"programId,omitempty"`
	AND            []SubjectFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR             []SubjectFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (SubjectFilter) FilterName() FilterType {
	return SubjectFilterType
}

func (SubjectFilter) ConnectionName() ConnectionType {
	return SubjectsConnectionType
}

type SubjectConnectionNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}

type SubjectsConnectionEdge struct {
	Cursor string                `json:"cursor" gqls:"cursor"`
	Node   SubjectConnectionNode `json:"node" gqls:"node"`
}

type SubjectsConnectionResponse struct {
	TotalCount int                      `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo       `json:"pageInfo" gqls:"pageInfo"`
	Edges      []SubjectsConnectionEdge `json:"edges" gqls:"edges"`
}

func (scr SubjectsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &scr.PageInfo
}

type AmsSubjectConnectionService struct {
	AmsSubjectService
}

func (scs AmsSubjectConnectionService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Subject, error) {
	condition := NewCondition(options...)

	filter := SubjectFilter{
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
	var subjects []*Subject
	var pages []SubjectsConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get subject by program failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	for _, p := range pages {
		for _, v := range p.Edges {
			obj := &Subject{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			subjects = append(subjects, obj)
		}
	}
	return subjects, nil
}
func (scs AmsSubjectConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Subject, error) {
	condition := NewCondition(options...)

	filter := SubjectFilter{
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
	var subjects []*Subject
	var pages []SubjectsConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get subject by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	for _, p := range pages {
		for _, v := range p.Edges {
			obj := &Subject{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			subjects = append(subjects, obj)
		}
	}
	return subjects, nil
}
