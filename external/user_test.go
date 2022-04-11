package external

import (
	"context"
	_ "embed"
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
