package model

import (
	"context"
	"sort"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IProgramGroupModel interface {
	GetByProgramID(ctx context.Context, id string, operator *entity.Operator) (*entity.ProgramGroup, error)
	Query(ctx context.Context, condition *da.ProgramGroupQueryCondition) ([]*entity.ProgramGroup, error)
	QueryMap(ctx context.Context, condition *da.ProgramGroupQueryCondition) (map[string]*entity.ProgramGroup, error)
	AllGroupNames(ctx context.Context) ([]string, error)
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
	err := da.GetProgramGroupDA().Get(ctx, id, pg)
	if err == dbo.ErrRecordNotFound {
		return nil, constant.ErrRecordNotFound
	}

	return pg, nil
}

func (s programGroupModel) QueryMap(ctx context.Context, condition *da.ProgramGroupQueryCondition) (map[string]*entity.ProgramGroup, error) {
	programGroups, err := s.Query(ctx, condition)
	if err != nil {
		return nil, err
	}

	dict := make(map[string]*entity.ProgramGroup, len(programGroups))
	for _, pg := range programGroups {
		dict[pg.ProgramID] = pg
	}

	return dict, nil
}

func (s programGroupModel) Query(ctx context.Context, condition *da.ProgramGroupQueryCondition) ([]*entity.ProgramGroup, error) {
	programGroups := []*entity.ProgramGroup{}
	err := da.GetProgramGroupDA().Query(ctx, condition, &programGroups)
	if err != nil {
		return nil, err
	}

	return programGroups, nil
}

func (s programGroupModel) AllGroupNames(ctx context.Context) ([]string, error) {
	programGroups, err := s.Query(ctx, &da.ProgramGroupQueryCondition{})
	if err != nil {
		return nil, err
	}

	groupNames := make([]string, 0, len(programGroups))
	for _, programGroup := range programGroups {
		groupNames = append(groupNames, string(programGroup.GroupName))
	}

	// add default group
	groupNames = append(groupNames, string(entity.ProgramGroupBadaMoreFeaturedContent))
	groupNames = utils.SliceDeduplication(groupNames)
	sort.Strings(groupNames)

	return groupNames, nil
}
