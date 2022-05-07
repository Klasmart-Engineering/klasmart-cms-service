package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ProgramFilter struct {
	ID             *UUIDFilter         `json:"id,omitempty" gqls:"id,omitempty"`
	Name           *StringFilter       `json:"name,omitempty" gqls:"name,omitempty"`
	Status         *StringFilter       `json:"status,omitempty" gqls:"status,omitempty"`
	System         *BooleanFilter      `json:"system,omitempty" gqls:"system,omitempty"`
	OrganizationID *UUIDFilter         `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	GradeID        *UUIDFilter         `json:"gradeId,omitempty" gqls:"gradeId,omitempty"`
	AgeRangeFrom   *AgeRangeTypeFilter `json:"ageRangeFrom,omitempty" gqls:"ageRangeFrom,omitempty"`
	AgeRangeTo     *AgeRangeTypeFilter `json:"ageRangeTo,omitempty" gqls:"ageRangeTo,omitempty"`
	SubjectID      *UUIDFilter         `json:"subjectId,omitempty" gqls:"subjectId,omitempty"`
	SchoolID       *UUIDFilter         `json:"schoolId,omitempty" gqls:"schoolId,omitempty"`
	ClassID        *UUIDFilter         `json:"classId,omitempty" gqls:"classId,omitempty"`
	AND            []ProgramFilter     `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR             []ProgramFilter     `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (ProgramFilter) FilterName() FilterType {
	return ProgramFilterType
}

func (ProgramFilter) ConnectionName() ConnectionType {
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
		Status: &StringFilter{Operator: StringOperator(OperatorTypeEq), Value: Active.String()},
		OR: []ProgramFilter{
			{OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(operator.OrgID)}},
			{System: &BooleanFilter{Operator: OperatorTypeEq, Value: true}},
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

	var pages []ProgramsConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get programs by ids failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}

	if len(pages) == 0 {
		log.Warn(ctx, "program is empty",
			log.Any("operator", operator),
			log.Any("filter", filter))
		return []*Program{}, nil
	}

	programs := make([]*Program, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, p := range pages {
		for _, v := range p.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Warn(ctx, "program exist",
					log.Any("program", v),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
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
