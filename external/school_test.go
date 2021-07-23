package external

import (
	"context"
	"testing"
)

func TestAmsSchoolService_GetByClasses(t *testing.T) {
	orgClassMap, err := GetClassServiceProvider().GetByOrganizationIDs(context.TODO(), testOperator, []string{testOperator.OrgID})
	if err != nil {
		t.Errorf("error = %v", err)
		return
	}
	orgClassList, ok := orgClassMap[testOperator.OrgID]
	if !ok || len(orgClassList) <= 0 {
		t.Errorf("error = %v", err)
		return
	}
	orgClassIDs := make([]string, len(orgClassList))
	for i, item := range orgClassList {
		orgClassIDs[i] = item.ID
	}
	t.Log(len(orgClassIDs))
	schools, err := GetSchoolServiceProvider().GetByClasses(context.TODO(),
		testOperator,
		orgClassIDs,
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
	userIDs = append(userIDs, userIDs...)
	userIDs = append(userIDs, userIDs...)
	schools, err := GetSchoolServiceProvider().GetByUsers(context.TODO(),
		testOperator,
		testOperator.OrgID,
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
	count := 0
	for key, school := range schools {
		if school == nil {
			t.Error("GetSchoolServiceProvider().GetByUsers() get null")
			return
		}
		t.Logf("%s:%d", key, len(school))
		if len(school) == 0 {
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
