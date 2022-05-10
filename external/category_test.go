package external

import (
	"context"
	"fmt"
	"testing"
	"time"
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
	time.Sleep(time.Second)
}

var catToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MjE0Njg3NCwiaXNzIjoia2lkc2xvb3AifQ.W5y6k4AnXeVSUnd12rRUSRYpKhigVY3h46lj8qoJXZtqLk6tRpCmg57E3_TI27rs1tsEqgR0dEXUl5dqJD3FHP7UjS9t-SWpY9aqvK3yYUoPd3YCbOyvsW3PkSd7UkrbSOVADsInDPT9H3DMjimPMrz_aTPyofpRCeOxYSiFOPW4iNc51qJ3Kb7RTfW-ljyVd2XLIB7gHXwimG4hppFxjOtlDmW3EoNVFnAUbhYghz1NJhpc5kiOLuz9VBDaEjv2gNC7sAsys8sqmmHZHDuEW-lYsyty_q1uInxbUFSBQonHtfjAuAFfXZFQS1hpQHIcxyqlvrc7IOFmWeOdFZMxNA"

func TestAmsCategoryService_GetByProgram(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = catToken
	id := "75004121-0c0d-486c-ba65-4c57deacb44b"
	provider := AmsCategoryConnectionService{}

	categories1, err := provider.AmsCategoryService.GetByProgram(ctx, testOperator, id)
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByProgram(() error = %v", err)
		return
	}

	categories2, err := provider.GetByProgram(ctx, testOperator, id)
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByProgram(() error = %v", err)
		return
	}
	fmt.Println("len:", len(categories1) == len(categories2))
}

func TestAmsCategoryService_GetByOrganization(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = catToken
	provider := AmsCategoryConnectionService{}

	categories1, err := provider.AmsCategoryService.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByProgram(() error = %v", err)
		return
	}

	categories2, err := provider.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByProgram(() error = %v", err)
		return
	}
	map1 := make(map[string]*Category)
	for _, c := range categories1 {
		if _, ok := map1[c.ID]; ok {
			fmt.Println("exists:", *c)
		}
		map1[c.ID] = c
	}
	map2 := make(map[string]*Category)
	for _, c := range categories2 {
		if _, ok := map2[c.ID]; ok {
			fmt.Println("exists:", c)
		}
		map2[c.ID] = c
	}
	for k, v := range map2 {
		if _, ok := map1[k]; !ok {
			fmt.Println("not exists:", *v)
		}
	}
	fmt.Println("len:", len(categories1) == len(categories2))
}

func TestAmsCategoryService_GetBySubjects(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = catToken
	IDs := []string{
		"2e922238-decb-438e-b960-a0e404e015a5",
		"44a5ecab-2d1c-4dd7-b20e-f67ec923ed02",
		"fab745e8-9e31-4d0c-b780-c40120c98b27",
		"66a453b0-d38f-472e-b055-7a94a94d66c4",
	}
	provider := AmsCategoryConnectionService{}
	categories1, err := provider.AmsCategoryService.GetBySubjects(ctx, testOperator, IDs, WithStatus(Active))
	if err != nil {
		t.Errorf("GetCategoryServiceProvider().GetBySubjects(() error = %v", err)
		return
	}
	categories2, err := provider.GetBySubjects(ctx, testOperator, IDs, WithStatus(Active))
	if err != nil {
		t.Errorf("GetCategoryServiceProvider().GetBySubjects(() error = %v", err)
		return
	}
	fmt.Println("len:", len(categories1) == len(categories2))
}
