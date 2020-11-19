package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IProgramSubjectDA interface {
	dbo.DataAccesser
	DeleteByProgramID(ctx context.Context, tx *dbo.DBContext, programID string) error
}

type programSubjectDA struct {
	dbo.BaseDA
}

func (p *programSubjectDA) DeleteByProgramID(ctx context.Context, tx *dbo.DBContext, programID string) error {
	if err := tx.Where("program_id = ?", programID).
		Delete(&entity.ProgramSubject{}).Error; err != nil {
		log.Error(ctx, "delete error",
			log.Err(err),
			log.String("programID", programID),
		)
		return err
	}
	return nil
}

var (
	_programSubjectOnce sync.Once
	_programSubjectDA   IProgramSubjectDA
)

func GetProgramSubjectDA() IProgramSubjectDA {
	_programSubjectOnce.Do(func() {
		_programSubjectDA = &programSubjectDA{}
	})
	return _programSubjectDA
}

type ProgramSubjectCondition struct {
}

func (c ProgramSubjectCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c ProgramSubjectCondition) GetOrderBy() string {
	return ""
}

func (c ProgramSubjectCondition) GetPager() *dbo.Pager {
	return nil
}
