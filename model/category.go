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
	CreateCategory(ctx context.Context, op *entity.Operator, data entity.CategoryObject) (*entity.CategoryObject, error)
	UpdateCategory(ctx context.Context, op *entity.Operator, data entity.CategoryObject) error
	DeleteCategory(ctx context.Context, op *entity.Operator, id string) error
	GetCategoryByID(ctx context.Context, op *entity.Operator, id string) (*entity.CategoryObject, error)

	SearchCategories(ctx context.Context, op *entity.Operator, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error)
	PageCategories(ctx context.Context, op *entity.Operator, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error)
}

type CategoryModel struct{}

// Repeated insertion with the same primary key will overwrite non-primary key data
func (cm *CategoryModel) CreateCategory(ctx context.Context, op *entity.Operator, data entity.CategoryObject) (*entity.CategoryObject, error) {
	now := time.Now().Unix()
	data.ID = utils.NewID()
	data.CreatedID = op.UserID
	data.UpdatedID = op.UserID
	data.CreatedAt = now
	data.UpdatedAt = now
	return da.GetCategoryDA().CreateCategory(ctx, data)
}

func (cm *CategoryModel) UpdateCategory(ctx context.Context, op *entity.Operator, data entity.CategoryObject) error {
	data.UpdatedID = op.UserID
	data.UpdatedAt = time.Now().Unix()
	return da.GetCategoryDA().UpdateCategory(ctx, data)
}

func (cm *CategoryModel) DeleteCategory(ctx context.Context, op *entity.Operator, id string) error {
	return da.GetCategoryDA().DeleteCategory(ctx, op, id)
}

func (cm *CategoryModel) GetCategoryByID(ctx context.Context, op *entity.Operator, id string) (*entity.CategoryObject, error) {
	return da.GetCategoryDA().GetCategoryByID(ctx, id)
}

func (cm *CategoryModel) SearchCategories(ctx context.Context, op *entity.Operator, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error) {

	return da.GetCategoryDA().SearchCategories(ctx, condition)

}

func (cm *CategoryModel) PageCategories(ctx context.Context, op *entity.Operator, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error) {
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
