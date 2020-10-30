package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/basicdata"
)

type ProgramServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*entity.Program, error)
}

func GetProgramServiceProvider() ProgramServiceProvider {
	return &programService{}
}

type mockProgramService struct{}

func (s mockProgramService) BatchGet(ctx context.Context, ids []string) ([]*entity.Program, error) {
	var programs []*entity.Program
	for _, option := range GetMockData().Options {
		programs = append(programs, option.Program)
	}
	return programs, nil
}

type programService struct{}

func (s programService) BatchGet(ctx context.Context, ids []string) ([]*entity.Program, error) {
	result, err := basicdata.GetProgramModel().Query(ctx, &da.ProgramCondition{
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
