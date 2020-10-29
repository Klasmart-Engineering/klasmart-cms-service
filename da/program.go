package da

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

type IProgramDA interface {
	dbo.DataAccesser
}

type programDA struct {
	dbo.BaseDA
}

var (
	_programOnce sync.Once
	_programDA   IProgramDA
)

func GetProgramDA() IProgramDA {
	_programOnce.Do(func() {
		_programDA = &programDA{}
	})
	return _programDA
}

type ProgramCondition struct {
	OrderBy ProgramOrderBy
	Pager   dbo.Pager
}

func (c ProgramCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c ProgramCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c ProgramCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type ProgramOrderBy int

const (
	ProgramOrderByNameAsc = iota + 1
)

func NewProgramOrderBy(orderBy string) ProgramOrderBy {
	return ProgramOrderByNameAsc
}

func (c ProgramOrderBy) ToSQL() string {
	switch c {
	default:
		return "name"
	}
}
