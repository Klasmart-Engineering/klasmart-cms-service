package da

import (
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ISkillDA interface {
	dbo.DataAccesser
}

type skillDA struct {
	dbo.BaseDA
}

var (
	_skillOnce sync.Once
	_skillDA   ISkillDA
)

func GetSkillDA() ISkillDA {
	_skillOnce.Do(func() {
		_skillDA = &skillDA{}
	})
	return _skillDA
}

type SkillCondition struct {
	IDs             entity.NullStrings
	DevelopmentalID sql.NullString

	OrderBy SkillOrderBy
	Pager   dbo.Pager
}

func (c SkillCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.DevelopmentalID.Valid {
		wheres = append(wheres, "developmental_id = ?")
		params = append(params, c.DevelopmentalID.String)
	}

	if c.IDs.Valid {
		wheres = append(wheres, fmt.Sprintf("id in (%s)", c.IDs.SQLPlaceHolder()))
		params = append(params, c.IDs.ToInterfaceSlice()...)
	}

	return wheres, params
}

func (c SkillCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c SkillCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type SkillOrderBy int

const (
	SkillOrderByNameAsc = iota + 1
)

func NewSkillOrderBy(orderBy string) SkillOrderBy {
	return SkillOrderByNameAsc
}

func (c SkillOrderBy) ToSQL() string {
	switch c {
	default:
		return "name"
	}
}
