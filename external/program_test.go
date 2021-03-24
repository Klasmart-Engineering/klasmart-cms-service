package external

import (
	"context"
	"testing"
)

func TestAmsProgramService_BatchGet(t *testing.T) {
	ids := []string{"b39edb9a-ab91-4245-94a4-eb2b5007c033", "14d350f1-a7ba-4f46-bef9-dc847f0cbac5"}
	programs, err := GetProgramServiceProvider().BatchGet(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetProgramServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(programs) != len(ids) {
		t.Errorf("GetProgramServiceProvider().BatchGet() want %d results got %d", len(ids), len(programs))
		return
	}

	for _, program := range programs {
		if program == nil {
			t.Error("GetProgramServiceProvider().BatchGet() get null")
			return
		}
	}
}

func TestAmsProgramService_GetByOrganization(t *testing.T) {
	programs, err := GetProgramServiceProvider().GetByOrganization(context.TODO(), testOperator)
	if err != nil {
		t.Errorf("GetProgramServiceProvider().GetByOrganization() error = %v", err)
		return
	}

	if len(programs) == 0 {
		t.Error("GetProgramServiceProvider().GetByOrganization() get empty slice")
		return
	}

	for _, program := range programs {
		if program == nil {
			t.Error("GetProgramServiceProvider().GetByOrganization() get null")
			return
		}
	}
}

func TestAmsProgramService_Query(t *testing.T) {
	programs, err := GetProgramServiceProvider().Query(context.TODO(),
		testOperator,
		WithOrganization(testOperator.OrgID),
		WithStatus(Inactive))
	if err != nil {
		t.Errorf("GetProgramServiceProvider().Query() error = %v", err)
		return
	}

	if len(programs) == 0 {
		t.Error("GetProgramServiceProvider().Query() get empty slice")
		return
	}

	for _, program := range programs {
		if program == nil {
			t.Error("GetProgramServiceProvider().Query() get null")
			return
		}
	}
}
