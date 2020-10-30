package da

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

type IVisibilitySettingDA interface {
	dbo.DataAccesser
}

type visibilitySettingDA struct {
	dbo.BaseDA
}

var (
	_visibilitySettingOnce sync.Once
	_visibilitySettingDA   IVisibilitySettingDA
)

func GetVisibilitySettingDA() IVisibilitySettingDA {
	_visibilitySettingOnce.Do(func() {
		_visibilitySettingDA = &visibilitySettingDA{}
	})
	return _visibilitySettingDA
}

type VisibilitySettingCondition struct {
	OrderBy VisibilitySettingOrderBy
	Pager   dbo.Pager
}

func (c VisibilitySettingCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c VisibilitySettingCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c VisibilitySettingCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type VisibilitySettingOrderBy int

const (
	VisibilitySettingOrderByNameAsc = iota + 1
)

func NewVisibilitySettingOrderBy(orderBy string) VisibilitySettingOrderBy {
	return VisibilitySettingOrderByNameAsc
}

func (c VisibilitySettingOrderBy) ToSQL() string {
	switch c {
	default:
		return "name"
	}
}
