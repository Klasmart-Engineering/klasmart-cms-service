package da

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

type IClassTypeDA interface {
	dbo.DataAccesser
}

type classTypeDA struct {
	dbo.BaseDA
}

var (
	_classTypeOnce sync.Once
	_classTypeDA   IClassTypeDA
)

func GetClassTypeDA() IClassTypeDA {
	_classTypeOnce.Do(func() {
		_classTypeDA = &classTypeDA{}
	})
	return _classTypeDA
}

type ClassTypeCondition struct {
	OrderBy ClassTypeOrderBy
	Pager   dbo.Pager
}

func (c ClassTypeCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c ClassTypeCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c ClassTypeCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type ClassTypeOrderBy int

const (
	ClassTypeOrderByNameAsc = iota + 1
)

func NewClassTypeOrderBy(orderBy string) ClassTypeOrderBy {
	return ClassTypeOrderByNameAsc
}

func (c ClassTypeOrderBy) ToSQL() string {
	switch c {
	default:
		return "name"
	}
}
