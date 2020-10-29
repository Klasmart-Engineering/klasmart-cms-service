package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IVisibilitySettingModel interface {
	Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.VisibilitySetting, error)
	GetByID(ctx context.Context, id string) (*entity.VisibilitySetting, error)
}

type visibilitySettingModel struct {
}

func (m *visibilitySettingModel) Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.VisibilitySetting, error) {
	panic("implement me")
}

func (m *visibilitySettingModel) GetByID(ctx context.Context, id string) (*entity.VisibilitySetting, error) {
	panic("implement me")
}

var (
	_visibilitySettingOnce  sync.Once
	_visibilitySettingModel IVisibilitySettingModel
)

func GetVisibilitySettingModel() IVisibilitySettingModel {
	_visibilitySettingOnce.Do(func() {
		_visibilitySettingModel = &visibilitySettingModel{}
	})
	return _visibilitySettingModel
}
