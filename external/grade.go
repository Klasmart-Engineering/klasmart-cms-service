package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/basicdata"
)

type GradeServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*entity.Grade, error)
}

func GetGradeServiceProvider() GradeServiceProvider {
	return &gradeService{}
}

type mockGradeService struct{}

func (s mockGradeService) BatchGet(ctx context.Context, ids []string) ([]*entity.Grade, error) {
	var grades []*entity.Grade
	for _, option := range GetMockData().Options {
		grades = append(grades, option.Grade...)
	}
	return grades, nil
}

type gradeService struct{}

func (s gradeService) BatchGet(ctx context.Context, ids []string) ([]*entity.Grade, error) {
	result, err := basicdata.GetGradeModel().Query(ctx, &da.GradeCondition{
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
