package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"math/rand"
	"testing"
	"time"
)

func TestScheduleModel_Add(t *testing.T) {
	id := utils.NewID()
	start := time.Now().AddDate(0, 0, rand.Intn(10))
	viewdata := &entity.ScheduleAddView{
		Title:        fmt.Sprintf("%s_%s", id[0:4], "title"),
		ClassID:      fmt.Sprintf("%d", rand.Intn(10)),
		LessonPlanID: fmt.Sprintf("%d", rand.Intn(10)),
		TeacherIDs:   []string{"1", "2", "3"},
		OrgID:        "1",
		StartAt:      start.Unix(),
		EndAt:        start.Add(20 * time.Minute).Unix(),
		ModeType:     entity.ModeTypeAllDay,
		SubjectID:    "1",
		ProgramID:    "2",
		ClassType:    "ClassType",
		DueAt:        0,
		Description:  "深刻理解的连接方式烦死了烦死了副书记",
		AttachmentID: fmt.Sprintf("%d", rand.Intn(100)),
		Version:      0,
		Repeat:       entity.RepeatOptions{},
	}
	_, err := GetScheduleModel().Add(context.Background(), &entity.Operator{
		UserID: "1",
		Role:   "",
	}, viewdata)
	if err != nil {
		fmt.Println(err)
	}
}
