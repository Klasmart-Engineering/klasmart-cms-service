package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type ISubjectModel interface {
	Query(ctx context.Context, condition *da.SubjectCondition) ([]*entity.Subject, error)
	GetByID(ctx context.Context, id string) (*entity.Subject, error)
	Add(ctx context.Context, data *entity.Subject) (string, error)
	Update(ctx context.Context, data *entity.Subject) (string, error)
}

type subjectModel struct {
}

func (m *subjectModel) Add(ctx context.Context, data *entity.Subject) (string, error) {
	data.ID = utils.NewID()
	_, err := da.GetSubjectDA().Insert(ctx, data)
	if err != nil {
		log.Error(ctx, "add error", log.Err(err), log.Any("data", data))
		return "", err
	}
	return data.ID, nil
}

func (m *subjectModel) Update(ctx context.Context, data *entity.Subject) (string, error) {
	var old = new(entity.Subject)
	err := da.GetSubjectDA().Get(ctx, data.ID, old)
	if err != nil {
		log.Error(ctx, "get error", log.Err(err), log.Any("data", data))
		return "", err
	}
	old.Name = data.Name
	_, err = da.GetSubjectDA().Update(ctx, old)
	if err != nil {
		log.Error(ctx, "update error", log.Err(err), log.Any("data", data))
		return "", err
	}
	return old.ID, nil
}

func (m *subjectModel) Query(ctx context.Context, condition *da.SubjectCondition) ([]*entity.Subject, error) {
	var result []*entity.Subject
	err := da.GetSubjectDA().Query(ctx, condition, &result)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

func (m *subjectModel) GetByID(ctx context.Context, id string) (*entity.Subject, error) {
	var result = new(entity.Subject)
	err := da.GetSubjectDA().Get(ctx, id, result)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "GetByID:not found", log.Err(err), log.String("id", id))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetByID error", log.Err(err), log.String("id", id))
		return nil, err
	}
	return result, nil
}

var (
	_subjectOnce  sync.Once
	_subjectModel ISubjectModel
)

func GetSubjectModel() ISubjectModel {
	_subjectOnce.Do(func() {
		_subjectModel = &subjectModel{}
	})
	return _subjectModel
}
