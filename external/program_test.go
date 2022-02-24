package external

import (
	"context"
	"testing"
)

func TestAmsProgramService_BatchGet(t *testing.T) {
	ids := []string{"04c630cc-fabe-4176-80f2-30a029907a33", "7565ae11-8130-4b7d-ac24-1d9dd6f792f2"}
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
	programs, err := GetProgramServiceProvider().GetByOrganization(context.TODO(), testOperator, WithStatus(Inactive))
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
