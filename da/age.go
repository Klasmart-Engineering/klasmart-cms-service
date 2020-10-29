package da

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

type IAgeDA interface {
	dbo.DataAccesser
}

type ageDA struct {
	dbo.BaseDA
}

var (
	_ageOnce sync.Once
	_ageDA   IAgeDA
)

func GetAgeDA() IAgeDA {
	_ageOnce.Do(func() {
		_ageDA = &ageDA{}
	})
	return _ageDA
}

type AgeCondition struct {
	OrderBy AgeOrderBy
	Pager   dbo.Pager
}

func (c AgeCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c AgeCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c AgeCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type AgeOrderBy int

const (
	AgeOrderByNameAsc = iota + 1
)

func NewAgeOrderBy(orderBy string) AgeOrderBy {
	return AgeOrderByNameAsc
}

func (c AgeOrderBy) ToSQL() string {
	switch c {
	default:
		return "name"
	}
}
