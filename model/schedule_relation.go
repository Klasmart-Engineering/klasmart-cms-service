package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IScheduleRelationModel interface {
	Query(ctx context.Context, op *entity.Operator, condition *da.ScheduleRelationCondition) ([]*entity.ScheduleRelation, error)
}
type scheduleRelationModel struct {
}

func (s *scheduleRelationModel) Query(ctx context.Context, op *entity.Operator, condition *da.ScheduleRelationCondition) ([]*entity.ScheduleRelation, error) {
	var result []*entity.ScheduleRelation
	err := da.GetScheduleRelationDA().Query(ctx, condition, &result)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

var (
	_scheduleRelationOnce  sync.Once
	_scheduleRelationModel IScheduleRelationModel
)

func GetScheduleRelationModel() IScheduleRelationModel {
	_scheduleRelationOnce.Do(func() {
		_scheduleRelationModel = &scheduleRelationModel{}
	})
	return _scheduleRelationModel
}
