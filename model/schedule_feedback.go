package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IScheduleFeedbackModel interface {
	Add(ctx context.Context, op *entity.Operator, input *entity.ScheduleFeedbackAddInput) (string, error)
	ExistByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string)
}

type scheduleFeedbackModel struct {
}

func (s *scheduleFeedbackModel) ExistByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) {
	panic("implement me")
}

func (s *scheduleFeedbackModel) Add(ctx context.Context, op *entity.Operator, input *entity.ScheduleFeedbackAddInput) (string, error) {
	panic("implement me")
}

var (
	_scheduleFeedbackOnce  sync.Once
	_scheduleFeedbackModel IScheduleFeedbackModel
)

func GetScheduleFeedbackModel() IScheduleFeedbackModel {
	_scheduleFeedbackOnce.Do(func() {
		_scheduleFeedbackModel = &scheduleFeedbackModel{}
	})
	return _scheduleFeedbackModel
}
