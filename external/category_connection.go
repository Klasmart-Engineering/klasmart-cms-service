package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
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
func (accs AmsCategoryConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Category, error) {
	result := make(map[string][]CategoriesConnectionResponse)
	err := subPageQuery(ctx, operator, "organizationNode", "categoriesConnection", []string{operator.OrgID}, result)
	if err != nil {
		log.Error(ctx, "GetBySubjects: subPageQuery failed",
			log.Err(err),
			log.Any("operator", operator))
		return nil, err
	}

	condition := NewCondition(options...)
	var categories []*Category
	for _, pages := range result {
		for _, page := range pages {
			for _, edge := range page.Edges {
				category := Category{
					ID:     edge.Node.ID,
					Name:   edge.Node.Name,
					System: edge.Node.System,
					Status: APStatus(edge.Node.Status),
				}
				if condition.System.Valid && category.System != condition.System.Bool {
					continue
				}
				if condition.Status.Valid && condition.Status.Status != category.Status {
					continue
				} else if condition.Status.Valid && category.Status != Active {
					// only status = "Active" data is returned by default
					continue
				}
				categories = append(categories, &category)
			}
		}
	}
	return categories, nil
}
func (accs AmsCategoryConnectionService) GetBySubjects(ctx context.Context, operator *entity.Operator, subjectIDs []string, options ...APOption) ([]*Category, error) {
	result := make(map[string][]CategoriesConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(subjectIDs)
	err := subPageQuery(ctx, operator, "subcategoryNode", "categoriesConnection", IDs, result)
	if err != nil {
		log.Error(ctx, "GetBySubjects: subPageQuery failed",
			log.Err(err),
			log.Strings("subject_ids", IDs))
		return nil, err
	}

	condition := NewCondition(options...)
	var categories []*Category
	for _, pages := range result {
		for _, page := range pages {
			for _, edge := range page.Edges {
				category := Category{
					ID:     edge.Node.ID,
					Name:   edge.Node.Name,
					Status: APStatus(edge.Node.Status),
				}
				if condition.System.Valid && category.System != condition.System.Bool {
					continue
				}
				if condition.Status.Valid && condition.Status.Status != category.Status {
					continue
				} else if condition.Status.Valid && category.Status != Active {
					// only status = "Active" data is returned by default
					continue
				}
				categories = append(categories, &category)
			}
		}
	}
	return categories, nil
}
