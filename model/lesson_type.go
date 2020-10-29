package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ILessonTypeModel interface {
	Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.LessonType, error)
	GetByID(ctx context.Context, id string) (*entity.LessonType, error)
}

type lessonTypeModel struct {
}

func (m *lessonTypeModel) Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.LessonType, error) {
	panic("implement me")
}

func (m *lessonTypeModel) GetByID(ctx context.Context, id string) (*entity.LessonType, error) {
	panic("implement me")
}

var (
	_lessonTypeOnce  sync.Once
	_lessonTypeModel ILessonTypeModel
)

func GetLessonTypeModel() ILessonTypeModel {
	_lessonTypeOnce.Do(func() {
		_lessonTypeModel = &lessonTypeModel{}
	})
	return _lessonTypeModel
}
