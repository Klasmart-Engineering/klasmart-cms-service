package basicdata

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ISubjectModel interface {
	Query(ctx context.Context, condition *da.SubjectCondition) ([]*entity.Subject, error)
	GetByID(ctx context.Context, id string) (*entity.Subject, error)
}

type subjectModel struct {
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
