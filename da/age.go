package da

import (
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
	IDs       entity.NullStrings
	ProgramID sql.NullString

	OrderBy AgeOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c AgeCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.IDs.Valid {
		wheres = append(wheres, fmt.Sprintf("id in (%s)", c.IDs.SQLPlaceHolder()))
		params = append(params, c.IDs.ToInterfaceSlice()...)
	}
	if c.ProgramID.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where program_id = ? and %s.id = %s.age_id)",
			constant.TableNameProgramAge, constant.TableNameAge, constant.TableNameProgramAge)
		wheres = append(wheres, sql)
		params = append(params, c.ProgramID.String)
	}

	if c.DeleteAt.Valid {
		wheres = append(wheres, "delete_at>0")
	} else {
		wheres = append(wheres, "(delete_at=0)")
	}
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
