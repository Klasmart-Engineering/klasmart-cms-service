package model

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IProgramModel interface {
	GetByOrganization(ctx context.Context, operator *entity.Operator) ([]*entity.Program, error)
	GetByGroup(ctx context.Context, operator *entity.Operator, groupName string) ([]*entity.Program, error)
}

var (
	_programOnce  sync.Once
	_programModel IProgramModel
)

func GetProgramModel() IProgramModel {
	_programOnce.Do(func() {
		_programModel = &programModel{}
	})
	return _programModel
}

type programModel struct{}

func (s programModel) GetByOrganization(ctx context.Context, operator *entity.Operator) ([]*entity.Program, error) {
	programs, err := external.GetProgramServiceProvider().GetByOrganization(ctx, operator)
	if err != nil {
		return nil, err
	}

	return s.fillGroupName(ctx, programs)
}

func (s programModel) fillGroupName(ctx context.Context, programs []*external.Program) ([]*entity.Program, error) {
	condition := &da.ProgramGroupQueryCondition{
		ProgramIDs: entity.NullStrings{
			Strings: make([]string, len(programs)),
			Valid:   true,
		},
	}

	newPrograms := make([]*entity.Program, 0, len(programs))
	for _, program := range programs {
		newPrograms = append(newPrograms, &entity.Program{
			ID:   program.ID,
			Name: program.Name,
		})
		condition.ProgramIDs.Strings = append(condition.ProgramIDs.Strings, program.ID)
	}
	condition.ProgramIDs.Strings = utils.SliceDeduplication(condition.ProgramIDs.Strings)

	if len(programs) == 0 {
		return newPrograms, nil
	}

	groupMap, err := GetProgramGroupModel().QueryMap(ctx, condition)
	if err != nil {
		return nil, err
	}

	for _, program := range newPrograms {
		group, found := groupMap[program.ID]
		if !found {
			// Math, ESL, Science and other custom programs are "More Featured Content"
			program.GroupName = entity.ProgramGroupBadaMoreFeaturedContent
			continue
		}

		program.GroupName = group.GroupName
	}

	return newPrograms, nil
}

func (s programModel) GetByGroup(ctx context.Context, operator *entity.Operator, groupName string) ([]*entity.Program, error) {
	programs, err := s.GetByOrganization(ctx, operator)
	if err != nil {
		return nil, err
	}

	// filter group one by one @_@
	resultPrograms := make([]*entity.Program, 0, len(programs))
	for _, program := range programs {
		if string(program.GroupName) != groupName {
			continue
		}

		resultPrograms = append(resultPrograms, program)
	}

	return resultPrograms, nil
}
