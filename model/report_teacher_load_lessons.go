package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ITeacherLoadLessonsModel interface {
	List(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error)
	Summary(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummary, error)
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
	GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		RelationIDs:  entity.NullStrings{Strings: args.ClassIDs, Valid: true},
		RelationType: sql.NullString{String: string(entity.ScheduleRelationTypeClassRosterClass), Valid: true},
	})
	panic("implement me")
}

func (t teacherLoadLessonModel) Summary(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummary, error) {
	panic("implement me")
}
