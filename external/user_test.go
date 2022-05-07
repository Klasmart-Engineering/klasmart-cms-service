package external

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func TestAmsUserService_FilterByPermission(t *testing.T) {
	ids := []string{"f2626a21-3e98-517d-ac4a-ed6f33231869", "0a6091d7-1014-595d-abbf-dad456692d15"}
	want := []string{"f2626a21-3e98-517d-ac4a-ed6f33231869"}
	filtered, err := GetUserServiceProvider().FilterByPermission(context.TODO(), testOperator, ids, CreateContentPage201)
	if err != nil {
		t.Errorf("GetUserServiceProvider().FilterByPermission() error = %v", err)
		return
	}

	if !reflect.DeepEqual(filtered, want) {
		t.Errorf("GetUserServiceProvider().FilterByPermission() want %+v results got %+v", want, filtered)
		return
	}
}

func TestAmsUserService_GetUserCount(t *testing.T) {
	ctx := context.Background()
	GetUserServiceProvider().GetUserCount(ctx, testOperator, &entity.GetUserCountCondition{
		OrgID: entity.NullString{
			String: "60c064cc-bbd8-4724-b3f6-b886dce4774f",
			Valid:  true,
		},
		RoleID: entity.NullString{
			String: "913995b6-d4a9-4797-a1f0-1b4035da2a4b",
			Valid:  true,
		},
	})

}

var usrToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTkwOTU1NiwiaXNzIjoia2lkc2xvb3AifQ.kE2Bp7yFXhFeFqKclBRNhupZimdXCJBol64lMqqR2y_CWnTwF0WaU_eqqkH5YXmHtPHSUpME6ZWoaHDQNcuvRQ2QXSTHAIWLeM9t6WyPVKAZYMsbURdlHSrvqqxW_RzZgsQlYl3-2ra6DGHGrFj2hb678J07wGixvwrz5xl9LoEVIb3IkZ3nFLRWy3FiJ31cbZxJX8WtkgjVRO-V9mcImbwcIJluh3JqkjGoApEacldnAvuxpebLAJkQdkBdezOJwx38RH97CP_8Q7YzoeNkgFp6bNPPC0hSxUzDjY86way0cry2l0NNgtS3zGkw7O4FpxzQZPO_dhGbDGJCOK7q1A"

// userID: 000d653d-7961-447c-8d66-ad8c4a40eae6
var orgID = "6300b3c5-8936-497e-ba1f-d67164b59c65"

func TestGetByOrganization(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = usrToken
	ID := "6300b3c5-8936-497e-ba1f-d67164b59c65"
	provider := AmsUserConnectionService{}
	users1, err := provider.AmsUserService.GetByOrganization(ctx, testOperator, ID)
	if err != nil {
		t.Errorf("GetUserServiceProvider().GetByProgram() error = %v", err)
	}
	users2, err := provider.GetByOrganization(ctx, testOperator, ID)
	if err != nil {
		t.Errorf("GetUserServiceProvider().GetByProgram() error = %v", err)
	}
	fmt.Println("len:", len(users1) == len(users2))
}

func TestGetOnlyUnderOrgUsers(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = usrToken
	provider := AmsUserConnectionService{}
	ID := "6300b3c5-8936-497e-ba1f-d67164b59c65"
	users1, err := provider.AmsUserService.GetOnlyUnderOrgUsers(ctx, testOperator, ID)
	if err != nil {
		t.Errorf("GetUserServiceProvider().GetByProgram() error = %v", err)
	}
	users2, err := provider.GetOnlyUnderOrgUsers(ctx, testOperator, ID)
	if err != nil {
		t.Errorf("GetUserServiceProvider().GetByProgram() error = %v", err)
	}
	fmt.Println("len:", len(users1) == len(users2))
}
