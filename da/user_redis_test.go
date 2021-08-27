package da

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func TestSetUsers(t *testing.T) {
	ctx := context.TODO()
	users := []*User{
		{
			User: external.User{
				ID: "user_1",
			},
		},
		{
			User: external.User{
				ID: "user_2",
			},
		},
		{
			User: external.User{
				ID: "user_3",
			},
		},
	}
	err := GetUserRedisDA().SetUsers(ctx, "org123", users)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetUsersByOrg(t *testing.T) {
	ctx := context.TODO()
	users, err := GetUserRedisDA().GetUsersByOrg(ctx, "org123")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range users {
		t.Log(v)
	}
}
