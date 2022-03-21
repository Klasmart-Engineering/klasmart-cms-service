package entity

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type TeacherLoadLessonRequest TeacherLoadLessonArgs

func (t TeacherLoadLessonRequest) Validate(ctx context.Context, op *Operator) (TeacherLoadLessonArgs, error) {
	if len(t.TeacherIDs) == 0 {
		return TeacherLoadLessonArgs{}, constant.ErrInvalidArgs
	}

	if len(t.ClassIDs) == 0 {
		return TeacherLoadLessonArgs{}, constant.ErrInvalidArgs
	}

	_, _, err := t.Duration.Value(ctx)
	if err != nil {
		log.Error(ctx, "Validate: duration not correct",
			log.Err(err),
			log.Any("duration", t.Duration))
		return TeacherLoadLessonArgs{}, constant.ErrInvalidArgs
	}
	return TeacherLoadLessonArgs(t), nil
}

type TeacherLoadLessonArgs struct {
	TeacherIDs []string  `json:"teacher_ids" form:"teacher_ids"`
	ClassIDs   []string  `json:"class_ids" form:"class_ids"`
	Duration   TimeRange `json:"duration" form:"duration"`
}

type TeacherLoadLesson struct {
	TeacherID               string `json:"teacher_id" gorm:"column:teacher_id"`
	NumberOfClasses         int    `json:"number_of_classes" gorm:"-"`
	NumberOfStudents        int    `json:"number_of_students" gorm:"-"`
	CompletedLiveLessons    int    `json:"completed_live_Lessons" gorm:"column:live_completed_count"`
	CompletedInClassLessons int    `json:"completed_in_class_lessons" gorm:"column:in_class_completed_count"`
	MissedLiveLessons       int    `json:"missed_live_lessons" gorm:"column:live_missed_count"`
	MissedInClassLessons    int    `json:"missed_in_class_lessons" gorm:"column:in_class_missed_count"`
	TotalScheduled          int    `json:"total_scheduled" gorm:"column:total_schedule"`
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

type TeacherLoadLessonSummaryFields struct {
	LiveCompletedCount       int `json:"live_completed_count" gorm:"column:live_completed_count"`
	LiveCompletedDuration    int `json:"live_completed_duration" gorm:"column:live_completed_duration"`
	InClassCompletedCount    int `json:"in_class_completed_count" gorm:"column:in_class_completed_count"`
	InClassCompletedDuration int `json:"in_class_completed_duration" gorm:"column:in_class_completed_duration"`
	LiveMissedCount          int `json:"live_missed_count" gorm:"column:live_missed_count"`
	LiveMissedDuration       int `json:"live_missed_duration" gorm:"column:live_missed_duration"`
	InClassMissedCount       int `json:"in_class_missed_count" gorm:"column:in_class_missed_count"`
	InClassMissedDuration    int `json:"in_class_missed_duration" gorm:"column:in_class_missed_duration"`
}

type TeacherLoadLessonListResponse struct {
	List []*TeacherLoadLesson `json:"list"`
}

type TeacherLoadLessonSummaryResponse struct {
	Summary *TeacherLoadLessonSummary `json:"summary"`
}

type TeacherLoadOverview struct {
	NumOfMissedLessons            int `json:"num_of_missed_lessons"`
	NumOfTeachersCompletedAll     int `json:"num_of_teachers_completed_all"`
	NumOfTeachersMissedSome       int `json:"num_of_teachers_missed_some"`
	NumOfTeachersMissedFrequently int `json:"num_of_teachers_missed_frequently"`
}
