package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type SubcategoryFilter struct {
	OrganizationID *UUIDFilter         `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	CategoryID     *UUIDFilter         `json:"categoryId,omitempty" gqls:"categoryId,omitempty"`
	System         *BooleanFilter      `json:"system,omitempty" gqls:"system,omitempty"`
	Status         *StringFilter       `json:"status,omitempty" gqls:"status,omitempty"`
	AND            []SubcategoryFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR             []SubcategoryFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (SubcategoryFilter) FilterName() FilterType {
	return SubcategoryFilterType
}

func (SubcategoryFilter) ConnectionName() ConnectionType {
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

func (sccs AmsSubCategoryConnectionService) pageNodes(ctx context.Context, operator *entity.Operator, pages []SubcategoriesConnectionResponse) []*SubCategory {
	if len(pages) == 0 {
		log.Warn(ctx, "pageNodes is empty",
			log.Any("operator", operator))
		return []*SubCategory{}
	}
	subcategories := make([]*SubCategory, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, edge := range page.Edges {
			if _, ok := exists[edge.Node.ID]; ok {
				log.Warn(ctx, "pageNodes: subcategory exist",
					log.Any("subcategory", edge.Node),
					log.Any("operator", operator))
				continue
			}
			exists[edge.Node.ID] = true
			obj := &SubCategory{
				ID:     edge.Node.ID,
				Name:   edge.Node.Name,
				Status: APStatus(edge.Node.Status),
				System: edge.Node.System,
			}
			subcategories = append(subcategories, obj)
		}
	}
	return subcategories
}

func (sccs AmsSubCategoryConnectionService) NewSubcategoryFilter(ctx context.Context, operator *entity.Operator, options ...APOption) *SubcategoryFilter {
	condition := NewCondition(options...)
	var filter SubcategoryFilter
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

func (sccs AmsSubCategoryConnectionService) GetByCategory(ctx context.Context, operator *entity.Operator, categoryID string, options ...APOption) ([]*SubCategory, error) {
	filter := sccs.NewSubcategoryFilter(ctx, operator, options...)
	filter.CategoryID = &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(categoryID)}

	var pages []SubcategoriesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get subcategory by category failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	if len(pages) == 0 {
		log.Debug(ctx, "subcategory is empty",
			log.Any("operator", operator),
			log.Any("filter", filter))
		return []*SubCategory{}, nil
	}
	subCategories := sccs.pageNodes(ctx, operator, pages)
	return subCategories, nil
}
func (sccs AmsSubCategoryConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*SubCategory, error) {
	filter := sccs.NewSubcategoryFilter(ctx, operator, options...)

	filter.OR = []SubcategoryFilter{
		{OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(operator.OrgID)}},
		{System: &BooleanFilter{Operator: OperatorTypeEq, Value: true}},
	}

	var pages []SubcategoriesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get subcategory by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}

	subCategories := sccs.pageNodes(ctx, operator, pages)
	return subCategories, nil
}
