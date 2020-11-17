package da

import (
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
	IDs entity.NullStrings

	OrderBy LessonTypeOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c LessonTypeCondition) GetConditions() ([]string, []interface{}) {
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
