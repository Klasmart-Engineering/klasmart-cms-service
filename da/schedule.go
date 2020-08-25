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
	"time"
)

type IScheduleDA interface {
	dbo.DataAccesser
	BatchInsert(context.Context, *dbo.DBContext, []*entity.Schedule) (int, error)
	SoftDelete(ctx context.Context, tx *dbo.DBContext, id string, operator *entity.Operator) error
	DeleteWithFollowing(ctx context.Context, tx *dbo.DBContext, repeatID string, startAt int64) error
}

type scheduleDA struct {
	dbo.BaseDA
}

func (s *scheduleDA) BatchInsert(ctx context.Context, dbContext *dbo.DBContext, schedules []*entity.Schedule) (int, error) {
	var data [][]interface{}
	for _, item := range schedules {
		data = append(data, []interface{}{
			item.ID,
			item.Title,
		})
	}
	sql := SQLBatchInsert(constant.TableNameSchedule, []string{"id", "title"}, data)
	execResult := dbContext.Exec(sql.Format, sql.Values...)
	if execResult.Error != nil {
		logger.Error(ctx, "db exec sql error", log.Any("sql", sql), log.Err(execResult.Error))
		return 0, execResult.Error
	}
	total := int(execResult.RowsAffected)
	return total, nil
}

func (s *scheduleDA) DeleteWithFollowing(ctx context.Context, tx *dbo.DBContext, repeatID string, startAt int64) error {
	if err := tx.Unscoped().
		Where("repeat_id = ?", repeatID).
		Where("start_at >= ?", startAt).
		Delete(&entity.Schedule{}).Error; err != nil {
		log.Error(ctx, "delete schedules with following: delete failed",
			log.String("repeat_id", repeatID),
			log.Int64("start_at", startAt),
		)
		return err
	}
	return nil
}

func (s *scheduleDA) SoftDelete(ctx context.Context, tx *dbo.DBContext, id string, operator *entity.Operator) error {
	if err := tx.Model(&entity.Schedule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_id": operator.UserID,
			"deleted_at": time.Now().Unix(),
		}).Error; err != nil {
		log.Error(ctx, "soft delete schedule: update failed")
		return err
	}
	return nil
}

var (
	_scheduleOnce sync.Once
	_scheduleDA   IScheduleDA
)

func GetScheduleDA() IScheduleDA {
	_scheduleOnce.Do(func() {
		_scheduleDA = &scheduleDA{}
	})
	return _scheduleDA
}

type ScheduleCondition struct {
	OrgID            sql.NullString
	StartAtLe        sql.NullInt64
	EndAtGe          sql.NullInt64
	TeacherID        sql.NullString
	TeacherIDs       entity.NullStrings
	StartAndEndRange []sql.NullInt64
	//ScheduleIDs entity.NullStrings

	OrderBy ScheduleOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c ScheduleCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.OrgID.Valid {
		wheres = append(wheres, "org_id = ?")
		params = append(params, c.OrgID.String)
	}

	if c.StartAtLe.Valid {
		wheres = append(wheres, "start_at <= ?")
		params = append(params, c.StartAtLe.Int64)
	}
	if len(c.StartAndEndRange) == 2 {
		startRange := c.StartAndEndRange[0]
		if startRange.Valid {
			wheres = append(wheres, "(start_at <= ? and end_at >= ?)")
			params = append(params, startRange.Int64, startRange.Int64)
		}
		endRange := c.StartAndEndRange[1]
		if endRange.Valid {
			wheres = append(wheres, "or (start_at <= ? and end_at >= ?)")
			params = append(params, endRange.Int64, endRange.Int64)
		}

	}
	if c.EndAtGe.Valid {
		wheres = append(wheres, "end_at >= ?")
		params = append(params, c.EndAtGe.Int64)
	}
	if c.TeacherID.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where teacher_id = ? and (deleted_at=0) and %s.id = %s.schedule_id)", constant.TableNameScheduleTeacher, constant.TableNameScheduleTeacher, constant.TableNameSchedule)
		wheres = append(wheres, sql)
		params = append(params, c.TeacherID.String)
	}
	if c.TeacherIDs.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where teacher_id in (%s) and (deleted_at=0) and %s.id = %s.schedule_id)", constant.TableNameScheduleTeacher, c.TeacherIDs.SQLPlaceHolder(), constant.TableNameScheduleTeacher, constant.TableNameSchedule)
		wheres = append(wheres, sql)
		params = append(params, c.TeacherID.String)
	}

	if c.DeleteAt.Valid {
		wheres = append(wheres, "deleted_at>0")
	} else {
		wheres = append(wheres, "(deleted_at=0)")
	}

	return wheres, params
}

func (c ScheduleCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c ScheduleCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type ScheduleOrderBy int

const (
	ScheduleOrderByCreateAtDesc = iota + 1
	ScheduleOrderByCreateAtAsc
	ScheduleOrderByStartAtAsc
	ScheduleOrderByStartAtDesc
)

func NewScheduleOrderBy(orderby string) ScheduleOrderBy {
	switch orderby {
	case "create_at":
		return ScheduleOrderByCreateAtAsc
	case "-create_at":
		return ScheduleOrderByCreateAtDesc
	case "start_at":
		return ScheduleOrderByStartAtAsc
	case "-start_at":
		return ScheduleOrderByStartAtDesc
	default:
		return ScheduleOrderByStartAtAsc
	}
}

func (c ScheduleOrderBy) ToSQL() string {
	switch c {
	case ScheduleOrderByCreateAtAsc:
		return "create_at"
	case ScheduleOrderByCreateAtDesc:
		return "create_at desc"
	case ScheduleOrderByStartAtAsc:
		return "start_at"
	case ScheduleOrderByStartAtDesc:
		return "start_at desc"
	default:
		return "start_at"
	}
}
