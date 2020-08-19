package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

const lockTimedOut = time.Hour * 12

type ILockLogModel interface {
	Allow(ctx context.Context, recordID string, operatorID string) (bool, error)
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

func (l *lockLogModel) Allow(ctx context.Context, recordID string, operatorID string) (bool, error) {
	item, err := da.GetLockLogDA().GetByRecordID(ctx, recordID)
	if err != nil {
		if err == constant.ErrRecordNotFound {
			return true, nil
		}
		log.Error(ctx, "allow: get lock by record id failed", log.Err(err))
		return false, err
	}
	if time.Unix(item.CreatedAt, 0).Add(lockTimedOut).After(time.Now()) {
		if err := l.Unlock(ctx, recordID); err != nil {
			log.Error(ctx, "allow: unlock overtime item failed", log.Err(err))
			return false, err
		}
		return false, nil
	}
	if item.OperatorID != operatorID {
		return false, nil
	}
	return true, nil
}

func (l *lockLogModel) Lock(ctx context.Context, recordID string, operatorID string) error {
	if err := da.GetLockLogDA().Insert(ctx, &entity.LockLog{
		ID:         utils.NewID(),
		RecordID:   recordID,
		OperatorID: operatorID,
		CreatedAt:  time.Now().Unix(),
	}); err != nil {
		log.Error(ctx, "lock: insert lock log failed", log.Err(err))
		return err
	}
	return nil
}

func (l *lockLogModel) Unlock(ctx context.Context, recordID string) error {
	if err := da.GetLockLogDA().SoftDeleteByRecordID(ctx, recordID); err != nil {
		log.Error(ctx, "unlock: soft delete by record id failed", log.Err(err))
		return err
	}
	return nil
}
