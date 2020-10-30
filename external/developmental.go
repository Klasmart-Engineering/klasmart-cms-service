package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/basicdata"
)

type DevelopmentalServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*entity.Developmental, error)
}

func GetDevelopmentalServiceProvider() DevelopmentalServiceProvider {
	return &developmentalService{}
}

type developmentalService struct{}

func (s developmentalService) BatchGet(ctx context.Context, ids []string) ([]*entity.Developmental, error) {
	result, err := basicdata.GetDevelopmentalModel().Query(ctx, &da.DevelopmentalCondition{
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

//type Developmental struct {
//	ID    string   `json:"id"`
//	Name  string   `json:"name"`
//	Skill []*Skill `json:"skills"`
//}
//type mockDevelopmentalService struct{}
//
//func (s mockDevelopmentalService) BatchGet(ctx context.Context, ids []string) ([]*Developmental, error) {
//	var developments []*Developmental
//	for _, option := range GetMockData().Options {
//		developments = append(developments, option.Developmental...)
//	}
//	return developments, nil
//}
