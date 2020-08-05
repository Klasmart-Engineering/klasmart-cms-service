package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ICategoryDA interface {
	CreateCategory(ctx context.Context, data entity.CategoryObject) (*entity.CategoryObject, error)
	UpdateCategory(ctx context.Context, data entity.CategoryObject) error
	DeleteCategory(ctx context.Context, id string) error
	GetCategoryByID(ctx context.Context, id string) (*entity.CategoryObject, error)

	SearchCategories(ctx context.Context, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error)
	PageCategories(ctx context.Context, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error)
}

var _categoryOnce sync.Once

var categoryDA *CategoryDynamoDA

func GetCategoryDA() ICategoryDA {
	_categoryOnce.Do(func() {
		categoryDA = new(CategoryDynamoDA)
	})
	return categoryDA
}
