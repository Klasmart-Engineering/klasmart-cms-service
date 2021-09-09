package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
	"time"
)

func TestClassesAssignmentsModel_CreateRecord(t *testing.T) {
	setup()
	ctx := context.Background()
	op := initOperator()
	args := &entity.AddClassAndLiveAssessmentArgs{
		ScheduleID:    "612fb63396a00415789ce44e",
		AttendanceIDs: []string{"487b1e29-7a68-42dc-b0c7-775ae116154f", "87cd7ff0-141d-4ab9-84ed-55d0810bebdd"},
		ClassEndTime:  time.Now().Unix(),
	}

	_, err := GetClassesAssignmentsModel().CreateRecord(ctx, op, args)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("ok")
}

func TestClassesAssignmentsModel_GetOverview(t *testing.T) {
	setup()
	ctx := context.Background()
	op := initOperator()
	request := &entity.ClassesAssignmentOverViewRequest{
		ClassIDs:  []string{"d04a2fb9-b6ba-4542-9872-eabebde756fb"},
		Durations: []entity.TimeRange{"1630554600-1631118121"},
	}
	result, err := GetClassesAssignmentsModel().GetOverview(ctx, op, request)
	if err != nil {
		t.Fatal(err)
	}
	for i := range result {
		fmt.Printf("%#v\n", result[i])
	}
}

func TestClassesAssignmentsModel_GetStatistic(t *testing.T) {
	setup()
	ctx := context.Background()
	op := initOperator()
	request := &entity.ClassesAssignmentsViewRequest{
		ClassIDs:  []string{"d04a2fb9-b6ba-4542-9872-eabebde756fb"},
		Durations: []entity.TimeRange{"1630554600-1631150516"},
		Type:      string(entity.LiveType),
	}
	result, err := GetClassesAssignmentsModel().GetStatistic(ctx, op, request)
	if err != nil {
		t.Fatal(err)
	}
	for i := range result {
		fmt.Printf("%#v\n", result[i])
	}
}

func TestClassesAssignmentsModel_GetUnattended(t *testing.T) {
	setup()
	ctx := context.Background()
	op := initOperator()
	request := &entity.ClassesAssignmentsUnattendedViewRequest{
		ClassID:   "d04a2fb9-b6ba-4542-9872-eabebde756fb",
		Durations: []entity.TimeRange{"1630554600-1631150516"},
		Type:      string(entity.LiveType),
	}
	result, err := GetClassesAssignmentsModel().GetUnattended(ctx, op, request)
	if err != nil {
		t.Fatal(err)
	}
	for i := range result {
		fmt.Printf("%#v\n", result[i])
	}
}
