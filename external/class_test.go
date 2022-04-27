package external

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestAmsClassService_GetByUserIDs(t *testing.T) {
	userInfos, err := GetUserServiceProvider().GetByOrganization(context.TODO(), testOperator, testOperator.OrgID)
	if err != nil {
		t.Error("GetUserServiceProvider.GetByOrganization error")
		return
	}
	userIDs := make([]string, len(userInfos))
	for i, item := range userInfos {
		userIDs[i] = item.ID
	}
	fmt.Println(len(userIDs))
	userIDs = append(userIDs, userIDs...)
	classes, err := GetClassServiceProvider().GetByUserIDs(context.TODO(), testOperator, userIDs, WithStatus(Active))
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

var clsToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTA2OTcwMSwiaXNzIjoia2lkc2xvb3AifQ.ipjdb1QNMjpeZcJtRgTSTezFBdDJqM1G8aG_e5RG6wjrxwOBw5MeKhyf0ZT5duuqDhimPnvaE0MGGHZxTadfXmaf-xJv9RJKWNWU0xMWsdFzJQ6D0eMmWM5emhpXDPHy9I3DbbghgElUT_h9RvdtS40W4YT9cv-lyomV-2-m5RWi9kv-3klgaJzBpWb-etjfj8mL9w-h11e04CH3b5YPp7eP9kCI2fvg56hf41s-9cZCgIAq4lUqGEern8dhCGu2z_6KmmUbmndra9m8ZKa2rJsADUstNjfOGZn8ie3Lku_26OykLdqTF5NFhw4E0KhOMBxANojuR-U7Jb6RLSBiJQ"

func TestAmsClassService_GetOnlyUnderOrgClasses(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = clsToken
	result, err := GetClassServiceProvider().GetOnlyUnderOrgClasses(ctx, testOperator, orgID)
	if err != nil {
		t.Fatal(err)
	}
	bs, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(bs)
}
