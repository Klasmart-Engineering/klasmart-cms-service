package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IProgramDevelopmentDA interface {
	dbo.DataAccesser
	DeleteByProgramID(ctx context.Context, tx *dbo.DBContext, programID string) error
}

type programDevelopmentDA struct {
	dbo.BaseDA
}

func (p *programDevelopmentDA) DeleteByProgramID(ctx context.Context, tx *dbo.DBContext, programID string) error {
	if err := tx.Where("program_id = ?", programID).
		Delete(&entity.ProgramDevelopment{}).Error; err != nil {
		log.Error(ctx, "delete error",
			log.Err(err),
			log.String("programID", programID),
		)
		return err
	}
	return nil
}

var (
	_programDevelopmentOnce sync.Once
	_programDevelopmentDA   IProgramDevelopmentDA
)

func GetProgramDevelopmentDA() IProgramDevelopmentDA {
	_programDevelopmentOnce.Do(func() {
		_programDevelopmentDA = &programDevelopmentDA{}
	})
	return _programDevelopmentDA
}

type ProgramDevelopmentCondition struct {
}

func (c ProgramDevelopmentCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c ProgramDevelopmentCondition) GetOrderBy() string {
	return ""
}

func (c ProgramDevelopmentCondition) GetPager() *dbo.Pager {
	return nil
}
