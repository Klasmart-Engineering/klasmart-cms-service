package model

import (
	"context"
	"fmt"
	"github.com/KL-Engineering/dbo"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

func TestLessoons(t *testing.T) {
	ctx := context.Background()
	request := entity.TeacherLoadMissedLessonsRequest{TeacherId: "f05327e0-a729-52e1-a0ed-641168e37ba4",
		ClassIDs: []string{"0b3f5f4d-3540-42ab-9fda-474fbbe8b51c", "e05187ad-63bf-43a6-9a5c-72f1a64bce01"}, Page: 2, PageSize: -1,
		Duration: "1608600600-1630419300"}
	response, err := GetReportModel().MissedLessonsList(ctx, &request)
	fmt.Println(err)
	fmt.Println(response)

}

func TestClassAttendance(t *testing.T) {
	ctx := context.Background()
	request := entity.ClassAttendanceRequest{ClassID: "3702b30e-f72b-4ddc-a2c5-064343fdddba",
		StudentID:               "3ce8a5fa-f1aa-4f6a-8f4e-8703bdec789d",
		SelectedSubjectIDList:   []string{"f037ee92-212c-4592-a171-ed32fb892162"},
		UnSelectedSubjectIDList: []string{},
		//UnSelectedSubjectIDList: []string{"aecc64ff-be08-4b8e-b282-59e4e430617d", "22fc5ed6-894b-45c2-8410-8823dfdd13df", "4814ff17-fe15-40e1-b61a-a917742ec580", "f12276a9-4331-4699-b0fa-68e8df172843", "5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", "20d6ca2f-13df-4a7a-8dcb-955908db7baa", "7cf8d3a3-5493-46c9-93eb-12f220d101d0", "fab745e8-9e31-4d0c-b780-c40120c98b27", "66a453b0-d38f-472e-b055-7a94a94d66c4", "36c4f793-9aa3-4fb8-84f0-68a2ab920d5a", "b997e0d1-2dd7-40d8-847a-b8670247e96b", "49c8d5ee-472b-47a6-8c57-58daf863c2e1", "b19f511e-a46b-488d-9212-22c0369c8afd", "29d24801-0089-4b8e-85d3-77688e961efb", ""},
		Durations: []entity.TimeRange{"1634486400-1635091200"}}
	response, err := GetReportModel().ClassAttendanceStatistics(ctx, nil, &request)
	fmt.Println(err)
	fmt.Println(response.Items[0].Duration, response.Items[0].AttendancePercentage,
		response.Items[0].ClassAverageAttendancePercentage,
		response.Items[0].UnSelectedSubjectsAverageAttendancePercentage,
		response.Items[0].AttendedCount,
		response.Items[0].ScheduledCount)
	//fmt.Println(response.Items[1].Duration, response.Items[1].ClassAverageAttendancePercentage)
}

func TestAppInsightMessage(t *testing.T) {
	ctx := context.Background()
	request := entity.AppInsightMessageRequest{ClassID: "3702b30e-f72b-4ddc-a2c5-064343fdddba",
		StudentID: "3ce8a5fa-f1aa-4f6a-8f4e-8703bdec789d",
		EndTime:   1635091200}
	op := &entity.Operator{
		UserID: "afdfc0d9-ada9-4e66-b225-20f956d1a399",
		OrgID:  "60c064cc-bbd8-4724-b3f6-b886dce4774f",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0NTc3ODIyNiwiaXNzIjoia2lkc2xvb3AifQ.pAQe9Iu0k7GCX_YW26rCqRHPpdBAEKRzL23qkVjdbpJzVLBSn7brep3JzIjqioA3OEx53JZ7JzaVnv7dAvabr4CIPtJwxdIvtM6RB0UfzcDTI0qSfEpAr-TVLvw2oomxwnt7YOEd3xRr-V7B-T9l0auGOdStJwWNG60q1gdwpg9t6q9KIqAlAuyUDIOthsUi7-sT-jPoZtpXV9Riog0pilEqqejo5y3wYE6U5Xu5tIupYbikpAPdsA1DCY4T5KC06j4ao1YEdumjGEbC2YUOS__THbEq-69R5Fgv1RiuL98nQESAmrGE0TItNEk0Bf1rhRNcC0xzxTukr-WgIP4Zqw"

	response, err := GetReportModel().GetAppInsightMessage(ctx, op, &request)
	fmt.Println(err)
	fmt.Println(response)
	//fmt.Println(response.Items[1].Duration, response.Items[1].ClassAverageAttendancePercentage)
}

func TestListStudentsReport(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "afdfc0d9-ada9-4e66-b225-20f956d1a399",
		OrgID:  "60c064cc-bbd8-4724-b3f6-b886dce4774f",
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1NDQzNzczNCwiaXNzIjoia2lkc2xvb3AifQ.fFaQ6laXgJ9U_7_K7CA1R_qmZ5oCeN5jkdQubKU4u8mGZxXgDIX1KfZ9NlGkN0Fkup5gRvmvIwFkTAKav5mh6a46suiZhFD0HVqh90Lg2SlrywPRhT4X5qTWMVw68Q1DJdWLQXQ-mScuPzB_Pde1fN5wOrp4D0m_Z1nx2TM05oU3LhNVN11mk-Mq-3HRFwk6-gIDs2qY0qdv0VZUmBHnUCrSanyGPaqOSEYLw06Vzn40JVTNU_6b3cNPZ1h6UNfA4ZahaS9Ow5hfhEGtfVCBylnyd0h23jK_qHpCR7E_WVSAC9C80AGVhB9vH6wOLckW4N8VzSsmNXdHC0-UVzqnhg",
	}
	request := entity.ListStudentsAchievementReportRequest{
		TeacherID:    "7f0cfdd6-95c4-497a-a9c2-5422c67cb0fb",
		ClassID:      "46499f7d-c9aa-4368-9613-18c5ab0f3746",
		LessonPlanID: "628da79e552ba3b9994c9200",
		Status:       entity.ReportOutcomeStatusOptionAll,
		SortBy:       "desc",
	}
	response, err := GetReportModel().ListStudentsReport(ctx, dbo.MustGetDB(ctx), op, request)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(response)
}
