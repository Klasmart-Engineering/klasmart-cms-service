package external

import (
	"context"
	"fmt"
	"testing"
)

func TestOrganizationService_BatchGet(t *testing.T) {
	orgs, err := GetOrganizationServiceProvider().BatchGet(context.TODO(),
		testOperator,
		[]string{"3f135b91-a616-4c80-914a-e4463104dbac", "3f135b91-a616-4c80-914a-e4463104dbad"})
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(orgs) == 0 {
		t.Error("GetOrganizationServiceProvider().BatchGet() get empty slice")
		return
	}

	for _, org := range orgs {
		if org == nil {
			t.Error("GetOrganizationServiceProvider().BatchGet() get null")
			return
		}
	}
}

func TestAmsOrganizationService_GetNameByOrganizationOrSchool(t *testing.T) {
	ids := []string{"92db7ddd-1f23-4f64-bd47-94f6d34a50c0", "92db7ddd-1f23-4f64-bd47-94f6d34a50c0"}
	names, err := GetOrganizationServiceProvider().GetNameByOrganizationOrSchool(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() error = %v", err)
		return
	}

	if len(names) != len(ids) {
		t.Error("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() get empty slice")
		return
	}
	for i := range names {
		fmt.Printf("%#v\n", names[i])
	}
}

func TestAmsOrganizationService_GetByPermission(t *testing.T) {
	orgs, err := GetOrganizationServiceProvider().GetByPermission(context.TODO(),
		testOperator,
		CreateContentPage201,
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetByPermission() error = %v", err)
		return
	}

	if len(orgs) == 0 {
		t.Error("GetOrganizationServiceProvider().GetByPermission() get empty slice")
		return
	}

	for _, org := range orgs {
		if org == nil {
			t.Error("GetOrganizationServiceProvider().GetByPermission() get null")
			return
		}
	}
}

func TestAmsOrganizationService_GetByUserID(t *testing.T) {
	orgs, err := GetOrganizationServiceProvider().GetByUserID(context.TODO(),
		testOperator,
		"335e0577-99cb-5d88-b5e1-dfdb14d5d4c2",
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetByUserID() error = %v", err)
		return
	}

	if len(orgs) == 0 {
		t.Error("GetOrganizationServiceProvider().GetByUserID() get empty slice")
		return
	}

	for _, org := range orgs {
		if org == nil {
			t.Error("GetOrganizationServiceProvider().GetByUserID() get null")
			return
		}
	}
}

func TestOrganizationService_QueryByIDs(t *testing.T) {
	orgs, err := GetOrganizationServiceProvider().QueryByIDs(context.TODO(),
		[]string{
			"ad26d555-e9ad-4582-8fd6-c5e180847844",
			"00a91b89-02f2-4c36-8afd-5e3cdcfd1c86",
			"16ab82c3-355a-4002-883f-eb37b78b10a7",
			"f27efd10-000e-4542-bef2-0ccda39b93d3",
			"0ee01c37-c014-4c22-bb81-84d4f2a53b36"},
		testOperator)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().QueryByIDs() error = %v", err)
		return
	}

	if len(orgs) == 0 {
		t.Error("GetOrganizationServiceProvider().QueryByIDs() get empty slice")
		return
	}

	for _, org := range orgs {
		if org == nil {
			t.Error("GetOrganizationServiceProvider().QueryByIDs() get null")
			return
		}
	}
}
