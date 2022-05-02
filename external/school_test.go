package external

import (
	"context"
	"fmt"
	"testing"
)

// 0ff7769f-cc94-4a80-a780-dcb0947db18b
// 00429737-f515-4348-b24f-919c2f82a2aa
func TestAmsSchoolService_GetByClasses(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = schToken
	IDs := []string{
		"0ff7769f-cc94-4a80-a780-dcb0947db18b",
		"00429737-f515-4348-b24f-919c2f82a2aa",
	}
	result, err := GetSchoolServiceProvider().GetByClasses(ctx, testOperator, IDs)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range result {
		fmt.Println(k, v)
	}
}

func TestAmsSchoolService_GetByOrganizationID(t *testing.T) {
	testOperator.Token = schToken
	schools, err := GetSchoolServiceProvider().GetByOrganizationID(context.TODO(),
		testOperator,
		orgID,
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByOrganizationID() error = %v", err)
		return
	}

	if len(schools) == 0 {
		t.Error("GetSchoolServiceProvider().GetByOrganizationID() get empty slice")
		return
	}

	for _, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().GetByOrganizationID() get null")
			return
		}
	}
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
	schools, err := GetSchoolServiceProvider().GetByOperator(context.TODO(),
		testOperator,
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByOperator() error = %v", err)
		return
	}

	// if len(schools) == 0 {
	// 	t.Error("GetSchoolServiceProvider().GetByOperator() get empty slice")
	// 	return
	// }

	for _, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().GetByOperator() get null")
			return
		}
	}
}

var schToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTI4NjQzMSwiaXNzIjoia2lkc2xvb3AifQ.hrlqM9vbxyHfoolKwlGTJwFBlReHzJiaEy2Gdepm1bIR2O9XjzBQoasyUGAoK_hS5I8uXuad_Ilo354MzugQJIXJN1GWRiw03UDcvJAqbxuXecHtdCkLzes48wCPRV2KDuo9gTJkTQCEpsYxdDzyu-bOKY5vTEALenn79-HnLQULaKRLcGLw6N8ePk0fKUEuLtvIFOc40QhJkNbGNtM4dj-Ux8kyR-j6wof7tCy6BZG7MyjUKPaCKzdFpgaw673pOWCia_G0xGCT29nD_OLROxZdb6SueFT0DXVE4Gpki3fywbgVJt4ZRoeuh2AqsP5nMhimGxrHRaxg7q8ygJKtHQ"

func TestAmsSchoolService_GetByUsers(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = schToken
	IDs := []string{
		"1dd1d0b4-df1a-4486-bd6c-a89f9b92f779",
		"1e200965-df57-461e-8af3-e255886e8e41",
	}
	result, err := GetSchoolServiceProvider().GetByUsers(ctx, testOperator, orgID, IDs)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range result {
		fmt.Println(k, v)
	}
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
