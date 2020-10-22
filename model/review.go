package model

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	mutex "gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
)

type IReviewerModel interface {
	Approve(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error
	Reject(ctx context.Context, tx *dbo.DBContext, cid string, reasons []string, remark string, user *entity.Operator) error
}

type Reviewer struct {
}

func (rv *Reviewer) Approve(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error {
	// TODO:
	// 1. check auth
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixContentReview, cid)
	if err != nil {
		log.Error(ctx, "Get lock failed", log.String("cid", cid), log.Err(err))
		return err
	}
	locker.Lock()
	defer locker.Unlock()

	// 2. get ContentModel
	cm := new(ContentModel)
	content, err := cm.GetContentByID(ctx, tx, cid, user)
	if err != nil {
		log.Error(ctx, "Approve: GetContentByID failed:", log.Err(err))
		return err
	}
	err = content.SetStatus(entity.ContentStatusPublished)
	if err != nil {
		log.Error(ctx, "Approve: SetStatus failed: ", log.Err(err))
		return err
	}
	err = cm.UpdateContentPublishStatus(ctx, tx, cid, []string{}, "", string(content.PublishStatus))
	if err != nil {
		log.Error(ctx, "Approve: Update Status failed: ", log.Err(err))
		return err
	}
	return nil
}

func (rv *Reviewer) Reject(ctx context.Context, tx *dbo.DBContext, cid string, reasons []string, remark string, user *entity.Operator) error {
	// TODO:
	// 1. check auth
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixContentReview, cid)
	if err != nil {
		log.Error(ctx, "Get lock failed", log.String("cid", cid), log.Err(err))
		return err
	}
	locker.Lock()
	defer locker.Unlock()

	// 2. get ContentModel
	cm := new(ContentModel)
	content, err := cm.GetContentByID(ctx, tx, cid, user)
	if err != nil {
		log.Error(ctx, "Reject: GetContentByID failed: ", log.Err(err))
		return err
	}
	err = content.SetStatus(entity.ContentStatusRejected)
	if err != nil {
		log.Error(ctx, "Reject: SetStatus failed: ", log.Err(err))
		return err
	}
	err = cm.UpdateContentPublishStatus(ctx, tx, cid, reasons, remark, string(content.PublishStatus))
	if err != nil {
		log.Error(ctx, "Reject: Update Status failed: ", log.Err(err))
		return err
	}
	return nil
}

var reviewer *Reviewer
var _reviewerOnce sync.Once

func GetReviewerModel() IReviewerModel {
	_reviewerOnce.Do(func() {
		reviewer = new(Reviewer)
	})
	return reviewer
}
