package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/basicdata"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type AgeServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*entity.Age, error)
}

func GetAgeServiceProvider() AgeServiceProvider {
	return &ageService{}
}

type mockAgeService struct{}

func (s mockAgeService) BatchGet(ctx context.Context, ids []string) ([]*entity.Age, error) {
	var ages []*entity.Age
	for _, option := range GetMockData().Options {
		ages = append(ages, option.Age...)
	}
	return ages, nil
}

type ageService struct{}

func (s ageService) BatchGet(ctx context.Context, ids []string) ([]*entity.Age, error) {
	result, err := basicdata.GetAgeModel().Query(ctx, &da.AgeCondition{
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
