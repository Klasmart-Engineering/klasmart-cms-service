package external

import (
	"context"
	"testing"
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
}

func TestAmsAgeService_GetByProgram(t *testing.T) {
	ages, err := GetAgeServiceProvider().GetByProgram(context.TODO(), testOperator, "75004121-0c0d-486c-ba65-4c57deacb44b")
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByProgram() error = %v", err)
		return
	}

	if len(ages) == 0 {
		t.Error("GetAgeServiceProvider().GetByProgram() get empty slice")
		return
	}

	for _, age := range ages {
		if age == nil {
			t.Error("GetAgeServiceProvider().GetByProgram() get null")
			return
		}
	}
}

func TestAmsAgeService_GetByOrganization(t *testing.T) {
	ages, err := GetAgeServiceProvider().GetByOrganization(context.TODO(), testOperator, WithStatus(Active))
	if err != nil {
		t.Errorf("GetAgeServiceProvider().GetByOrganization() error = %v", err)
		return
	}

	if len(ages) == 0 {
		t.Error("GetAgeServiceProvider().GetByOrganization() get empty slice")
		return
	}

	for _, age := range ages {
		if age == nil {
			t.Error("GetAgeServiceProvider().GetByOrganization() get null")
			return
		}
	}
}
