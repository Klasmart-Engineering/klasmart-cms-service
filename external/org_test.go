package external

import (
	"context"
	"fmt"
	"testing"
)

func TestOrganizationService_BatchGet(t *testing.T) {
	orgs, err := GetOrganizationServiceProvider().BatchGet(context.TODO(),
		testOperator,
		[]string{"3f135b91-a616-4c80-914a-e4463104dbac", "3f135b91-a616-4c80-914a-e4463104dbad"})
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(orgs) == 0 {
		t.Error("GetOrganizationServiceProvider().BatchGet() get empty slice")
		return
	}

	for _, org := range orgs {
		if org == nil {
			t.Error("GetOrganizationServiceProvider().BatchGet() get null")
			return
		}
	}
}

func TestAmsOrganizationService_GetNameByOrganizationOrSchool(t *testing.T) {
	ids := []string{"92db7ddd-1f23-4f64-bd47-94f6d34a50c0", "92db7ddd-1f23-4f64-bd47-94f6d34a50c0"}
	names, err := GetOrganizationServiceProvider().GetNameByOrganizationOrSchool(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() error = %v", err)
		return
	}

	if len(names) != len(ids) {
		t.Error("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() get empty slice")
		return
	}
	for i := range names {
		fmt.Printf("%#v\n", names[i])
	}
}

func TestAmsOrganizationService_GetByPermission(t *testing.T) {
	orgs, err := GetOrganizationServiceProvider().GetByPermission(context.TODO(),
		testOperator,
		CreateContentPage201,
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetByPermission() error = %v", err)
		return
	}

	if len(orgs) == 0 {
		t.Error("GetOrganizationServiceProvider().GetByPermission() get empty slice")
		return
	}

	for _, org := range orgs {
		if org == nil {
			t.Error("GetOrganizationServiceProvider().GetByPermission() get null")
			return
		}
	}
}

//var orgID = "60c064cc-bbd8-4724-b3f6-b886dce4774f"
var userID = "000d653d-7961-447c-8d66-ad8c4a40eae6"
var orgToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTA1OTk4NCwiaXNzIjoia2lkc2xvb3AifQ.hwwRFLcJTqD6Ohr80GF95de1yVgTO_7mg5sTHVRIJTTrzZGzRD-J6nAgBxBGZUv2zWIN-FYgholBrxiv1P2JmbDFlSq8P5KRp67R3N8Lp5NMNcFpOL5qEDHTzUlfm0RhhFsF_MvVGu5LXsdxvPWx2t8FVxz3ofN1frJd4TRCuXBus2O9G5Uk_VOd6imSoTNKhbTXmt0U5WnuOhKT8jmKu4iHodhJfxWbmOPUvTMwR2_L4XSwKs6Nv5AbvduZqyUUebkCtLT5KKEhM3JfDGRsikjkACtwetQhb_8zAAFimBRoF2A2AHw_AkvhME5eJ98p451fLnIMARnwUcIYIloS7A"

func TestAmsOrganizationService_GetByUserID(t *testing.T) {
	testOperator.Token = orgToken
	//testOperator.OrgID = orgID
	orgs, err := GetOrganizationServiceProvider().GetByUserID(context.TODO(),
		testOperator,
		userID,
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetByUserID() error = %v", err)
		return
	}

	if len(orgs) == 0 {
		t.Error("GetOrganizationServiceProvider().GetByUserID() get empty slice")
		return
	}

	for _, org := range orgs {
		if org == nil {
			t.Error("GetOrganizationServiceProvider().GetByUserID() get null")
			return
		}
	}
}

func TestOrganizationService_QueryByIDs(t *testing.T) {
	orgs, err := GetOrganizationServiceProvider().QueryByIDs(context.TODO(),
		[]string{
			"ad26d555-e9ad-4582-8fd6-c5e180847844",
			"00a91b89-02f2-4c36-8afd-5e3cdcfd1c86",
			"16ab82c3-355a-4002-883f-eb37b78b10a7",
			"f27efd10-000e-4542-bef2-0ccda39b93d3",
			"0ee01c37-c014-4c22-bb81-84d4f2a53b36"},
		testOperator)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().QueryByIDs() error = %v", err)
		return
	}

	if len(orgs) == 0 {
		t.Error("GetOrganizationServiceProvider().QueryByIDs() get empty slice")
		return
	}

	for _, org := range orgs {
		if org == nil {
			t.Error("GetOrganizationServiceProvider().QueryByIDs() get null")
			return
		}
	}
}
