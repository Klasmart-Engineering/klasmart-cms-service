package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IAgeModel interface {
	Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Age, error)
	GetByID(ctx context.Context, id string) (*entity.Age, error)
}

type ageModel struct {
}

func (m *ageModel) Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Age, error) {
	panic("implement me")
}

func (m *ageModel) GetByID(ctx context.Context, id string) (*entity.Age, error) {
	panic("implement me")
}

var (
	_ageOnce  sync.Once
	_ageModel IAgeModel
)

func GetAgeModel() IAgeModel {
	_ageOnce.Do(func() {
		_ageModel = &ageModel{}
	})
	return _ageModel
}
