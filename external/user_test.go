package external

import (
	"context"
	"reflect"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"

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
	config.Get().AMS.EndPoint = "https://api.alpha.kidsloop.net/user//user"
	GetUserServiceProvider().GetUserCount(ctx, &entity.Operator{
		Token: "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0NTU4NDU1NCwiaXNzIjoia2lkc2xvb3AifQ.i1Fl8RGmR5D_VLfK9x9A2DUmY1hO1nNUaw8-uxMTeRh_aZJMyHRJfGb33GnyLuM3V54MJVwj-FUBmgMagQz4ePiIHGS0vVbuhkD_ua2S-UhUG31MZVdIS1dywQ_olZG7emFMmTVoPobvhny6Wegi4Srhz861O5da7togly8BftyAjEAnD7LViosm8pIT1p5gqgC6X57VS19I8rDe6D03J6luT4aCOO-Hm5aRZH4ctwIiKaNWBb8YpQhsaU0kxu3X8B3pM88T1q9Ex1P3LlkSauuIYaiu7t2DRLleFwQDTDVflx9e3r92AuV7IhdlogQRJ2vL0eO3Am8yot8Bf0Di4g",
	}, entity.GetUserCountCondition{
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
