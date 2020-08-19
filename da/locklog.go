package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ILockLogDA interface {
	Insert(ctx context.Context, log *entity.LockLog) error
	GetByID(ctx context.Context, id string) (*entity.LockLog, error)
	GetByRecordID(ctx context.Context, userID string) (*entity.LockLog, error)
	SoftDeleteByID(ctx context.Context, id string) error
	SoftDeleteByRecordID(ctx context.Context, recordID string) error
}

var (
	lockLogInstance     ILockLogDA
	lockLogInstanceOnce = sync.Once{}
)

func GetLockLogDA() ILockLogDA {
	lockLogInstanceOnce.Do(func() {
		lockLogInstance = &lockLogDA{}
	})
	return lockLogInstance
}
