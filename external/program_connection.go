package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ProgramFilter struct {
	ID             *UUIDFilter         `gqls:"id,omitempty"`
	Name           *StringFilter       `gqls:"name,omitempty"`
	Status         *StringFilter       `gqls:"status,omitempty"`
	System         *BooleanFilter      `gqls:"system,omitempty"`
	OrganizationID *UUIDFilter         `gqls:"organizationId,omitempty"`
	GradeID        *UUIDFilter         `gqls:"gradeId,omitempty"`
	AgeRangeFrom   *AgeRangeTypeFilter `gqls:"ageRangeFrom,omitempty"`
	AgeRangeTo     *AgeRangeTypeFilter `gqls:"ageRangeTo,omitempty"`
	SubjectID      *UUIDFilter         `gqls:"subjectId,omitempty"`
	SchoolID       *UUIDFilter         `gqls:"schoolId,omitempty"`
	ClassID        *UUIDFilter         `gqls:"classId,omitempty"`
	AND            []ProgramFilter     `gqls:"AND,omitempty"`
	OR             []ProgramFilter     `gqls:"OR,omitempty"`
}

func (ProgramFilter) FilterType() FilterOfType {
	return ProgramsConnectionType
}

type ProgramConnectionNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}
type ProgramsConnectionEdge struct {
	Cursor string                `json:"cursor" gqls:"cursor"`
	Node   ProgramConnectionNode `json:"node" gqls:"node"`
}

type ProgramsConnectionResponse struct {
	TotalCount int                      `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo       `json:"pageInfo" gqls:"pageInfo"`
	Edges      []ProgramsConnectionEdge `json:"edges" gqls:"edges"`
}

func (pcs ProgramsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &pcs.PageInfo
}

type AmsProgramConnectionService struct {
	AmsProgramService
}

func (pcs AmsProgramConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Program, error) {
	condition := NewCondition(options...)

	filter := ProgramFilter{
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

	var programs []*Program
	var pages []ProgramsConnectionResponse
	err := pageQuery(ctx, operator, filter.FilterType(), filter, &pages)
	if err != nil {
		log.Error(ctx, "get programs by ids failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	for _, p := range pages {
		for _, v := range p.Edges {
			obj := &Program{
				ID:   v.Node.ID,
				Name: v.Node.Name,
				//GroupName:
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			programs = append(programs, obj)
		}
	}
	return programs, nil
}
