package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

func TestGetClassEventBusModel(t *testing.T) {
	ctx := context.Background()
	err := GetClassEventBusModel().PubAddMembers(BusTopicClassAddMembers, ctx, &entity.Operator{UserID: "123"}, &entity.ScheduleClassEvent{
		ClassID: "33333",
		Users: []*entity.ScheduleClassUserEvent{
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
