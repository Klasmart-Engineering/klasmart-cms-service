package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IGradeModel interface {
	Query(ctx context.Context, condition *da.GradeCondition) ([]*entity.Grade, error)
	GetByID(ctx context.Context, id string) (*entity.Grade, error)
}

type gradeModel struct {
}

func (m *gradeModel) Query(ctx context.Context, condition *da.GradeCondition) ([]*entity.Grade, error) {
	var result []*entity.Grade
	err := da.GetGradeDA().Query(ctx, condition, &result)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

func (m *gradeModel) GetByID(ctx context.Context, id string) (*entity.Grade, error) {
	var result = new(entity.Grade)
	err := da.GetGradeDA().Get(ctx, id, result)
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
	_gradeOnce  sync.Once
	_gradeModel IGradeModel
)

func GetGradeModel() IGradeModel {
	_gradeOnce.Do(func() {
		_gradeModel = &gradeModel{}
	})
	return _gradeModel
}
