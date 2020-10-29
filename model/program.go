package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IProgramModel interface {
	Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Program, error)
	GetByID(ctx context.Context, id string) (*entity.Program, error)
}

type programModel struct {
}

func (m *programModel) Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Program, error) {
	panic("implement me")
}

func (m *programModel) GetByID(ctx context.Context, id string) (*entity.Program, error) {
	panic("implement me")
}

var (
	_programOnce  sync.Once
	_programModel IProgramModel
)

func GetProgramModel() IProgramModel {
	_programOnce.Do(func() {
		_programModel = &programModel{}
	})
	return _programModel
}
