package da

import (
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
	IDs entity.NullStrings

	OrderBy ProgramOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c ProgramCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.IDs.Valid {
		wheres = append(wheres, fmt.Sprintf("id in (%s)", c.IDs.SQLPlaceHolder()))
		params = append(params, c.IDs.ToInterfaceSlice()...)
	}

	if c.DeleteAt.Valid {
		wheres = append(wheres, "delete_at>0")
	} else {
		wheres = append(wheres, "(delete_at=0)")
	}
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
