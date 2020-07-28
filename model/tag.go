package model

import (
	"calmisland/kidsloop2/entity"
	"context"
	"sync"
)

type ITagModel interface{
	Add(ctx context.Context, tag *entity.TagAddView) (string,error)
	BatchAdd(ctx context.Context, tag []*entity.TagAddView) error
	Update(ctx context.Context, tag *entity.TagUpdateView) error
	Query(ctx context.Context, condition *entity.TagCondition) ([]*entity.TagView,error)
	GetByID(ctx context.Context,tagID string)(*entity.TagView,error)
}

type tagModel struct{}

var (
	_tagOnce sync.Once
	_tagModel ITagModel
)

func GetTagModel() ITagModel{
	_tagOnce.Do(func(){
		_tagModel = &tagModel{}
	})
	return _tagModel
}

func (t tagModel) Add(ctx context.Context, tag *entity.TagAddView) (string,error){

	return "",nil
}

func (t tagModel) BatchAdd(ctx context.Context, tag []*entity.TagAddView) error{
	return nil
}

func (t tagModel) Update(ctx context.Context, tag *entity.TagUpdateView) error{
	return nil
}

func (t tagModel) Query(ctx context.Context, condition *entity.TagCondition) ([]*entity.TagView,error){
	return nil,nil
}

func (t tagModel)GetByID(ctx context.Context,tagID string)(*entity.TagView,error){
	return nil,nil
}