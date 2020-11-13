package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type ISkillModel interface {
	Query(ctx context.Context, condition *da.SkillCondition) ([]*entity.Skill, error)
	GetByID(ctx context.Context, id string) (*entity.Skill, error)
	Add(ctx context.Context, data *entity.Skill) (string, error)
	Update(ctx context.Context, data *entity.Skill) (string, error)
}

type skillModel struct {
}

func (m *skillModel) Add(ctx context.Context, data *entity.Skill) (string, error) {
	data.ID = utils.NewID()
	_, err := da.GetSkillDA().Insert(ctx, data)
	if err != nil {
		log.Error(ctx, "add error", log.Err(err), log.Any("data", data))
		return "", err
	}
	return data.ID, nil
}

func (m *skillModel) Update(ctx context.Context, data *entity.Skill) (string, error) {
	var old = new(entity.Skill)
	err := da.GetSkillDA().Get(ctx, data.ID, old)
	if err != nil {
		log.Error(ctx, "get error", log.Err(err), log.Any("data", data))
		return "", err
	}
	old.Name = data.Name
	_, err = da.GetSkillDA().Update(ctx, old)
	if err != nil {
		log.Error(ctx, "update error", log.Err(err), log.Any("data", data))
		return "", err
	}
	return old.ID, nil
}

func (m *skillModel) Query(ctx context.Context, condition *da.SkillCondition) ([]*entity.Skill, error) {
	var result []*entity.Skill
	err := da.GetSkillDA().Query(ctx, condition, &result)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

func (m *skillModel) GetByID(ctx context.Context, id string) (*entity.Skill, error) {
	var result = new(entity.Skill)
	err := da.GetSkillDA().Get(ctx, id, result)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "GetByID:not found", log.Err(err), log.String("id", id))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetByID error", log.Err(err), log.String("id", id))
		return nil, err
	}
	return result, nil
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
