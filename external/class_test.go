package external

import (
	"context"
	"testing"
)

func TestAmsClassService_GetByUserIDs(t *testing.T) {
	classes, err := GetClassServiceProvider().GetByUserIDs(context.TODO(), testOperator, []string{"335e0577-99cb-5d88-b5e1-dfdb14d5d4c2", "335e0577-99cb-5d88-b5e1-dfdb14d5d4c2"}, WithStatus(Active))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByUserIDs() error = %v", err)
		return
	}

	if len(classes) == 0 {
		t.Error("GetClassServiceProvider().GetByUserIDs() get empty slice")
		return
	}

	for _, class := range classes {
		if class == nil {
			t.Error("GetClassServiceProvider().GetByUserIDs() get null")
			return
		}
	}
}

func TestAmsClassService_GetByOrganizationIDs(t *testing.T) {
	classes, err := GetClassServiceProvider().GetByOrganizationIDs(context.TODO(), testOperator, []string{"9e285fc9-50fd-4cf2-ba5b-3f191c3338b4", "9e285fc9-50fd-4cf2-ba5b-3f191c3338b4"}, WithStatus(Inactive))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByOrganizationIDs() error = %v", err)
		return
	}

	if len(classes) == 0 {
		t.Error("GetClassServiceProvider().GetByOrganizationIDs() get empty slice")
		return
	}

	for _, class := range classes {
		if class == nil {
			t.Error("GetClassServiceProvider().GetByOrganizationIDs() get null")
			return
		}
	}
}

func TestAmsClassService_GetBySchoolIDs(t *testing.T) {
	classes, err := GetClassServiceProvider().GetBySchoolIDs(context.TODO(), testOperator, []string{"7215eab9-4b1c-437c-9f2f-0fdba5b0acb3", "fe8fdf9b-466f-45ba-a76a-51db173d02d4", "7215eab9-4b1c-437c-9f2f-0fdba5b0acb3"}, WithStatus(Active))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetBySchoolIDs() error = %v", err)
		return
	}

	if len(classes) == 0 {
		t.Error("GetClassServiceProvider().GetBySchoolIDs() get empty slice")
		return
	}

	for _, class := range classes {
		if class == nil {
			t.Error("GetClassServiceProvider().GetBySchoolIDs() get null")
			return
		}
	}
}
