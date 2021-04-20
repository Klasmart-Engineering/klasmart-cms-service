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

func TestAmsCategoryService_GetByOrganization(t *testing.T) {
	categories, err := GetCategoryServiceProvider().GetByOrganization(context.TODO(), testOperator, WithStatus(Active))
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByOrganization(() error = %v", err)
		return
	}

	if len(categories) == 0 {
		t.Error("GetAgeServiceProvider().GetByOrganization(() get empty slice")
		return
	}

	for _, category := range categories {
		if category == nil {
			t.Error("GetAgeServiceProvider().GetByOrganization(() get null")
			return
		}
	}
}

func TestAmsCategoryService_GetBySubjects(t *testing.T) {
	result, err := GetCategoryServiceProvider().GetBySubjects(context.TODO(), testOperator, []string{
		"2e922238-decb-438e-b960-a0e404e015a5",
		"44a5ecab-2d1c-4dd7-b20e-f67ec923ed02",
		"fab745e8-9e31-4d0c-b780-c40120c98b27",
		"66a453b0-d38f-472e-b055-7a94a94d66c4",
	}, WithStatus(Active))
	if err != nil {
		t.Errorf("GetCategoryServiceProvider().GetBySubjects(() error = %v", err)
		return
	}
	for _, item := range result {
		t.Log(*item)
	}
}
