package da

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IScheduleDA interface {
	dbo.DataAccesser
	BatchInsert(context.Context, *dbo.DBContext, []*entity.Schedule) (int64, error)
	MultipleBatchInsert(ctx context.Context, tx *dbo.DBContext, schedules []*entity.Schedule) (int64, error)
	SoftDelete(ctx context.Context, tx *dbo.DBContext, id string, operator *entity.Operator) error
	DeleteWithFollowing(ctx context.Context, tx *dbo.DBContext, repeatID string, startAt int64) error
	GetLessonPlanIDsByCondition(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]string, error)
	UpdateProgram(ctx context.Context, tx *sql.Tx, orgID string, oldProgramID string, newProgramID string) error
	UpdateSubject(ctx context.Context, tx *sql.Tx, orgID string, oldSubjectID string, oldProgramID string, newSubjectID string) error
	GetPrograms(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]string, error)
	GetSubjects(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]string, error)
	GetClassTypes(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]string, error)
}

type scheduleDA struct {
	dbo.BaseDA
}

func (s *scheduleDA) GetPrograms(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]string, error) {
	wheres, parameters := condition.GetConditions()
	whereSql := strings.Join(wheres, " and ")
	var scheduleList []*entity.Schedule
	err := tx.Table(constant.TableNameSchedule).Select("distinct program_id").Where(whereSql, parameters...).Find(&scheduleList).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "get programs ids from db error",
			log.Err(err),
			log.Any("condition", condition),
		)
		return nil, err
	}
	var result = make([]string, 0, len(scheduleList))
	for _, item := range scheduleList {
		if item.ProgramID == "" {
			continue
		}
		result = append(result, item.ProgramID)
	}
	log.Debug(ctx, "ProgramIDs", log.Strings("ProgramIDs", result))
	return result, nil
}
func (s *scheduleDA) GetSubjects(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]string, error) {
	wheres, parameters := condition.GetConditions()
	whereSql := strings.Join(wheres, " and ")
	var scheduleList []*entity.Schedule
	err := tx.Table(constant.TableNameSchedule).Select("distinct subject_id").Where(whereSql, parameters...).Find(&scheduleList).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "get programs ids from db error",
			log.Err(err),
			log.Any("condition", condition),
		)
		return nil, err
	}
	var result = make([]string, 0, len(scheduleList))
	for _, item := range scheduleList {
		if item.SubjectID == "" {
			continue
		}
		result = append(result, item.SubjectID)
	}
	log.Debug(ctx, "SubjectIDs", log.Strings("SubjectIDs", result))
	return result, nil
}
func (s *scheduleDA) GetClassTypes(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]string, error) {
	wheres, parameters := condition.GetConditions()
	whereSql := strings.Join(wheres, " and ")
	var scheduleList []*entity.Schedule
	err := tx.Table(constant.TableNameSchedule).Select("distinct class_type").Where(whereSql, parameters...).Find(&scheduleList).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "get class_type from db error",
			log.Err(err),
			log.Any("condition", condition),
		)
		return nil, err
	}
	var result = make([]string, 0, len(scheduleList))
	for _, item := range scheduleList {
		if item.ClassType == "" {
			continue
		}
		result = append(result, item.ClassType.String())
	}
	log.Debug(ctx, "class_type", log.Strings("class_type", result))
	return result, nil
}

func (s *scheduleDA) MultipleBatchInsert(ctx context.Context, tx *dbo.DBContext, schedules []*entity.Schedule) (int64, error) {
	total := len(schedules)
	pageSize := constant.ScheduleBatchInsertCount
	pageCount := (total + pageSize - 1) / pageSize
	var rowsAffected int64
	for i := 0; i < pageCount; i++ {
		start := i * pageSize
		end := (i + 1) * pageSize
		if end >= total {
			end = total
		}
		data := schedules[start:end]
		row, err := s.BatchInsert(ctx, tx, data)
		if err != nil {
			return rowsAffected, err
		}
		rowsAffected += row
	}
	return rowsAffected, nil
}

func (s *scheduleDA) BatchInsert(ctx context.Context, dbContext *dbo.DBContext, schedules []*entity.Schedule) (int64, error) {
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
			item.IsHomeFun,
			item.IsHidden,
		})
	}
	format, values := SQLBatchInsert(constant.TableNameSchedule, []string{
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
		"`is_home_fun`",
		"`is_hidden`",
	}, data)
	execResult := dbContext.Exec(format, values...)
	if execResult.Error != nil {
		logger.Error(ctx, "db exec sql error", log.Any("format", format), log.Any("values", values), log.Err(execResult.Error))
		return 0, execResult.Error
	}
	return execResult.RowsAffected, nil
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
		Where("status = ?", entity.ScheduleStatusNotStart).
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

func (s *scheduleDA) UpdateProgram(ctx context.Context, tx *sql.Tx, orgID string, oldProgramID string, newProgramID string) error {
	_, err := tx.
		Exec(fmt.Sprintf("UPDATE `%s` SET `program_id` = ?, `updated_at` = ?  WHERE (program_id = ? and org_id = ?)", constant.TableNameSchedule), newProgramID, time.Now().Unix(), oldProgramID, orgID)

	if err != nil {
		log.Error(ctx, "update schedule program: update failed",
			log.Err(err),
			log.String("oldProgramID", oldProgramID),
			log.String("newProgramID", newProgramID),
			log.Any("orgID", orgID),
		)
		return err
	}
	return nil
}

func (s *scheduleDA) UpdateSubject(ctx context.Context, tx *sql.Tx, orgID string, oldSubjectID string, oldProgramID string, newSubjectID string) error {
	_, err := tx.
		Exec(fmt.Sprintf("UPDATE `%s` SET `subject_id` = ?, `updated_at` = ?  WHERE (subject_id = ? and program_id = ? and org_id = ?)", constant.TableNameSchedule), newSubjectID, time.Now().Unix(), oldSubjectID, oldProgramID, orgID)

	if err != nil {
		log.Error(ctx, "update schedule subject: update failed",
			log.Err(err),
			log.String("oldSubjectID", oldSubjectID),
			log.String("newSubjectID", newSubjectID),
			log.Any("orgID", orgID),
		)
		return err
	}
	return nil
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

func NewNotStartScheduleCondition(classRosterID string, userIDs []string) *ScheduleCondition {
	condition := &ScheduleCondition{
		RosterClassID: sql.NullString{
			String: classRosterID,
			Valid:  classRosterID != "",
		},
		NotStart: sql.NullBool{
			Bool:  true,
			Valid: true,
		},
		RelationIDs: entity.NullStrings{
			Strings: userIDs,
			Valid:   len(userIDs) > 0,
		},
	}
	return condition
}

type ScheduleCondition struct {
	IDs                      entity.NullStrings
	OrgID                    sql.NullString
	StartAtGe                sql.NullInt64
	StartAtLt                sql.NullInt64
	EndAtLe                  sql.NullInt64
	EndAtLt                  sql.NullInt64
	EndAtGe                  sql.NullInt64
	StartAndEndRange         []sql.NullInt64
	StartAndEndTimeViewRange []sql.NullInt64
	LessonPlanID             sql.NullString
	LessonPlanIDs            entity.NullStrings
	RepeatID                 sql.NullString
	Status                   sql.NullString
	SubjectIDs               entity.NullStrings
	ProgramIDs               entity.NullStrings
	RelationID               sql.NullString
	RelationIDs              entity.NullStrings
	RelationSchoolIDs        entity.NullStrings
	ClassTypes               entity.NullStrings
	DueToEq                  sql.NullInt64
	AnyTime                  sql.NullBool
	RosterClassID            sql.NullString
	NotStart                 sql.NullBool

	OrderBy ScheduleOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c ScheduleCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.IDs.Valid {
		wheres = append(wheres, "id in (?)")
		params = append(params, c.IDs.Strings)
	}

	//if c.OrgID.Valid {
	//	wheres = append(wheres, "org_id = ?")
	//	params = append(params, c.OrgID.String)
	//}

	if c.StartAtGe.Valid {
		wheres = append(wheres, "start_at >= ?")
		params = append(params, c.StartAtGe.Int64)
	}

	if len(c.StartAndEndRange) == 2 {
		startRange := c.StartAndEndRange[0]
		endRange := c.StartAndEndRange[1]
		wheres = append(wheres, "((start_at <= ? and end_at > ?) or (start_at <= ? and end_at > ?))")
		params = append(params, startRange.Int64, startRange.Int64, endRange.Int64, endRange.Int64)
	}
	if len(c.StartAndEndTimeViewRange) == 2 {
		startRange := c.StartAndEndTimeViewRange[0]
		endRange := c.StartAndEndTimeViewRange[1]
		wheres = append(wheres, "((start_at >= ? and start_at <= ?) or (end_at >= ? and end_at <= ?) or (due_at>=? and due_at<=?))")
		params = append(params, startRange.Int64, endRange.Int64, startRange.Int64, endRange.Int64, startRange.Int64, endRange.Int64)
	}
	if c.EndAtLe.Valid {
		wheres = append(wheres, "end_at <= ?")
		params = append(params, c.EndAtLe.Int64)
	}
	if c.EndAtLt.Valid {
		wheres = append(wheres, "end_at < ?")
		params = append(params, c.EndAtLt.Int64)
	}
	if c.EndAtGe.Valid {
		wheres = append(wheres, "end_at >= ?")
		params = append(params, c.EndAtGe.Int64)
	}
	if c.StartAtLt.Valid {
		wheres = append(wheres, "start_at < ?")
		params = append(params, c.StartAtLt.Int64)
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

	if c.SubjectIDs.Valid {
		wheres = append(wheres, fmt.Sprintf("subject_id in (%s)", c.SubjectIDs.SQLPlaceHolder()))
		params = append(params, c.SubjectIDs.ToInterfaceSlice()...)
	}
	if c.ProgramIDs.Valid {
		wheres = append(wheres, fmt.Sprintf("program_id in (%s)", c.ProgramIDs.SQLPlaceHolder()))
		params = append(params, c.ProgramIDs.ToInterfaceSlice()...)
	}

	if c.RelationID.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where relation_id = ? and %s.id = %s.schedule_id)",
			constant.TableNameScheduleRelation, constant.TableNameSchedule, constant.TableNameScheduleRelation)
		wheres = append(wheres, sql)
		params = append(params, c.RelationID.String)
	}
	if c.RelationIDs.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where relation_id in (?) and %s.id = %s.schedule_id)",
			constant.TableNameScheduleRelation, constant.TableNameSchedule, constant.TableNameScheduleRelation)
		wheres = append(wheres, sql)
		params = append(params, c.RelationIDs.Strings)
	}
	if c.RelationSchoolIDs.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where relation_id in (?) and %s.id = %s.schedule_id)",
			constant.TableNameScheduleRelation, constant.TableNameSchedule, constant.TableNameScheduleRelation)
		wheres = append(wheres, sql)
		params = append(params, c.RelationSchoolIDs.Strings)
	}
	if c.OrgID.Valid {
		wheres = append(wheres, "org_id = ?")
		params = append(params, c.OrgID.String)
	}
	if c.ClassTypes.Valid {
		wheres = append(wheres, "class_type in (?)")
		params = append(params, c.ClassTypes.Strings)
	}
	if c.DueToEq.Valid {
		wheres = append(wheres, "due_at = ?")
		params = append(params, c.DueToEq.Int64)
	}
	if c.AnyTime.Valid {
		wheres = append(wheres, "due_at=0 and start_at=0 and end_at=0")
	}

	if c.RosterClassID.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where relation_id = ? and relation_type = ? and %s.id = %s.schedule_id)",
			constant.TableNameScheduleRelation, constant.TableNameSchedule, constant.TableNameScheduleRelation)
		wheres = append(wheres, sql)
		params = append(params, c.RosterClassID.String, entity.ScheduleRelationTypeClassRosterClass)
	}
	if c.NotStart.Valid {
		notEditAt := time.Now().Add(constant.ScheduleAllowEditTime).Unix()
		wheres = append(wheres, " ((due_at=0 and start_at=0 and end_at=0) || (start_at = 0 and due_at > ?) || (start_at > ?)) ")
		params = append(params, time.Now().Unix(), notEditAt)
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
		return "created_at"
	case ScheduleOrderByCreateAtDesc:
		return "created_at desc"
	case ScheduleOrderByStartAtAsc:
		return "start_at"
	case ScheduleOrderByStartAtDesc:
		return "start_at desc"
	default:
		return "start_at"
	}
}
