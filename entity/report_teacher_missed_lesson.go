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
	PageNumber int       `json:"pageNumber" form:"pageNumber"`
	PageSize   int       `json:"pageSize" form:"pageSize"`
}
type TeacherLoadMissedLessonsResponse struct {
	List []*TeacherLoadMissedLesson `json:"list"`
}

type TeacherLoadMissedLesson struct {
	LessonType  []string  `json:"lesson_type" form:"lesson_type"`
	LessonName  []string  `json:"lesson_name" form:"lesson_name"`
	ClassName   TimeRange `json:"class_name" form:"class_name"`
	NoOfStudent TimeRange `json:"no_of_student" form:"no_of_student"`
	StartDate   TimeRange `json:"start_date" form:"start_date"`
	EndDate     TimeRange `json:"end_date" form:"end_date"`
}
