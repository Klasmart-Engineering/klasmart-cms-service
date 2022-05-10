package external

import (
	"context"
	"fmt"
	"testing"
)

var clsToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MjEwNTEzNiwiaXNzIjoia2lkc2xvb3AifQ.OlhWfjZUMkId4qNSg6bIpxouJ1oqct3EHrqFWgynixZezYbsH3zVcQL8mr1Bn6tQZ9XQUNpU0DRI398bezZKnkYf5_z4MSi8FKANeHlCCnABgDB1PZbmgTZ-8iodIA5hq0ZA0sWmOndOBCvvKMvxtEAmocr8nP2BxwZKIl-TxZLZQvVfBjOKEEDciTu7GGGKUBJZtxZxtSqZ4f6Q2QN9BA-dcMWlC6ZadVCwaMxLRrQ7YPrK2OkKTthUsqt_9eASnZem5CEzrdyeMjeBVo3nLPZP1Pww4G9eQbk-As0emr-GWf_ZtSlErjP-xQ--rFzwsBySygUneIoeLoI6RkTFdQ"
var usrIDs = []string{
	"000d653d-7961-447c-8d66-ad8c4a40eae6",
}

func TestAmsClassService_GetByUserIDs(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = clsToken
	provider := AmsClassConnectionService{}
	classes1, err := provider.AmsClassService.GetByUserIDs(ctx, testOperator, usrIDs, WithStatus(Active))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByUserIDs() error = %v", err)
		return
	}
	classes2, err := provider.GetByUserIDs(context.TODO(), testOperator, usrIDs, WithStatus(Active))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByUserIDs() error = %v", err)
		return
	}

	fmt.Println("len:", len(classes1) == len(classes2))
}

func TestAmsClassService_GetByOrganizationIDs(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = clsToken
	provider := AmsClassConnectionService{}
	classes1, err := provider.AmsClassService.GetByOrganizationIDs(ctx, testOperator, orgIDs)
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByOrganizationIDs() error = %v", err)
		return
	}

	classes2, err := provider.GetByOrganizationIDs(ctx, testOperator, orgIDs)
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByOrganizationIDs() error = %v", err)
		return
	}
	fmt.Println("len:", len(classes1) == len(classes2))
}

func TestAmsClassService_GetBySchoolIDs(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = clsToken
	provider := AmsClassConnectionService{}
	classes1, err := provider.AmsClassService.GetBySchoolIDs(ctx, testOperator, schIDs, WithStatus(Inactive))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByOrganizationIDs() error = %v", err)
		return
	}

	classes2, err := provider.GetBySchoolIDs(ctx, testOperator, schIDs, WithStatus(Inactive))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByOrganizationIDs() error = %v", err)
		return
	}
	fmt.Println("len:", len(classes1) == len(classes2))
}

func TestAmsClassService_GetOnlyUnderOrgClasses(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = clsToken
	provider := AmsClassConnectionService{}
	result1, err := provider.AmsClassService.GetOnlyUnderOrgClasses(ctx, testOperator, orgID)
	if err != nil {
		t.Fatal(err)
	}
	result2, err := provider.GetOnlyUnderOrgClasses(ctx, testOperator, orgID)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("len:", len(result1) == len(result2))
}
