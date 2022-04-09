package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
	"time"
)

var tokenOp = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImExZmFhNTc1LWVhMGMtNGEzMC04YmI2LTViYjM3M2MwYjA5NCIsImVtYWlsIjoiYWxsMTEyNEB5b3BtYWlsLmNvbSIsImV4cCI6Mjc0ODcyMDcwMSwiaXNzIjoiY2FsbWlkLWRlYnVnIn0.qVfuPzeQFKvHlOg3aPh45rQ878LrGif5I3yb3eZj7Z8"

func TestAmsProgramService_BatchGet(t *testing.T) {
	op := &entity.Operator{
		Token: tokenOp,
	}
	ids := []string{"4591423a-2619-4ef8-a900-f5d924939d02", "14d350f1-a7ba-4f46-bef9-dc847f0cbac5"}
	programs, err := GetProgramServiceProvider().BatchGet(context.TODO(), op, ids)
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
