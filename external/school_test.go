package external

import (
	"context"
	"fmt"
	"testing"
)

func TestAmsSchoolService_GetByClasses(t *testing.T) {
	orgClassMap, err := GetClassServiceProvider().GetByOrganizationIDs(context.TODO(), testOperator, []string{testOperator.OrgID})
	if err != nil {
		t.Errorf("error = %v", err)
		return
	}
	orgClassList, ok := orgClassMap[testOperator.OrgID]
	if !ok || len(orgClassList) <= 0 {
		t.Errorf("error = %v", err)
		return
	}
	orgClassIDs := make([]string, len(orgClassList))
	for i, item := range orgClassList {
		orgClassIDs[i] = item.ID
	}
	t.Log(len(orgClassIDs))
	schools, err := GetSchoolServiceProvider().GetByClasses(context.TODO(),
		testOperator,
		orgClassIDs,
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByClasses() error = %v", err)
		return
	}

	if len(schools) == 0 {
		t.Error("GetSchoolServiceProvider().GetByClasses() get empty slice")
		return
	}

	for _, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().GetByClasses() get null")
			return
		}
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

var schToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTI0NTc1NywiaXNzIjoia2lkc2xvb3AifQ.o-LJMTvAF_OFTZ-oaSnwJcsRIyGpUqeZ5AZr1BR8v_OQ_EiMJnj-ONWpJDMdXq9Z3DWNBu78nIWtsit34k1Rrqk4vA8BzPd19KN6rrLQ9pJ4y2zKT3IKS68_qu3PvI-JhUscqw_xDQG7Av8SrSB8QHVUoKyder_hh1ErUdUzl0CXWbeUqHoHEXMlj8REkNv3mOedEW-f2Vea-gfCqlOVWb-9d5Up6qfpo79ynbXJM3fLpIPBnQVMnO0EWvA_YsBz1QYDRPFYUcqw5pd0eUmpZB-4YatFnpcr9VRZk3TuWgHcNEijL56CyIM7k-WKsh0Q-Ijom6o3KSJYsuIrIqdjxQ"

func TestAmsSchoolService_GetByUsers(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = schToken
	//orgID := "6300b3c5-8936-497e-ba1f-d67164b59c65"
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
