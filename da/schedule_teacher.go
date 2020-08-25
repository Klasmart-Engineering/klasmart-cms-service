package da

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IScheduleTeacherDA interface {
	dbo.DataAccesser
	BatchInsert(context.Context, *dbo.DBContext, []*entity.TeacherSchedule) (int, error)
	DeleteByScheduleID(ctx context.Context, tx *dbo.DBContext, scheduleID string) error
}

type scheduleTeacherDA struct {
	dbo.BaseDA
}

func (s *scheduleTeacherDA) BatchInsert(ctx context.Context, dbContext *dbo.DBContext, schedules []*entity.TeacherSchedule) (int, error) {
	var data [][]interface{}
	for _, item := range schedules {
		data = append(data, []interface{}{
			item.ID,
			item.ScheduleID,
			item.TeacherID,
			item.DeletedAt,
		})
	}
	sql := utils.SQLBatchInsert(constant.TableNameTeacherSchedule, []string{"id", "schedule_id", "teacher_id", "deleted_at"}, data)
	execResult := dbContext.Exec(sql.Format, sql.Values...)
	if execResult.Error != nil {
		logger.Error(ctx, "db exec sql error", log.Any("sql", sql))
		return 0, execResult.Error
	}
	total := int(execResult.RowsAffected)
	return total, nil
}

func (s *scheduleTeacherDA) DeleteByScheduleID(ctx context.Context, tx *dbo.DBContext, scheduleID string) error {
	if err := tx.Unscoped().
		Where("schedule_id = ?", scheduleID).
		Delete(&entity.TeacherSchedule{}).Error; err != nil {
		log.Error(ctx, "delete teacher_schedule by schedule id: delete failed", log.Err(err))
		return err
	}
	return nil
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
	TeacherID  sql.NullString
	ScheduleID sql.NullString
	OrderBy    ScheduleTeacherOrderBy
	Pager      dbo.Pager

	DeleteAt sql.NullInt64
}

func (c ScheduleTeacherCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.ScheduleID.Valid {
		wheres = append(wheres, "schedule_id = ?")
		params = append(params, c.ScheduleID.String)
	}

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
