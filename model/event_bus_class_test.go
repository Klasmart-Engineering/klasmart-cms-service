package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

func TestGetClassEventBusModel(t *testing.T) {
	ctx := context.Background()
	err := GetClassEventBusModel().PubAddMembers(ctx, &entity.Operator{UserID: "123"}, &entity.ClassUpdateMembersEvent{
		ClassID: "33333",
		Members: []*entity.ClassMember{
			{
				ID:       "55555",
				RoleType: entity.ClassUserRoleTypeEventStudent,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
