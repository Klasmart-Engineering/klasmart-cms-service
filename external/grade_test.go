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
	grades, err := GetGradeServiceProvider().GetByProgram(context.TODO(), testOperator, "75004121-0c0d-486c-ba65-4c57deacb44b")
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
