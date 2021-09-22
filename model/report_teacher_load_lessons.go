package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ITeacherLoadLessonsModel interface {
	List(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error)
	Summary(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummary, error)
	MissedLessonsList(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadMissedLessonsArgs) ([]*entity.TeacherLoadMissedLesson, error)
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

func (t teacherLoadLessonModel) List(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error) {
	panic("implement me")
}

func (t teacherLoadLessonModel) Summary(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummary, error) {
	panic("implement me")
}
func (t teacherLoadLessonModel) MissedLessonsList(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadMissedLessonsArgs) ([]*entity.TeacherLoadMissedLesson, error) {
	panic("implement me")
}
