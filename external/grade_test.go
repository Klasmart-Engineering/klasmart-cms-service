package external

import (
	"context"
	"fmt"
	"testing"
)

func TestAmsGradeService_BatchGet(t *testing.T) {
	ids := []string{"98461ca1-06a1-432a-97d0-4e1dff33e1a5", "a9f0217d-f7ec-4add-950d-4e8986ab2c82"}
	grades, err := GetGradeServiceProvider().BatchGet(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetGradeServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(grades) != len(ids) {
		t.Errorf("GetGradeServiceProvider().BatchGet() want %d results got %d", len(ids), len(grades))
		return
	}

	for _, grade := range grades {
		if grade == nil {
			t.Error("GetGradeServiceProvider().BatchGet() get null")
			return
		}
	}
}

var grdToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTkwNzc1NCwiaXNzIjoia2lkc2xvb3AifQ.Z92QPHIbeJ8fJd4rdobVanaHJpNXMbYHNgBRo_HTMaXE8XtRaM85SzguFVzpaTySGR0HQl7dV6V3XJV2wAiqJStbYT5ad84WwA_EC5d_qwoGgZCFd0avVO75jY_z7DjSmmSVwD1b7x9Ob13G58OQaDp6KErPlCzDeb3uc12VuRuqBlVx1AD7xZlMONRyYq4h7VOK4B6YtTdf_bq75AfQycnGmse5tY-yFuHf4-K-c3KWIIUlxpBroFM045k-sIV2CttZduBeI-wgx3XdtTQtxVmqrJrYxySlJPpCMPqJhRK7b0aRZOptNLH2NWBjN-JbuOkkDCimUlF4ROCfivUSEw"

func TestAmsGradeService_GetByProgram(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = grdToken
	ID := "04c630cc-fabe-4176-80f2-30a029907a33"
	provider := AmsGradeConnectionService{}

	grades1, err := provider.AmsGradeService.GetByProgram(ctx, testOperator, ID)
	if err != nil {
		t.Errorf("GetGradeServiceProvider().GetByProgram() error = %v", err)
		return
	}

	grades2, err := provider.GetByProgram(ctx, testOperator, ID)
	if err != nil {
		t.Errorf("GetGradeServiceProvider().GetByProgram() error = %v", err)
		return
	}

	fmt.Println("len:", len(grades1) == len(grades2))
}

func TestAmsGradeService_GetByOrganization(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = grdToken
	provider := AmsGradeConnectionService{}

	grades1, err := provider.AmsGradeService.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetGradeServiceProvider().GetByProgram() error = %v", err)
		return
	}

	grades2, err := provider.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetGradeServiceProvider().GetByProgram() error = %v", err)
		return
	}

	fmt.Println("len:", len(grades1) == len(grades2))
}
