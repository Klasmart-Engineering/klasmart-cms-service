package da

import (
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
	IDs entity.NullStrings

	OrderBy ClassTypeOrderBy
	Pager   dbo.Pager
}

func (c ClassTypeCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.IDs.Valid {
		wheres = append(wheres, fmt.Sprintf("id in (%s)", c.IDs.SQLPlaceHolder()))
		params = append(params, c.IDs.ToInterfaceSlice()...)
	}

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
