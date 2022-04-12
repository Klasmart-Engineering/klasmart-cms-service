package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type AmsAgeConnectionService struct {
	AmsAgeService
}

type AgeRangeFilter struct {
	AgeRangeValueFrom *AgeRangeValueFilter `json:"ageRangeValueFrom,omitempty" gqls:"ageRangeValueFrom,omitempty"`
	AgeRangeUnitFrom  *AgeRangeUnitFilter  `json:"ageRangeUnitFrom,omitempty" gqls:"ageRangeUnitFrom,omitempty"`
	AgeRangeValueTo   *AgeRangeValueFilter `json:"ageRangeValueTo,omitempty" gqls:"ageRangeValueTo,omitempty"`
	AgeRangeUnitTo    *AgeRangeUnitFilter  `json:"ageRangeUnitTo,omitempty" gqls:"ageRangeUnitTo,omitempty"`
	Status            *StringFilter        `json:"status,omitempty" gqls:"status,omitempty"`
	System            *BooleanFilter       `json:"system,omitempty" gqls:"system,omitempty"`
	OrganizationID    *UUIDFilter          `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	ProgramID         *UUIDFilter          `json:"programId,omitempty" gqls:"programId,omitempty"`
	AND               []AgeRangeFilter     `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR                []AgeRangeFilter     `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (AgeRangeFilter) FilterType() FilterOfType {
	return AgeRangesConnectionType
}

type AgeRangeConnectionNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}

type AgeRangesConnectionEdge struct {
	Cursor string                 `json:"cursor" gqls:"cursor"`
	Node   AgeRangeConnectionNode `json:"node" gqls:"node"`
}

type AgesConnectionResponse struct {
	TotalCount int                       `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo        `json:"pageInfo" gqls:"pageInfo"`
	Edges      []AgeRangesConnectionEdge `json:"edges" gqls:"edges"`
}

func (acs AgesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &acs.PageInfo
}
func (acs AmsAgeConnectionService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Age, error) {
	condition := NewCondition(options...)
	filter := AgeRangeFilter{
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
	var pages []AgesConnectionResponse
	err := pageQuery(ctx, operator, filter.FilterType(), filter, &pages)
	if err != nil {
		log.Error(ctx, "get age by program failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	var ages []*Age
	for _, page := range pages {
		for _, v := range page.Edges {
			age := Age{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			ages = append(ages, &age)
		}
	}
	return ages, nil
}
func (acs AmsAgeConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Age, error) {
	condition := NewCondition(options...)

	filter := AgeRangeFilter{
		ProgramID: &UUIDFilter{
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

	var pages []AgesConnectionResponse
	err := pageQuery(ctx, operator, filter.FilterType(), filter, &pages)
	if err != nil {
		log.Error(ctx, "get age by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	var ages []*Age
	for _, page := range pages {
		for _, v := range page.Edges {
			age := Age{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			ages = append(ages, &age)
		}
	}
	return ages, nil
}
