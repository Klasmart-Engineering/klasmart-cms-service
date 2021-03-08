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
	GetRelationTypeByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) (entity.ScheduleRoleType, error)
}
type scheduleRelationModel struct {
}

func (s *scheduleRelationModel) GetRelationTypeByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) (entity.ScheduleRoleType, error) {
	condition := &da.ScheduleRelationCondition{
		RelationID: sql.NullString{
			String: op.UserID,
			Valid:  true,
		},
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
	}
	var relations []*entity.ScheduleRelation
	err := da.GetScheduleRelationDA().Query(ctx, condition, &relations)
	if err != nil {
		log.Error(ctx, "get relation count error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return "", err
	}
	if len(relations) <= 0 {
		log.Info(ctx, "not found", log.Any("op", op), log.Any("condition", condition))
		return entity.ScheduleRoleTypeUnknown, nil
	}
	relation := relations[0]
	switch relation.RelationType {
	case entity.ScheduleRelationTypeParticipantTeacher, entity.ScheduleRelationTypeClassRosterTeacher:
		return entity.ScheduleRoleTypeTeacher, nil
	case entity.ScheduleRelationTypeParticipantStudent, entity.ScheduleRelationTypeClassRosterStudent:
		return entity.ScheduleRoleTypeStudent, nil
	default:
		return entity.ScheduleRoleTypeUnknown, nil
	}
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
