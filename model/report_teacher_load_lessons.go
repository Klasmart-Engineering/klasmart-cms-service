package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (t *reportModel) ListTeacherLoadLessons(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error) {
	panic("implement me")
}

func (t *reportModel) SummaryTeacherLoadLessons(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummary, error) {
	panic("implement me")
}
func (t *reportModel) MissedLessonsList(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadMissedLessonsArgs) (*entity.TeacherLoadMissedLessonsResponse, error) {
	panic("implement me")
}
