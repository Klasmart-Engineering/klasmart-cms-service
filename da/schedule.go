package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IScheduleDA interface {
	Insert(ctx context.Context, schedule *entity.Schedule) error
	BatchInsert(ctx context.Context, schedule []*entity.Schedule) error

	Update(ctx context.Context, schedule *entity.Schedule) error
	BatchUpdate(ctx context.Context, schedule []*entity.Schedule) error

	Query(ctx context.Context, condition *ScheduleCondition) ([]*entity.Schedule, error)
	Page(ctx context.Context, condition *ScheduleCondition) (int64, []*entity.Schedule, error)
	GetByID(ctx context.Context, id string) (*entity.Schedule, error)

	SoftDelete(ctx context.Context, id string) error
	BatchSoftDelete(ctx context.Context, op *entity.Operator, condition *ScheduleCondition) error
}
