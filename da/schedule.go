package da

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IScheduleDA interface {
	dbo.DataAccesser
	SoftDelete(ctx context.Context, tx *dbo.DBContext, id string, operator *entity.Operator) error
	DeleteWithFollowing(ctx context.Context, tx *dbo.DBContext, repeatID string, startAt int64) error
	GetLessonPlanIDsByCondition(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]string, error)
	UpdateProgram(ctx context.Context, tx *sql.Tx, orgID string, oldProgramID string, newProgramID string) error
	UpdateSubject(ctx context.Context, tx *sql.Tx, orgID string, oldSubjectID string, oldProgramID string, newSubjectID string) error
	GetProgramIDs(ctx context.Context, tx *dbo.DBContext, orgID string, relationIDs []string) ([]string, error)
	GetClassTypes(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]string, error)
	GetTeachLoadByCondition(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]*ScheduleTeachLoadDBResult, error)
	UpdateLiveLessonPlan(ctx context.Context, tx *dbo.DBContext, scheduleID string, liveMaterials *entity.ScheduleLiveLessonPlan) error
	UpdateScheduleReviewStatus(ctx context.Context, tx *dbo.DBContext, scheduleID string, reviewStatus entity.ScheduleReviewStatus) error
}

type scheduleDA struct {
	dbo.BaseDA
}

func (s *scheduleDA) GetProgramIDs(ctx context.Context, tx *dbo.DBContext, orgID string, relationIDs []string) ([]string, error) {
	tx.ResetCondition()
	var programIDs []string
	db := tx.Table(constant.TableNameSchedule).
		Select("distinct program_id").
		Where("schedules.org_id = ? and delete_at = 0", orgID)

	if len(relationIDs) > 0 {
		db = db.Joins("left join schedules_relations ON schedules.id = schedules_relations.schedule_id").
			Where("schedules_relations.relation_id in (?)", relationIDs)
	}

	err := db.Scan(&programIDs).Error
	if err != nil {
		log.Error(ctx, "GetProgramIDs error",
			log.Err(err),
			log.String("orgID", orgID),
			log.Strings("relation_ids", relationIDs),
		)
		return nil, err
	}

	return programIDs, nil
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
		UpdateColumns(entity.Schedule{
			DeletedID: operator.UserID,
			DeleteAt:  time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
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

type ScheduleTeachLoadDBResult struct {
	TeacherID string
	ClassType entity.ScheduleClassType
	Durations []int64
}

//SELECT schedules_relations.relation_id,schedules.class_type,
//sum((case (start_at > 1605456000 and end_at<1605544740) when 1 then end_at-start_at else
//if ((start_at<1605456000 and end_at>1605456000),end_at-1605456000,
//if ((start_at<1605544740 and end_at>1605544740),1605544740-start_at, 0)
//) end))  as result0
//FROM `schedules` inner join  schedules_relations on (schedules.id = schedules_relations.schedule_id) WHERE (org_id = '72e47ef0-92bf-4429-a06f-2014e3d3df4b' and class_type in ('OfflineClass','OnlineClass') and exists(select 1 from schedules_relations where relation_id in ('5751555a-cc18-4662-9ae5-a5ad90569f79','49b3be6a-d139-4f82-9b77-0acc89525d3f') and relation_type in ('class_roster_class','participant_class') and schedules.id = schedules_relations.schedule_id) and schedules_relations.relation_id in ('4fde6e1b-8efe-58e9-a404-51fb98ebf9b8','42098862-28b1-5417-9800-3b89e557a2b9') and (delete_at=0))
//GROUP BY schedules_relations.relation_id,schedules.class_type
func (s *scheduleDA) GetTeachLoadByCondition(ctx context.Context, tx *dbo.DBContext, condition *ScheduleCondition) ([]*ScheduleTeachLoadDBResult, error) {
	if len(condition.TeachLoadTimeRanges) <= 0 || len(condition.TeachLoadTeacherIDs.Strings) <= 0 {
		log.Info(ctx, "params invalid", log.Any("condition", condition))
		return nil, constant.ErrInvalidArgs
	}
	result := make([]*ScheduleTeachLoadDBResult, 0, len(condition.TeachLoadTeacherIDs.Strings))
	wheres, parameters := condition.GetConditions()
	whereSql := strings.Join(wheres, " and ")

	sql := new(strings.Builder)
	sql.WriteString(fmt.Sprintf("%s.relation_id,%s.class_type,", constant.TableNameScheduleRelation, constant.TableNameSchedule))
	for i, item := range condition.TeachLoadTimeRanges {
		sql.WriteString(fmt.Sprintf(`
		sum((case (start_at >= %d and end_at<=%d) when 1 then end_at-start_at else 
		if ((start_at<%d and end_at>%d),end_at-%d, 
			if ((start_at<%d and end_at>%d),%d-start_at, 0)
		) end))  as %s%d 	
		`,
			item.StartAt,
			item.EndAt,
			item.StartAt,
			item.StartAt,
			item.StartAt,
			item.EndAt,
			item.EndAt,
			item.EndAt,
			"result",
			i,
		))
		if i < len(condition.TeachLoadTimeRanges)-1 {
			sql.WriteString(",")
		}
	}

	rows, err := tx.Table(constant.TableNameSchedule).
		Select(sql.String()).
		Joins(fmt.Sprintf("inner join  %s on (%s.id = %s.schedule_id)",
			constant.TableNameScheduleRelation,
			constant.TableNameSchedule,
			constant.TableNameScheduleRelation)).
		Where(whereSql, parameters...).
		Group(fmt.Sprintf("%s.relation_id,%s.class_type", constant.TableNameScheduleRelation, constant.TableNameSchedule)).
		Rows()

	if gorm.IsRecordNotFoundError(err) {
		log.Error(ctx, "get teach load not found from db",
			log.Err(err),
			log.Any("condition", condition))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetTeachLoadByCondition: from db error",
			log.Err(err),
			log.Any("condition", condition),
		)
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Error(ctx, "rows columns error",
			log.Err(err),
			log.Any("condition", condition),
		)
	}
	if len(cols) < 3 {
		log.Error(ctx, "rows columns is invalid",
			log.Strings("cols", cols),
			log.Any("condition", condition),
		)
		return nil, constant.ErrInternalServer
	}
	values := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range values {
		scans[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(scans...); err != nil {
			log.Error(ctx, "rows scan error", log.Any("condition", condition), log.Err(err))
			return nil, err
		}

		loadItem := new(ScheduleTeachLoadDBResult)
		loadItem.TeacherID = string(values[0])

		classType := entity.ScheduleClassType(values[1])
		if !classType.Valid() {
			log.Debug(ctx, "class type invalid", log.String("classType", classType.String()))
			continue
		}
		loadItem.ClassType = classType
		loadItem.Durations = make([]int64, 0, len(condition.TeachLoadTimeRanges))

		for i := 2; i < len(values); i++ {
			valStr := string(values[i])
			valInt, err := strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				log.Error(ctx, "string parse int error", log.Any("valStr", valStr), log.Err(err))
				return nil, err
			}
			loadItem.Durations = append(loadItem.Durations, valInt)
		}
		result = append(result, loadItem)
	}

	log.Debug(ctx, "result", log.Any("result", result))
	return result, nil
}

func (s *scheduleDA) UpdateLiveLessonPlan(ctx context.Context, tx *dbo.DBContext, scheduleID string, liveLessonPlan *entity.ScheduleLiveLessonPlan) error {
	tx.ResetCondition()

	err := tx.Table(constant.TableNameSchedule).
		Where("id = ?", scheduleID).
		Update("live_lesson_plan", liveLessonPlan).Error
	if err != nil {
		log.Error(ctx, "UpdateLiveMaterials error",
			log.Err(err),
			log.String("scheduleID", scheduleID),
			log.Any("liveLessonPlan", liveLessonPlan))
		return err
	}

	return nil
}

func (s *scheduleDA) UpdateScheduleReviewStatus(ctx context.Context, tx *dbo.DBContext, scheduleID string, reviewStatus entity.ScheduleReviewStatus) error {
	tx.ResetCondition()

	err := tx.Table(constant.TableNameSchedule).
		Where("id = ?", scheduleID).
		Update("review_status", reviewStatus).Error
	if err != nil {
		log.Error(ctx, "UpdateScheduleReviewStatus error",
			log.Err(err),
			log.String("scheduleID", scheduleID),
			log.Any("reviewStatus", reviewStatus))
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

func NewScheduleTeachLoadCondition(input *entity.ScheduleTeachingLoadInput) *ScheduleCondition {
	return &ScheduleCondition{
		OrgID: sql.NullString{
			String: input.OrgID,
			Valid:  true,
		},
		RelationSchoolIDs: entity.NullStrings{
			Strings: input.SchoolIDs,
			Valid:   len(input.SchoolIDs) > 0,
		},
		RelationClassIDs: entity.NullStrings{
			Strings: input.ClassIDs,
			Valid:   len(input.ClassIDs) > 0,
		},
		TeachLoadTeacherIDs: entity.NullStrings{
			Strings: input.TeacherIDs,
			Valid:   len(input.TeacherIDs) > 0,
		},
		ClassTypes: entity.NullStrings{
			Strings: []string{
				entity.ScheduleClassTypeOfflineClass.String(),
				entity.ScheduleClassTypeOnlineClass.String(),
			},
			Valid: true,
		},
		TeachLoadTimeRanges: input.TimeRanges,
	}
}

type ScheduleCondition struct {
	IDs                      entity.NullStrings
	OrgID                    sql.NullString
	StartAtGe                sql.NullInt64
	StartAtAndDueAtGe        sql.NullInt64
	StartAtOrEndAtOrDueAtGe  sql.NullInt64
	StartAtOrEndAtOrDueAtLe  sql.NullInt64
	StartAtLt                sql.NullInt64
	EndAtLe                  sql.NullInt64
	EndAtAndDueAtLe          sql.NullInt64
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
	IsHomefun                sql.NullBool
	ReviewStatus             entity.NullStrings
	SuccessReviewStudentID   sql.NullString
	DueToEq                  sql.NullInt64
	AnyTime                  sql.NullBool
	RosterClassID            sql.NullString
	NotStart                 sql.NullBool

	RelationClassIDs    entity.NullStrings
	RelationTeacherIDs  entity.NullStrings
	RelationStudentIDs  entity.NullStrings
	RelationUserIDs     entity.NullStrings
	TeachLoadTimeRanges []*entity.ScheduleTimeRange
	TeachLoadTeacherIDs entity.NullStrings

	OrderBy ScheduleOrderBy
	Pager   dbo.Pager

	DeleteAt   sql.NullInt64
	CreateAtGe sql.NullInt64
	CreateAtLt sql.NullInt64
}

func (c ScheduleCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.IDs.Valid {
		wheres = append(wheres, "id in (?)")
		params = append(params, c.IDs.Strings)
	}

	if c.StartAtGe.Valid {
		wheres = append(wheres, "start_at >= ?")
		params = append(params, c.StartAtGe.Int64)
	}

	if c.StartAtAndDueAtGe.Valid {
		wheres = append(wheres, "(start_at >= ? or due_at >= ?)")
		params = append(params, c.StartAtAndDueAtGe.Int64, c.StartAtAndDueAtGe.Int64)
	}

	if c.StartAtOrEndAtOrDueAtGe.Valid {
		wheres = append(wheres, "(start_at >= ? or end_at >= ? or due_at >= ?)")
		params = append(params, c.StartAtOrEndAtOrDueAtGe.Int64, c.StartAtOrEndAtOrDueAtGe.Int64, c.StartAtOrEndAtOrDueAtGe.Int64)
	}

	if c.StartAtOrEndAtOrDueAtLe.Valid {
		wheres = append(wheres, "(start_at <= ? or end_at <= ? or (due_at <= ? and due_at > 0))")
		params = append(params, c.StartAtOrEndAtOrDueAtLe.Int64, c.StartAtOrEndAtOrDueAtLe.Int64, c.StartAtOrEndAtOrDueAtLe.Int64)
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
		wheres = append(wheres, "((start_at >= ? and start_at <= ?) or (end_at >= ? and end_at <= ?) or (start_at <= ? and end_at >= ?) or (due_at>=? and due_at<=?))")
		params = append(params, startRange.Int64, endRange.Int64, startRange.Int64, endRange.Int64, startRange.Int64, endRange.Int64, startRange.Int64, endRange.Int64)
	}
	if c.EndAtLe.Valid {
		wheres = append(wheres, "end_at <= ?")
		params = append(params, c.EndAtLe.Int64)
	}
	if c.EndAtAndDueAtLe.Valid {
		wheres = append(wheres, "(end_at <= ? or (due_at <= ? and due_at > 0))")
		params = append(params, c.EndAtAndDueAtLe.Int64, c.EndAtAndDueAtLe.Int64)
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

	if c.CreateAtGe.Valid {
		wheres = append(wheres, "created_at >= ?")
		params = append(params, c.CreateAtGe.Int64)
	}
	if c.CreateAtLt.Valid {
		wheres = append(wheres, "created_at < ?")
		params = append(params, c.CreateAtLt.Int64)
	}

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
		sql := fmt.Sprintf("exists(select 1 from %s where relation_id in (?) and relation_type = ? and %s.id = %s.schedule_id)",
			constant.TableNameScheduleRelation, constant.TableNameSchedule, constant.TableNameScheduleRelation)
		wheres = append(wheres, sql)
		params = append(params, c.SubjectIDs.Strings, entity.ScheduleRelationTypeSubject)
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
		sql := fmt.Sprintf("exists(select 1 from %s where relation_id in (?) and %s.id = %s.schedule_id and %s.relation_type = ?)",
			constant.TableNameScheduleRelation, constant.TableNameSchedule, constant.TableNameScheduleRelation, constant.TableNameScheduleRelation)
		wheres = append(wheres, sql)
		params = append(params, c.RelationSchoolIDs.Strings, entity.ScheduleRelationTypeSchool)
	}
	if c.OrgID.Valid {
		wheres = append(wheres, "org_id = ?")
		params = append(params, c.OrgID.String)
	}
	if c.ClassTypes.Valid {
		wheres = append(wheres, "class_type in (?)")
		params = append(params, c.ClassTypes.Strings)
	}

	if c.IsHomefun.Valid {
		wheres = append(wheres, "((is_home_fun = ? and class_type = ?) or (class_type != ?))")
		params = append(params, c.IsHomefun.Bool, entity.ScheduleClassTypeHomework, entity.ScheduleClassTypeHomework)
	}

	if c.ReviewStatus.Valid {
		wheres = append(wheres, "review_status in (?)")
		params = append(params, c.ReviewStatus.Strings)
	}

	if c.SuccessReviewStudentID.Valid {
		sql := fmt.Sprintf("not exists(select 1 from %s where %s.is_review = 1 and %s.schedule_id = %s.id and %s.student_id = ? and %s.review_status = ?)",
			constant.TableNameScheduleReview, constant.TableNameSchedule, constant.TableNameScheduleReview, constant.TableNameSchedule, constant.TableNameScheduleReview, constant.TableNameScheduleReview)
		wheres = append(wheres, sql)
		params = append(params, c.SuccessReviewStudentID.String, entity.ScheduleReviewStatusSuccess)
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

	if c.RelationClassIDs.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where relation_id in (?) and relation_type in (?) and %s.id = %s.schedule_id)",
			constant.TableNameScheduleRelation, constant.TableNameSchedule, constant.TableNameScheduleRelation)
		wheres = append(wheres, sql)
		params = append(params, c.RelationClassIDs.Strings, []entity.ScheduleRelationType{
			entity.ScheduleRelationTypeClassRosterClass,
			entity.ScheduleRelationTypeParticipantClass,
		})
	}

	if c.RelationTeacherIDs.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where relation_id in (?) and relation_type in (?) and %s.id = %s.schedule_id)",
			constant.TableNameScheduleRelation, constant.TableNameSchedule, constant.TableNameScheduleRelation)
		wheres = append(wheres, sql)
		params = append(params, c.RelationTeacherIDs.Strings, []entity.ScheduleRelationType{
			entity.ScheduleRelationTypeClassRosterTeacher,
			entity.ScheduleRelationTypeParticipantTeacher,
		})
	}

	if c.RelationStudentIDs.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where relation_id in (?) and relation_type in (?) and %s.id = %s.schedule_id)",
			constant.TableNameScheduleRelation, constant.TableNameSchedule, constant.TableNameScheduleRelation)
		wheres = append(wheres, sql)
		params = append(params, c.RelationStudentIDs.Strings, []entity.ScheduleRelationType{
			entity.ScheduleRelationTypeClassRosterStudent,
			entity.ScheduleRelationTypeParticipantStudent,
		})
	}

	if c.RelationUserIDs.Valid {
		sql := fmt.Sprintf("exists(select 1 from %s where relation_id in (?) and relation_type in (?) and %s.id = %s.schedule_id)",
			constant.TableNameScheduleRelation, constant.TableNameSchedule, constant.TableNameScheduleRelation)
		wheres = append(wheres, sql)
		params = append(params, c.RelationUserIDs.Strings, []entity.ScheduleRelationType{
			entity.ScheduleRelationTypeClassRosterTeacher,
			entity.ScheduleRelationTypeParticipantTeacher,
			entity.ScheduleRelationTypeClassRosterStudent,
			entity.ScheduleRelationTypeParticipantStudent,
		})
	}

	if c.TeachLoadTeacherIDs.Valid {
		sql := fmt.Sprintf("%s.relation_id in (?) and %s.relation_type in (?)",
			constant.TableNameScheduleRelation, constant.TableNameScheduleRelation)
		wheres = append(wheres, sql)
		params = append(params, c.TeachLoadTeacherIDs.Strings,
			[]string{
				entity.ScheduleRelationTypeParticipantTeacher.String(),
				entity.ScheduleRelationTypeClassRosterTeacher.String(),
			})
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
	ScheduleOrderByScheduleAtAsc
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
	case "schedule_at":
		return ScheduleOrderByScheduleAtAsc
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
	case ScheduleOrderByScheduleAtAsc:
		return "start_at = 0 AND due_at = 0, CASE WHEN start_at = 0 THEN due_at ELSE start_at END ASC"
	default:
		return "start_at"
	}
}
