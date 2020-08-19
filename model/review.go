package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IReviewerModel interface {
	Approve(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error
	Reject(ctx context.Context, tx *dbo.DBContext, cid string, reason string, user *entity.Operator) error
}

type Reviewer struct {
}

func (rv *Reviewer) Approve(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error {
	// TODO:
	// 1. check auth
	// 2. get ContentModel
	cm := new(ContentModel)
	content, err := cm.GetContentById(ctx, tx, cid, user)
	if err != nil {
		log.Error(ctx, "Approve: GetContentById failed: ", log.Err(err))
		return err
	}
	err = content.SetStatus(entity.ContentStatusPublished)
	if err != nil {
		log.Error(ctx, "Approve: SetStatus failed: ", log.Err(err))
		return err
	}
	err = cm.UpdateContentPublishStatus(ctx, tx, cid, "", string(content.PublishStatus))
	if err != nil {
		log.Error(ctx, "Approve: Update Status failed: ", log.Err(err))
		return err
	}
	return nil
}

func (rv *Reviewer) Reject(ctx context.Context, tx *dbo.DBContext, cid string, reason string, user *entity.Operator) error {
	// TODO:
	// 1. check auth
	// 2. get ContentModel
	cm := new(ContentModel)
	content, err := cm.GetContentById(ctx, tx, cid, user)
	if err != nil {
		log.Error(ctx, "Reject: GetContentById failed: ", log.Err(err))
		return err
	}
	err = content.SetStatus(entity.ContentStatusRejected)
	if err != nil {
		log.Error(ctx, "Reject: SetStatus failed: ", log.Err(err))
		return err
	}
	err = cm.UpdateContentPublishStatus(ctx, tx, cid, reason, string(content.PublishStatus))
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
