package external

import (
	"context"
	"testing"
)

func TestAmsCategoryService_BatchGet(t *testing.T) {
	ids := []string{"84b8f87a-7b61-4580-a190-a9ce3fe90dd3", "2d5ea951-836c-471e-996e-76823a992689"}
	categories, err := GetCategoryServiceProvider().BatchGet(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetCategoryServiceProvider.BatchGet() error = %v", err)
		return
	}

	if len(categories) != len(ids) {
		t.Errorf("GetCategoryServiceProvider.BatchGet() want %d results got %d", len(ids), len(categories))
		return
	}

	for _, category := range categories {
		if category == nil {
			t.Error("GetCategoryServiceProvider.BatchGet() get null")
			return
		}
	}
}

func TestAmsCategoryService_GetByProgram(t *testing.T) {
	categories, err := GetCategoryServiceProvider().GetByProgram(context.TODO(), testOperator, "75004121-0c0d-486c-ba65-4c57deacb44b")
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByProgram(() error = %v", err)
		return
	}

	if len(categories) == 0 {
		t.Error("GetAgeServiceProvider().GetByProgram(() get empty slice")
		return
	}

	for _, category := range categories {
		if category == nil {
			t.Error("GetAgeServiceProvider().GetByProgram(() get null")
			return
		}
	}
}
