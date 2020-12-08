package da

import (
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
	IDs entity.NullStrings

	OrderBy VisibilitySettingOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c VisibilitySettingCondition) GetConditions() ([]string, []interface{}) {
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
		return "number desc, name asc"
	}
}
