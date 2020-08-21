package dynamodbhelper

import (
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

func TestGetUpdateBuilder(t *testing.T) {
	s := entity.Schedule{
		ID:           "1001",
		Title:        "schedule",
		ClassID:      "1",
		LessonPlanID: "2",
		TeacherIDs:   []string{"1", "2"},
		OrgID:        "3",
		StartAt:      0,
		EndAt:        0,
		ModeType:     "",
		SubjectID:    "",
		ProgramID:    "",
		ClassType:    "",
		DueAt:        0,
		Description:  "",
		AttachmentID: "",
		Version:      0,
		Repeat:       entity.RepeatOptions{},
		CreatedID:    "2",
		UpdatedID:    "3",
		DeletedID:    "2",
		CreatedAt:    0,
		UpdatedAt:    0,
		DeletedAt:    0,
	}
	result, err := GetUpdateBuilder(s)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)
}
