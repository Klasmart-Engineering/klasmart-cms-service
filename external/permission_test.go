package external

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

func TestHasAnyOrganizationPermission(t *testing.T) {
	config.LoadEnvConfig()
	has, err := GetPermissionServiceProvider().HasAnyOrganizationPermission(context.TODO(), &entity.Operator{
		UserID: "acebe3a4-9c76-5b14-bd90-1d1a0eb53e89",
		OrgID:  "",
		Token:  token,
	}, []string{"7e97287c-5e8b-4e78-9a4a-70b237bb5af5", "ae630b2e-59f8-4c35-8d17-57d6b9994f4e"}, "view_my_unpublished_learning_outcome_410")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(has)
}

func TestHasAnySchoolPermission(t *testing.T) {
	config.LoadEnvConfig()
	has, err := GetPermissionServiceProvider().HasAnySchoolPermission(context.TODO(), &entity.Operator{
		UserID: "acebe3a4-9c76-5b14-bd90-1d1a0eb53e89",
		OrgID:  "",
		Token:  token,
	}, []string{"7e97287c-5e8b-4e78-9a4a-70b237bb5af5", "ae630b2e-59f8-4c35-8d17-57d6b9994f4e"}, "view_my_unpublished_learning_outcome_410")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(has)
}
