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
		ScheduleID:    "612f95c1249ae63b75456a04",
		AttendanceIDs: []string{"3235698f-3d5b-4a44-b79c-c6df48b3dc29", "be9dbfbe-8437-405d-8e6b-3e18dbb6a349"},
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
		ClassIDs:  []string{"d04a2fb9-b6ba-4542-9872-eabebde756fb", "968a820a-111c-40bd-82dc-9c2af4fe2129"},
		Durations: []entity.TimeRange{"1620554600-1631150516"},
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
		ClassIDs:  []string{"d04a2fb9-b6ba-4542-9872-eabebde756fb", "968a820a-111c-40bd-82dc-9c2af4fe2129"},
		Durations: []entity.TimeRange{"1620554600-1631150516"},
		Type:      string(entity.HomeFunType),
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
