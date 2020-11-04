package da

import (
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
	IDs       entity.NullStrings
	ProgramID sql.NullString

	OrderBy DevelopmentalOrderBy
	Pager   dbo.Pager
}

func (c DevelopmentalCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.IDs.Valid {
		wheres = append(wheres, fmt.Sprintf("id in (%s)", c.IDs.SQLPlaceHolder()))
		params = append(params, c.IDs.ToInterfaceSlice()...)
	}
	if c.ProgramID.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where program_id = ? and %s.id = %s.development_id)",
			constant.TableNameProgramDevelopment, constant.TableNameDevelopmental, constant.TableNameProgramDevelopment)
		wheres = append(wheres, sql)
		params = append(params, c.ProgramID.String)
	}

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
