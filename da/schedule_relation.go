package da

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/jinzhu/gorm"
)

type IScheduleRelationDA interface {
	dbo.DataAccesser
	Delete(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) error
	// Deprecated: Only used to cmd/schedule_migrate/schedule_migrate.go
	MultipleBatchInsert(ctx context.Context, tx *dbo.DBContext, relations []*entity.ScheduleRelation) (int64, error)
	GetRelationIDsByCondition(ctx context.Context, tx *dbo.DBContext, condition *ScheduleRelationCondition) ([]string, error)
	DeleteByIDs(ctx context.Context, tx *dbo.DBContext, ids []string) error
	GetSubjectIDsByProgramID(ctx context.Context, tx *dbo.DBContext, orgID, programID string, schedulePermissionRelationIDs []string) ([]string, error)
}

type scheduleRelationDA struct {
	dbo.BaseDA
}

func (s *scheduleRelationDA) GetRelationIDsByCondition(ctx context.Context, tx *dbo.DBContext, condition *ScheduleRelationCondition) ([]string, error) {
	wheres, parameters := condition.GetConditions()
	whereSql := strings.Join(wheres, " and ")
	var relationList []*entity.ScheduleRelation
	err := tx.Table(constant.TableNameScheduleRelation).Select("distinct relation_id").Where(whereSql, parameters...).Find(&relationList).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetRelationIDsByCondition:get from db error",
			log.Err(err),
			log.Any("condition", condition),
		)
		return nil, err
	}
	var result = make([]string, len(relationList))
	for i, item := range relationList {
		result[i] = item.RelationID
	}
	log.Debug(ctx, "RelationIDs", log.Strings("RelationIDs", result))
	return result, nil
}

// Deprecated: Only used to cmd/schedule_migrate/schedule_migrate.go
func (s *scheduleRelationDA) MultipleBatchInsert(ctx context.Context, tx *dbo.DBContext, relations []*entity.ScheduleRelation) (int64, error) {
	total := len(relations)
	pageSize := constant.ScheduleRelationBatchInsertCount
	pageCount := (total + pageSize - 1) / pageSize
	var rowsAffected int64
	for i := 0; i < pageCount; i++ {
		start := i * pageSize
		end := (i + 1) * pageSize
		if end >= total {
			end = total
		}
		data := relations[start:end]
		batchSize := len(data)
		_, err := s.InsertInBatchesTx(ctx, tx, data, batchSize)
		if err != nil {
			return rowsAffected, err
		}

		rowsAffected += int64(batchSize)
	}

	return rowsAffected, nil
}

func (s *scheduleRelationDA) Delete(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) error {
	if err := tx.Unscoped().
		Where("schedule_id in (?)", scheduleIDs).
		Delete(&entity.ScheduleRelation{}).Error; err != nil {
		log.Error(ctx, "delete schedules relation delete failed",
			log.Strings("schedule_id", scheduleIDs),
		)
		return err
	}
	return nil
}

func (s *scheduleRelationDA) DeleteByIDs(ctx context.Context, tx *dbo.DBContext, ids []string) error {
	if err := tx.Unscoped().
		Where("id in (?)", ids).
		Delete(&entity.ScheduleRelation{}).Error; err != nil {
		log.Error(ctx, "delete schedules relation delete failed",
			log.Strings("ids", ids),
		)
		return err
	}
	return nil
}

func (*scheduleRelationDA) GetSubjectIDsByProgramID(ctx context.Context, tx *dbo.DBContext, orgID, programID string, relationIDs []string) ([]string, error) {
	tx.ResetCondition()
	var subjectIDs []string
	db := tx.Table(constant.TableNameScheduleRelation).
		Select("distinct schedules_relations.relation_id").
		Joins("left join schedules on schedules.id = schedules_relations.schedule_id").
		Where("schedules_relations.relation_type = ?", entity.ScheduleRelationTypeSubject).
		Where("schedules.org_id = ? and schedules.program_id = ? and schedules.delete_at = 0", orgID, programID)
	if len(relationIDs) > 0 {
		db = db.Joins("left join schedules_relations AS b ON schedules.id = b.schedule_id ").
			Where("b.relation_id in (?)", relationIDs)
	}

	err := db.Scan(&subjectIDs).Error
	if err != nil {
		log.Error(ctx, "GetSubjectIDsByProgramID error",
			log.Err(err),
			log.String("orgID", orgID),
			log.String("programID", programID))
		return nil, err
	}

	return subjectIDs, nil
}

var (
	_scheduleRelationOnce sync.Once
	_scheduleRelationDA   IScheduleRelationDA
)

func GetScheduleRelationDA() IScheduleRelationDA {
	_scheduleRelationOnce.Do(func() {
		_scheduleRelationDA = &scheduleRelationDA{}
	})
	return _scheduleRelationDA
}

type ConflictCondition struct {
	IgnoreScheduleID   sql.NullString
	IgnoreRepeatID     sql.NullString
	RelationIDs        []string
	ConflictTime       []*ConflictTime
	ScheduleClassTypes entity.NullStrings
	OrgID              sql.NullString
}

type ConflictTime struct {
	StartAt int64
	EndAt   int64
}

func GetNotStartCondition(classRosterID string, userIDs []string) *ScheduleRelationCondition {
	condition := &NotStartCondition{
		RosterClassID: sql.NullString{
			String: classRosterID,
			Valid:  classRosterID != "",
		},
		NotStart: sql.NullBool{
			Bool:  true,
			Valid: true,
		},
	}
	return &ScheduleRelationCondition{
		NotStartCondition: condition,
		RelationIDs: entity.NullStrings{
			Strings: userIDs,
			Valid:   true,
		},
	}
}

type NotStartCondition struct {
	RosterClassID sql.NullString
	NotStart      sql.NullBool
}

type ScheduleFilterSubject struct {
	ProgramID   sql.NullString
	OrgID       sql.NullString
	RelationIDs entity.NullStrings
}

type ScheduleRelationCondition struct {
	ConflictCondition     *ConflictCondition
	ScheduleFilterSubject *ScheduleFilterSubject
	NotStartCondition     *NotStartCondition
	RelationID            sql.NullString
	RelationType          sql.NullString
	ScheduleID            sql.NullString
	ScheduleAndRelations  []*ScheduleAndRelations
	RelationTypes         entity.NullStrings
	RelationIDs           entity.NullStrings
	ScheduleIDs           entity.NullStrings
}

type ScheduleAndRelations struct {
	ScheduleID  string
	RelationIDs []string
}

//select * from schedules_relations where
//relation_id in ('') and
//exists(select 1 from schedules where
//((start_at <= 0 and end_at > 0) or (start_at <= 0 and end_at > 0))
//and
//schedules.id = schedules_relations.schedule_id);
func (c ScheduleRelationCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}
	if c.ConflictCondition != nil {
		wheres = append(wheres, "relation_type in (?)")
		params = append(params, []entity.ScheduleRelationType{
			entity.ScheduleRelationTypeParticipantTeacher,
			entity.ScheduleRelationTypeParticipantStudent,
			entity.ScheduleRelationTypeClassRosterTeacher,
			entity.ScheduleRelationTypeClassRosterStudent,
		})

		if c.ConflictCondition.IgnoreScheduleID.Valid {
			wheres = append(wheres, "schedule_id <> ?")
			params = append(params, c.ConflictCondition.IgnoreScheduleID.String)
		}

		wheres = append(wheres, "relation_id in (?)")
		params = append(params, c.ConflictCondition.RelationIDs)

		sql := new(strings.Builder)
		sql.WriteString(fmt.Sprintf("exists(select 1 from %s where ", constant.TableNameSchedule))

		sql.WriteString("(")
		var timeWheres []string
		for _, item := range c.ConflictCondition.ConflictTime {
			timeWheres = append(timeWheres, fmt.Sprintf("(%s.start_at < ? and %s.end_at > ?)",
				constant.TableNameSchedule, constant.TableNameSchedule))
			params = append(params, item.EndAt, item.StartAt)
		}
		sql.WriteString(strings.Join(timeWheres, " or "))
		sql.WriteString(")")

		if c.ConflictCondition.IgnoreRepeatID.Valid {
			sql.WriteString(" and repeat_id <> ? ")
			params = append(params, c.ConflictCondition.IgnoreRepeatID.String)
		}
		if c.ConflictCondition.ScheduleClassTypes.Valid {
			sql.WriteString(" and class_type in (?) ")
			params = append(params, c.ConflictCondition.ScheduleClassTypes.Strings)
		}
		if c.ConflictCondition.OrgID.Valid {
			sql.WriteString(" and org_id = ? ")
			params = append(params, c.ConflictCondition.OrgID.String)
		}

		sql.WriteString(fmt.Sprintf(" and %s.id = %s.schedule_id)", constant.TableNameSchedule, constant.TableNameScheduleRelation))
		wheres = append(wheres, sql.String())
	}

	if c.ScheduleID.Valid {
		wheres = append(wheres, "schedule_id = ?")
		params = append(params, c.ScheduleID.String)
	}
	if c.RelationID.Valid {
		wheres = append(wheres, "relation_id = ?")
		params = append(params, c.RelationID.String)
	}
	if c.RelationType.Valid {
		wheres = append(wheres, "relation_type = ?")
		params = append(params, c.RelationType.String)
	}
	if c.RelationTypes.Valid {
		wheres = append(wheres, "relation_type in (?)")
		params = append(params, c.RelationTypes.Strings)
	}
	if c.RelationIDs.Valid {
		wheres = append(wheres, "relation_id in (?)")
		params = append(params, c.RelationIDs.Strings)
	}

	if c.NotStartCondition != nil {
		notEditAt := time.Now().Add(constant.ScheduleAllowEditTime).Unix()
		sql := fmt.Sprintf(`
exists(select 1 from %s where 
((due_at=0 and start_at=0 and end_at=0) || (start_at = 0 and due_at > ?) || (start_at > ?)) 
and (delete_at=0) 
and exists(select 1 from schedules_relations where relation_id = ? and relation_type = ?  and schedules.id = schedules_relations.schedule_id)
and %s.schedule_id = %s.id)`,
			constant.TableNameSchedule, constant.TableNameScheduleRelation, constant.TableNameSchedule)

		wheres = append(wheres, sql)
		params = append(params, time.Now().Unix(), notEditAt, c.NotStartCondition.RosterClassID.String, entity.ScheduleRelationTypeClassRosterClass.String())
	}

	if c.ScheduleFilterSubject != nil {
		wheres = append(wheres, "relation_type = ?")
		params = append(params, entity.ScheduleRelationTypeSubject)

		if c.ScheduleFilterSubject.ProgramID.Valid {
			sql := fmt.Sprintf(`
exists(select 1 from %s where 
org_id = ? 
and program_id = ? 
and (delete_at=0)
and %s.schedule_id = %s.id)`,
				constant.TableNameSchedule, constant.TableNameScheduleRelation, constant.TableNameSchedule)
			wheres = append(wheres, sql)
			params = append(params, c.ScheduleFilterSubject.OrgID.String, c.ScheduleFilterSubject.ProgramID.String)
		}
		if c.ScheduleFilterSubject.RelationIDs.Valid {
			sql := fmt.Sprintf(`
exists(select 1 from %s as b where b.relation_id in (?) and schedule_id = b.schedule_id) 
`,
				constant.TableNameScheduleRelation)
			wheres = append(wheres, sql)
			params = append(params, c.ScheduleFilterSubject.RelationIDs.Strings)
		}
	}
	if c.ScheduleIDs.Valid {
		wheres = append(wheres, "schedule_id in (?)")
		params = append(params, c.ScheduleIDs.Strings)
	}

	if len(c.ScheduleAndRelations) > 0 {
		var tempWhere = make([]string, len(c.ScheduleAndRelations))
		for i, item := range c.ScheduleAndRelations {
			tempWhere[i] = " (schedule_id = ? and relation_id in (?)) "
			params = append(params, item.ScheduleID, item.RelationIDs)
		}
		sql := strings.Join(tempWhere, " or ")
		wheres = append(wheres, fmt.Sprintf("(%s)", sql))
	}

	return wheres, params
}

func (c ScheduleRelationCondition) GetOrderBy() string {
	return ""
}

func (c ScheduleRelationCondition) GetPager() *dbo.Pager {
	return nil
}
