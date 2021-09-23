package entity

import (
	"context"
)

type TeacherLoadMissedLessonsRequest TeacherLoadMissedLessonsArgs

func (t TeacherLoadMissedLessonsRequest) Validate(ctx context.Context, op *Operator) (TeacherLoadMissedLessonsArgs, error) {
	return TeacherLoadMissedLessonsArgs(t), nil
}

type TeacherLoadMissedLessonsArgs struct {
	TeacherId  string    `json:"teacher_id" form:"teacher_id"`
	ClassIDs   []string  `json:"class_ids" form:"class_ids"`
	Duration   TimeRange `json:"duration" form:"duration"`
	PageNumber int       `json:"page_number" form:"page_number"`
	PageSize   int       `json:"page_size" form:"page_size"`
}
type TeacherLoadMissedLessonsResponse struct {
	List []*TeacherLoadMissedLesson `json:"list"`
}

type TeacherLoadMissedLesson struct {
	LessonType  string `json:"lesson_type" form:"lesson_type"`
	LessonName  string `json:"lesson_name" form:"lesson_name"`
	ClassName   string `json:"class_name" form:"class_name"`
	NoOfStudent int    `json:"no_of_student" form:"no_of_student"`
	StartDate   int64  `json:"start_date" form:"start_date"`
	EndDate     int64  `json:"end_date" form:"end_date"`
}
