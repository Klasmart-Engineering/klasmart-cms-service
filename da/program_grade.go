package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IProgramGradeDA interface {
	dbo.DataAccesser
	DeleteByProgramID(ctx context.Context, tx *dbo.DBContext, programID string) error
}

type programGradeDA struct {
	dbo.BaseDA
}

func (p *programGradeDA) DeleteByProgramID(ctx context.Context, tx *dbo.DBContext, programID string) error {
	if err := tx.Where("program_id = ?", programID).
		Delete(&entity.ProgramGrade{}).Error; err != nil {
		log.Error(ctx, "delete error",
			log.Err(err),
			log.String("programID", programID),
		)
		return err
	}
	return nil
}

var (
	_programGradeOnce sync.Once
	_programGradeDA   IProgramGradeDA
)

func GetProgramGradeDA() IProgramGradeDA {
	_programGradeOnce.Do(func() {
		_programGradeDA = &programGradeDA{}
	})
	return _programGradeDA
}

type ProgramGradeCondition struct {
}

func (c ProgramGradeCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c ProgramGradeCondition) GetOrderBy() string {
	return ""
}

func (c ProgramGradeCondition) GetPager() *dbo.Pager {
	return nil
}
