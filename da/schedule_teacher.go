package da

import (
	"context"
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IScheduleTeacherDA interface {
	dbo.DataAccesser
	BatchInsert(context.Context, *dbo.DBContext, []*entity.ScheduleTeacher) (int, error)
	DeleteByScheduleID(ctx context.Context, tx *dbo.DBContext, scheduleID string) error
	BatchDelByScheduleIDs(ctx context.Context, tx *dbo.DBContext, scheduleID []string) error
}

type scheduleTeacherDA struct {
	dbo.BaseDA
}

func (s *scheduleTeacherDA) BatchInsert(ctx context.Context, dbContext *dbo.DBContext, schedules []*entity.ScheduleTeacher) (int, error) {
	var data [][]interface{}
	for _, item := range schedules {
		data = append(data, []interface{}{
			item.ID,
			item.ScheduleID,
			item.TeacherID,
			item.DeleteAt,
		})
	}
	sql := SQLBatchInsert(constant.TableNameScheduleTeacher, []string{"id", "schedule_id", "teacher_id", "delete_at"}, data)
	execResult := dbContext.Exec(sql.Format, sql.Values...)
	if execResult.Error != nil {
		logger.Error(ctx, "db exec sql error", log.Any("sql", sql), log.Err(execResult.Error))
		return 0, execResult.Error
	}
	total := int(execResult.RowsAffected)
	return total, nil
}

func (s *scheduleTeacherDA) DeleteByScheduleID(ctx context.Context, tx *dbo.DBContext, scheduleID string) error {
	if err := tx.Unscoped().
		Where("schedule_id = ?", scheduleID).
		Delete(&entity.ScheduleTeacher{}).Error; err != nil {
		log.Error(ctx, "delete teacher_schedule by schedule id: delete failed",
			log.Err(err),
			log.String("schedule_id", scheduleID),
		)
		return err
	}
	return nil
}

func (s *scheduleTeacherDA) BatchDelByScheduleIDs(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) error {
	if err := tx.Unscoped().
		Where("schedule_id in (?)", scheduleIDs).
		Delete(&entity.ScheduleTeacher{}).Error; err != nil {
		log.Error(ctx, "delete teacher_schedule by schedule id: delete failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
		)
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
	TeacherID   sql.NullString
	ScheduleID  sql.NullString
	ScheduleIDs entity.NullStrings
	//OrderBy    ScheduleTeacherOrderBy
	Pager dbo.Pager

	DeleteAt sql.NullInt64
}

func (c ScheduleTeacherCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.ScheduleID.Valid {
		wheres = append(wheres, "schedule_id = ?")
		params = append(params, c.ScheduleID.String)
	}
	if c.ScheduleIDs.Valid {
		sql := fmt.Sprintf("schedule_id in (%s)", c.ScheduleIDs.SQLPlaceHolder())
		wheres = append(wheres, sql)
		params = append(params, c.ScheduleIDs.ToInterfaceSlice()...)
	}

	if c.DeleteAt.Valid {
		wheres = append(wheres, "delete_at>0")
	} else {
		wheres = append(wheres, "delete_at=0")
	}

	return wheres, params
}

func (c ScheduleTeacherCondition) GetOrderBy() string {
	return ""
}

func (c ScheduleTeacherCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

//type ScheduleTeacherOrderBy int
//
//const (
//	ScheduleTeacherOrderByCreateAtDesc = iota + 1
//	ScheduleTeacherOrderByCreateAtAsc
//)
//
//func NewScheduleTeacherOrderBy(orderby string) ScheduleTeacherOrderBy {
//	//switch orderby {
//	//case "create_at":
//	//	return ScheduleTeacherOrderByCreateAtAsc
//	//case "-create_at":
//	//	return ScheduleTeacherOrderByCreateAtDesc
//	//default:
//	//	return ScheduleTeacherOrderByCreateAtDesc
//	//}
//	return ScheduleTeacherOrderByCreateAtDesc
//}
//
//func (c ScheduleTeacherOrderBy) ToSQL() string {
//	return ""
//	//switch c {
//	//case ScheduleTeacherOrderByCreateAtAsc:
//	//	return "create_at"
//	//case ScheduleTeacherOrderByCreateAtDesc:
//	//	return "create_at desc"
//	//default:
//	//	return "create_at desc"
//	//}
//}
