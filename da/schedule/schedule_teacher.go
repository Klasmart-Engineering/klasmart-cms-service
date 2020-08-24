package schedule

import (
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

type IScheduleTeacherDA interface {
	dbo.DataAccesser
}

type scheduleTeacherDA struct {
	dbo.BaseDA
}

var (
	_scheduleTeacherOnce sync.Once
	_scheduleTeacherDA   IScheduleTeacherDA
)

func GetScheduleTeacherDA() IScheduleTeacherDA {
	_scheduleTeacherOnce.Do(func() {
		_scheduleTeacherDA = &scheduleTeacherDA{}
	})
	return _scheduleTeacherDA
}

type ScheduleTeacherCondition struct {
	OrderBy ScheduleTeacherOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c ScheduleTeacherCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.DeleteAt.Valid {
		wheres = append(wheres, "delete_at>0")
	} else {
		wheres = append(wheres, "(delete_at=0 || delete_at is null)")
	}

	return wheres, params
}

func (c ScheduleTeacherCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c ScheduleTeacherCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type ScheduleTeacherOrderBy int

const (
	ScheduleTeacherOrderByCreateAtDesc = iota + 1
	ScheduleTeacherOrderByCreateAtAsc
)

func NewScheduleTeacherOrderBy(orderby string) ScheduleTeacherOrderBy {
	//switch orderby {
	//case "create_at":
	//	return ScheduleTeacherOrderByCreateAtAsc
	//case "-create_at":
	//	return ScheduleTeacherOrderByCreateAtDesc
	//default:
	//	return ScheduleTeacherOrderByCreateAtDesc
	//}
	return ScheduleTeacherOrderByCreateAtDesc
}

func (c ScheduleTeacherOrderBy) ToSQL() string {
	return ""
	//switch c {
	//case ScheduleTeacherOrderByCreateAtAsc:
	//	return "create_at"
	//case ScheduleTeacherOrderByCreateAtDesc:
	//	return "create_at desc"
	//default:
	//	return "create_at desc"
	//}
}
