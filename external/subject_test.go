package external

import (
	"context"
	"fmt"
	"testing"
)

func TestAmsSubjectService_BatchGet(t *testing.T) {
	ids := []string{"5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", "fab745e8-9e31-4d0c-b780-c40120c98b27"}
	subjects, err := GetSubjectServiceProvider().BatchGet(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetSubjectServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(subjects) != len(ids) {
		t.Errorf("GetSubjectServiceProvider().BatchGet() want %d results got %d", len(ids), len(subjects))
		return
	}

	for _, subject := range subjects {
		if subject == nil {
			t.Error("GetSubjectServiceProvider().BatchGet() get null")
			return
		}
	}
}

var sbjToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTkwNDE0MywiaXNzIjoia2lkc2xvb3AifQ.DOPhGxQp4pCzONmp92GJWES5CJr-dz7gjGO_KyGDS6lLUqYtv-e74r3GC8s4OWFdPnYUJEp_zi5HIX-RAe-_w1Was1WSWJXjc2c9JM2CJ3OOKK-y8nTHFdFcoQc0UFwKJRfqNDc0hnU_mY1rtCUouAa7Ktpu6C8YExcNx191GFJcu4LRk-kz9GFyvdT8v5jTY1dO5RPLANdgaGVrVGnTur0ASJfh9whlMATbMlDThJXWbwA2NSIMhd8Y8ZkFDoI-3ZQoQyoVHgzscDaiK_UnfwBXRK4NPHONKcOIfXXtUvjrHYt5lfheOrx3FnkCi448dT3oCfl5V-9HSiXSKSHR6g"

func TestAmsSubjectService_GetByProgram(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = sbjToken
	ID := "04c630cc-fabe-4176-80f2-30a029907a33"
	provider := AmsSubjectConnectionService{}
	subjects1, err := provider.AmsSubjectService.GetByProgram(ctx, testOperator, ID, WithStatus(Active))
	if err != nil {
		t.Errorf("GetSubjectServiceProvider().GetByProgram() error = %v", err)
		return
	}

	subjects2, err := provider.GetByProgram(ctx, testOperator, ID, WithStatus(Active))
	if err != nil {
		t.Errorf("GetSubjectServiceProvider().GetByProgram() error = %v", err)
		return
	}

	fmt.Println("len:", len(subjects1) == len(subjects2))
}

func TestAmsSubjectService_GetByOrganization(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = sbjToken
	provider := AmsSubjectConnectionService{}
	subjects1, err := provider.AmsSubjectService.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetSubjectServiceProvider().GetByProgram() error = %v", err)
		return
	}

	subjects2, err := provider.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetSubjectServiceProvider().GetByProgram() error = %v", err)
		return
	}

	fmt.Println("len:", len(subjects1) == len(subjects2))
}
