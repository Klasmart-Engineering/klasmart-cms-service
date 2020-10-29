package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IClassTypeModel interface {
	Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.ClassType, error)
	GetByID(ctx context.Context, id string) (*entity.ClassType, error)
}

type classTypeModel struct {
}

func (m *classTypeModel) Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.ClassType, error) {
	panic("implement me")
}

func (m *classTypeModel) GetByID(ctx context.Context, id string) (*entity.ClassType, error) {
	panic("implement me")
}

var (
	_classTypeOnce  sync.Once
	_classTypeModel IClassTypeModel
)

func GetClassTypeModel() IClassTypeModel {
	_classTypeOnce.Do(func() {
		_classTypeModel = &classTypeModel{}
	})
	return _classTypeModel
}
