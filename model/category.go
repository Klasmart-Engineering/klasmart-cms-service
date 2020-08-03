package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ICategoryModel interface {
	CreateCategory(ctx context.Context, data entity.CategoryObject) (*entity.CategoryObject, error)
	UpdateCategory(ctx context.Context, data entity.CategoryObject) error
	DeleteCategory(ctx context.Context, id string) error
	GetCategoryById(ctx context.Context, id string) (*entity.CategoryObject, error)

	SearchCategories(ctx context.Context, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error)
	PageCategories(ctx context.Context, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error)
}

type CategoryModel struct{}

// Repeated insertion with the same primary key will overwrite non-primary key data
func (cm *CategoryModel) CreateCategory(ctx context.Context, data entity.CategoryObject) (*entity.CategoryObject, error) {
	now := time.Now().Unix()
	data.ID = utils.NewID()
	data.CreatedAt = now
	data.UpdatedAt = now
	return da.GetCategoryDA().CreateCategory(ctx, data)
}

func (cm *CategoryModel) UpdateCategory(ctx context.Context, data entity.CategoryObject) error {
	return da.GetCategoryDA().UpdateCategory(ctx, data)
}

func (cm *CategoryModel) DeleteCategory(ctx context.Context, id string) error {
	return da.GetCategoryDA().DeleteCategory(ctx, id)
}

func (cm *CategoryModel) GetCategoryById(ctx context.Context, id string) (*entity.CategoryObject, error) {
	return da.GetCategoryDA().GetCategoryById(ctx, id)
}

func (cm *CategoryModel) SearchCategories(ctx context.Context, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error) {

	return da.GetCategoryDA().SearchCategories(ctx, condition)

}

func (cm *CategoryModel) PageCategories(ctx context.Context, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error) {
	return da.GetCategoryDA().PageCategories(ctx, condition)
}

var categoryModel *CategoryModel
var _categoryOnce sync.Once

func GetCategoryModel() ICategoryModel {
	_categoryOnce.Do(func() {
		categoryModel = new(CategoryModel)
	})
	return categoryModel
}
