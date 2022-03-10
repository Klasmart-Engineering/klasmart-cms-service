package model

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/test/utils"
)

func TestAdd(t *testing.T) {
	op := initOperator()

	schedule := &entity.ScheduleAddView{
		Title:      "ky's schedule",
		ClassType:  entity.ScheduleClassTypeHomework,
		IsHomeFun:  true,
		OutcomeIDs: []string{"60a36fd8de590052a3c5de00"},
	}
	outcomeIDs, err := GetScheduleModel().Add(context.TODO(), op, schedule)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(outcomeIDs)
}

func TestScheduleModel_GetByID(t *testing.T) {
	tt := time.Now().Add(1 * time.Hour).Unix()
	t.Log(tt)
	tt2 := time.Now().Add(2 * time.Hour).Unix()
	t.Log(tt2)
}

func TestTemp(t *testing.T) {
	diff := constant.ScheduleAllowGoLiveTime
	loc, _ := time.LoadLocation("Africa/Algiers")
	now := time.Now().In(loc)
	temp := now.Add(diff)
	fmt.Println(temp.Unix())
	now = time.Now()
	temp = now.Add(-diff)
	fmt.Println(temp)
}

func TestGetTeachLoadByCondition(t *testing.T) {
	cfg := &config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: "root:Badanamu123456@tcp(192.168.1.234:3306)/kidsloop2?parseTime=true&charset=utf8mb4",
		},
	}
	config.Set(cfg)
	option := dbo.WithConnectionString(cfg.DBConfig.ConnectionString)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		log.Println("connection mysql db error:", err)
		return
	}
	dbo.ReplaceGlobal(newDBO)
	ctx := context.Background()
	input := &entity.ScheduleTeachingLoadInput{
		OrgID:      "72e47ef0-92bf-4429-a06f-2014e3d3df4b",
		ClassIDs:   []string{"5751555a-cc18-4662-9ae5-a5ad90569f79", "49b3be6a-d139-4f82-9b77-0acc89525d3f"},
		TeacherIDs: []string{"4fde6e1b-8efe-58e9-a404-51fb98ebf9b8", "42098862-28b1-5417-9800-3b89e557a2b9"},
		TimeRanges: []*entity.ScheduleTimeRange{
			{
				StartAt: 1605456000,
				EndAt:   1605544740,
			},
			{
				StartAt: 1605456000,
				EndAt:   1605544740,
			},
			{
				StartAt: 1605456000,
				EndAt:   1605544740,
			},
		},
	}
	result, _ := GetScheduleModel().GetTeachingLoad(ctx, input)
	for _, item := range result {
		t.Log(item.TeacherID, ":", item.ClassType, item.Durations)
	}
}

func TestGetLearningOutcomeIDs(t *testing.T) {
	result, err := GetScheduleModel().GetLearningOutcomeIDs(context.TODO(), &entity.Operator{}, []string{"60f92a4cc9b482f7b98259a6", "6099c496e05f6e940027387c"})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

func TestQueryUnsafe(t *testing.T) {
	schedules, err := GetScheduleModel().QueryUnsafe(context.TODO(), &entity.ScheduleQueryCondition{
		IDs:               entity.NullStrings{Strings: []string{"60f929bb7604f720d1943c33", "60f92bd0f964d549bf922b46"}, Valid: true},
		RelationSchoolIDs: entity.NullStrings{Strings: []string{"60f929bb7604f720d1943c33", "60f92bd0f964d549bf922b46"}, Valid: true},
		ClassTypes:        entity.NullStrings{Strings: []string{string(entity.ScheduleClassTypeHomework), string(entity.ScheduleClassTypeLabelOnlineClass)}, Valid: true},
		IsHomefun:         sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range schedules {
		t.Log(v)
	}
}

func TestRemoveResourceMetadata(t *testing.T) {
	ctx := context.TODO()
	err := removeResourceMetadata(ctx, "schedule_attachment-6125d6e648bee33deac23bcc.jpg")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetRepeatResult(t *testing.T) {
	ctx := context.TODO()
	m := &scheduleModel{}
	r, err := m.getRepeatResult(ctx, 1632639194, 1632649980, &entity.RepeatOptions{
		Type: "daily",
		Daily: entity.RepeatDaily{
			Interval: 1,
			End: entity.RepeatEnd{
				Type:       "after_time",
				AfterCount: 1,
				AfterTime:  1642649980,
			},
		},
	}, time.Local)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range r {
		t.Log(v)
	}
}

func TestGetScheduleLiveLessonPlan(t *testing.T) {
	ctx := context.TODO()
	op := &entity.Operator{}
	scheduleID := "60a1d40a03b03c3acdb4f946"
	result, err := GetScheduleModel().GetScheduleLiveLessonPlan(ctx, op, scheduleID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

func TestGetScheduleRelationIDs(t *testing.T) {
	ctx := context.TODO()
	op := &entity.Operator{}
	scheduleID := "60b452aa0b409a6320720f3e"
	result, err := GetScheduleModel().GetScheduleRelationIDs(ctx, op, scheduleID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result.OrgID)
	t.Log(result.ClassRosterClassID)
	t.Log(result.ClassRosterTeacherIDs)
	t.Log(result.ClassRosterStudentIDs)
	t.Log(result.ParticipantTeacherIDs)
	t.Log(result.ParticipantStudentIDs)
}

func TestQueryScheduleTimeView(t *testing.T) {
	ctx := context.TODO()
	token := "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjRhMGNjZWU0LTA4MmItNDM5OC05NzMxLWEyNTUwNDE2MzkzNCIsImVtYWlsIjoidGVjaDFAeW9wbWFpbC5jb20iLCJleHAiOjE2NDY3NTkxMzEsImlzcyI6ImtpZHNsb29wIn0.bx5IXoruQs_ATYs0mJ4ODUvEjS9tRKxHVKdK85LlseujyYFTc2WidB3k9Twl7OaRFl9_EYl7PDhWH77Rg1ocR84GnOJAzNUIM9wpt5H4urnKEzWabtWdBjNBtbaRVPY4bKEeP5NaOF5Z72kKzYCsXWgTr9fIL_v2E4WGroXeUw3pRVNGvRcCWnLTM5fXWP7fy2NGdrJee4da-CSGWWCuRkF2nmveLujFTvpfiBQzFCZ0fW1xZ_ugYQMOeG_AWZ3PPdayivijYP9STphXCB3YXoXrLdWevzWJAu6JctXqMdbP7gJCORn6J5dZuen9iFak-9C8BB3uYHiBg8Lc0kLKEg"
	orgID := "92db7ddd-1f23-4f64-bd47-94f6d34a50c0"
	op := utils.InitOperator(ctx, token, orgID)
	query := &entity.ScheduleTimeViewListRequest{
		ViewType: "month",
		TimeAt:   1638316800,
	}
	count, result, err := GetScheduleModel().QueryScheduleTimeView(ctx, query, op, time.Local)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(count)
	t.Log(result)
}
