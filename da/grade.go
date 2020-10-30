package da

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
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
	OrderBy GradeOrderBy
	Pager   dbo.Pager
}

func (c GradeCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

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
