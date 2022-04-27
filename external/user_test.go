package external

import (
	"context"
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

// userID: 000d653d-7961-447c-8d66-ad8c4a40eae6
var orgID = "6300b3c5-8936-497e-ba1f-d67164b59c65"

func TestGetByOrganization(t *testing.T) {
	ctx := context.Background()
	//uf := UserFilter{
	//	UserID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: "000d653d-7961-447c-8d66-ad8c4a40eae6"},
	//}
	GetUserServiceProvider().GetByOrganization(ctx, testOperator, orgID)
}

var utestToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTA1ODE4NCwiaXNzIjoia2lkc2xvb3AifQ.Abs8v_zbaKfC7WQmS3-AXsddZruG8jq-AgyBrtaVyN3XGeSIFZvvSeJ8xiJug2QEWl8Afjfer5ndZph__qZ2rBCL-SMcF-rw-375cfvh6zOyZv1n7r8SereMnLugkY5nS2XCU-BPPgBVauf0ugONoLn_pKpN_nZ7jHr7U_OyKEuUeQ_CAmwRVGwpSFbiAtTdqvAYvrBr5-o52y0G_-LWiZ8GlM725NW87gUEHNb8Y7oSRN1xSz0mPObC1xjV_xlrT3g736mFmXBa0IbppDUFsqplEB-q_fHVBWyokF8G_qhESOeq167LGRyX0g_DPxk4HAGPBwkYe5qUOF-N5_yloQ"

func TestGetBy(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = utestToken
	GetUserServiceProvider().GetOnlyUnderOrgUsers(ctx, testOperator, orgID)
}
