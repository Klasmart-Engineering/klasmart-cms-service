package da

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

type ILessonTypeDA interface {
	dbo.DataAccesser
}

type lessonTypeDA struct {
	dbo.BaseDA
}

var (
	_lessonTypeOnce sync.Once
	_lessonTypeDA   ILessonTypeDA
)

func GetLessonTypeDA() ILessonTypeDA {
	_lessonTypeOnce.Do(func() {
		_lessonTypeDA = &lessonTypeDA{}
	})
	return _lessonTypeDA
}

type LessonTypeCondition struct {
	OrderBy LessonTypeOrderBy
	Pager   dbo.Pager
}

func (c LessonTypeCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c LessonTypeCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c LessonTypeCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type LessonTypeOrderBy int

const (
	LessonTypeOrderByNameAsc = iota + 1
)

func NewLessonTypeOrderBy(orderBy string) LessonTypeOrderBy {
	return LessonTypeOrderByNameAsc
}

func (c LessonTypeOrderBy) ToSQL() string {
	switch c {
	default:
		return "name"
	}
}
