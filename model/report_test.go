package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

func TestLessoons(t *testing.T) {
	ctx := context.Background()
	request := entity.TeacherLoadMissedLessonsRequest{TeacherId: "0e6b5f9d-0383-5ac6-b13d-8af02697fa8b",
		ClassIDs: []string{"0b3f5f4d-3540-42ab-9fda-474fbbe8b51c", "e05187ad-63bf-43a6-9a5c-72f1a64bce01"}, Page: 2, PageSize: -1,
		Duration: "1608600600-1630419300"}
	response, err := GetReportModel().MissedLessonsList(ctx, &request)
	fmt.Println(err)
	fmt.Println(response)

}


func TestClassAttendance(t *testing.T) {
	ctx := context.Background()
	request := entity.ClassAttendanceRequest{ClassID: "0a29685b-73dd-4225-b40f-cf27b89ba50a",
		StudentID:               "ffea3bde-7e39-4f46-b082-951a00a407c9",
		SelectedSubjectIDList:   []string{"20d6ca2f-13df-4a7a-8dcb-955908db7baa", "5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71"},
		UnSelectedSubjectIDList: []string{"1b622730-1135-43aa-b956-786365561f21", "19d8ebc4-7703-4d73-a44d-76052f729e40"},
		Durations:               []entity.TimeRange{"1633924080-1634528880", "1633232880-1633837680"}}
	response, err := GetReportModel().ClassAttendanceStatistics(ctx, &request)
	fmt.Println(err)
	fmt.Println(response.Items[0].ClassAverageAttendancePercentage)
	fmt.Println(response.Items[0].AttendedCount)
}