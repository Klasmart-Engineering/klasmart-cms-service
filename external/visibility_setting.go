package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/basicdata"
)

type VisibilitySettingServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*entity.VisibilitySetting, error)
}

func GetVisibilitySettingProvider() VisibilitySettingServiceProvider {
	return &visibilitySettingService{}
}

type visibilitySettingService struct{}

func (s visibilitySettingService) BatchGet(ctx context.Context, ids []string) ([]*entity.VisibilitySetting, error) {
	result, err := basicdata.GetVisibilitySettingModel().Query(ctx, &da.VisibilitySettingCondition{
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
