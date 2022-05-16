package external

import (
	"context"
	"fmt"
	"testing"
)

func TestAmsSubCategoryService_BatchGet(t *testing.T) {
	ids := []string{"40a232cd-d6e8-4ec1-97ec-4e4df7d00a78", "4114f381-a7c5-4e88-be84-2bef4eb04ad0", "96f81756-70e3-41e5-9143-740376574e35"}
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

var sbcToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTg5Mzc4NSwiaXNzIjoia2lkc2xvb3AifQ.LUX1vs6Mq0NwopG09WwrBqnyFsJAngjUBvKFu4iG-vVY8F7rh8acViUSt-sTAHm01yTzvjkebubcyi_kO8kShTTwVMgBxWNgW1Zz3b4yNLyNxq4Bm_1sN4R1beWI9ygfKjOUtiaJzFqXmTt8A5fVQIoNfNS9rS99CdAXHQ60hzFu5wog-UGG97ve82jrURU3_y1puPnRwy4_ZonXNnclq0ygNsbrpKMeoz-piHE1gqrrFKzfwjOeKvxIEepK02_dWODsacJXeEtmTFOWKe6pPlPl5-OpsAq6Urfrtq1uTBKV57EAk9HAwLxvFpQSVtIGXj9PcC9Scaw2llSfT6jGAw"

func TestAmsSubCategoryService_GetByCategory(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = sbcToken
	ID := "84b8f87a-7b61-4580-a190-a9ce3fe90dd3"
	provider := AmsSubCategoryConnectionService{}
	subCategories1, err := provider.AmsSubCategoryService.GetByCategory(ctx, testOperator, ID)
	if err != nil {
		t.Errorf("GetSubCategoryServiceProvider().GetByCategory(() error = %v", err)
		return
	}
	subCategories2, err := provider.GetByCategory(ctx, testOperator, ID)
	if err != nil {
		t.Errorf("GetSubCategoryServiceProvider().GetByCategory(() error = %v", err)
		return
	}

	fmt.Println("len:", len(subCategories1) == len(subCategories2))
}

func TestAmsSubCategoryService_GetByOrganization(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = sbcToken
	provider := AmsSubCategoryConnectionService{}
	subCategories1, err := provider.AmsSubCategoryService.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetSubCategoryServiceProvider().GetByCategory(() error = %v", err)
		return
	}
	subCategories2, err := provider.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetSubCategoryServiceProvider().GetByCategory(() error = %v", err)
		return
	}

	fmt.Println("len:", len(subCategories1) == len(subCategories2))
}
