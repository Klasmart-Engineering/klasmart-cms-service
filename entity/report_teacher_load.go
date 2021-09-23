package entity

import (
	"context"
)

type TeacherLoadLessonRequest TeacherLoadLessonArgs

func (t TeacherLoadLessonRequest) Validate(ctx context.Context, op *Operator) (TeacherLoadLessonArgs, error) {
	return TeacherLoadLessonArgs(t), nil
}

type TeacherLoadLessonArgs struct {
	TeacherIDs []string  `json:"teacher_ids" form:"teacher_ids"`
	ClassIDs   []string  `json:"class_ids" form:"class_ids"`
	Duration   TimeRange `json:"duration" form:"duration"`
}

type TeacherLoadLesson struct {
	TeacherID               string `json:"teacher_id"`
	NumberOfClasses         int    `json:"number_of_classes"`
	NumberOfStudents        int    `json:"number_of_students"`
	CompletedLiveLessons    int    `json:"completed_live_Lessons"`
	CompletedInClassLessons int    `json:"completed_in_class_lessons"`
	MissedLiveLessons       int    `json:"missed_live_lessons"`
	MissedInClassLessons    int    `json:"missed_in_class_lessons"`
	TotalScheduled          int    `json:"total_scheduled"`
}

type SummaryNode struct {
	Count    int `json:"count"`
	Duration int `json:"duration"`
}
type TeacherLoadLessonSummary struct {
	CompletedLiveLessons    SummaryNode `json:"completed_live_lessons"`
	CompletedInClassLessons SummaryNode `json:"completed_in_class_lessons"`
	MissedLiveLessons       SummaryNode `json:"missed_live_lessons"`
	MissedInClassLessons    SummaryNode `json:"missed_in_class_lessons"`
}

type TeacherLoadLessonListResponse struct {
	List []*TeacherLoadLesson `json:"list"`
}

type TeacherLoadLessonSummaryResponse struct {
	Summary *TeacherLoadLessonSummary `json:"summary"`
}
