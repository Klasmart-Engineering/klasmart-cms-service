package da

import (
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type scheduleDynamoDA struct{}

func (s scheduleDynamoDA) Insert(ctx context.Context, schedule *entity.Schedule) error {
	panic("implement me")
}

func (s scheduleDynamoDA) BatchInsert(ctx context.Context, schedule []*entity.Schedule) error {
	panic("implement me")
}

func (s scheduleDynamoDA) Update(ctx context.Context, schedule *entity.Schedule) error {
	panic("implement me")
}

func (s scheduleDynamoDA) BatchUpdate(ctx context.Context, schedule []*entity.Schedule) error {
	panic("implement me")
}

func (s scheduleDynamoDA) Query(ctx context.Context, condition *ScheduleCondition) ([]*entity.Schedule, error) {
	panic("implement me")
}

func (s scheduleDynamoDA) Page(ctx context.Context, condition *ScheduleCondition) (int64, []*entity.Schedule, error) {
	panic("implement me")
}

func (s scheduleDynamoDA) GetByID(ctx context.Context, id string) (*entity.Schedule, error) {
	panic("implement me")
}

func (s scheduleDynamoDA) DeleteSoft(ctx context.Context, id string) error {
	panic("implement me")
}

func (s scheduleDynamoDA) BatchDeleteSoft(ctx context.Context, op *entity.Operator, condition *ScheduleCondition) error {
	panic("implement me")
}

type ScheduleCondition struct {
	Pager utils.Pager

	DeleteAt entity.NullInt
}

func (s ScheduleCondition) GetCondition() (expression.Expression, error) {
	return expression.Expression{}, nil
}

var (
	_scheduleOnce sync.Once
	_scheduleDA   IScheduleDA
)

func GetScheduleDA() IScheduleDA {
	_scheduleOnce.Do(func() {
		_scheduleDA = scheduleDynamoDA{}
	})
	return _scheduleDA
}
