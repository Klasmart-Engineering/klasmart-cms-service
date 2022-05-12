package external

import (
	"context"
	"fmt"
	"strings"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cache/cache"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
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

func (sccs AmsSubCategoryConnectionService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$subcategory_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: subcategoryNode(id: $subcategory_id_%d) {id name status system}\n", index, index)
	}
	sb.WriteString("}")

	request := NewRequest(sb.String(), RequestToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("subcategory_id_%d", index), id)
	}

	data := map[string]*SubCategory{}
	response := &GraphQLSubResponse{
		Data: &data,
	}

	_, err = GetAmsConnection().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get subCategories by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get subCategories by ids failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Strings("ids", ids))
		return nil, response.Errors
	}

	subCategories := make([]cache.Object, 0, len(data))
	for index := range ids {
		subCategory := data[fmt.Sprintf("q%d", indexMapping[index])]
		if subCategory == nil {
			log.Error(ctx, "subCategory not found", log.String("id", ids[index]))
			return nil, constant.ErrRecordNotFound
		}
		subCategories = append(subCategories, subCategory)
	}

	log.Info(ctx, "get subCategories by ids success",
		log.Strings("ids", ids),
		log.Any("subCategories", subCategories))

	return subCategories, nil
}
