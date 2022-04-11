package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
	"time"
)

func TestAmsSubjectService_BatchGet(t *testing.T) {
	ids := []string{"5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", "fab745e8-9e31-4d0c-b780-c40120c98b27"}
	subjects, err := GetSubjectServiceProvider().BatchGet(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetSubjectServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(subjects) != len(ids) {
		t.Errorf("GetSubjectServiceProvider().BatchGet() want %d results got %d", len(ids), len(subjects))
		return
	}

	for _, subject := range subjects {
		if subject == nil {
			t.Error("GetSubjectServiceProvider().BatchGet() get null")
			return
		}
	}
	time.Sleep(time.Second)
}

func TestAmsSubjectService_GetByProgram(t *testing.T) {
	op := &entity.Operator{
		Token: tokenOp,
	}
	subjects, err := GetSubjectServiceProvider().GetByProgram(context.TODO(), op, "04c630cc-fabe-4176-80f2-30a029907a33", WithStatus(Active))
	if err != nil {
		t.Errorf("GetSubjectServiceProvider().GetByProgram() error = %v", err)
		return
	}

	if len(subjects) == 0 {
		t.Error("GetSubjectServiceProvider().GetByProgram() get empty slice")
		return
	}

	for _, subject := range subjects {
		if subject == nil {
			t.Error("GetSubjectServiceProvider().GetByProgram() get null")
			return
		}
	}
}

func TestAmsSubjectService_GetByOrganization(t *testing.T) {
	subjects, err := GetSubjectServiceProvider().GetByOrganization(context.TODO(), testOperator, WithStatus(Active))
	if err != nil {
		t.Errorf("GetSubjectServiceProvider().GetByOrganization() error = %v", err)
		return
	}

	if len(subjects) == 0 {
		t.Error("GetSubjectServiceProvider().GetByOrganization() get empty slice")
		return
	}

	for _, subject := range subjects {
		if subject == nil {
			t.Error("GetSubjectServiceProvider().GetByOrganization() get null")
			return
		}
	}
}
