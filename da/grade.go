package da

import (
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IGradeDA interface {
	dbo.DataAccesser
}

type gradeDA struct {
	dbo.BaseDA
}

var (
	_gradeOnce sync.Once
	_gradeDA   IGradeDA
)

func GetGradeDA() IGradeDA {
	_gradeOnce.Do(func() {
		_gradeDA = &gradeDA{}
	})
	return _gradeDA
}

type GradeCondition struct {
	IDs       entity.NullStrings
	ProgramID sql.NullString

	OrderBy GradeOrderBy
	Pager   dbo.Pager
}

func (c GradeCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.IDs.Valid {
		wheres = append(wheres, fmt.Sprintf("id in (%s)", c.IDs.SQLPlaceHolder()))
		params = append(params, c.IDs.ToInterfaceSlice()...)
	}
	if c.ProgramID.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where program_id = ? and %s.id = %s.grade_id)",
			constant.TableNameProgramGrade, constant.TableNameGrade, constant.TableNameProgramGrade)
		wheres = append(wheres, sql)
		params = append(params, c.ProgramID.String)
	}

	return wheres, params
}

func (c GradeCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c GradeCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type GradeOrderBy int

const (
	GradeOrderByNameAsc = iota + 1
)

func NewGradeOrderBy(orderBy string) GradeOrderBy {
	return GradeOrderByNameAsc
}

func (c GradeOrderBy) ToSQL() string {
	switch c {
	default:
		return "name"
	}
}
