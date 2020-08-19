package model

import (
	"context"
	"sync"
)

type ILockLogModel interface {
	IsLock(ctx context.Context, recordID string) (bool, error)
	Lock(ctx context.Context, recordID string, operatorID string) error
	Unlock(ctx context.Context, recordID string) error
}

var (
	lockLogInstance     ILockLogModel
	lockLogInstanceOnce = sync.Once{}
)

func GetLockLogModel() ILockLogModel {
	lockLogInstanceOnce.Do(func() {
		lockLogInstance = &lockLogModel{}
	})
	return lockLogInstance
}

type lockLogModel struct{}

func (l *lockLogModel) IsLock(ctx context.Context, recordID string) (bool, error) {
	panic("implement me")
}

func (l *lockLogModel) Lock(ctx context.Context, recordID string, operatorID string) error {
	panic("implement me")
}

func (l *lockLogModel) Unlock(ctx context.Context, recordID string) error {
	panic("implement me")
}
