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

var schToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTI0NDg1NywiaXNzIjoia2lkc2xvb3AifQ.tlvFpzi2zm9R6NOwTYc4ctGtrkgpqqGABvnzTB6K9G1xrLQ5tSdZm6o6UGTJZZVLS3-QLgB4Nn4zhwX95yi6o4IC1ehRyvcK8gUgoDgD1BiI0pxjQwtId6-8yXhTkC5VyRQyh13nzkFAIOcP7W_JlNJEpbaOHTASRsu6CJndKnSZGvCrYU1FEo1HX28k1mzzo1NHmlrmobvIlzwUpM-Qcskd_qXlP7XHdz399kTHuedXeMbLzIMee1wluAJFMT0lBwpjBN5JGH39RhpFl65JSUfWj9h_7_UjSlKqrjdH2o04dt77GVkjsBADCNKnQtwu8VRIrETqNykkkA4_lOuqjQ"

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
