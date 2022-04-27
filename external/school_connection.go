package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type SchoolFilter struct {
	SchoolID       *UUIDFilter    `json:"schoolId,omitempty" gqls:"schoolId,omitempty"`
	Name           *StringFilter  `json:"name,omitempty" gqls:"name,omitempty"`
	Shortcode      *StringFilter  `json:"shortcode,omitempty" gqls:"shortcode,omitempty"`
	Status         *StringFilter  `json:"status,omitempty" gqls:"status,omitempty"`
	OrganizationId *UUIDFilter    `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	UserID         *UUIDFilter    `json:"userId,omitempty" gqls:"userId,omitempty"`
	ProgramId      *UUIDFilter    `json:"programId,omitempty" gqls:"programId,omitempty"`
	AND            []SchoolFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR             []SchoolFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (SchoolFilter) FilterName() FilterType {
	return SchoolFilterType
}

func (SchoolFilter) ConnectionName() ConnectionType {
	return SchoolsConnectionType
}

type AmsSchoolConnectionService struct {
	AmsSchoolService
}

type SchoolConnectionNode struct {
	ID             string `json:"id" gqls:"id"`
	Name           string `json:"name" gqls:"name"`
	ShortCode      string `json:"shortCode" gqls:"shortCode"`
	Status         string `json:"status" gqls:"status"`
	OrganizationId string `json:"organizationId" gqls:"organizationId"`
}

type SchoolsConnectionEdge struct {
	Cursor string               `json:"cursor" gqls:"cursor"`
	Node   SchoolConnectionNode `json:"node" gqls:"node"`
}

type SchoolsConnectionResponse struct {
	TotalCount int                     `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo      `json:"pageInfo" gqls:"pageInfo"`
	Edges      []SchoolsConnectionEdge `json:"edges" gqls:"edges"`
}

func (pcs SchoolsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &pcs.PageInfo
}

func (ascs AmsSchoolConnectionService) GetByOrganizationID(ctx context.Context, operator *entity.Operator, organizationID string, options ...APOption) ([]*School, error) {
	condition := NewCondition(options...)
	filter := SchoolFilter{
		OrganizationId: &UUIDFilter{
			Operator: UUIDOperator(OperatorTypeEq),
			Value:    UUID(organizationID),
		},
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		},
	}

	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	var pages []SchoolsConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get age by program failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	var organizations []*School
	for _, page := range pages {
		for _, v := range page.Edges {
			org := School{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
			}
			organizations = append(organizations, &org)
		}
	}
	return organizations, nil
}
func (ascs AmsSchoolConnectionService) GetByUserID(ctx context.Context, operator *entity.Operator, id string, options ...APOption) ([]*School, error) {
	condition := NewCondition(options...)
	filter := SchoolFilter{
		UserID: &UUIDFilter{
			Operator: UUIDOperator(OperatorTypeEq),
			Value:    UUID(id),
		},
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		},
	}

	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	var pages []SchoolsConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get age by program failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	var organizations []*School
	for _, page := range pages {
		for _, v := range page.Edges {
			org := School{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
			}
			organizations = append(organizations, &org)
		}
	}
	return organizations, nil
}
