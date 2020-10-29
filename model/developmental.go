package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IDevelopmentalModel interface {
	Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Developmental, error)
	GetByID(ctx context.Context, id string) (*entity.Developmental, error)
}

type developmentalModel struct {
}

func (m *developmentalModel) Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Developmental, error) {
	panic("implement me")
}

func (m *developmentalModel) GetByID(ctx context.Context, id string) (*entity.Developmental, error) {
	panic("implement me")
}

var (
	_developmentalOnce  sync.Once
	_developmentalModel IDevelopmentalModel
)

func GetDevelopmentalModel() IDevelopmentalModel {
	_developmentalOnce.Do(func() {
		_developmentalModel = &developmentalModel{}
	})
	return _developmentalModel
}
