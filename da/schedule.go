package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
)

type IScheduleDA interface {
	Insert(ctx context.Context, schedule *entity.Schedule) error
	BatchInsert(ctx context.Context, schedule []*entity.Schedule) error

	Update(ctx context.Context, schedule *entity.Schedule) error
	//BatchUpdate(ctx context.Context, schedule []*entity.Schedule) error

	Query(ctx context.Context, condition *dynamodbhelper.Condition) ([]*entity.Schedule, error)
	Page(ctx context.Context, condition *dynamodbhelper.Condition) ([]*entity.Schedule, error)
	GetByID(ctx context.Context, id string) (*entity.Schedule, error)

	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
}
