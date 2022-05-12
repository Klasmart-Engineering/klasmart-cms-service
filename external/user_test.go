package external

import (
	"context"
	"fmt"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

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

func TestAmsUserService_QueryByIDs(t *testing.T) {
	ids := []string{
		"fda18438-a998-48d2-b3fa-7be85707716f",
		"fe225a37-ee64-4ed2-969e-8846dad568e7",
		"f82b4f73-44de-4b16-9abc-be11d0fed30c",
		"cdc2cac4-10c4-4989-b39b-7c7867abee5d",
		"cccccccc-cccc-cccc-cccc-cccccccccccc"}
	users, err := GetUserServiceProvider().QueryByIDs(context.TODO(), ids, testOperator)
	if err != nil {
		t.Errorf("GetUserServiceProvider().QueryByIDs() error = %v", err)
		return
	}

	if len(users) != len(ids) {
		t.Errorf("GetUserServiceProvider().QueryByIDs() want %d results got %d", len(ids), len(users))
		return
	}

	for _, subject := range users {
		if subject == nil {
			t.Error("GetUserServiceProvider().QueryByIDs() get null")
			return
		}
	}
}
