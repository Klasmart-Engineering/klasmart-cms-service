package model

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IProgramGroupModel interface {
	GetByProgramID(ctx context.Context, id string, operator *entity.Operator) (*entity.ProgramGroup, error)
	QueryMap(ctx context.Context, condition *da.ProgramGroupQueryCondition) (map[string]*entity.ProgramGroup, error)
}

var (
	_programGroupOnce  sync.Once
	_programGroupModel IProgramGroupModel
)

func GetProgramGroupModel() IProgramGroupModel {
	_programGroupOnce.Do(func() {
		_programGroupModel = &programGroupModel{}
	})
	return _programGroupModel
}

type programGroupModel struct {
}

func (s programGroupModel) GetByProgramID(ctx context.Context, id string, operator *entity.Operator) (*entity.ProgramGroup, error) {
	pg := new(entity.ProgramGroup)
	err := da.GetProgramDA().Get(ctx, id, pg)
	if err == dbo.ErrRecordNotFound {
		return nil, constant.ErrRecordNotFound
	}

	return pg, nil
}

func (s programGroupModel) QueryMap(ctx context.Context, condition *da.ProgramGroupQueryCondition) (map[string]*entity.ProgramGroup, error) {
	programGroups := []*entity.ProgramGroup{}
	err := da.GetProgramGroupDA().Query(ctx, condition, &programGroups)
	if err != nil {
		return nil, err
	}

	dict := make(map[string]*entity.ProgramGroup, len(programGroups))
	for _, pg := range programGroups {
		dict[pg.ProgramID] = pg
	}

	return dict, nil
}
