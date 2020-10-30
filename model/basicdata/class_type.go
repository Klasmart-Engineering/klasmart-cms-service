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

type IClassTypeModel interface {
	Query(ctx context.Context, condition *da.ClassTypeCondition) ([]*entity.ClassType, error)
	GetByID(ctx context.Context, id string) (*entity.ClassType, error)
}

type classTypeModel struct {
}

func (m *classTypeModel) Query(ctx context.Context, condition *da.ClassTypeCondition) ([]*entity.ClassType, error) {
	var result []*entity.ClassType
	err := da.GetClassTypeDA().Query(ctx, condition, &result)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

func (m *classTypeModel) GetByID(ctx context.Context, id string) (*entity.ClassType, error) {
	var result = new(entity.ClassType)
	err := da.GetClassTypeDA().Get(ctx, id, result)
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
	_classTypeOnce  sync.Once
	_classTypeModel IClassTypeModel
)

func GetClassTypeModel() IClassTypeModel {
	_classTypeOnce.Do(func() {
		_classTypeModel = &classTypeModel{}
	})
	return _classTypeModel
}
