package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (t *reportModel) ListTeacherLoadLessons(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error) {
	panic("implement me")
}

func (t *reportModel) SummaryTeacherLoadLessons(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummary, error) {
	panic("implement me")
}
func (t *reportModel) MissedLessonsList(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (response *entity.TeacherLoadMissedLessonsResponse, err error) {
	response = new(entity.TeacherLoadMissedLessonsResponse)
	da := da.GetReportDA()
	list, err := da.MissedLessonsListInfo(ctx, request)
	if err != nil {
		return nil, err
	}
	total, err := da.MissedLessonsListTotal(ctx, request)
	if err != nil {
		return nil, err
	}
	response.List = list
	response.Total = total
	return
}
