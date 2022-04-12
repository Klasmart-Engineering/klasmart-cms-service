package external

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/test/utils"
)

func TestCreateScheduleReview(t *testing.T) {
	ctx := context.TODO()
	op := &entity.Operator{}
	utils.InitConfig(ctx)
	createScheduleReviewRequest := CreateScheduleReviewRequest{
		ScheduleID:     "abc-131",
		DueAt:          123789,
		TimeZoneOffset: 0,
		ProgramID:      "class_id_1",
		SubjectIDs:     []string{"subject1"},
		ClassID:        "class_id_1",
		StudentIDs:     []string{"79d78876-79bb-4b79-9868-4a99e03ca757"},
		ContentStartAt: 0,
		ContentEndAt:   0,
	}
	err := GetScheduleReviewServiceProvider().CreateScheduleReview(ctx, op, createScheduleReviewRequest)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckScheduleReview(t *testing.T) {
	ctx := context.TODO()
	op := &entity.Operator{}
	utils.InitConfig(ctx)
	checkScheduleReviewRequest := CheckScheduleReviewRequest{
		TimeZoneOffset: 0,
		ProgramID:      "class_id_1",
		SubjectIDs:     []string{"subject1"},
		StudentIDs:     []string{"79d78876-79bb-4b79-9868-4a99e03ca757"},
		ContentStartAt: 0,
		ContentEndAt:   0,
	}
	resp, err := GetScheduleReviewServiceProvider().CheckScheduleReview(ctx, op, checkScheduleReviewRequest)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}

func TestDeleteScheduleReview(t *testing.T) {
	ctx := context.TODO()
	op := &entity.Operator{}
	utils.InitConfig(ctx)
	deleteScheduleReviewRequest := DeleteScheduleReviewRequest{
		ScheduleIDs: []string{"abc-131"},
	}
	resp, err := GetScheduleReviewServiceProvider().DeleteScheduleReview(ctx, op, deleteScheduleReviewRequest)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}
