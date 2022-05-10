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

type CategoryFilter struct {
	Status *StringFilter    `json:"status,omitempty" gqls:"status,omitempty"`
	System *BooleanFilter   `json:"system,omitempty" gqls:"system,omitempty"`
	AND    []CategoryFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR     []CategoryFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (CategoryFilter) FilterName() FilterType {
	return CategoryFilterType
}

func (CategoryFilter) ConnectionName() ConnectionType {
	return CategoriesConnectionType
}

type CategoryConnectionNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}

type CategoriesConnectionEdge struct {
	Cursor string                 `json:"cursor" gqls:"cursor"`
	Node   CategoryConnectionNode `json:"node" gqls:"node"`
}

type CategoriesConnectionResponse struct {
	TotalCount int                        `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo         `json:"pageInfo" gqls:"pageInfo"`
	Edges      []CategoriesConnectionEdge `json:"edges" gqls:"edges"`
}

func (ccr CategoriesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &ccr.PageInfo
}

type AmsCategoryConnectionService struct {
	AmsCategoryService
}

func (accs AmsCategoryConnectionService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Category, error) {
	subjects, err := GetSubjectServiceProvider().GetByProgram(ctx, operator, programID)
	if err != nil {
		log.Error(ctx, "GetByProgram: get subjects failed",
			log.Err(err),
			log.String("program", programID),
			log.Any("operator", operator))
		return nil, err
	}
	subjectIDs := make([]string, 0, len(subjects))
	for _, s := range subjects {
		subjectIDs = append(subjectIDs, s.ID)
	}
	if len(subjectIDs) == 0 {
		log.Debug(ctx, "GetByProgram: subject is empty",
			log.String("program", programID),
			log.Any("operator", operator))
		return []*Category{}, nil
	}
	categories, err := accs.GetBySubjects(ctx, operator, subjectIDs, options...)
	if err != nil {
		log.Error(ctx, "GetByProgram: get by subjects failed",
			log.Err(err),
			log.Strings("subjects", subjectIDs),
			log.String("program", programID),
			log.Any("operator", operator))
		return nil, err
	}
	return categories, nil
}

func (accs AmsCategoryConnectionService) pageNodes(ctx context.Context, operator *entity.Operator, pages []CategoriesConnectionResponse) []*Category {
	if len(pages) == 0 {
		log.Warn(ctx, "pageNodes is empty",
			log.Any("operator", operator))
		return []*Category{}
	}
	categories := make([]*Category, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, edge := range page.Edges {
			if _, ok := exists[edge.Node.ID]; ok {
				log.Warn(ctx, "pageNodes: category exist",
					log.Any("category", edge.Node),
					log.Any("operator", operator))
				continue
			}
			exists[edge.Node.ID] = true
			category := Category{
				ID:     edge.Node.ID,
				Name:   edge.Node.Name,
				System: edge.Node.System,
				Status: APStatus(edge.Node.Status),
			}
			categories = append(categories, &category)
		}
	}
	return categories
}

func (accs AmsCategoryConnectionService) NewCategoryFilter(ctx context.Context, operator *entity.Operator, options ...APOption) *CategoryFilter {
	condition := NewCondition(options...)
	var filter CategoryFilter
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

func (accs AmsCategoryConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Category, error) {
	filter := accs.NewCategoryFilter(ctx, operator, options...)
	result := make(map[string][]CategoriesConnectionResponse)
	err := subPageQuery(ctx, operator, "organizationNode", filter, []string{operator.OrgID}, result)
	if err != nil {
		log.Error(ctx, "GetByOrganization: subPageQuery failed",
			log.Err(err),
			log.Any("operator", operator))
		return nil, err
	}

	var categories []*Category
	for k, pages := range result {
		cats := accs.pageNodes(ctx, operator, pages)
		if len(cats) == 0 {
			log.Warn(ctx, "GetByOrganization: category is empty",
				log.String("organization", k),
				log.Any("operator", operator))
			continue
		}
		categories = append(categories, cats...)
	}
	return categories, nil
}

func (accs AmsCategoryConnectionService) GetBySubjects(ctx context.Context, operator *entity.Operator, subjectIDs []string, options ...APOption) ([]*Category, error) {
	filter := accs.NewCategoryFilter(ctx, operator, options...)
	result := make(map[string][]CategoriesConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(subjectIDs)
	err := subPageQuery(ctx, operator, "subjectNode", filter, IDs, result)
	if err != nil {
		log.Error(ctx, "GetBySubjects: subPageQuery failed",
			log.Err(err),
			log.Strings("subject_ids", IDs))
		return nil, err
	}

	var categories []*Category
	for _, pages := range result {
		cats := accs.pageNodes(ctx, operator, pages)
		if len(cats) == 0 {
			continue
		}
		categories = append(categories, cats...)
	}
	return categories, nil
}

func (accs AmsCategoryConnectionService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$category_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: categoryNode(id: $category_id_%d) {id name status system}\n", index, index)
	}
	sb.WriteString("}")

	request := NewRequest(sb.String(), RequestToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("category_id_%d", index), id)
	}

	data := map[string]*Category{}

	response := &GraphQLSubResponse{
		Data: &data,
	}

	_, err = GetAmsConnection().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get categories by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get categories by ids failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Strings("ids", ids))
		return nil, response.Errors
	}

	categories := make([]cache.Object, 0, len(data))
	for index := range ids {
		category := data[fmt.Sprintf("q%d", indexMapping[index])]
		if category == nil {
			log.Error(ctx, "category not found", log.String("id", ids[index]))
			return nil, constant.ErrRecordNotFound
		}
		categories = append(categories, category)
	}

	log.Info(ctx, "get categories by ids success",
		log.Strings("ids", ids),
		log.Any("categories", categories))

	return categories, nil
}
