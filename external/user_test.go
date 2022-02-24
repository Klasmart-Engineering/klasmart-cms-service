package external

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

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

func TestAmsUserService_GetUserCount(t *testing.T) {
	ctx := context.Background()
	GetUserServiceProvider().GetUserCount(ctx, testOperator, entity.GetUserCountCondition{
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
