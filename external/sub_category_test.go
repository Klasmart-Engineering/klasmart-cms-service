package external

import (
	"context"
	"testing"
)

func TestAmsSubCategoryService_BatchGet(t *testing.T) {
	ids := []string{"40a232cd-d6e8-4ec1-97ec-4e4df7d00a78", "2b6b5d54-0243-4c7e-917a-1627f107f198"}
	subCategories, err := GetSubCategoryServiceProvider().BatchGet(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetSubCategoryServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(subCategories) != len(ids) {
		t.Errorf("GetSubCategoryServiceProvider().BatchGet() want %d results got %d", len(ids), len(subCategories))
		return
	}

	for _, category := range subCategories {
		if category == nil {
			t.Error("GetSubCategoryServiceProvider().BatchGet() get null")
			return
		}
	}
}

func TestAmsSubCategoryService_GetByCategory(t *testing.T) {
	subCategories, err := GetSubCategoryServiceProvider().GetByCategory(context.TODO(), testOperator, "84b8f87a-7b61-4580-a190-a9ce3fe90dd3")
	if err != nil {
		t.Errorf("GetSubCategoryServiceProvider().GetByCategory(() error = %v", err)
		return
	}

	if len(subCategories) == 0 {
		t.Error("GetSubCategoryServiceProvider().GetByCategory(() get empty slice")
		return
	}

	for _, subCategory := range subCategories {
		if subCategory == nil {
			t.Error("GetSubCategoryServiceProvider().GetByCategory(() get null")
			return
		}
	}
}
