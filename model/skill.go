package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ISkillModel interface {
	Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Skill, error)
	GetByID(ctx context.Context, id string) (*entity.Skill, error)
}

type skillModel struct {
}

func (m *skillModel) Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Skill, error) {
	panic("implement me")
}

func (m *skillModel) GetByID(ctx context.Context, id string) (*entity.Skill, error) {
	panic("implement me")
}

var (
	_skillOnce  sync.Once
	_skillModel ISkillModel
)

func GetSkillModel() ISkillModel {
	_skillOnce.Do(func() {
		_skillModel = &skillModel{}
	})
	return _skillModel
}
