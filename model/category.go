package model

import (
	"calmisland/kidsloop2/entity"
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"sync"
)

type ICategoryModel interface {
	CreateCategory(ctx context.Context, data entity.CategoryObject) (string, error)
	UpdateCategory(ctx context.Context, data entity.CategoryObject) error
	DeleteCategory(ctx context.Context, id string) error
	GetCategoryById(ctx context.Context, id string) (*entity.CategoryObject, error)
	
	SearchCategories(ctx context.Context, condition *SearchCategoryCondition) ([]*entity.CategoryObject, error)
}

type SearchCategoryCondition struct {
	IDs        []string `json:"ids"`
	Names      []string `json:"names"`

	PageSize int64 `json:"page_size"`
	Page     int64 `json:"page"`
	OrderBy	 string `json:"order_by"`
}

func (s *SearchCategoryCondition) getConditions() []expression.ConditionBuilder {
	conditions := make([]expression.ConditionBuilder, 0)
	if len(s.IDs) > 0 {
		condition := expression.Name("_id").In(expression.Value(s.IDs))
		conditions = append(conditions, condition)
	}
	if len(s.Names) > 0 {
		condition := expression.Name("name").In(expression.Value(s.Names))
		conditions = append(conditions, condition)
	}

	return conditions
}

type CategoryModel struct{}

func (cm *CategoryModel) CreateCategory(ctx context.Context, data entity.CategoryObject) (string, error) {
	panic("implement me")
}

func (cm *CategoryModel) UpdateCategory(ctx context.Context, data entity.CategoryObject) error {
	panic("implement me")
}

func (cm *CategoryModel) DeleteCategory(ctx context.Context, id string) error {
	panic("implement me")
}

func (cm *CategoryModel) GetCategoryById(ctx context.Context, id string) (*entity.CategoryObject, error) {
	panic("implement me")
}

func (cm *CategoryModel) SearchCategories(ctx context.Context, condition *SearchCategoryCondition) ([]*entity.CategoryObject, error) {
	panic("implement me")
}

var categoryModel *CategoryModel
var _categoryOnce sync.Once

func GetCategoryModel() ICategoryModel {
	_categoryOnce.Do(func() {
		categoryModel = new(CategoryModel)
	})
	return categoryModel
}
