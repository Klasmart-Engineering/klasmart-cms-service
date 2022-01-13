package external

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

var token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0MTk5OTE2MywiaXNzIjoia2lkc2xvb3AifQ.SAoVGuuqP84Y4fr0WNlgED3JEKElmGwrZJDHhtIVPdTtD7lMWV02612k1hfP6Tvl1GcMl_pCMrcy60KiHpCxmp10cVrbf9oOFiezdfFql49gQm15Skng5S4vaWMhZeKaV5lfDfwIIp8dx4kugHrHTL2o5zolFeJFlSJLjV2BnyHM7h_Y5oZdLAyuMaG1c4hv6FsZDGvLenLlbf0M-B8CkGVIuYmfyJ82V_GDCRCGgniP8Nog1XIy4vGYRoiMWzN4eSURePGU9utoXBeBr63Ty397su7HNzgw8_OLU-5_YFOTinifnL9djvlIyfBedxSq97f1NQLvVg7te2ro4vpmAw"

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
