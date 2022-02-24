package external

import (
	"context"
	"testing"
)

func TestAmsGradeService_BatchGet(t *testing.T) {
	ids := []string{"9d3e591d-06a6-4fc4-9714-cf155a15b415", "98461ca1-06a1-432a-97d0-4e1dff33e1a5"}
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
	grades, err := GetGradeServiceProvider().GetByProgram(context.TODO(), testOperator, "75004121-0c0d-486c-ba65-4c57deacb44b", WithStatus(Active))
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
