package external

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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
	subCategories := make([]*SubCategory, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, p := range pages {
		for _, v := range p.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Warn(ctx, "subcategory exist",
					log.Any("subcategory", v),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
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
func (sccs AmsSubCategoryConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*SubCategory, error) {
	condition := NewCondition(options...)

	filter := SubcategoryFilter{
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		},
		OR: []SubcategoryFilter{
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

	var pages []SubcategoriesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get subcategory by organization failed",
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
	subCategories := make([]*SubCategory, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, p := range pages {
		for _, v := range p.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Warn(ctx, "subcategory exist",
					log.Any("subcategory", v),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
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
