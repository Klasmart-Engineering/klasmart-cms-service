package external

import (
	"context"
	"testing"
)

func TestAmsSchoolService_GetByClasses(t *testing.T) {
	schools, err := GetSchoolServiceProvider().GetByClasses(context.TODO(),
		testOperator,
		[]string{"7a8b9212-d4ae-4bb8-bf5a-b3a425626790", "7a8b9212-d4ae-4bb8-bf5a-b3a425626790"},
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByClasses() error = %v", err)
		return
	}

	if len(schools) == 0 {
		t.Error("GetSchoolServiceProvider().GetByClasses() get empty slice")
		return
	}

	for _, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().GetByClasses() get null")
			return
		}
	}
}

func TestAmsSchoolService_GetByOrganizationID(t *testing.T) {
	schools, err := GetSchoolServiceProvider().GetByOrganizationID(context.TODO(),
		testOperator,
		"9e285fc9-50fd-4cf2-ba5b-3f191c3338b4",
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByOrganizationID() error = %v", err)
		return
	}

	if len(schools) == 0 {
		t.Error("GetSchoolServiceProvider().GetByOrganizationID() get empty slice")
		return
	}

	for _, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().GetByOrganizationID() get null")
			return
		}
	}
}

func TestAmsSchoolService_GetByPermission(t *testing.T) {
	schools, err := GetSchoolServiceProvider().GetByPermission(context.TODO(),
		testOperator,
		CreateContentPage201,
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByPermission() error = %v", err)
		return
	}

	// if len(schools) == 0 {
	// 	t.Error("GetSchoolServiceProvider().GetByPermission() get empty slice")
	// 	return
	// }

	for _, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().GetByPermission() get null")
			return
		}
	}
}

func TestAmsSchoolService_GetByOperator(t *testing.T) {
	schools, err := GetSchoolServiceProvider().GetByOperator(context.TODO(),
		testOperator,
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByOperator() error = %v", err)
		return
	}

	// if len(schools) == 0 {
	// 	t.Error("GetSchoolServiceProvider().GetByOperator() get empty slice")
	// 	return
	// }

	for _, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().GetByOperator() get null")
			return
		}
	}
}

func TestAmsSchoolService_GetByUsers(t *testing.T) {
	userInfos, err := GetUserServiceProvider().GetByOrganization(context.TODO(), testOperator, testOperator.OrgID)
	if err != nil {
		t.Error("GetUserServiceProvider.GetByOrganization error")
		return
	}
	userIDs := make([]string, len(userInfos))
	for i, item := range userInfos {
		userIDs[i] = item.ID
	}

	schools, err := GetSchoolServiceProvider().GetByUsers(context.TODO(),
		testOperator,
		"32bfe7ba-d897-4504-955c-8d6b484549c6",
		userIDs,
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().GetByUsers() error = %v", err)
		return
	}

	if len(schools) == 0 {
		t.Error("GetSchoolServiceProvider().GetByUsers() get empty slice")
		return
	}
	count:=0
	for key, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().GetByUsers() get null")
			return
		}
		t.Logf("%s:%d",key,len(school))
		if len(school)==0{
			count++
		}
	}
	t.Log(count)
}

func TestAmsSchoolService_BatchGet(t *testing.T) {
	ids := []string{"4460deda-ed0c-42d1-902d-8513070493be", "dc910c1e-23ac-4e29-a77e-b11b445be25a"}
	schools, err := GetSchoolServiceProvider().BatchGet(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetSchoolServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(schools) != len(ids) {
		t.Error("GetSchoolServiceProvider().BatchGet() get invalid slice")
		return
	}

	for _, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().BatchGet() get null")
			return
		}
	}
}