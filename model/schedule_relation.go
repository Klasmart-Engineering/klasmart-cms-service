package model

import (
	"context"
	"database/sql"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

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
	GetIDs(ctx context.Context, op *entity.Operator, condition *da.ScheduleRelationCondition) ([]string, error)
	GetUsers(ctx context.Context, op *entity.Operator, scheduleID string) (*entity.ScheduleUserRelation, error)
	GetSubjects(ctx context.Context, op *entity.Operator, scheduleID string) ([]*entity.ScheduleShortInfo, error)
	GetSubjectIDs(ctx context.Context, scheduleID string) ([]string, error)
	GetSubjectsByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string][]*entity.ScheduleShortInfo, error)
	GetOutcomeIDs(ctx context.Context, scheduleID string) ([]string, error)
	// scheduleID:classInfo
	GetClassRosterMap(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]*entity.ScheduleShortInfo, error)
	// // scheduleID:Relations
	GetRelationMap(ctx context.Context, op *entity.Operator, scheduleIDs []string, types []entity.ScheduleRelationType) (map[string][]*entity.ScheduleRelation, error)
}

type scheduleRelationModel struct {
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

func (s *scheduleRelationModel) GetSubjectsByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string][]*entity.ScheduleShortInfo, error) {
	var scheduleRelations []*entity.ScheduleRelation
	relationCondition := da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
		RelationTypes: entity.NullStrings{
			Strings: []string{
				string(entity.ScheduleRelationTypeSubject),
			},
			Valid: true,
		},
	}
	err := da.GetScheduleRelationDA().Query(ctx, relationCondition, &scheduleRelations)
	if err != nil {
		log.Error(ctx, "get users relation error", log.Err(err), log.Any("relationCondition", relationCondition))
		return nil, err
	}
	result := make(map[string][]*entity.ScheduleShortInfo)
	subjectIDMap := make(map[string]bool)
	subjectIDs := make([]string, 0, len(scheduleRelations))

	for _, item := range scheduleRelations {
		if _, ok := subjectIDMap[item.RelationID]; !ok {
			subjectIDMap[item.RelationID] = true
			subjectIDs = append(subjectIDs, item.RelationID)
		}
	}

	subjectMap, err := GetScheduleModel().GetSubjectsBySubjectIDs(ctx, op, subjectIDs)
	if err != nil {
		return nil, err
	}
	for _, item := range scheduleRelations {
		if _, ok := result[item.ScheduleID]; !ok {
			result[item.ScheduleID] = make([]*entity.ScheduleShortInfo, 0, len(scheduleRelations))
		}
		if subject, ok := subjectMap[item.RelationID]; ok {
			result[item.ScheduleID] = append(result[item.ScheduleID], subject)
		}
	}

	return result, nil
}

func (s *scheduleRelationModel) GetSubjectIDs(ctx context.Context, scheduleID string) ([]string, error) {
	var scheduleRelations []*entity.ScheduleRelation
	relationCondition := da.ScheduleRelationCondition{
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
		RelationTypes: entity.NullStrings{
			Strings: []string{
				string(entity.ScheduleRelationTypeSubject),
			},
			Valid: true,
		},
	}
	err := da.GetScheduleRelationDA().Query(ctx, relationCondition, &scheduleRelations)
	if err != nil {
		log.Error(ctx, "get users relation error", log.Err(err), log.Any("relationCondition", relationCondition))
		return nil, err
	}
	subjectIDs := make([]string, len(scheduleRelations))
	for i, item := range scheduleRelations {
		subjectIDs[i] = item.RelationID
	}
	return subjectIDs, nil
}

func (s *scheduleRelationModel) GetSubjects(ctx context.Context, op *entity.Operator, scheduleID string) ([]*entity.ScheduleShortInfo, error) {
	subjectIDs, err := s.GetSubjectIDs(ctx, scheduleID)
	if err != nil {
		log.Error(ctx, "get subjects error", log.Err(err), log.Any("scheduleID", scheduleID))
		return nil, err
	}
	subjects, err := external.GetSubjectServiceProvider().BatchGet(ctx, op, subjectIDs)
	if err != nil {
		return nil, err
	}
	result := make([]*entity.ScheduleShortInfo, len(subjects))
	for i, item := range subjects {
		result[i] = &entity.ScheduleShortInfo{
			ID:   item.ID,
			Name: item.Name,
		}
	}
	return result, nil
}

func (s *scheduleRelationModel) GetUsers(ctx context.Context, op *entity.Operator, scheduleID string) (*entity.ScheduleUserRelation, error) {
	result := new(entity.ScheduleUserRelation)

	_, err := GetSchedulePermissionModel().HasScheduleOrgPermissions(ctx, op, []external.PermissionName{
		external.ScheduleCreateMyEvent,
		external.ScheduleCreateMySchoolEvent,
		external.ScheduleCreateEvent,
	})
	if err == constant.ErrInternalServer {
		return nil, err
	}
	relations, err := GetScheduleRelationModel().GetUsersByScheduleID(ctx, op, scheduleID)
	if err != nil {
		log.Error(ctx, "get users by schedule id error", log.Err(err), log.Any("op", op), log.String("scheduleID", scheduleID))
		return nil, err
	}
	teacherIDs := make([]string, 0, len(relations))
	studentIDs := make([]string, 0, len(relations))
	userIDs := make([]string, 0, len(relations))
	for _, relationItem := range relations {
		switch relationItem.RelationType {
		case entity.ScheduleRelationTypeClassRosterTeacher, entity.ScheduleRelationTypeParticipantTeacher:
			teacherIDs = append(teacherIDs, relationItem.RelationID)
			userIDs = append(userIDs, relationItem.RelationID)
		case entity.ScheduleRelationTypeClassRosterStudent, entity.ScheduleRelationTypeParticipantStudent:
			if err == nil {
				studentIDs = append(studentIDs, relationItem.RelationID)
				userIDs = append(userIDs, relationItem.RelationID)
			}
		}
	}

	userMap, err := external.GetUserServiceProvider().BatchGetMap(ctx, op, userIDs)
	if err != nil {
		return nil, err
	}
	log.Info(ctx, "external.GetUserServiceProvider().BatchGetMap",
		log.Strings("userIDs", userIDs),
		log.Any("userMap", userMap))

	result.Teachers = make([]*entity.ScheduleShortInfo, 0, len(teacherIDs))
	for _, id := range teacherIDs {
		if user, ok := userMap[id]; ok {
			result.Teachers = append(result.Teachers, &entity.ScheduleShortInfo{
				ID:   user.ID,
				Name: user.Name,
			})
		}
	}

	result.Students = make([]*entity.ScheduleShortInfo, 0, len(studentIDs))
	for _, id := range studentIDs {
		if user, ok := userMap[id]; ok {
			result.Students = append(result.Students, &entity.ScheduleShortInfo{
				ID:   user.ID,
				Name: user.Name,
			})
		}
	}
	return result, nil
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
		log.Error(ctx, "da.GetScheduleRelationDA().Query error",
			log.Err(err),
			log.Any("op", op),
			log.Any("condition", condition))
		return "", err
	}
	if len(relations) == 0 {
		log.Debug(ctx, "schedule relation not found",
			log.Any("condition", condition))
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

func (s *scheduleRelationModel) GetIDs(ctx context.Context, op *entity.Operator, condition *da.ScheduleRelationCondition) ([]string, error) {
	relations, err := s.Query(ctx, op, condition)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(relations))
	for i, item := range relations {
		result[i] = item.ID
	}
	return result, nil
}

func (s *scheduleRelationModel) GetOutcomeIDs(ctx context.Context, scheduleID string) ([]string, error) {
	var scheduleRelations []*entity.ScheduleRelation
	relationCondition := da.ScheduleRelationCondition{
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
		RelationType: sql.NullString{
			String: string(entity.ScheduleRelationTypeLearningOutcome),
			Valid:  true,
		},
	}
	err := da.GetScheduleRelationDA().Query(ctx, relationCondition, &scheduleRelations)
	if err != nil {
		log.Error(ctx, "da.GetScheduleRelationDA().Query error",
			log.Err(err),
			log.Any("relationCondition", relationCondition))
		return nil, err
	}

	outcomeIDs := make([]string, len(scheduleRelations))
	for i := range scheduleRelations {
		outcomeIDs[i] = scheduleRelations[i].RelationID
	}

	return outcomeIDs, nil
}

func (s *scheduleRelationModel) GetClassRosterMap(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	condition := &da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
		RelationType: sql.NullString{
			String: string(entity.ScheduleRelationTypeClassRosterClass),
			Valid:  true,
		},
	}
	var classRelations []*entity.ScheduleRelation
	err := da.GetScheduleRelationDA().Query(ctx, condition, &classRelations)
	if err != nil {
		log.Error(ctx, "GetClassRosterMap error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return nil, err
	}

	classIDs := make([]string, 0, len(classRelations))
	classIDMap := make(map[string]struct{})
	scheduleClassMap := make(map[string]string, len(classRelations))
	for _, item := range classRelations {
		scheduleClassMap[item.ScheduleID] = item.RelationID

		if _, ok := classIDMap[item.RelationID]; !ok {
			classIDMap[item.RelationID] = struct{}{}
			classIDs = append(classIDs, item.RelationID)
		}
	}

	classMap, err := external.GetClassServiceProvider().BatchGetMap(ctx, op, classIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*entity.ScheduleShortInfo, len(classRelations))
	for _, item := range classRelations {
		if classInfo, ok := classMap[scheduleClassMap[item.ScheduleID]]; ok {
			result[item.ScheduleID] = &entity.ScheduleShortInfo{
				ID:   classInfo.ID,
				Name: classInfo.Name,
			}
		}
	}

	return result, nil
}

func (s *scheduleRelationModel) GetRelationMap(ctx context.Context, op *entity.Operator, scheduleIDs []string, types []entity.ScheduleRelationType) (map[string][]*entity.ScheduleRelation, error) {
	typesStr := make([]string, len(types))
	for i, item := range types {
		typesStr[i] = item.String()
	}

	condition := &da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
		RelationTypes: entity.NullStrings{
			Strings: typesStr,
			Valid:   true,
		},
	}
	var relations []*entity.ScheduleRelation
	err := da.GetScheduleRelationDA().Query(ctx, condition, &relations)
	if err != nil {
		log.Error(ctx, "GetClassRosterMap error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return nil, err
	}

	result := make(map[string][]*entity.ScheduleRelation, len(scheduleIDs))

	for _, item := range relations {
		result[item.ScheduleID] = append(result[item.ScheduleID], item)
	}

	return result, nil
}
