package da

import (
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
	IDs       entity.NullStrings
	ProgramID sql.NullString

	OrderBy SubjectOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c SubjectCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.IDs.Valid {
		wheres = append(wheres, fmt.Sprintf("id in (%s)", c.IDs.SQLPlaceHolder()))
		params = append(params, c.IDs.ToInterfaceSlice()...)
	}
	if c.ProgramID.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where program_id = ? and %s.id = %s.subject_id)",
			constant.TableNameProgramSubject, constant.TableNameSubject, constant.TableNameProgramSubject)
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
		return "number asc, name asc"
	}
}
