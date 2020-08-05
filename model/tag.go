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

type ITagModel interface {
	Add(ctx context.Context, tag *entity.TagAddView) (string, error)
	Update(ctx context.Context, tag *entity.TagUpdateView) error
	Query(ctx context.Context, condition *da.TagCondition) ([]*entity.TagView, error)
	GetByID(ctx context.Context, id string) (*entity.TagView, error)
	GetByIDs(ctx context.Context, ids []string) ([]*entity.TagView, error)
	GetByName(ctx context.Context, name string) (*entity.TagView, error)
	Delete(ctx context.Context, id string) error
	Page(ctx context.Context, condition *da.TagCondition) (int64, []*entity.TagView, error)
}

type tagModel struct{}

var (
	_tagOnce  sync.Once
	_tagModel ITagModel
)

func GetTagModel() ITagModel {
	_tagOnce.Do(func() {
		_tagModel = &tagModel{}
	})
	return _tagModel
}

func (t tagModel) Add(ctx context.Context, tag *entity.TagAddView) (string, error) {
	old, err := t.GetByName(ctx, tag.Name)
	if err != nil && err != constant.ErrRecordNotFound {
		log.Info(ctx, "get tag by name", log.Err(err), log.String("tagName", tag.Name))
		return "", err
	}
	if old != nil {
		log.Info(ctx, "tag name duplicate record", log.String("tagName", tag.Name))
		return "", constant.ErrDuplicateRecord
	}
	in := &entity.Tag{
		ID:        utils.NewId(),
		Name:      tag.Name,
		States:    constant.Enable,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: 0,
		DeletedAt: 0,
	}
	err = da.GetTagDA().Insert(ctx, in)
	if err != nil {
		log.Error(ctx, "insert tag error", log.Err(err), log.Any("tagInfo", in))
		return "", utils.ConvertDynamodbError(err)
	}

	return in.ID, nil
}

func (t tagModel) Update(ctx context.Context, view *entity.TagUpdateView) error {
	old, _ := t.GetByName(ctx, view.Name)

	if old != nil && old.ID != view.ID {
		log.Info(ctx, "tag name duplicate record", log.String("tagName", view.Name))
		return constant.ErrDuplicateRecord
	}

	tag := &entity.Tag{
		ID:     view.ID,
		Name:   view.Name,
		States: view.States,
	}
	err := da.GetTagDA().Update(ctx, tag)

	return utils.ConvertDynamodbError(err)
}

func (t tagModel) Query(ctx context.Context, condition *da.TagCondition) ([]*entity.TagView, error) {
	tags, err := da.GetTagDA().Query(ctx, condition)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.TagView, len(tags))
	for i, item := range tags {
		result[i] = &entity.TagView{
			ID:       item.ID,
			Name:     item.Name,
			States:   item.States,
			CreateAt: item.CreatedAt,
		}
	}
	return result, nil
}

func (t tagModel) GetByID(ctx context.Context, id string) (*entity.TagView, error) {
	tag, err := da.GetTagDA().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tagView := &entity.TagView{
		ID:       tag.ID,
		Name:     tag.Name,
		CreateAt: tag.CreatedAt,
	}
	err = utils.ConvertDynamodbError(err)
	return tagView, err
}

func (t tagModel) GetByIDs(ctx context.Context, ids []string) ([]*entity.TagView, error) {
	tags, err := da.GetTagDA().GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make([]*entity.TagView, len(tags))
	for i, item := range tags {
		result[i] = &entity.TagView{
			ID:       item.ID,
			Name:     item.Name,
			States:   item.States,
			CreateAt: item.CreatedAt,
		}
	}
	return result, nil
}

func (t tagModel) GetByName(ctx context.Context, name string) (*entity.TagView, error) {
	tags, err := da.GetTagDA().Query(ctx, &da.TagCondition{
		Name:     name,
		DeleteAt: 0,
	})
	if err != nil {
		log.Error(ctx, "get tag by name error",log.Err(err), log.String("tagName", name))
		return nil, err
	}
	if len(tags) > 0 {
		tag := tags[0]
		return &entity.TagView{
			ID:       tag.ID,
			Name:     tag.Name,
			States:   tag.States,
			CreateAt: tag.CreatedAt,
		}, nil
	}
	return nil, constant.ErrRecordNotFound
}

func (t tagModel) Delete(ctx context.Context, id string) error {
	err := da.GetTagDA().Delete(ctx, id)
	err = utils.ConvertDynamodbError(err)
	if err == constant.ErrRecordNotFound {
		return nil
	}
	return err
}

func (t tagModel) Page(ctx context.Context, condition *da.TagCondition) (int64, []*entity.TagView, error) {
	total, tags, err := da.GetTagDA().Page(ctx, condition)
	err = utils.ConvertDynamodbError(err)
	if err != nil {
		return 0, nil, err
	}
	result := make([]*entity.TagView, len(tags))
	for i, item := range tags {
		result[i] = &entity.TagView{
			ID:       item.ID,
			Name:     item.Name,
			States:   item.States,
			CreateAt: item.CreatedAt,
		}
	}
	return total, result, nil
}
