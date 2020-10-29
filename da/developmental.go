package da

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

type IDevelopmentalDA interface {
	dbo.DataAccesser
}

type developmentalDA struct {
	dbo.BaseDA
}

var (
	_developmentalOnce sync.Once
	_developmentalDA   IDevelopmentalDA
)

func GetDevelopmentalDA() IDevelopmentalDA {
	_developmentalOnce.Do(func() {
		_developmentalDA = &developmentalDA{}
	})
	return _developmentalDA
}

type DevelopmentalCondition struct {
	OrderBy DevelopmentalOrderBy
	Pager   dbo.Pager
}

func (c DevelopmentalCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c DevelopmentalCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c DevelopmentalCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type DevelopmentalOrderBy int

const (
	DevelopmentalOrderByNameAsc = iota + 1
)

func NewDevelopmentalOrderBy(orderBy string) DevelopmentalOrderBy {
	return DevelopmentalOrderByNameAsc
}

func (c DevelopmentalOrderBy) ToSQL() string {
	switch c {
	default:
		return "name"
	}
}
