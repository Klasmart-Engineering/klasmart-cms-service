package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IGradeModel interface {
	Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Grade, error)
	GetByID(ctx context.Context, id string) (*entity.Grade, error)
}

type gradeModel struct {
}

func (m *gradeModel) Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Grade, error) {
	panic("implement me")
}

func (m *gradeModel) GetByID(ctx context.Context, id string) (*entity.Grade, error) {
	panic("implement me")
}

var (
	_gradeOnce  sync.Once
	_gradeModel IGradeModel
)

func GetGradeModel() IGradeModel {
	_gradeOnce.Do(func() {
		_gradeModel = &gradeModel{}
	})
	return _gradeModel
}
