package external

import (
	"context"
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
	ids := []string{"9e285fc9-50fd-4cf2-ba5b-3f191c3338b4", "ac341b5f-a5f8-44a4-8b43-c0ff21d337b2"}
	names, err := GetOrganizationServiceProvider().GetNameByOrganizationOrSchool(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() error = %v", err)
		return
	}

	if len(names) != len(ids) {
		t.Error("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() get empty slice")
		return
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
