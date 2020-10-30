package da

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

type ISubjectDA interface {
	dbo.DataAccesser
}

type subjectDA struct {
	dbo.BaseDA
}

var (
	_subjectOnce sync.Once
	_subjectDA   ISubjectDA
)

func GetSubjectDA() ISubjectDA {
	_subjectOnce.Do(func() {
		_subjectDA = &subjectDA{}
	})
	return _subjectDA
}

type SubjectCondition struct {
	OrderBy SubjectOrderBy
	Pager   dbo.Pager
}

func (c SubjectCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	return wheres, params
}

func (c SubjectCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c SubjectCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type SubjectOrderBy int

const (
	SubjectOrderByNameAsc = iota + 1
)

func NewSubjectOrderBy(orderBy string) SubjectOrderBy {
	return SubjectOrderByNameAsc
}

func (c SubjectOrderBy) ToSQL() string {
	switch c {
	default:
		return "name"
	}
}
