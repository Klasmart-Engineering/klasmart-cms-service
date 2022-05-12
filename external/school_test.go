package external

import (
	"context"
	"fmt"
	"testing"
)

var schToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MjA1OTIwMCwiaXNzIjoia2lkc2xvb3AifQ.S-rNdS57cN3T3slBy4MTqzQJa0bs7mq5wMPn6493vAGflKZRny8ooGDy9f2TRb7h-ud5E_Xcqf_l4aRzBmgCEWlrvBBDBPYXk5Zrav8y9p5BeyvY35j4S9iCR6uv2-IPAsFr_uu8S41l81vG0JKH6VgdBVlLYWmAb_fxxj-krKYrS3C8Ppp7DNw783Ov67Vt6nGHV3LQ6M9HbagbjXL-MJYQJGSCUqbLNGwCZzZG1NnbpV7bzJfpsH8XYBLCXb2z38BrV7dqTkluk9mftiqH3mNITpYkqsqOKJIz6zf_02vaU31BUkJ8ajdf-Queq4OIOwrde_wB1sU6S0CXWIjTOw"

// 0ff7769f-cc94-4a80-a780-dcb0947db18b
// 00429737-f515-4348-b24f-919c2f82a2aa
var clsIDs = []string{
	"0ff7769f-cc94-4a80-a780-dcb0947db18b",
	"00429737-f515-4348-b24f-919c2f82a2aa",
}

func TestAmsSchoolService_GetByClasses(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = schToken
	IDs := make([]string, 0, len(clsIDs))
	IDs = append(IDs, clsIDs...)
	provider := AmsSchoolConnectionService{}

	result1, err := provider.AmsSchoolService.GetByClasses(ctx, testOperator, IDs)
	if err != nil {
		t.Fatal(err)
	}
	result2, err := provider.GetByClasses(ctx, testOperator, IDs)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("len:", len(result1) == len(result2))
}

func TestAmsSchoolService_GetByOrganizationID(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = schToken
	provider := AmsSchoolConnectionService{}
	schools1, err := provider.AmsSchoolService.GetByOrganizationID(ctx, testOperator, orgID)
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByOrganizationID() error = %v", err)
		return
	}
	schools2, err := provider.GetByOrganizationID(ctx, testOperator, orgID)
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByOrganizationID() error = %v", err)
		return
	}
	fmt.Println("len:", len(schools1) == len(schools2))
}

func TestAmsSchoolService_GetByPermission(t *testing.T) {
	schools, err := GetSchoolServiceProvider().GetByPermission(context.TODO(),
		testOperator,
		CreateContentPage201,
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByPermission() error = %v", err)
		return
	}

	// if len(schools) == 0 {
	// 	t.Error("GetSchoolServiceProvider().GetByPermission() get empty slice")
	// 	return
	// }

	for _, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().GetByPermission() get null")
			return
		}
	}
}

func TestAmsSchoolService_GetByOperator(t *testing.T) {
	testOperator.Token = schToken
	provider := AmsSchoolConnectionService{}
	schools1, err := provider.AmsSchoolService.GetByOperator(context.TODO(), testOperator, WithStatus(Active))
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByOperator() error = %v", err)
		return
	}

	schools2, err := provider.GetByOperator(context.TODO(), testOperator, WithStatus(Active))
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByOperator() error = %v", err)
		return
	}

	fmt.Println("len:", len(schools1) == len(schools2))
}

func TestAmsSchoolService_GetByUsers(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = schToken
	IDs := []string{
		"1dd1d0b4-df1a-4486-bd6c-a89f9b92f779",
		"1e200965-df57-461e-8af3-e255886e8e41",
	}
	provider := AmsSchoolConnectionService{}
	schools1, err := provider.AmsSchoolService.GetByUsers(ctx, testOperator, orgID, IDs)
	if err != nil {
		t.Fatal(err)
	}
	schools2, err := provider.GetByUsers(ctx, testOperator, orgID, IDs)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("len:", len(schools1) == len(schools2))
}

func TestAmsSchoolService_BatchGet(t *testing.T) {
	ids := []string{"4460deda-ed0c-42d1-902d-8513070493be", "dc910c1e-23ac-4e29-a77e-b11b445be25a"}
	schools, err := GetSchoolServiceProvider().BatchGet(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(schools) != len(ids) {
		t.Error("GetSchoolServiceProvider().BatchGet() get invalid slice")
		return
	}

	for _, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().BatchGet() get null")
			return
		}
	}
}
