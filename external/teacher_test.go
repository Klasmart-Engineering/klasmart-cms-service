package external

import (
	"context"
	"fmt"
	"testing"
)

var tchToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MjEwNjkzNiwiaXNzIjoia2lkc2xvb3AifQ.BmR9uQuk0zWtwsb2d2ZHHL2OIvVQTD0TWYSU4svm_lfzAqSXGCi_bSPqf22CjIAQ2CtB_knFIQQEhudkjcxybut2GZ6T537EIaYJWaO3Erw2WBxlsXUv932gjNl13ZDZ36-rt1PHrcrLFt9467I3vaCq6pUgQ2FjapgKnmyek3Uq2S8EYGdwXfpoBvNnnU5rQDAnDQqP0NmfP9dkvKKtKgy0Qdqet0_a74t3giYRroAQCNLtmphthZ_41cjzQ4MtY_EtRLOhsvXEvAiZESuOBBJ1dBuJjF-4TxvMLqcnAGwVAj3g_Sq6FNvYwH5Db6x84RoPoTYD2DjXR9fpPAAiqA"

func TestAmsTeacherService_GetByOrganization(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = tchToken
	provider := AmsTeacherConnectionService{}
	teachers1, err := provider.AmsTeacherService.GetByOrganization(ctx, testOperator, orgID)
	if err != nil {
		t.Fatal(err)
	}
	teachers2, err := provider.GetByOrganization(ctx, testOperator, orgID)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("len:", len(teachers1) == len(teachers2))

}

func TestAmsTeacherService_GetByOrganizations(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = tchToken
	provider := AmsTeacherConnectionService{}
	teachers1, err := provider.AmsTeacherService.GetByOrganizations(ctx, testOperator, orgIDs)
	if err != nil {
		t.Fatal(err)
	}
	teachers2, err := provider.GetByOrganizations(ctx, testOperator, orgIDs)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("len:", len(teachers1) == len(teachers2))

}

func TestAmsTeacherService_GetBySchools(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = tchToken
	IDs := []string{
		"0adee5ec-9454-44a9-b894-05ca1768b01e",
		"0bf25570-337d-42fd-a594-09821f0d59fb",
	}
	provider := AmsTeacherConnectionService{}
	result1, err := provider.AmsTeacherService.GetBySchools(ctx, testOperator, IDs)
	if err != nil {
		t.Fatal(err)
	}
	result2, err := provider.GetBySchools(ctx, testOperator, IDs)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("len:", len(result1) == len(result2))
}

func TestAmsTeacherService_GetByClasses(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = tchToken
	IDs := []string{
		"ee8a4dd5-e806-4d85-b098-cc5a04399b1b",
		"510506d6-92fd-46a8-a585-f64ca59422ab",
	}
	provider := AmsTeacherConnectionService{}
	teachers1, err := provider.AmsTeacherService.GetByClasses(ctx, testOperator, IDs)
	if err != nil {
		t.Fatal(err)
	}
	teachers2, err := provider.GetByClasses(ctx, testOperator, IDs)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("len:", len(teachers1) == len(teachers2))

	for k, tch1s := range teachers1 {
		tch2s := teachers2[k]
		for _, t1 := range tch1s {
			match := false
			for _, t2 := range tch2s {
				if t1.ID == t2.ID {
					match = true
					break
				}
			}
			if !match {
				fmt.Println("mismatch:", *t1)
			}
		}
	}
}
