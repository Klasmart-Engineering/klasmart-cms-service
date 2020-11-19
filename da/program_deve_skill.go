package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IProgramDeveSkillDA interface {
	dbo.DataAccesser
	DeleteByProgramIDAndDevelopmentID(ctx context.Context, tx *dbo.DBContext, programID string, developmentID string) error
}

type programDeveSkillDA struct {
	dbo.BaseDA
}

func (p *programDeveSkillDA) DeleteByProgramIDAndDevelopmentID(ctx context.Context, tx *dbo.DBContext, programID string, developmentID string) error {
	if err := tx.Where("program_id = ? and development_id = ?", programID, developmentID).
		Delete(&entity.DevelopmentSkill{}).Error; err != nil {
		log.Error(ctx, "delete error",
			log.Err(err),
			log.String("programID", programID),
		)
		return err
	}
	return nil
}

var (
	_programDeveSkillOnce sync.Once
	_programDeveSkillDA   IProgramDeveSkillDA
)

func GetProgramDeveSkillDA() IProgramDeveSkillDA {
	_programDeveSkillOnce.Do(func() {
		_programDeveSkillDA = &programDeveSkillDA{}
	})
	return _programDeveSkillDA
}

type ProgramDeveSkillCondition struct {
}

func (c ProgramDeveSkillCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c ProgramDeveSkillCondition) GetOrderBy() string {
	return ""
}

func (c ProgramDeveSkillCondition) GetPager() *dbo.Pager {
	return nil
}
