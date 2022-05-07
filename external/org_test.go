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

var orgIDs = []string{
	"60c064cc-bbd8-4724-b3f6-b886dce4774f",
	"6300b3c5-8936-497e-ba1f-d67164b59c65",
}
var schIDs = []string{
	"0adee5ec-9454-44a9-b894-05ca1768b01e",
	"0bf25570-337d-42fd-a594-09821f0d59fb",
	"2a1b4505-98c5-4809-8bf2-3a73663e2fca",
	"2f86886b-c84a-49be-8208-46060a95d54d",
	"433bba1e-5af2-4747-9e21-182d7e1eae96",
}

var orgToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTkzOTg5MywiaXNzIjoia2lkc2xvb3AifQ.YQmaKGViaG7bxpxPIHXGy9pRsKkkA61EwqOtT6jZMSntSpVUfqgx694PEt23qzmb-VB4rVIbbqk6TfzdQkx5pf7VexSVjHaH-a8cgzfyO8SBawH_58w7NUgg_KunYmklvnVBPufrcZQgOyptTpv0JdAJr5H3VdNAxMkbZZftI0UTFqD8Mtib2P5fcwcdjVyHa_zDBAhyebnADih9OfXZP6QnMgKwXLQHzdx3aFarEoh13Yl9wTHc9TvYqNA3MTr8JDL4gGaT6JI-ys0B3YGL2geIY-iseo99IWoMms5jJj05raowtODSZVbm0W9G-Jo1MV4WdFlK9iCa2dmkBwbvZg"

func TestAmsOrganizationService_GetByClasses(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = orgToken
	ids := []string{
		"00429737-f515-4348-b24f-919c2f82a2aa",
		"00beaf3c-a4c4-4927-8547-c75dc47947d8",
		"a4dfcb90-6523-4f72-b246-1eaaf26304fd",
	}

	provider := AmsOrganizationConnectionService{}

	organizations1, err := provider.AmsOrganizationService.GetByClasses(ctx, testOperator, ids)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() error = %v", err)
		return
	}

	organizations2, err := provider.GetByClasses(ctx, testOperator, ids)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() error = %v", err)
		return
	}
	fmt.Println("len:", len(organizations1) == len(organizations2))
}

func TestAmsOrganizationService_GetNameByOrganizationOrSchool(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = orgToken
	ids := make([]string, 0, len(orgIDs)+len(schIDs))
	ids = append(ids, orgIDs...)
	ids = append(ids, schIDs...)

	provider := AmsOrganizationConnectionService{}

	names1, err := provider.AmsOrganizationService.GetNameByOrganizationOrSchool(ctx, testOperator, ids)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() error = %v", err)
		return
	}

	names2, err := provider.GetNameByOrganizationOrSchool(ctx, testOperator, ids)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() error = %v", err)
		return
	}
	fmt.Println("len:", len(names1) == len(names2))
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

func TestAmsOrganizationService_GetByUserID(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = orgToken
	id := "000d653d-7961-447c-8d66-ad8c4a40eae6"
	provider := AmsOrganizationConnectionService{}

	organizations1, err := provider.AmsOrganizationService.GetByUserID(ctx, testOperator, id)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() error = %v", err)
		return
	}

	organizations2, err := provider.GetByUserID(ctx, testOperator, id)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() error = %v", err)
		return
	}
	fmt.Println("len:", len(organizations1) == len(organizations2))
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
