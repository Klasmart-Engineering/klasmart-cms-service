package external

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestAmsAgeService_BatchGet(t *testing.T) {
	ids := []string{"023eeeb1-5f72-4fa3-a2a7-63603607ac2b", "bb7982cd-020f-4e1a-93fc-4a6874917f07"}
	ages, err := GetAgeServiceProvider().BatchGet(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetAgeServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(ages) != len(ids) {
		t.Errorf("GetAgeServiceProvider().BatchGet() want %d results got %d", len(ids), len(ages))
		return
	}

	for _, age := range ages {
		if age == nil {
			t.Error("GetAgeServiceProvider().BatchGet() get null")
			return
		}
	}

	time.Sleep(time.Second)
}

var ageToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTgyNDE2OSwiaXNzIjoia2lkc2xvb3AifQ.dGK8xXolSSdx-cf7uarPqVJe2oOkrXHfP0-GuBA0XGzU5Yb4e2F2KkEsVQTAY78M7KemvhliFujntppg1JmOu3xNDNo6mNeqtBzcfI5HSnit8kHCmAX5MkFGSSzuK_mAHtvCwGvdEyqj6-8FhMWkompKtH78o3EMvItutOU2mhv-V24p5pdFbE1KK_7RI15Yxdej_Uu_UJCl57LpNog_vS0ANd_QyqF-QJEb0esffYr66s1X1_JFqekeZezckl27-0AwyOX3-Jg_SyOy7bn-Sk1WWys2-aSdZ4N3I41SUKRHCrdrEromsFmqT9DUvUl5gixL-sZKJq1zhecw8GEwWA"

func TestAmsAgeService_GetByProgram(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = ageToken
	programID := "04c630cc-fabe-4176-80f2-30a029907a33"
	provider := AmsAgeConnectionService{}
	ages1, err := provider.AmsAgeService.GetByProgram(ctx, testOperator, programID)
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByProgram() error = %v", err)
		return
	}

	ages2, err := provider.GetByProgram(ctx, testOperator, programID)
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByProgram() error = %v", err)
		return
	}
	fmt.Println("len:", len(ages1) == len(ages2))
}

func TestAmsAgeService_GetByOrganization(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = ageToken
	provider := AmsAgeConnectionService{}
	ages1, err := provider.AmsAgeService.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByProgram() error = %v", err)
		return
	}

	ages2, err := provider.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByProgram() error = %v", err)
		return
	}
	fmt.Println("len:", len(ages1) == len(ages2))
}
