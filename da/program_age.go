package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IProgramAgeDA interface {
	dbo.DataAccesser
	DeleteByProgramID(ctx context.Context, tx *dbo.DBContext, programID string) error
}

type programAgeDA struct {
	dbo.BaseDA
}

func (p *programAgeDA) DeleteByProgramID(ctx context.Context, tx *dbo.DBContext, programID string) error {
	if err := tx.Where("program_id = ?", programID).
		Delete(&entity.ProgramAge{}).Error; err != nil {
		log.Error(ctx, "delete error",
			log.Err(err),
			log.String("programID", programID),
		)
		return err
	}
	return nil
}

var (
	_programAgeOnce sync.Once
	_programAgeDA   IProgramAgeDA
)

func GetProgramAgeDA() IProgramAgeDA {
	_programAgeOnce.Do(func() {
		_programAgeDA = &programAgeDA{}
	})
	return _programAgeDA
}

type ProgramAgeCondition struct {
}

func (c ProgramAgeCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c ProgramAgeCondition) GetOrderBy() string {
	return ""
}

func (c ProgramAgeCondition) GetPager() *dbo.Pager {
	return nil
}
