package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IScheduleRelationModel interface {
	Query(ctx context.Context, op *entity.Operator, condition *da.ScheduleRelationCondition) ([]*entity.ScheduleRelation, error)
	IsTeacher(ctx context.Context, op *entity.Operator, scheduleID string) (bool, error)
}
type scheduleRelationModel struct {
}

func (s *scheduleRelationModel) IsTeacher(ctx context.Context, op *entity.Operator, scheduleID string) (bool, error) {
	condition := &da.ScheduleRelationCondition{
		RelationID: sql.NullString{
			String: op.UserID,
			Valid:  true,
		},
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
		RelationTypes: entity.NullStrings{
			Strings: []string{string(entity.ScheduleRelationTypeClassRosterTeacher), string(entity.ScheduleRelationTypeParticipantTeacher)},
			Valid:   true,
		},
	}
	count, err := da.GetScheduleRelationDA().Count(ctx, condition, &entity.ScheduleRelation{})
	if err != nil {
		log.Error(ctx, "get relation count error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return false, err
	}
	return count > 0, nil
}

func (s *scheduleRelationModel) Query(ctx context.Context, op *entity.Operator, condition *da.ScheduleRelationCondition) ([]*entity.ScheduleRelation, error) {
	var result []*entity.ScheduleRelation
	err := da.GetScheduleRelationDA().Query(ctx, condition, &result)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

var (
	_scheduleRelationOnce  sync.Once
	_scheduleRelationModel IScheduleRelationModel
)

func GetScheduleRelationModel() IScheduleRelationModel {
	_scheduleRelationOnce.Do(func() {
		_scheduleRelationModel = &scheduleRelationModel{}
	})
	return _scheduleRelationModel
}
