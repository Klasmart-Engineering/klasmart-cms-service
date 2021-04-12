package da

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
	"sync"
)

type IScheduleRelationDA interface {
	dbo.DataAccesser
	Delete(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) error
	BatchInsert(ctx context.Context, tx *dbo.DBContext, relations []*entity.ScheduleRelation) (int64, error)
	MultipleBatchInsert(ctx context.Context, tx *dbo.DBContext, relations []*entity.ScheduleRelation) (int64, error)
	//GetRelationIDsByCondition(ctx context.Context, tx *dbo.DBContext, condition *ScheduleRelationCondition) ([]string, error)
	GetRelationIDsByCondition(ctx context.Context, tx *dbo.DBContext, condition *ScheduleRelationCondition) ([]string, error)
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
		row, err := s.BatchInsert(ctx, tx, data)
		if err != nil {
			return rowsAffected, err
		}
		rowsAffected += row
	}
	return rowsAffected, nil
}

func (s *scheduleRelationDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, relations []*entity.ScheduleRelation) (int64, error) {
	var sqlData [][]interface{}
	for _, item := range relations {
		sqlData = append(sqlData, []interface{}{
			item.ID,
			item.ScheduleID,
			item.RelationID,
			item.RelationType,
		})
	}
	if len(sqlData) <= 0 {
		return 0, nil
	}
	format, values := SQLBatchInsert(constant.TableNameScheduleRelation, []string{
		"`id`",
		"`schedule_id`",
		"`relation_id`",
		"`relation_type`",
	}, sqlData)

	execResult := tx.Exec(format, values...)
	if execResult.Error != nil {
		return 0, execResult.Error
	}
	return execResult.RowsAffected, nil
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

type ScheduleRelationCondition struct {
	ConflictCondition *ConflictCondition
	ScheduleID        sql.NullString
	RelationID        sql.NullString
	RelationType      sql.NullString
	RelationTypes     entity.NullStrings
	RelationIDs       entity.NullStrings
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
			timeWheres = append(timeWheres, fmt.Sprintf("((%s.start_at <= ? and %s.end_at > ?) or (%s.start_at <= ? and %s.end_at > ?))",
				constant.TableNameSchedule, constant.TableNameSchedule,
				constant.TableNameSchedule, constant.TableNameSchedule))
			params = append(params, item.StartAt, item.StartAt, item.EndAt, item.EndAt)
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
	return wheres, params
}

func (c ScheduleRelationCondition) GetOrderBy() string {
	return ""
}

func (c ScheduleRelationCondition) GetPager() *dbo.Pager {
	return nil
}
