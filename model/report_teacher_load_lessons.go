package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ITeacherLoadLessonsModel interface {
	Summary(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonSummaryArgs) (*entity.TeacherLoadLessonSummaryRes, error)
}

var (
	_teacherLoadLessonModel     ITeacherLoadLessonsModel
	_teacherLoadLessonModelOnce sync.Once
)

func GetTeacherLoadLessonsModel() ITeacherLoadLessonsModel {
	_teacherLoadLessonModelOnce.Do(func() {
		_teacherLoadLessonModel = new(teacherLoadLessonModel)
	})
	return _teacherLoadLessonModel
}

type teacherLoadLessonModel struct{}

func (t teacherLoadLessonModel) Summary(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonSummaryArgs) (*entity.TeacherLoadLessonSummaryRes, error) {
	panic("implement me")
}
