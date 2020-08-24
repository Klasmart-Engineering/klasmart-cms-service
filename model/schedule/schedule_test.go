package dyschedule

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"math/rand"
	"testing"
	"time"
)

func TestScheduleModel_Add(t *testing.T) {
	id := "1"
	start := time.Now().AddDate(0, 0, rand.Intn(10))
	viewdata := &entity.ScheduleAddView{
		Title:        fmt.Sprintf("%s_%s", id, "title"),
		ClassID:      "Class-1",
		LessonPlanID: fmt.Sprintf("%d", rand.Intn(10)),
		TeacherIDs:   []string{"Teacher-1", "Teacher-2"},
		OrgID:        "org-1",
		StartAt:      start.Unix(),
		EndAt:        start.Add(20 * time.Minute).Unix(),
		ModeType:     entity.ModeTypeAllDay,
		SubjectID:    "Subject-1",
		ProgramID:    "Program-2",
		ClassType:    "ClassType",
		DueAt:        0,
		Description:  "深刻理解的连接方式烦死了烦死了副书记",
		AttachmentID: "",
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

func TestScheduleModel_GetByID(t *testing.T) {

}
