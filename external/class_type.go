package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/basicdata"
)

type ClassTypeServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*entity.ClassType, error)
}

func GetClassTypeServiceProvider() ClassTypeServiceProvider {
	return &classTypeService{}
}

type mockClassTypeService struct{}

func (s mockClassTypeService) BatchGet(ctx context.Context, ids []string) ([]*entity.ClassType, error) {
	return GetMockData().ClassTypes, nil
}

type classTypeService struct{}

func (s classTypeService) BatchGet(ctx context.Context, ids []string) ([]*entity.ClassType, error) {
	result, err := basicdata.GetClassTypeModel().Query(ctx, &da.ClassTypeCondition{
		IDs: entity.NullStrings{
			Strings: ids,
			Valid:   len(ids) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "BatchGet:error", log.Err(err), log.Strings("ids", ids))
		return nil, err
	}
	return result, nil
}
