package da

import (
	"database/sql"
	"fmt"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IOrganizationPropertyDA interface {
	dbo.DataAccesser
}

var (
	organizationPropertyDA    IOrganizationPropertyDA
	_organizationPropertyOnce sync.Once
)

func GetOrganizationPropertyDA() IOrganizationPropertyDA {
	_organizationPropertyOnce.Do(func() {
		organizationPropertyDA = new(OrganizationPropertyMySQLDA)
	})

	return organizationPropertyDA
}

type OrganizationPropertyCondition struct {
	IDs            entity.NullStrings
	Types          entity.NullStrings
	IncludeDeleted sql.NullBool
	OrderBy        OrganizationPropertyOrderBy
	Pager          dbo.Pager
}

func (c OrganizationPropertyCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.IDs.Valid {
		wheres = append(wheres, fmt.Sprintf("id in (%s)", c.IDs.SQLPlaceHolder()))
		params = append(params, c.IDs.ToInterfaceSlice()...)
	}

	if c.Types.Valid {
		wheres = append(wheres, fmt.Sprintf("`type` in (%s)", c.Types.SQLPlaceHolder()))
		params = append(params, c.Types.ToInterfaceSlice()...)
	}

	if !c.IncludeDeleted.Valid || !c.IncludeDeleted.Bool {
		wheres = append(wheres, "(delete_at=0)")
	}

	return wheres, params
}

func (c OrganizationPropertyCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c OrganizationPropertyCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type OrganizationPropertyOrderBy int

const (
	OrganizationPropertyOrderByTypeThenName = iota + 1
)

func NewOrganizationPropertyOrderBy(orderBy string) OrganizationPropertyOrderBy {
	return OrganizationPropertyOrderByTypeThenName
}

func (c OrganizationPropertyOrderBy) ToSQL() string {
	return "`type` asc, name asc"
}
