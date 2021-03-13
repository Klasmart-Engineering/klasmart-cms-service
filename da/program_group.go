package da

import (
	"fmt"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IProgramGroupDA interface {
	dbo.DataAccesser
}

type ProgramGroupMySQLDA struct {
	dbo.BaseDA
}

var (
	_programGroupOnce sync.Once
	_programGroupDA   IProgramGroupDA
)

func GetProgramGroupDA() IProgramGroupDA {
	_programGroupOnce.Do(func() {
		_programGroupDA = &ProgramGroupMySQLDA{}
	})
	return _programGroupDA
}

type ProgramGroupQueryCondition struct {
	ProgramIDs entity.NullStrings
}

func (c ProgramGroupQueryCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.ProgramIDs.Valid {
		wheres = append(wheres, fmt.Sprintf("program_id in (%s)", c.ProgramIDs.SQLPlaceHolder()))
		params = append(params, c.ProgramIDs.ToInterfaceSlice()...)
	}

	return wheres, params
}

func (c ProgramGroupQueryCondition) GetOrderBy() string {
	return "group_name"
}

func (c ProgramGroupQueryCondition) GetPager() *dbo.Pager {
	return nil
}
