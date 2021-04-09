package model

import (
	"context"
	"database/sql"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IScheduleRelationModel interface {
	Query(ctx context.Context, op *entity.Operator, condition *da.ScheduleRelationCondition) ([]*entity.ScheduleRelation, error)
	IsTeacher(ctx context.Context, op *entity.Operator, scheduleID string) (bool, error)
	GetRelationTypeByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) (entity.ScheduleRoleType, error)
	GetTeacherIDs(ctx context.Context, op *entity.Operator, scheduleID string) ([]string, error)
	GetClassRosterID(ctx context.Context, op *entity.Operator, scheduleID string) (string, error)
	GetUsersByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) ([]*entity.ScheduleRelation, error)
	Count(ctx context.Context, op *entity.Operator, condition *da.ScheduleRelationCondition) (int, error)
	HasScheduleByRelationIDs(ctx context.Context, op *entity.Operator, relationIDs []string) (bool, error)
}
type scheduleRelationModel struct {
}

func (s *scheduleRelationModel) HasScheduleByRelationIDs(ctx context.Context, op *entity.Operator, relationIDs []string) (bool, error) {
	condition := &da.ScheduleRelationCondition{
		RelationIDs: entity.NullStrings{
			Valid:   true,
			Strings: relationIDs,
		},
	}
	count, err := GetScheduleRelationModel().Count(ctx, op, condition)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *scheduleRelationModel) Count(ctx context.Context, op *entity.Operator, condition *da.ScheduleRelationCondition) (int, error) {
	count, err := da.GetScheduleRelationDA().Count(ctx, condition, &entity.ScheduleRelation{})
	if err != nil {
		log.Error(ctx, "schedule relation count error", log.Err(err), log.Any("condition", condition))
		return 0, err
	}
	return count, nil
}

func (s *scheduleRelationModel) GetUsersByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) ([]*entity.ScheduleRelation, error) {
	var scheduleRelations []*entity.ScheduleRelation
	relationCondition := da.ScheduleRelationCondition{
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
		RelationTypes: entity.NullStrings{
			Strings: []string{
				string(entity.ScheduleRelationTypeParticipantTeacher),
				string(entity.ScheduleRelationTypeParticipantStudent),
				string(entity.ScheduleRelationTypeClassRosterTeacher),
				string(entity.ScheduleRelationTypeClassRosterStudent),
			},
			Valid: true,
		},
	}
	err := da.GetScheduleRelationDA().Query(ctx, relationCondition, &scheduleRelations)
	if err != nil {
		log.Error(ctx, "get users relation error", log.Err(err), log.Any("relationCondition", relationCondition))
		return nil, err
	}
	return scheduleRelations, nil
}

func (s *scheduleRelationModel) IsTeacher(ctx context.Context, op *entity.Operator, scheduleID string) (bool, error) {
	condition := &da.ScheduleRelationCondition{
		RelationID: sql.NullString{
			String: op.UserID,
			Valid:  true,
		},
		RelationTypes: entity.NullStrings{
			Strings: []string{string(entity.ScheduleRelationTypeClassRosterTeacher), string(entity.ScheduleRelationTypeParticipantTeacher)},
			Valid:   true,
		},
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
	}
	count, err := da.GetScheduleRelationDA().Count(ctx, condition, &entity.ScheduleRelation{})
	if err != nil {
		log.Error(ctx, "get relation count error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return false, err
	}
	return count > 0, nil
}

func (s *scheduleRelationModel) GetClassRosterID(ctx context.Context, op *entity.Operator, scheduleID string) (string, error) {
	condition := &da.ScheduleRelationCondition{
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
		RelationType: sql.NullString{
			String: string(entity.ScheduleRelationTypeClassRosterClass),
			Valid:  true,
		},
	}
	var classRelations []*entity.ScheduleRelation
	err := da.GetScheduleRelationDA().Query(ctx, condition, &classRelations)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return "", err
	}

	if len(classRelations) <= 0 {
		log.Info(ctx, "schedule no class roster", log.Any("op", op), log.Any("condition", condition))
		return "", nil
	}
	return classRelations[0].RelationID, nil
}

func (s *scheduleRelationModel) GetTeacherIDs(ctx context.Context, op *entity.Operator, scheduleID string) ([]string, error) {
	condition := &da.ScheduleRelationCondition{
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
		RelationTypes: entity.NullStrings{
			Strings: []string{
				entity.ScheduleRelationTypeClassRosterTeacher.String(),
				entity.ScheduleRelationTypeParticipantTeacher.String(),
			},
			Valid: true,
		},
	}
	var relations []*entity.ScheduleRelation
	err := da.GetScheduleRelationDA().Query(ctx, condition, &relations)
	if err != nil {
		log.Error(ctx, "get relation count error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return nil, err
	}
	result := make([]string, len(relations))
	for i, item := range relations {
		result[i] = item.RelationID
	}
	return result, nil
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
