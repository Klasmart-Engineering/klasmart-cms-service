package da

import (
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

type IScheduleFeedbackDA interface {
	dbo.DataAccesser
}

type scheduleFeedbackDA struct {
	dbo.BaseDA
}

var (
	_scheduleFeedbackOnce sync.Once
	_scheduleFeedbackDA   IScheduleFeedbackDA
)

func GetScheduleFeedbackDA() IScheduleFeedbackDA {
	_scheduleFeedbackOnce.Do(func() {
		_scheduleFeedbackDA = &scheduleFeedbackDA{}
	})
	return _scheduleFeedbackDA
}

type ScheduleFeedbackCondition struct {
	ScheduleID sql.NullString
	UserID     sql.NullString

	OrderBy ScheduleFeedbackOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c ScheduleFeedbackCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.ScheduleID.Valid {
		wheres = append(wheres, "schedule_id = ?")
		params = append(params, c.ScheduleID.String)
	}

	if c.DeleteAt.Valid {
		wheres = append(wheres, "delete_at>0")
	} else {
		wheres = append(wheres, "(delete_at=0)")
	}

	return wheres, params
}

func (c ScheduleFeedbackCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c ScheduleFeedbackCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type ScheduleFeedbackOrderBy int

const (
	ScheduleFeedbackOrderByNameAsc = iota + 1
)

func NewScheduleFeedbackOrderBy(orderBy string) ScheduleFeedbackOrderBy {
	return ScheduleFeedbackOrderByNameAsc
}

func (c ScheduleFeedbackOrderBy) ToSQL() string {
	switch c {
	default:
		return "create_at desc"
	}
}
