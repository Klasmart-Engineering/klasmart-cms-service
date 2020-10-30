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

type IDevelopmentalModel interface {
	Query(ctx context.Context, condition *da.DevelopmentalCondition) ([]*entity.Developmental, error)
	GetByID(ctx context.Context, id string) (*entity.Developmental, error)
}

type developmentalModel struct {
}

func (m *developmentalModel) Query(ctx context.Context, condition *da.DevelopmentalCondition) ([]*entity.Developmental, error) {
	var result []*entity.Developmental
	err := da.GetDevelopmentalDA().Query(ctx, condition, &result)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

func (m *developmentalModel) GetByID(ctx context.Context, id string) (*entity.Developmental, error) {
	var result = new(entity.Developmental)
	err := da.GetDevelopmentalDA().Get(ctx, id, result)
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
	_developmentalOnce  sync.Once
	_developmentalModel IDevelopmentalModel
)

func GetDevelopmentalModel() IDevelopmentalModel {
	_developmentalOnce.Do(func() {
		_developmentalModel = &developmentalModel{}
	})
	return _developmentalModel
}
