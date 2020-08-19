package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type lockLogDA struct{}

func (da *lockLogDA) Insert(ctx context.Context, log *entity.LockLog) error {
	panic("implement me")
}

func (da *lockLogDA) GetByID(ctx context.Context, id string) (*entity.LockLog, error) {
	panic("implement me")
}

func (da *lockLogDA) GetByRecordID(ctx context.Context, userID string) (*entity.LockLog, error) {
	panic("implement me")
}

func (da *lockLogDA) SoftDeleteByID(ctx context.Context, id string) error {
	panic("implement me")
}

func (da *lockLogDA) SoftDeleteByRecordID(ctx context.Context, recordID string) error {
	panic("implement me")
}
