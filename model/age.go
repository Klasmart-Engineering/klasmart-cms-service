package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IAgeModel interface {
	Query(ctx context.Context, condition *da.AgeCondition) ([]*entity.Age, error)
	GetByID(ctx context.Context, id string) (*entity.Age, error)
}

type ageModel struct{}

func (m *ageModel) Query(ctx context.Context, condition *da.AgeCondition) ([]*entity.Age, error) {
	var result []*entity.Age
	err := da.GetAgeDA().Query(ctx, condition, &result)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

func (m *ageModel) GetByID(ctx context.Context, id string) (*entity.Age, error) {
	var result = new(entity.Age)
	err := da.GetAgeDA().Get(ctx, id, result)
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
	_ageOnce  sync.Once
	_ageModel IAgeModel
)

func GetAgeModel() IAgeModel {
	_ageOnce.Do(func() {
		_ageModel = &ageModel{}
	})
	return _ageModel
}
