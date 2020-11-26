package da

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
	"sync"
	"time"
)

type IScheduleDA interface {
	dbo.DataAccesser
	BatchInsert(context.Context, *dbo.DBContext, []*entity.Schedule) (int, error)
	SoftDelete(ctx context.Context, tx *dbo.DBContext, id string, operator *entity.Operator) error
	DeleteWithFollowing(ctx context.Context, tx *dbo.DBContext, repeatID string, startAt int64) error
	GetParticipateClass(ctx context.Context, tx *dbo.DBContext, teacherID string) ([]string, error)
	GetLessonPlanIDsByCondition(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]string, error)
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
			item.ClassID,
			item.LessonPlanID,
			item.OrgID,
			item.StartAt,
			item.EndAt,
			item.Status,
			item.IsAllDay,
			item.SubjectID,
			item.ProgramID,
			item.ClassType,
			item.DueAt,
			item.Description,
			item.Attachment,
			item.ScheduleVersion,
			item.RepeatID,
			item.RepeatJson,
			item.CreatedID,
			item.UpdatedID,
			item.DeletedID,
			item.CreatedAt,
			item.UpdatedAt,
			item.DeleteAt,
		})
	}
	sql := SQLBatchInsert(constant.TableNameSchedule, []string{
		"`id`",
		"`title`",
		"`class_id`",
		"`lesson_plan_id`",
		"`org_id`",
		"`start_at`",
		"`end_at`",
		"`status`",
		"`is_all_day`",
		"`subject_id`",
		"`program_id`",
		"`class_type`",
		"`due_at`",
		"`description`",
		"`attachment`",
		"`version`",
		"`repeat_id`",
		"`repeat`",
		"`created_id`",
		"`updated_id`",
		"`deleted_id`",
		"`created_at`",
		"`updated_at`",
		"`delete_at`",
	}, data)
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
		Where("status = ?", entity.ScheduleStatusNotStart).
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
			"delete_at":  time.Now().Unix(),
		}).Error; err != nil {
		log.Error(ctx, "soft delete schedule: update failed",
			log.Err(err),
			log.String("id", id),
			log.Any("operator", operator),
		)
		return err
	}
	return nil
}

func (s *scheduleDA) GetParticipateClass(ctx context.Context, tx *dbo.DBContext, teacherID string) ([]string, error) {
	sql := fmt.Sprintf("exists(select 1 from %s where teacher_id = ? and (delete_at=0) and %s.id = %s.schedule_id)",
		constant.TableNameScheduleTeacher, constant.TableNameSchedule, constant.TableNameScheduleTeacher)
	var scheduleList []*entity.Schedule
	err := tx.Table(constant.TableNameSchedule).Select("distinct class_id").Where(sql, teacherID).Find(&scheduleList).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetParticipateClass:get participate  class from db error", log.Err(err), log.String("teacherID", teacherID))
		return nil, err
	}
	var result = make([]string, len(scheduleList))
	for i, item := range scheduleList {
		result[i] = item.ClassID
	}
	log.Debug(ctx, "classIDs", log.Strings("classIDs", result))
	return result, nil
}

func (s *scheduleDA) GetLessonPlanIDsByCondition(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]string, error) {
	wheres, parameters := condition.GetConditions()
	whereSql := strings.Join(wheres, " and ")
	var scheduleList []*entity.Schedule
	err := tx.Table(constant.TableNameSchedule).Select("distinct lesson_plan_id").Where(whereSql, parameters...).Find(&scheduleList).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetLessonPlanIDsByCondition:get lessonPlan ids from db error",
			log.Err(err),
			log.Any("condition", condition),
		)
		return nil, err
	}
	var result = make([]string, len(scheduleList))
	for i, item := range scheduleList {
		result[i] = item.LessonPlanID
	}
	log.Debug(ctx, "lessonPlanIDs", log.Strings("lessonPlanIDs", result))
	return result, nil
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
	ID        sql.NullString
	OrgID     sql.NullString
	StartAtGe sql.NullInt64
	EndAtLe   sql.NullInt64
	//TeacherID                sql.NullString
	//TeacherIDs               entity.NullStrings
	StartAndEndRange         []sql.NullInt64
	StartAndEndTimeViewRange []sql.NullInt64
	LessonPlanID             sql.NullString
	LessonPlanIDs            entity.NullStrings
	RepeatID                 sql.NullString
	Status                   sql.NullString
	ClassIDs                 entity.NullStrings
	SubjectIDs               entity.NullStrings
	ProgramIDs               entity.NullStrings
	OrgIDs                   entity.NullStrings
	ClassID                  sql.NullString

	OrderBy ScheduleOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c ScheduleCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.ID.Valid {
		wheres = append(wheres, "id = ?")
		params = append(params, c.ID.String)
	}

	if c.OrgID.Valid {
		wheres = append(wheres, "org_id = ?")
		params = append(params, c.OrgID.String)
	}

	if c.StartAtGe.Valid {
		wheres = append(wheres, "start_at >= ?")
		params = append(params, c.StartAtGe.Int64)
	}

	if len(c.StartAndEndRange) == 2 {
		startRange := c.StartAndEndRange[0]
		endRange := c.StartAndEndRange[1]
		wheres = append(wheres, "((start_at <= ? and end_at >= ?) or (start_at <= ? and end_at >= ?))")
		params = append(params, startRange.Int64, startRange.Int64, endRange.Int64, endRange.Int64)
	}
	if len(c.StartAndEndTimeViewRange) == 2 {
		startRange := c.StartAndEndTimeViewRange[0]
		endRange := c.StartAndEndTimeViewRange[1]
		wheres = append(wheres, "((start_at >= ? and start_at <= ?) or (end_at >= ? and end_at <= ?))")
		params = append(params, startRange.Int64, endRange.Int64, startRange.Int64, endRange.Int64)
	}
	if c.EndAtLe.Valid {
		wheres = append(wheres, "end_at <= ?")
		params = append(params, c.EndAtLe.Int64)
	}
	//if c.TeacherID.Valid {
	//	sql := fmt.Sprintf("exists(select 1 from %s where teacher_id = ? and (delete_at=0) and %s.id = %s.schedule_id)",
	//		constant.TableNameScheduleTeacher, constant.TableNameSchedule, constant.TableNameScheduleTeacher)
	//	wheres = append(wheres, sql)
	//	params = append(params, c.TeacherID.String)
	//}
	//if c.TeacherIDs.Valid {
	//	sql := fmt.Sprintf("exists(select 1 from %s where teacher_id in (%s) and (delete_at=0) and %s.id = %s.schedule_id)",
	//		constant.TableNameScheduleTeacher, c.TeacherIDs.SQLPlaceHolder(), constant.TableNameSchedule, constant.TableNameScheduleTeacher)
	//	wheres = append(wheres, sql)
	//	params = append(params, c.TeacherIDs.ToInterfaceSlice()...)
	//}
	if c.LessonPlanID.Valid {
		wheres = append(wheres, "lesson_plan_id = ?")
		params = append(params, c.LessonPlanID.String)
	}
	if c.LessonPlanIDs.Valid {
		wheres = append(wheres, "lesson_plan_id in (?)")
		params = append(params, c.LessonPlanIDs.Strings)
	}
	if c.RepeatID.Valid {
		wheres = append(wheres, "repeat_id = ?")
		params = append(params, c.RepeatID.String)
	}
	if c.Status.Valid {
		wheres = append(wheres, "status = ?")
		params = append(params, c.Status.String)
	}

	if c.OrgIDs.Valid {
		wheres = append(wheres, fmt.Sprintf("org_id in (%s)", c.OrgIDs.SQLPlaceHolder()))
		params = append(params, c.OrgIDs.ToInterfaceSlice()...)
	}
	if c.ClassIDs.Valid {
		wheres = append(wheres, fmt.Sprintf("class_id in (%s)", c.ClassIDs.SQLPlaceHolder()))
		params = append(params, c.ClassIDs.ToInterfaceSlice()...)
	}
	if c.SubjectIDs.Valid {
		wheres = append(wheres, fmt.Sprintf("subject_id in (%s)", c.SubjectIDs.SQLPlaceHolder()))
		params = append(params, c.SubjectIDs.ToInterfaceSlice()...)
	}
	if c.ProgramIDs.Valid {
		wheres = append(wheres, fmt.Sprintf("program_id in (%s)", c.ProgramIDs.SQLPlaceHolder()))
		params = append(params, c.ProgramIDs.ToInterfaceSlice()...)
	}
	if c.ClassID.Valid {
		wheres = append(wheres, "class_id = ?")
		params = append(params, c.ClassID.String)
	}

	if c.DeleteAt.Valid {
		wheres = append(wheres, "delete_at>0")
	} else {
		wheres = append(wheres, "(delete_at=0)")
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
