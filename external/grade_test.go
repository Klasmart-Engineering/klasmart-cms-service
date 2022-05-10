package external

import (
	"context"
	"testing"
)

func TestAmsGradeService_BatchGet(t *testing.T) {
	ids := []string{"98461ca1-06a1-432a-97d0-4e1dff33e1a5", "a9f0217d-f7ec-4add-950d-4e8986ab2c82"}
	grades, err := GetGradeServiceProvider().BatchGet(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetGradeServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(grades) != len(ids) {
		t.Errorf("GetGradeServiceProvider().BatchGet() want %d results got %d", len(ids), len(grades))
		return
	}

	for _, grade := range grades {
		if grade == nil {
			t.Error("GetGradeServiceProvider().BatchGet() get null")
			return
		}
	}
}

func TestAmsGradeService_GetByProgram(t *testing.T) {
	grades, err := GetGradeServiceProvider().GetByProgram(context.TODO(), testOperator, "04c630cc-fabe-4176-80f2-30a029907a33", WithStatus(Active))
	if err != nil {
		t.Errorf("GetGradeServiceProvider().GetByProgram() error = %v", err)
		return
	}

	if len(grades) == 0 {
		t.Error("GetGradeServiceProvider().GetByProgram() get empty slice")
		return
	}

	for _, grade := range grades {
		if grade == nil {
			t.Error("GetGradeServiceProvider().GetByProgram() get null")
			return
		}
	}
}

func TestAmsGradeService_GetByOrganization(t *testing.T) {
	grades, err := GetGradeServiceProvider().GetByOrganization(context.TODO(), testOperator, WithStatus(Active))
	if err != nil {
		t.Errorf("GetGradeServiceProvider().GetByOrganization() error = %v", err)
		return
	}

	if len(grades) == 0 {
		t.Error("GetGradeServiceProvider().GetByOrganization() get empty slice")
		return
	}

	for _, grade := range grades {
		if grade == nil {
			t.Error("GetGradeServiceProvider().GetByOrganization() get null")
			return
		}
	}
}
