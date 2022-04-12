package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type SubcategoryFilter struct {
	OrganizationID *UUIDFilter         `json:"organizationId" gqls:"organizationId"`
	CategoryID     *UUIDFilter         `json:"categoryId" gqls:"categoryId"`
	System         *BooleanFilter      `json:"system" gqls:"system"`
	Status         *StringFilter       `json:"status" gqls:"status"`
	AND            []SubcategoryFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR             []SubcategoryFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (SubcategoryFilter) FilterType() FilterOfType {
	return SubcategoriesConnectionType
}

type SubcategoryConnectionNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}

type SubcategoriesConnectionEdge struct {
	Cursor string                    `json:"cursor" gqls:"cursor"`
	Node   SubcategoryConnectionNode `json:"node" gqls:"node"`
}

type SubcategoriesConnectionResponse struct {
	TotalCount int                           `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo            `json:"pageInfo" gqls:"pageInfo"`
	Edges      []SubcategoriesConnectionEdge `json:"edges" gqls:"edges"`
}

func (scs SubcategoriesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &scs.PageInfo
}

type AmsSubCategoryConnectionService struct {
	AmsSubCategoryService
}

func (sccs AmsSubCategoryConnectionService) GetByCategory(ctx context.Context, operator *entity.Operator, categoryID string, options ...APOption) ([]*SubCategory, error) {
	condition := NewCondition(options...)

	filter := SubcategoryFilter{
		CategoryID: &UUIDFilter{
			Operator: UUIDOperator(OperatorTypeEq),
			Value:    UUID(categoryID),
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

	var subCategories []*SubCategory
	var pages []SubcategoriesConnectionResponse
	err := pageQuery(ctx, operator, filter.FilterType(), filter, &pages)
	if err != nil {
		log.Error(ctx, "get subcategory by category failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	for _, p := range pages {
		for _, v := range p.Edges {
			obj := &SubCategory{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			subCategories = append(subCategories, obj)
		}
	}
	return subCategories, nil
}
func (sccs AmsGradeConnectionService) AmsSubCategoryConnectionService(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*SubCategory, error) {
	condition := NewCondition(options...)

	filter := SubcategoryFilter{
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

	var subCategories []*SubCategory
	var pages []SubcategoriesConnectionResponse
	err := pageQuery(ctx, operator, filter.FilterType(), filter, &pages)
	if err != nil {
		log.Error(ctx, "get subcategory by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	for _, p := range pages {
		for _, v := range p.Edges {
			obj := &SubCategory{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			subCategories = append(subCategories, obj)
		}
	}
	return subCategories, nil
}
