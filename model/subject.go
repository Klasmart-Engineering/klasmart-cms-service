package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ISubjectModel interface {
	Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Subject, error)
	GetByID(ctx context.Context, id string) (*entity.Subject, error)
}

type subjectModel struct {
}

func (m *subjectModel) Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.Subject, error) {
	panic("implement me")
}

func (m *subjectModel) GetByID(ctx context.Context, id string) (*entity.Subject, error) {
	panic("implement me")
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
