package external

import (
	"context"
	"testing"
	"time"
)

func TestAmsProgramService_BatchGet(t *testing.T) {
	ids := []string{"4591423a-2619-4ef8-a900-f5d924939d02", "14d350f1-a7ba-4f46-bef9-dc847f0cbac5"}
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
	time.Sleep(time.Second)
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
