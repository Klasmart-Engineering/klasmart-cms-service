package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IContentFilterModel interface{
	FilterPublishContent(ctx context.Context, c entity.ContentConditionRequest) (entity.ContentConditionRequest, error)
	FilterPendingContent(ctx context.Context, c entity.ContentConditionRequest) (entity.ContentConditionRequest, error)
	FilterArchivedContent(ctx context.Context, c entity.ContentConditionRequest) (entity.ContentConditionRequest, error)
}

type ContentFilterModel struct {

}

func (cf *ContentFilterModel) FilterPublishContent(ctx context.Context, c entity.ContentConditionRequest) (entity.ContentConditionRequest, error) {
	return c, nil
}
func (cf *ContentFilterModel) FilterPendingContent(ctx context.Context, c entity.ContentConditionRequest) (entity.ContentConditionRequest, error){
	return c, nil
}
func (cf *ContentFilterModel) FilterArchivedContent(ctx context.Context, c entity.ContentConditionRequest) (entity.ContentConditionRequest, error){
	return c, nil
}

var (
	_contentFilterModel IContentFilterModel
	_contentFilterModelOnce sync.Once
)

func GetContentFilterModel()IContentFilterModel{
	_contentFilterModelOnce.Do(func() {
		_contentFilterModel = new(ContentFilterModel)
	})
	return _contentFilterModel
}