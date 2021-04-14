package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrScheduleEditMissTime         = errors.New("editable time has expired")
	ErrScheduleLessonPlanUnAuthed   = errors.New("schedule content data unAuthed")
	ErrScheduleEditMissTimeForDueAt = errors.New("editable time has expired for due at")
	ErrScheduleAlreadyHidden        = errors.New("schedule already hidden")
	ErrScheduleAlreadyFeedback      = errors.New("students already submitted feedback")
)

type IScheduleModel interface {
	Add(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error)
	AddTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error)
	Update(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleUpdateView) (string, error)
	Delete(ctx context.Context, op *entity.Operator, id string, editType entity.ScheduleEditType) error
	Query(ctx context.Context, operator *entity.Operator, condition *da.ScheduleCondition, loc *time.Location) ([]*entity.ScheduleListView, error)
	QueryScheduledDates(ctx context.Context, operator *entity.Operator, condition *da.ScheduleCondition, loc *time.Location) ([]string, error)
	Page(ctx context.Context, operator *entity.Operator, condition *da.ScheduleCondition) (int, []*entity.ScheduleSearchView, error)
	GetByID(ctx context.Context, operator *entity.Operator, id string) (*entity.ScheduleDetailsView, error)
	ConflictDetection(ctx context.Context, op *entity.Operator, input *entity.ScheduleConflictInput) (*entity.ScheduleConflictView, error)
	GetOrgClassIDsByUserIDs(ctx context.Context, operator *entity.Operator, userIDs []string, orgID string) ([]string, error)
	GetTeacherByName(ctx context.Context, operator *entity.Operator, OrgID, name string) ([]*external.Teacher, error)
	ExistScheduleAttachmentFile(ctx context.Context, attachmentPath string) bool
	ExistScheduleByLessonPlanID(ctx context.Context, lessonPlanID string) (bool, error)
	ExistScheduleByID(ctx context.Context, id string) (bool, error)
	GetPlainByID(ctx context.Context, id string) (*entity.SchedulePlain, error)
	UpdateScheduleStatus(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string, status entity.ScheduleStatus) error
	GetLessonPlanByCondition(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, condition *da.ScheduleCondition) ([]*entity.ScheduleShortInfo, error)
	GetScheduleIDsByCondition(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, condition *entity.ScheduleIDsCondition) ([]string, error)
	GetScheduleIDsByOrgID(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, orgID string) ([]string, error)
	VerifyLessonPlanAuthed(ctx context.Context, operator *entity.Operator, lessonPlanID string) (bool, error)
	GetScheduleRealTimeStatus(ctx context.Context, op *entity.Operator, id string) (*entity.ScheduleRealTimeView, error)
	GetByIDs(ctx context.Context, op *entity.Operator, ids []string) ([]*entity.SchedulePlain, error)
	UpdateScheduleShowOption(ctx context.Context, op *entity.Operator, scheduleID string, option entity.ScheduleShowOption) (string, error)
	GetPrograms(ctx context.Context, op *entity.Operator) ([]*entity.ScheduleShortInfo, error)
	GetSubjects(ctx context.Context, op *entity.Operator, programID string) ([]*entity.ScheduleShortInfo, error)
	GetClassTypes(ctx context.Context, op *entity.Operator) ([]*entity.ScheduleShortInfo, error)
	GetRosterClassNotStartScheduleIDs(ctx context.Context, rosterClassID string, userIDs []string) ([]string, error)
	GetScheduleViewByID(ctx context.Context, op *entity.Operator, id string) (*entity.SchedulePopup, error)
}
type scheduleModel struct {
	testScheduleRepeatFlag bool
}

func (s *scheduleModel) GetScheduleRealTimeStatus(ctx context.Context, op *entity.Operator, id string) (*entity.ScheduleRealTimeView, error) {
	panic("implement me")
}

func (s *scheduleModel) UpdateScheduleShowOption(ctx context.Context, op *entity.Operator, scheduleID string, option entity.ScheduleShowOption) (string, error) {
	if !option.IsValid() {
		log.Info(ctx, "option is invalid", log.String("option", string(option)), log.String("scheduleID", scheduleID))
		return "", constant.ErrInvalidArgs
	}
	_, err := GetSchedulePermissionModel().HasScheduleOrgPermissions(ctx, op, []external.PermissionName{
		external.ScheduleCreateEvent,
		external.ScheduleCreateMySchoolEvent,
		external.ScheduleCreateMyEvent,
	})
	if err != nil {
		log.Error(ctx, "no permission", log.Any("op", op), log.Any("option", option), log.String("scheduleID", scheduleID))
		return "", err
	}
	var schedule = new(entity.Schedule)
	err = da.GetScheduleDA().Get(ctx, scheduleID, schedule)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "get schedule by id failed, schedule not found", log.Err(err), log.String("scheduleID", scheduleID))
		return "", constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "get schedule by id failed",
			log.Err(err),
			log.String("id", scheduleID),
		)
		return "", err
	}
	if schedule.DeleteAt != 0 {
		log.Error(ctx, "get schedule by id failed, schedule not found",
			log.String("id", scheduleID),
		)
		return "", constant.ErrRecordNotFound
	}
	schedule.IsHidden = option == entity.ScheduleShowOptionHidden
	_, err = da.GetScheduleDA().Update(ctx, schedule)
	if err != nil {
		log.Error(ctx, "get schedule by id failed, schedule not found",
			log.Any("schedule", schedule),
			log.Any("op", op),
		)
		return "", err
	}
	err = da.GetScheduleRedisDA().Clean(ctx, op.OrgID)
	if err != nil {
		log.Warn(ctx, "Add:GetScheduleRedisDA.Clean error", log.String("orgID", op.OrgID), log.Err(err))
	}
	return schedule.ID, nil
}

func (s *scheduleModel) GetOrgClassIDsByUserIDs(ctx context.Context, operator *entity.Operator, userIDs []string, orgID string) ([]string, error) {
	userClassInfos, err := external.GetClassServiceProvider().GetByUserIDs(ctx, operator, userIDs)
	if err != nil {
		log.Error(ctx, "GetMyOrgClassIDs:GetClassServiceProvider.GetByUserID error",
			log.Err(err),
			log.String("org", orgID),
			log.Strings("userIDs", userIDs))
		return nil, err
	}

	myClassIDs := make([]string, 0)
	for _, classInfos := range userClassInfos {
		for _, item := range classInfos {
			myClassIDs = append(myClassIDs, item.ID)
		}
	}
	orgClassInfoMap, err := external.GetClassServiceProvider().GetByOrganizationIDs(ctx, operator, []string{orgID})
	if err != nil {
		log.Error(ctx, "GetMyOrgClassIDs:GetClassServiceProvider.GetByOrganizationIDs error",
			log.Err(err),
			log.Strings("userIDs", userIDs),
			log.String("orgID", orgID),
		)
		return nil, err
	}
	orgClassInfos := orgClassInfoMap[orgID]
	orgClassIDs := make([]string, len(orgClassInfos))
	for i, item := range orgClassInfos {
		orgClassIDs[i] = item.ID
	}
	result := utils.IntersectAndDeduplicateStrSlice(myClassIDs, orgClassIDs)
	log.Debug(ctx, "my org class ids",
		log.String("orgID", orgID),
		log.Strings("userIDs", userIDs),
		log.Strings("myClassIDs", myClassIDs),
		log.Strings("orgClassIDs", orgClassIDs),
		log.Strings("result", result),
	)
	return result, nil
}

func (s *scheduleModel) getRepeatResult(ctx context.Context, startAt int64, endAt int64, options *entity.RepeatOptions, location *time.Location) ([]*RepeatBaseTimeStamp, error) {
	if options == nil || !options.Type.Valid() {
		return nil, constant.ErrInvalidArgs
	}

	cfg := NewRepeatConfig(options, location)
	plan, err := NewRepeatCyclePlan(ctx, startAt, endAt, cfg)
	if err != nil {
		log.Error(ctx, "StartScheduleRepeat:NewRepeatCyclePlan error", log.Err(err), log.Any("cfg", cfg))
		return nil, err
	}
	endRule, err := NewEndRepeatCycleRule(options)
	if err != nil {
		log.Error(ctx, "StartScheduleRepeat:NewEndRepeatCycleRule error", log.Err(err), log.Any("options", options))
		return nil, err
	}
	planResult, err := plan.GenerateTimeByEndRule(endRule)
	if err != nil {
		log.Error(ctx, "StartScheduleRepeat:GenerateTimeByEndRule error", log.Err(err),
			log.Any("plan", plan),
			log.Any("endRule", endRule),
		)
		return nil, err
	}
	return planResult, nil
}

func (s *scheduleModel) buildConflictCondition(ctx context.Context, op *entity.Operator, input *entity.ScheduleConflictInput) (*da.ConflictCondition, error) {
	// participants
	partUserLen := len(input.ParticipantsTeacherIDs) + len(input.ParticipantsStudentIDs)
	partUsers := make([]*entity.ScheduleUserInput, 0, partUserLen)
	for _, id := range input.ParticipantsTeacherIDs {
		partUsers = append(partUsers, &entity.ScheduleUserInput{
			ID:   id,
			Type: entity.ScheduleRelationTypeParticipantTeacher,
		})
	}
	for _, id := range input.ParticipantsStudentIDs {
		partUsers = append(partUsers, &entity.ScheduleUserInput{
			ID:   id,
			Type: entity.ScheduleRelationTypeParticipantStudent,
		})
	}
	accessiblePartUser, err := s.AccessibleParticipantUser(ctx, op, partUsers)
	if err != nil {
		log.Error(ctx, "AccessibleParticipantUser error", log.Err(err), log.Any("input", input), log.Any("op", op))
		return nil, err
	}
	userList := make([]string, 0, len(accessiblePartUser))
	for _, item := range accessiblePartUser {
		if item.Enable {
			userList = append(userList, item.ID)
		}
	}
	// class roster
	if input.ClassID != "" {
		isClassAccessible, err := s.AccessibleClass(ctx, op, input.ClassID)
		if err != nil {
			log.Error(ctx, "AccessibleClass error", log.Err(err), log.Any("input", input), log.Any("op", op))
			return nil, err
		}
		if isClassAccessible {
			userList = append(userList, input.ClassRosterTeacherIDs...)
			userList = append(userList, input.ClassRosterStudentIDs...)
		}
	}
	conflictCondition := &da.ConflictCondition{
		RelationIDs: userList,
	}
	if input.IsRepeat {
		repeatResult, err := s.getRepeatResult(ctx, input.StartAt, input.EndAt, &input.RepeatOptions, input.Location)
		if err != nil {
			log.Error(ctx, "get repeat result error", log.Err(err), log.Any("input", input), log.Any("op", op))
			return nil, err
		}
		conflictCondition.ConflictTime = make([]*da.ConflictTime, len(repeatResult))
		for i, item := range repeatResult {
			conflictCondition.ConflictTime[i] = &da.ConflictTime{
				StartAt: item.Start,
				EndAt:   item.End,
			}
		}
	} else {
		conflictCondition.ConflictTime = []*da.ConflictTime{
			&da.ConflictTime{
				StartAt: input.StartAt,
				EndAt:   input.EndAt,
			},
		}
	}
	conflictCondition.IgnoreScheduleID = sql.NullString{
		String: input.IgnoreScheduleID,
		Valid:  input.IgnoreScheduleID != "",
	}
	if conflictCondition.IgnoreScheduleID.Valid {
		var schedule = new(entity.Schedule)
		err := da.GetScheduleDA().Get(ctx, conflictCondition.IgnoreScheduleID.String, schedule)
		if err == dbo.ErrRecordNotFound {
			log.Error(ctx, "get schedule by id failed, schedule not found", log.Err(err), log.Any("conflictCondition", conflictCondition))
			return nil, constant.ErrRecordNotFound
		}
		if err != nil {
			log.Error(ctx, "get schedule by id failed", log.Err(err), log.Any("conflictCondition", conflictCondition))
			return nil, err
		}
		conflictCondition.IgnoreRepeatID = sql.NullString{
			String: schedule.RepeatID,
			Valid:  schedule.RepeatID != "",
		}
	}
	conflictCondition.ScheduleClassTypes = entity.NullStrings{
		Strings: []string{string(entity.ScheduleClassTypeOfflineClass), string(entity.ScheduleClassTypeOnlineClass)},
		Valid:   true,
	}
	conflictCondition.OrgID = sql.NullString{
		String: op.OrgID,
		Valid:  true,
	}
	return conflictCondition, nil
}

func (s *scheduleModel) ConflictDetection(ctx context.Context, op *entity.Operator, input *entity.ScheduleConflictInput) (*entity.ScheduleConflictView, error) {
	log.Debug(ctx, "ConflictDetection data", log.Any("input", input), log.Any("op", op))
	result := new(entity.ScheduleConflictView)
	conflictCondition, err := s.buildConflictCondition(ctx, op, input)
	if err != nil {
		log.Error(ctx, "buildConflictCondition error", log.Err(err), log.Any("op", op), log.Any("input", input))
		return nil, err
	}
	condition := &da.ScheduleRelationCondition{
		ConflictCondition: conflictCondition,
	}
	userIDs, err := da.GetScheduleRelationDA().GetRelationIDsByCondition(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		log.Error(ctx, "ConflictDetection:GetScheduleRelationDA GetRelationIDsByCondition error",
			log.Any("input", input),
			log.Any("op", op),
			log.Any("condition", condition),
			log.Err(err),
		)
		return nil, err
	}

	if len(userIDs) <= 0 {
		log.Info(ctx, "not conflict", log.Any("input", input), log.Any("op", op))
		return nil, nil
	}

	userInfos, err := external.GetUserServiceProvider().BatchGet(ctx, op, userIDs)
	if err != nil {
		log.Error(ctx, "ConflictDetection:GetScheduleRelationDA Query error",
			log.Any("input", input),
			log.Any("op", op),
			log.Strings("userIDs", userIDs),
			log.Err(err),
		)
		return nil, err
	}

	// user map
	var partTeachersMap = make(map[string]bool, len(input.ParticipantsTeacherIDs))
	for _, item := range input.ParticipantsTeacherIDs {
		partTeachersMap[item] = true
	}
	var partStudentsMap = make(map[string]bool, len(input.ParticipantsStudentIDs))
	for _, item := range input.ParticipantsStudentIDs {
		partStudentsMap[item] = true
	}
	var classTeachersMap = make(map[string]bool, len(input.ClassRosterTeacherIDs))
	for _, item := range input.ClassRosterTeacherIDs {
		classTeachersMap[item] = true
	}
	var classStudentsMap = make(map[string]bool, len(input.ClassRosterStudentIDs))
	for _, item := range input.ClassRosterStudentIDs {
		classStudentsMap[item] = true
	}

	for _, item := range userInfos {
		if !item.Valid {
			log.Info(ctx, "user is invalid", log.Any("user", item), log.Any("op", op))
			return nil, constant.ErrInvalidArgs
		}
		if _, ok := classTeachersMap[item.ID]; ok {
			result.ClassRosterTeachers = append(result.ClassRosterTeachers, entity.ScheduleConflictUserView{
				ID:   item.ID,
				Name: item.Name,
			})
			continue
		}
		if _, ok := classStudentsMap[item.ID]; ok {
			result.ClassRosterStudents = append(result.ClassRosterStudents, entity.ScheduleConflictUserView{
				ID:   item.ID,
				Name: item.Name,
			})
			continue
		}
		if _, ok := partTeachersMap[item.ID]; ok {
			result.ParticipantsTeachers = append(result.ParticipantsTeachers, entity.ScheduleConflictUserView{
				ID:   item.ID,
				Name: item.Name,
			})
			continue
		}
		if _, ok := partStudentsMap[item.ID]; ok {
			result.ParticipantsStudents = append(result.ParticipantsStudents, entity.ScheduleConflictUserView{
				ID:   item.ID,
				Name: item.Name,
			})
			continue
		}
	}
	return result, constant.ErrConflict
}

func (s *scheduleModel) ExistScheduleAttachmentFile(ctx context.Context, attachmentPath string) bool {
	_, exist := storage.DefaultStorage().ExistFile(ctx, storage.ScheduleAttachmentStoragePartition, attachmentPath)
	if !exist {
		log.Info(ctx, "add schedule: attachment is not exits", log.Any("attachmentPath", attachmentPath))
		return false
	}
	return true
}

func (s *scheduleModel) GetSchoolIDsByUserIDs(ctx context.Context, op *entity.Operator, userIDs []string) ([]string, error) {
	userSchoolMap, err := external.GetSchoolServiceProvider().GetByUsers(ctx, op, op.OrgID, userIDs)
	if err != nil {
		log.Error(ctx, "GetSchoolIDsByUserIDs.GetSchoolServiceProvider.GetByUsers",
			log.Err(err),
			log.Any("op", op),
			log.Strings("userIDs", userIDs),
		)
		return nil, err
	}
	schoolIDMap := make(map[string]bool)
	for _, schools := range userSchoolMap {
		for _, schoolItem := range schools {
			schoolIDMap[schoolItem.ID] = true
		}
	}
	result := make([]string, 0, len(schoolIDMap))
	for id := range schoolIDMap {
		result = append(result, id)
	}
	return result, nil
}
func (s *scheduleModel) GetSchoolIDsByClassIDs(ctx context.Context, op *entity.Operator, classIDs []string) ([]string, error) {
	classSchoolMap, err := external.GetSchoolServiceProvider().GetByClasses(ctx, op, classIDs)
	if err != nil {
		log.Error(ctx, "GetSchoolIDsByClassIDs.GetSchoolServiceProvider.GetByClasses error",
			log.Err(err),
			log.Any("op", op),
			log.Strings("classIDs", classIDs),
		)
		return nil, err
	}
	schoolIDMap := make(map[string]bool)
	for _, schools := range classSchoolMap {
		for _, schoolItem := range schools {
			schoolIDMap[schoolItem.ID] = true
		}
	}
	result := make([]string, 0, len(schoolIDMap))
	for id := range schoolIDMap {
		result = append(result, id)
	}
	return result, nil
}

func (s *scheduleModel) prepareScheduleRelationAddData(ctx context.Context, op *entity.Operator, input *entity.ScheduleRelationInput) ([]*entity.ScheduleRelation, error) {
	var scheduleRelations = make([]*entity.ScheduleRelation, 0)
	// org relation
	scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
		RelationID:   op.OrgID,
		RelationType: entity.ScheduleRelationTypeOrg,
	})

	//rosterLen := len(input.ClassRosterTeacherIDs) + len(input.ClassRosterStudentIDs)
	schoolIDs := make([]string, 0)
	if input.ClassRosterClassID != "" {
		// class relation
		scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
			RelationID:   input.ClassRosterClassID,
			RelationType: entity.ScheduleRelationTypeClassRosterClass,
		})
		classSchoolIDs, err := s.GetSchoolIDsByClassIDs(ctx, op, []string{input.ClassRosterClassID})
		if err != nil {
			log.Error(ctx, "prepareScheduleRelationAddData:GetSchoolIDsByClassIDs error", log.Err(err), log.Any("op", op), log.Any("input", input))
			return nil, err
		}
		schoolIDs = append(schoolIDs, classSchoolIDs...)
	}

	partUserLen := len(input.ParticipantsTeacherIDs) + len(input.ParticipantsStudentIDs)
	partUserIDs := make([]string, 0, partUserLen)
	partUserIDs = append(partUserIDs, input.ParticipantsTeacherIDs...)
	partUserIDs = append(partUserIDs, input.ParticipantsStudentIDs...)
	if partUserLen != 0 {
		partUserSchoolIDs, err := s.GetSchoolIDsByUserIDs(ctx, op, partUserIDs)
		if err != nil {
			log.Error(ctx, "prepareScheduleRelationAddData:GetSchoolIDsByUserIDs error",
				log.Err(err),
				log.Any("op", op),
				log.Strings("partUserIDs", partUserIDs),
				log.Any("input", input))
			return nil, err
		}
		schoolIDs = append(schoolIDs, partUserSchoolIDs...)
	}
	schoolIDs = utils.SliceDeduplication(schoolIDs)
	// school relation
	for _, id := range schoolIDs {
		if id == "" {
			continue
		}
		scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
			RelationID:   id,
			RelationType: entity.ScheduleRelationTypeSchool,
		})
	}

	// participants class relation
	userClassMap, err := external.GetClassServiceProvider().GetByUserIDs(ctx, op, partUserIDs)
	if err != nil {
		log.Error(ctx, "prepareScheduleRelationAddData:GetClassServiceProvider  GetByUserIDs error",
			log.Err(err),
			log.Any("op", op),
			log.Strings("partUserIDs", partUserIDs),
			log.Any("input", input))
		return nil, err
	}
	classIDMap := make(map[string]bool)
	for _, classInfos := range userClassMap {
		for _, classItem := range classInfos {
			classIDMap[classItem.ID] = true
		}
	}
	for classID := range classIDMap {
		scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
			RelationID:   classID,
			RelationType: entity.ScheduleRelationTypeParticipantClass,
		})
	}

	// user relation
	for _, item := range input.ClassRosterTeacherIDs {
		scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
			RelationID:   item,
			RelationType: entity.ScheduleRelationTypeClassRosterTeacher,
		})
	}
	for _, item := range input.ClassRosterStudentIDs {
		scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
			RelationID:   item,
			RelationType: entity.ScheduleRelationTypeClassRosterStudent,
		})
	}
	for _, item := range input.ParticipantsTeacherIDs {
		scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
			RelationID:   item,
			RelationType: entity.ScheduleRelationTypeParticipantTeacher,
		})
	}
	for _, item := range input.ParticipantsStudentIDs {
		scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
			RelationID:   item,
			RelationType: entity.ScheduleRelationTypeParticipantStudent,
		})
	}

	for _, item := range input.SubjectIDs {
		scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
			RelationID:   item,
			RelationType: entity.ScheduleRelationTypeSubject,
		})
	}
	return scheduleRelations, nil
}

func (s *scheduleModel) prepareScheduleRelationUpdateData(ctx context.Context, op *entity.Operator, input *entity.ScheduleRelationInput) ([]*entity.ScheduleRelation, error) {
	// get relation from db
	oldRelations, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		ScheduleID: sql.NullString{
			String: input.ScheduleID,
			Valid:  true,
		},
		RelationTypes: entity.NullStrings{
			Strings: []string{
				string(entity.ScheduleRelationTypeClassRosterTeacher),
				string(entity.ScheduleRelationTypeClassRosterStudent),
				string(entity.ScheduleRelationTypeParticipantTeacher),
				string(entity.ScheduleRelationTypeParticipantStudent),
			},
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}
	oldClassRosterTeacherIDs := make([]string, 0)
	oldClassRosterStudentIDs := make([]string, 0)
	oldPartUsers := make([]*entity.ScheduleUserInput, 0)
	for _, item := range oldRelations {
		switch item.RelationType {
		case entity.ScheduleRelationTypeClassRosterTeacher:
			oldClassRosterTeacherIDs = append(oldClassRosterTeacherIDs, item.RelationID)
		case entity.ScheduleRelationTypeClassRosterStudent:
			oldClassRosterStudentIDs = append(oldClassRosterStudentIDs, item.RelationID)
		case entity.ScheduleRelationTypeParticipantTeacher, entity.ScheduleRelationTypeParticipantStudent:
			oldPartUsers = append(oldPartUsers, &entity.ScheduleUserInput{
				ID:   item.RelationID,
				Type: item.RelationType,
			})
		}
	}
	conditionGetClass := &da.ScheduleRelationCondition{
		ScheduleID: sql.NullString{
			String: input.ScheduleID,
			Valid:  true,
		},
		RelationType: sql.NullString{
			String: string(entity.ScheduleRelationTypeClassRosterClass),
			Valid:  true,
		},
	}
	classRelations, err := GetScheduleRelationModel().Query(ctx, op, conditionGetClass)
	if err != nil {
		return nil, err
	}
	var classID string
	if len(classRelations) > 0 {
		classID = classRelations[0].RelationID
	}
	if classID != "" {
		isClassAccessible, err := s.AccessibleClass(ctx, op, classID)
		if err != nil {
			return nil, err
		}
		if !isClassAccessible {
			input.ClassRosterClassID = classID
			input.ClassRosterTeacherIDs = utils.SliceDeduplication(oldClassRosterTeacherIDs)
			input.ClassRosterStudentIDs = utils.SliceDeduplication(oldClassRosterStudentIDs)
		}
	}

	partUserAccessible, err := s.AccessibleParticipantUser(ctx, op, oldPartUsers)
	if err != nil {
		return nil, err
	}
	for _, item := range partUserAccessible {
		if !item.Enable {
			if item.Type == entity.ScheduleRelationTypeParticipantTeacher {
				input.ParticipantsTeacherIDs = append(input.ParticipantsTeacherIDs, item.ID)
			}
			if item.Type == entity.ScheduleRelationTypeParticipantStudent {
				input.ParticipantsStudentIDs = append(input.ParticipantsStudentIDs, item.ID)
			}
		}
	}
	input.ParticipantsTeacherIDs = utils.SliceDeduplication(input.ParticipantsTeacherIDs)
	input.ParticipantsStudentIDs = utils.SliceDeduplication(input.ParticipantsStudentIDs)
	return s.prepareScheduleRelationAddData(ctx, op, input)
}

func (s *scheduleModel) Add(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error) {
	id, err := dbo.GetTransResult(ctx, func(ctx context.Context, tx *dbo.DBContext) (interface{}, error) {
		return s.AddTx(ctx, tx, op, viewData)
	})
	if err != nil {
		log.Error(ctx, "add schedule error",
			log.Err(err),
			log.Any("viewData", viewData),
		)
		return "", err
	}
	err = da.GetScheduleRedisDA().Clean(ctx, op.OrgID)
	if err != nil {
		log.Warn(ctx, "clean schedule cache error", log.String("orgID", op.OrgID), log.Err(err))
	}
	return id.(string), nil
}
func (s *scheduleModel) AddTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error) {
	// verify data
	err := s.verifyData(ctx, op, &entity.ScheduleVerify{
		ClassID:      viewData.ClassID,
		SubjectIDs:   viewData.SubjectIDs,
		ProgramID:    viewData.ProgramID,
		LessonPlanID: viewData.LessonPlanID,
		ClassType:    viewData.ClassType,
		IsHomeFun:    viewData.IsHomeFun,
	})
	if err != nil {
		log.Error(ctx, "add schedule: verify data error",
			log.Err(err),
			log.Any("viewData", viewData))
		return "", err
	}
	if viewData.ClassType == entity.ScheduleClassTypeTask {
		viewData.LessonPlanID = ""
		viewData.ProgramID = ""
		viewData.SubjectIDs = nil
	}

	schedule, err := viewData.ToSchedule(ctx)
	schedule.CreatedID = op.UserID
	relationInput := &entity.ScheduleRelationInput{
		ClassRosterClassID:     viewData.ClassID,
		ClassRosterTeacherIDs:  viewData.ClassRosterTeacherIDs,
		ClassRosterStudentIDs:  viewData.ClassRosterStudentIDs,
		ParticipantsTeacherIDs: viewData.ParticipantsTeacherIDs,
		ParticipantsStudentIDs: viewData.ParticipantsStudentIDs,
		SubjectIDs:             viewData.SubjectIDs,
	}
	relations, err := s.prepareScheduleRelationAddData(ctx, op, relationInput)
	if err != nil {
		log.Error(ctx, "prepareScheduleRelationAddData error", log.Err(err), log.Any("op", op), log.Any("relationInput", relationInput))
		return "", err
	}

	scheduleID, err := s.addSchedule(ctx, tx, schedule, &viewData.Repeat, viewData.Location, relations)
	if err != nil {
		log.Error(ctx, "add schedule: error",
			log.Err(err),
			log.Any("viewData", viewData),
			log.Any("schedule", schedule),
		)
		return "", err
	}
	return scheduleID, nil
}

func (s *scheduleModel) addSchedule(ctx context.Context, tx *dbo.DBContext, schedule *entity.Schedule, options *entity.RepeatOptions, location *time.Location, relations []*entity.ScheduleRelation) (string, error) {
	scheduleList, err := s.StartScheduleRepeat(ctx, schedule, options, location)
	if err != nil {
		log.Error(ctx, "schedule repeat error", log.Err(err), log.Any("schedule", schedule), log.Any("options", options))
		return "", err
	}
	// add to schedules
	_, err = da.GetScheduleDA().MultipleBatchInsert(ctx, tx, scheduleList)
	if err != nil {
		log.Error(ctx, "schedule batchInsert error", log.Err(err), log.Any("scheduleList", scheduleList))
		return "", err
	}
	if len(scheduleList) <= 0 {
		log.Error(ctx, "schedules batchInsert error,schedules is empty", log.Any("schedule", schedule), log.Any("options", options))
		return "", constant.ErrRecordNotFound
	}
	// add schedules relation
	allRelations := make([]*entity.ScheduleRelation, 0)
	for _, item := range scheduleList {
		for _, relation := range relations {
			if relation.RelationID == "" {
				continue
			}
			allRelations = append(allRelations, &entity.ScheduleRelation{
				ID:           utils.NewID(),
				ScheduleID:   item.ID,
				RelationID:   relation.RelationID,
				RelationType: relation.RelationType,
			})
		}
	}
	_, err = da.GetScheduleRelationDA().MultipleBatchInsert(ctx, tx, allRelations)
	if err != nil {
		log.Error(ctx, "schedules_relations batchInsert error", log.Err(err), log.Any("allRelations", allRelations))
		return "", err
	}
	return scheduleList[0].ID, nil
}
func (s *scheduleModel) checkScheduleStatus(ctx context.Context, op *entity.Operator, id string) (*entity.Schedule, error) {
	// get old schedule by id
	var schedule = new(entity.Schedule)
	err := da.GetScheduleDA().Get(ctx, id, schedule)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "checkScheduleStatus: get schedule by id failed, schedule not found", log.Err(err), log.String("id", id))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "checkScheduleStatus: get schedule by id failed",
			log.Err(err),
			log.String("id", id),
		)
		return nil, err
	}
	if schedule.DeleteAt != 0 {
		log.Error(ctx, "checkScheduleStatus: get schedule by id failed, schedule not found",
			log.String("id", id),
		)
		return nil, constant.ErrRecordNotFound
	}
	if schedule.Status != entity.ScheduleStatusNotStart {
		log.Warn(ctx, "checkScheduleStatus: schedule status error",
			log.String("id", id),
			log.Any("schedule", schedule),
		)
		return nil, constant.ErrOperateNotAllowed
	}
	if schedule.ClassType == entity.ScheduleClassTypeHomework &&
		schedule.IsHomeFun &&
		schedule.IsHidden {
		log.Info(ctx, "schedule already hidden", log.Any("schedule", schedule))
		return nil, ErrScheduleAlreadyHidden
	}
	if schedule.ClassType == entity.ScheduleClassTypeHomework && schedule.IsHomeFun {
		exist, err := GetScheduleFeedbackModel().ExistByScheduleID(ctx, op, schedule.ID)
		if err != nil {
			log.Error(ctx, "update schedule: get schedule feedback error",
				log.Any("schedule", schedule),
				log.Err(err),
			)
			return nil, err
		}
		if exist {
			log.Info(ctx, "ErrScheduleAlreadyAssignments", log.Any("schedule", schedule))
			return nil, ErrScheduleAlreadyFeedback
		}
	}
	switch schedule.ClassType {
	case entity.ScheduleClassTypeHomework, entity.ScheduleClassTypeTask:
		if schedule.DueAt > 0 {
			now := time.Now().Unix()
			dueAtEnd := utils.TodayEndByTimeStamp(schedule.DueAt, time.Local).Unix()
			if dueAtEnd < now {
				log.Warn(ctx, "checkScheduleStatus: the due_at time has expired",
					log.Any("schedule", schedule),
					log.Any("now", now),
					log.Any("dueAtEnd", dueAtEnd),
				)
				return nil, ErrScheduleEditMissTimeForDueAt
			}
		}

	case entity.ScheduleClassTypeOnlineClass, entity.ScheduleClassTypeOfflineClass:
		diff := utils.TimeStampDiff(schedule.StartAt, time.Now().Unix())
		if diff <= constant.ScheduleAllowEditTime {
			log.Warn(ctx, "checkScheduleStatus: GetDiffToMinutesByTimeStamp warn",
				log.Any("schedule", schedule),
				log.Int64("schedule.StartAt", schedule.StartAt),
				log.Any("diff", diff),
				log.Any("ScheduleAllowEditTime", constant.ScheduleAllowEditTime),
			)
			return nil, ErrScheduleEditMissTime
		}
	}

	return schedule, nil
}
func (s *scheduleModel) Update(ctx context.Context, operator *entity.Operator, viewData *entity.ScheduleUpdateView) (string, error) {
	schedule, err := s.checkScheduleStatus(ctx, operator, viewData.ID)
	if err != nil {
		log.Error(ctx, "update schedule: get schedule by id error",
			log.Any("viewData", viewData),
			log.Err(err),
		)
		return "", err
	}

	// verify data
	err = s.verifyData(ctx, operator, &entity.ScheduleVerify{
		ClassID:      viewData.ClassID,
		SubjectIDs:   viewData.SubjectIDs,
		ProgramID:    viewData.ProgramID,
		LessonPlanID: viewData.LessonPlanID,
		ClassType:    viewData.ClassType,
		IsHomeFun:    viewData.IsHomeFun,
	})
	if err != nil {
		log.Error(ctx, "update schedule: verify data error",
			log.Err(err),
			log.Any("viewData", viewData))
		return "", err
	}

	if viewData.ClassType == entity.ScheduleClassTypeTask {
		viewData.LessonPlanID = ""
		viewData.ProgramID = ""
		viewData.SubjectIDs = nil
	}

	// update schedule
	var id string
	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		relationInput := &entity.ScheduleRelationInput{
			ScheduleID:             viewData.ID,
			ClassRosterClassID:     viewData.ClassID,
			ClassRosterTeacherIDs:  viewData.ClassRosterTeacherIDs,
			ClassRosterStudentIDs:  viewData.ClassRosterStudentIDs,
			ParticipantsTeacherIDs: viewData.ParticipantsTeacherIDs,
			ParticipantsStudentIDs: viewData.ParticipantsStudentIDs,
			SubjectIDs:             viewData.SubjectIDs,
		}
		var err error
		relations, err := s.prepareScheduleRelationUpdateData(ctx, operator, relationInput)
		if err != nil {
			return err
		}
		// delete schedule
		if err = s.deleteScheduleTx(ctx, tx, operator, schedule, viewData.EditType); err != nil {
			log.Error(ctx, "update schedule: delete failed",
				log.Err(err),
				log.String("id", viewData.ID),
				log.String("edit_type", string(viewData.EditType)),
			)
			return err
		}
		// delete relation
		err = s.deleteScheduleRelationTx(ctx, tx, operator, schedule, viewData.EditType)
		if err != nil {
			log.Error(ctx, "delete schedule relation error",
				log.Err(err),
				log.Any("schedule", schedule),
				log.String("edit_type", string(viewData.EditType)),
			)
			return err
		}
		// add schedule,update old schedule fields that need to be updated
		schedule.ID = utils.NewID()
		schedule.LessonPlanID = viewData.LessonPlanID
		schedule.ProgramID = viewData.ProgramID
		schedule.ClassID = viewData.ClassID
		schedule.StartAt = viewData.StartAt
		schedule.EndAt = viewData.EndAt
		schedule.Title = viewData.Title
		schedule.IsAllDay = viewData.IsAllDay
		schedule.Description = viewData.Description
		schedule.DueAt = viewData.DueAt
		schedule.ClassType = viewData.ClassType
		schedule.CreatedID = operator.UserID
		schedule.CreatedAt = time.Now().Unix()
		schedule.UpdatedID = operator.UserID
		schedule.UpdatedAt = time.Now().Unix()
		schedule.DeletedID = ""
		schedule.DeleteAt = 0
		schedule.IsHomeFun = viewData.IsHomeFun
		if viewData.ClassType != entity.ScheduleClassTypeHomework {
			schedule.IsHomeFun = false
		}
		// attachment
		b, err := json.Marshal(viewData.Attachment)
		if err != nil {
			log.Warn(ctx, "update schedule:marshal attachment error", log.Any("attachment", viewData.Attachment))
			return err
		}
		schedule.Attachment = string(b)

		// update repeat rule
		var repeatOptions *entity.RepeatOptions
		// if repeat selected, use repeat rule
		if viewData.IsRepeat {
			b, err := json.Marshal(viewData.Repeat)
			if err != nil {
				return err
			}
			schedule.RepeatJson = string(b)
			// if following selected, set repeat rule
			if viewData.EditType == entity.ScheduleEditWithFollowing {
				repeatOptions = &viewData.Repeat
			}
			if schedule.RepeatID == "" {
				schedule.RepeatID = utils.NewID()
			}
		} else {
			// if repeat not selected,but need to update follow schedule, use old schedule repeat rule
			if viewData.EditType == entity.ScheduleEditWithFollowing {
				var repeat = new(entity.RepeatOptions)
				if err := json.Unmarshal([]byte(schedule.RepeatJson), repeat); err != nil {
					log.Error(ctx, "update schedule:unmarshal schedule repeatJson error",
						log.Err(err),
						log.Any("viewData", viewData),
						log.Any("schedule", schedule),
					)
					return err
				}
				repeatOptions = repeat
			}
		}
		id, err = s.addSchedule(ctx, tx, schedule, repeatOptions, viewData.Location, relations)
		if err != nil {
			log.Error(ctx, "update schedule: add failed",
				log.Err(err),
				log.Any("schedule", schedule),
				log.Any("viewData", viewData),
			)
			return err
		}
		return nil
	}); err != nil {
		log.Error(ctx, "update schedule: tx failed", log.Err(err))
		return "", err
	}
	err = da.GetScheduleRedisDA().Clean(ctx, operator.OrgID)
	if err != nil {
		log.Warn(ctx, "clean schedule cache error", log.String("orgID", operator.OrgID), log.Err(err))
	}
	return id, nil
}
func (s *scheduleModel) Delete(ctx context.Context, op *entity.Operator, id string, editType entity.ScheduleEditType) error {
	schedule, err := s.checkScheduleStatus(ctx, op, id)
	if err == constant.ErrRecordNotFound {
		log.Warn(ctx, "DeleteTx:schedule not found",
			log.Err(err),
			log.String("id", id),
			log.String("edit_type", string(editType)),
		)
		return nil
	}
	if err != nil {
		log.Error(ctx, "DeleteTx:delete schedule by id error",
			log.Err(err),
			log.String("id", id),
			log.String("edit_type", string(editType)),
		)
		return err
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// delete schedule
		err := s.deleteScheduleTx(ctx, tx, op, schedule, editType)
		if err != nil {
			log.Error(ctx, "delete schedule error",
				log.Err(err),
				log.Any("schedule", schedule),
				log.String("edit_type", string(editType)),
			)
			return err
		}
		// delete relation
		err = s.deleteScheduleRelationTx(ctx, tx, op, schedule, editType)
		if err != nil {
			log.Error(ctx, "delete schedule relation error",
				log.Err(err),
				log.Any("schedule", schedule),
				log.String("edit_type", string(editType)),
			)
			return err
		}
		return nil
	})
	if err != nil {
		log.Error(ctx, "delete schedule error",
			log.Err(err),
			log.String("id", id),
			log.String("edit_type", string(editType)),
		)
		return err
	}
	err = da.GetScheduleRedisDA().Clean(ctx, op.OrgID)
	if err != nil {
		log.Warn(ctx, "clean schedule cache error", log.String("orgID", op.OrgID), log.Err(err))
	}

	return nil
}

func (s *scheduleModel) deleteScheduleRelationTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, schedule *entity.Schedule, editType entity.ScheduleEditType) error {
	if editType == entity.ScheduleEditOnlyCurrent {
		return da.GetScheduleRelationDA().Delete(ctx, tx, []string{schedule.ID})
	}
	if editType == entity.ScheduleEditWithFollowing {
		if schedule.RepeatID == "" {
			return da.GetScheduleRelationDA().Delete(ctx, tx, []string{schedule.ID})
		}
		var scheduleList []*entity.Schedule
		condition := da.ScheduleCondition{
			StartAtGe: sql.NullInt64{
				Int64: schedule.StartAt,
				Valid: true,
			},
			RepeatID: sql.NullString{
				String: schedule.RepeatID,
				Valid:  true,
			},
		}
		err := da.GetScheduleDA().Query(ctx, condition, &scheduleList)
		if err != nil {
			log.Error(ctx, "delete schedule relation error",
				log.Err(err),
				log.Any("op", op),
				log.Any("condition", condition),
			)
			return err
		}
		scheduleIDs := make([]string, len(scheduleList))
		for i, item := range scheduleList {
			scheduleIDs[i] = item.ID
		}
		return da.GetScheduleRelationDA().Delete(ctx, tx, scheduleIDs)
	}
	return constant.ErrInvalidArgs
}

func (s *scheduleModel) deleteScheduleTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, schedule *entity.Schedule, editType entity.ScheduleEditType) error {
	switch editType {
	case entity.ScheduleEditOnlyCurrent:
		if err := da.GetScheduleDA().SoftDelete(ctx, tx, schedule.ID, op); err != nil {
			log.Error(ctx, "delete schedule: soft delete failed",
				log.Any("schedule", schedule),
				log.String("edit_type", string(editType)),
			)
			return err
		}

	case entity.ScheduleEditWithFollowing:
		if schedule.RepeatID == "" {
			if err := da.GetScheduleDA().SoftDelete(ctx, tx, schedule.ID, op); err != nil {
				log.Error(ctx, "delete schedule: soft delete failed",
					log.Any("schedule", schedule),
					log.String("edit_type", string(editType)),
				)
				return err
			}
			return nil
		}
		if err := da.GetScheduleDA().DeleteWithFollowing(ctx, tx, schedule.RepeatID, schedule.StartAt); err != nil {
			log.Error(ctx, "delete schedule: delete with following failed",
				log.Err(err),
				log.String("repeat_id", schedule.RepeatID),
				log.Int64("start_at", schedule.StartAt),
				log.String("edit_type", string(editType)),
			)
			return err
		}
	}
	return nil
}

func (s *scheduleModel) Page(ctx context.Context, operator *entity.Operator, condition *da.ScheduleCondition) (int, []*entity.ScheduleSearchView, error) {
	var scheduleList []*entity.Schedule
	total, err := da.GetScheduleDA().Page(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "Page: schedule query error", log.Err(err), log.Any("condition", condition))
		return 0, nil, err
	}

	result := make([]*entity.ScheduleSearchView, 0, len(scheduleList))
	basicInfoInput := make([]*entity.ScheduleBasicDataInput, len(scheduleList))
	for i, item := range scheduleList {
		relations, err := GetScheduleRelationModel().GetUsersByScheduleID(ctx, operator, item.ID)
		if err != nil {
			return 0, nil, err
		}
		teacherIDs := make([]string, 0, len(relations))
		studentIDs := make([]string, 0, len(relations))
		for _, relationItem := range relations {
			switch relationItem.RelationType {
			case entity.ScheduleRelationTypeClassRosterTeacher, entity.ScheduleRelationTypeParticipantTeacher:
				teacherIDs = append(teacherIDs, relationItem.RelationID)
			case entity.ScheduleRelationTypeClassRosterStudent, entity.ScheduleRelationTypeParticipantStudent:
				studentIDs = append(studentIDs, relationItem.RelationID)
			}
		}
		subjectIDs, err := GetScheduleRelationModel().GetSubjectIDs(ctx, operator, item.ID)
		if err != nil {
			return 0, nil, err
		}
		basicInfoInput[i] = &entity.ScheduleBasicDataInput{
			ScheduleID:   item.ID,
			ClassID:      item.ClassID,
			ProgramID:    item.ProgramID,
			LessonPlanID: item.LessonPlanID,
			SubjectIDs:   subjectIDs,
			TeacherIDs:   teacherIDs,
			StudentIDs:   studentIDs,
		}
	}
	basicInfo, err := s.getBasicInfo(ctx, operator, basicInfoInput)
	if err != nil {
		log.Error(ctx, "Page: get basic info error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("scheduleList", scheduleList))
		return 0, nil, err
	}
	for _, item := range scheduleList {
		if item.ClassType == entity.ScheduleClassTypeHomework && item.DueAt <= 0 {
			log.Info(ctx, "schedule type is homework", log.Any("schedule", item))
			continue
		}
		viewData := &entity.ScheduleSearchView{
			ID:      item.ID,
			StartAt: item.StartAt,
			Title:   item.Title,
			EndAt:   item.EndAt,
		}
		if v, ok := basicInfo[item.ID]; ok {
			viewData.ScheduleBasic = *v
		}
		result = append(result, viewData)
	}

	return total, result, nil
}

func (s *scheduleModel) QueryByCache(ctx context.Context, op *entity.Operator, condition *da.ScheduleCondition) ([]*entity.ScheduleListView, error) {
	result, err := da.GetScheduleRedisDA().SearchToListView(ctx, op.OrgID, condition)
	if err != nil {
		log.Info(ctx, "Query from cache error",
			log.Err(err),
			log.Any("op", op),
			log.Any("condition", condition),
		)
		return nil, err
	}
	if len(result) == 0 {
		log.Info(ctx, "Query from cache not found",
			log.Err(err),
			log.Any("op", op),
			log.Any("condition", condition),
		)
		return nil, constant.ErrRecordNotFound
	}

	for _, item := range result {
		item.Status = item.Status.GetScheduleStatus(entity.ScheduleStatusInput{
			EndAt:     item.EndAt,
			DueAt:     item.DueAt,
			ClassType: item.ClassType,
		})
	}
	return result, nil
}

func (s *scheduleModel) ProcessQueryData(ctx context.Context, op *entity.Operator, scheduleList []*entity.Schedule, loc *time.Location) ([]*entity.ScheduleListView, error) {
	result := make([]*entity.ScheduleListView, 0, len(scheduleList))

	for _, item := range scheduleList {
		temp := &entity.ScheduleListView{
			ID:           item.ID,
			Title:        item.Title,
			StartAt:      item.StartAt,
			EndAt:        item.EndAt,
			IsRepeat:     item.RepeatID != "",
			LessonPlanID: item.LessonPlanID,
			Status:       item.Status,
			ClassType:    item.ClassType,
			ClassID:      item.ClassID,
			DueAt:        item.DueAt,
			IsHidden:     item.IsHidden,
			IsHomeFun:    item.IsHomeFun,
		}
		temp.Status = temp.Status.GetScheduleStatus(entity.ScheduleStatusInput{
			EndAt:     temp.EndAt,
			DueAt:     temp.DueAt,
			ClassType: temp.ClassType,
		})
		if temp.ClassType == entity.ScheduleClassTypeHomework {
			temp.StartAt = utils.TodayZeroByTimeStamp(temp.DueAt, loc).Unix()
			temp.EndAt = utils.TodayEndByTimeStamp(temp.DueAt, loc).Unix()
		}
		// get role type
		roleType, err := GetScheduleRelationModel().GetRelationTypeByScheduleID(ctx, op, item.ID)
		if err != nil {
			log.Error(ctx, "get relation type error", log.Any("op", op), log.Any("schedule", item), log.Err(err))
			return nil, err
		}
		temp.RoleType = roleType
		// verify is exist feedback
		existFeedback, err := GetScheduleFeedbackModel().ExistByScheduleID(ctx, op, item.ID)
		if err != nil {
			log.Error(ctx, "exist by schedule id error", log.Any("op", op), log.Any("schedule_id", item.ID), log.Err(err))
			return nil, err
		}
		temp.ExistFeedback = existFeedback

		result = append(result, temp)
	}

	return result, nil
}

func (s *scheduleModel) QueryByDB(ctx context.Context, op *entity.Operator, condition *da.ScheduleCondition) ([]*entity.Schedule, error) {
	var scheduleList []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "schedule query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}

	return scheduleList, nil
}
func (s *scheduleModel) Query(ctx context.Context, op *entity.Operator, condition *da.ScheduleCondition, loc *time.Location) ([]*entity.ScheduleListView, error) {
	cacheData, err := s.QueryByCache(ctx, op, condition)
	if err == nil {
		log.Debug(ctx, "Query:using cache",
			log.Any("op", op),
			log.Any("condition", condition),
		)
		return cacheData, nil
	}

	scheduleData, err := s.QueryByDB(ctx, op, condition)
	if err != nil {
		return nil, err
	}

	result, err := s.ProcessQueryData(ctx, op, scheduleData, loc)
	if err != nil {
		return nil, err
	}

	// cache
	err = da.GetScheduleRedisDA().Add(ctx, op.OrgID, &da.ScheduleCacheCondition{Condition: condition}, result)
	if err != nil {
		log.Info(ctx, "schedule add cache error", log.Err(err))
	}

	return result, nil
}

func (s *scheduleModel) getBasicInfo(ctx context.Context, op *entity.Operator, input []*entity.ScheduleBasicDataInput) (map[string]*entity.ScheduleBasic, error) {
	scheduleBasicMap := make(map[string]*entity.ScheduleBasic)
	if len(input) == 0 {
		return scheduleBasicMap, nil
	}
	var other []*entity.ScheduleBasicDataInput
	for _, item := range input {
		cacheData, err := da.GetScheduleRedisDA().SearchToBasicData(ctx, op.OrgID, item.ScheduleID)
		if err != nil {
			other = append(other, item)
			continue
		}
		log.Debug(ctx, "get basic data:using cache",
			log.Err(err),
			log.Any("op", op),
			log.Any("item", item),
			log.Any("cacheData", cacheData),
		)
		scheduleBasicMap[item.ScheduleID] = cacheData
	}

	var (
		classIDs      []string
		classMap      map[string]*entity.ScheduleAccessibleUserView
		subjectIDs    []string
		subjectMap    map[string]*entity.ScheduleShortInfo
		programIDs    []string
		programMap    map[string]*entity.ScheduleShortInfo
		lessonPlanIDs []string
		lessonPlanMap map[string]*entity.ScheduleShortInfo
		teacherIDs    []string
		teacherMap    = make(map[string]*entity.ScheduleShortInfo)
		scheduleIDs   = make([]string, len(other))
	)
	for i, item := range other {
		if item.ClassID != "" {
			classIDs = append(classIDs, item.ClassID)
		}
		if len(item.SubjectIDs) != 0 {
			subjectIDs = append(subjectIDs, item.SubjectIDs...)
		}
		if item.ProgramID != "" {
			programIDs = append(programIDs, item.ProgramID)
		}
		if item.LessonPlanID != "" {
			lessonPlanIDs = append(lessonPlanIDs, item.LessonPlanID)
		}
		if len(item.TeacherIDs) != 0 {
			for _, teacherID := range item.TeacherIDs {
				if _, ok := teacherMap[teacherID]; ok {
					continue
				}
				teacherMap[teacherID] = &entity.ScheduleShortInfo{}
			}
		}

		scheduleIDs[i] = item.ScheduleID
	}
	for key, _ := range teacherMap {
		teacherIDs = append(teacherIDs, key)
	}
	if len(teacherIDs) != 0 {
		teachers, err := external.GetUserServiceProvider().BatchGet(ctx, op, teacherIDs)
		if err != nil {
			return nil, err
		}
		for _, item := range teachers {
			teacherMap[item.ID] = &entity.ScheduleShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}

	classIDs = utils.SliceDeduplication(classIDs)
	classMap, err := s.getClassInfoMapByClassIDs(ctx, op, classIDs)
	if err != nil {
		log.Error(ctx, "get class info error", log.Err(err), log.Strings("classIDs", classIDs))
		return nil, err
	}

	subjectMap, err = s.getSubjectsByIDs(ctx, op, subjectIDs)
	if err != nil {
		log.Error(ctx, "get subject info error", log.Err(err), log.Strings("subjectIDs", subjectIDs))
		return nil, err
	}

	programMap, err = s.getProgramsByIDs(ctx, op, programIDs)
	if err != nil {
		log.Error(ctx, "get program info error", log.Err(err), log.Strings("programIDs", programIDs))
		return nil, err
	}
	lessonPlanMap, err = s.getLessonPlanByIDs(ctx, dbo.MustGetDB(ctx), lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "get lesson plan info error", log.Err(err), log.Any("lessonPlanIDs", lessonPlanIDs))
		return nil, err
	}

	for _, item := range other {
		scheduleBasic := &entity.ScheduleBasic{}
		if v, ok := classMap[item.ClassID]; ok {
			scheduleBasic.Class = v
		}
		scheduleBasic.Subjects = make([]*entity.ScheduleShortInfo, len(item.SubjectIDs))
		for i, subID := range item.SubjectIDs {
			scheduleBasic.Subjects[i] = subjectMap[subID]
		}
		if v, ok := programMap[item.ProgramID]; ok {
			scheduleBasic.Program = v
		}
		if v, ok := lessonPlanMap[item.LessonPlanID]; ok {
			scheduleBasic.LessonPlan = v
		}
		scheduleBasic.MemberTeachers = make([]*entity.ScheduleShortInfo, len(item.TeacherIDs))
		for i, teacherID := range item.TeacherIDs {
			scheduleBasic.MemberTeachers[i] = teacherMap[teacherID]
		}
		scheduleBasic.StudentCount = len(item.StudentIDs)
		scheduleBasic.Members = scheduleBasic.MemberTeachers
		scheduleBasicMap[item.ScheduleID] = scheduleBasic

		da.GetScheduleRedisDA().Add(ctx, op.OrgID, &da.ScheduleCacheCondition{ScheduleID: item.ScheduleID}, scheduleBasic)
	}

	return scheduleBasicMap, nil
}

func (s *scheduleModel) getLessonPlanByIDs(ctx context.Context, tx *dbo.DBContext, lessonPlanIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	lessonPlanMap := make(map[string]*entity.ScheduleShortInfo)
	if len(lessonPlanIDs) != 0 {
		lessonPlanIDs = utils.SliceDeduplication(lessonPlanIDs)
		lessonPlans, err := GetContentModel().GetContentNameByIDList(ctx, tx, lessonPlanIDs)
		if err != nil {
			log.Error(ctx, "get lesson plan info error", log.Err(err), log.Strings("lessonPlanIDs", lessonPlanIDs))
			return nil, err
		}

		for _, item := range lessonPlans {
			lessonPlanMap[item.ID] = &entity.ScheduleShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}
	return lessonPlanMap, nil
}

func (s *scheduleModel) getLessonPlanWithMaterial(ctx context.Context, op *entity.Operator, lessonPlanID string) (*entity.ScheduleLessonPlan, error) {
	result := new(entity.ScheduleLessonPlan)
	if lessonPlanID != "" {
		lessonInfo, err := GetContentModel().GetContentNameByID(ctx, dbo.MustGetDB(ctx), lessonPlanID)
		if err != nil {
			log.Error(ctx, "get content name by id error",
				log.Err(err),
				log.Any("lessonPlanID", lessonPlanID))
			return nil, err
		}
		result.ID = lessonInfo.ID
		result.Name = lessonInfo.Name

		isAuth, err := s.VerifyLessonPlanAuthed(ctx, op, lessonPlanID)
		if err != nil {
			return nil, err
		}
		result.IsAuth = isAuth

		contentList, err := GetContentModel().GetContentSubContentsByID(ctx, dbo.MustGetDB(ctx), lessonPlanID, op)
		if err == dbo.ErrRecordNotFound {
			log.Error(ctx, "getMaterials:get content sub by id not found",
				log.Err(err),
				log.Any("lessonPlanID", lessonPlanID))
			return nil, constant.ErrRecordNotFound
		}
		if err != nil {
			log.Error(ctx, "getMaterials:get content sub by id error",
				log.Err(err),
				log.Any("lessonPlanID", lessonPlanID))
			return nil, err
		}
		result.Materials = make([]*entity.ScheduleLessonPlanMaterial, len(contentList))
		for i, item := range contentList {
			materialItem := &entity.ScheduleLessonPlanMaterial{
				ID:   item.ID,
				Name: item.Name,
			}
			result.Materials[i] = materialItem
		}
	}
	return result, nil
}

func (s *scheduleModel) getClassInfoMapByClassIDs(ctx context.Context, operator *entity.Operator, classIDs []string) (map[string]*entity.ScheduleAccessibleUserView, error) {
	var classMap = make(map[string]*entity.ScheduleAccessibleUserView)
	if len(classIDs) != 0 {
		classService := external.GetClassServiceProvider()
		classInfos, err := classService.BatchGet(ctx, operator, classIDs)
		if err != nil {
			log.Error(ctx, "GetClassServiceProvider BatchGet error", log.Err(err), log.Strings("classIDs", classIDs))
			return nil, err
		}
		for _, item := range classInfos {
			if item != nil {
				classMap[item.ID] = &entity.ScheduleAccessibleUserView{
					ID:   item.ID,
					Name: item.Name,
				}
			}
		}
	}
	return classMap, nil
}

func (s *scheduleModel) getSubjectsByIDs(ctx context.Context, operator *entity.Operator, subjectIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	var subjectMap = make(map[string]*entity.ScheduleShortInfo)
	if len(subjectIDs) != 0 {
		subjectIDs = utils.SliceDeduplication(subjectIDs)
		subjectInfos, err := external.GetSubjectServiceProvider().BatchGet(ctx, operator, subjectIDs)
		if err != nil {
			log.Error(ctx, "GetSubjectServiceProvider BatchGet error", log.Err(err), log.Strings("subjectIDs", subjectIDs))
			return nil, err
		}

		for _, item := range subjectInfos {
			subjectMap[item.ID] = &entity.ScheduleShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}
	return subjectMap, nil
}

func (s *scheduleModel) getProgramsByIDs(ctx context.Context, operator *entity.Operator, programIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	var programMap = make(map[string]*entity.ScheduleShortInfo)
	if len(programIDs) != 0 {
		programIDs = utils.SliceDeduplication(programIDs)
		programInfos, err := external.GetProgramServiceProvider().BatchGet(ctx, operator, programIDs)
		if err != nil {
			log.Error(ctx, "GetProgramServiceProvider BatchGet error", log.Err(err), log.Strings("programIDs", programIDs))
			return nil, err
		}

		for _, item := range programInfos {
			programMap[item.ID] = &entity.ScheduleShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}
	return programMap, nil
}

func (s *scheduleModel) getByIDFormDB(ctx context.Context, operator *entity.Operator, id string) (*entity.Schedule, error) {
	var schedule = new(entity.Schedule)
	err := da.GetScheduleDA().Get(ctx, id, schedule)
	if err == dbo.ErrRecordNotFound {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetByID error", log.Err(err), log.String("id", id))
		return nil, err
	}
	if schedule.DeleteAt != 0 {
		return nil, constant.ErrRecordNotFound
	}
	return schedule, nil
}

func (s *scheduleModel) processSingleSchedule(ctx context.Context, operator *entity.Operator, schedule *entity.Schedule) (*entity.ScheduleDetailsView, error) {
	realTimeData, err := s.getLessonPlanAuthed(ctx, operator, schedule.ID, schedule.LessonPlanID)
	if err != nil {
		log.Error(ctx, "GetByID using cache:GetScheduleRealTimeStatus error",
			log.Err(err),
			log.Any("operator", operator),
		)
		return nil, err
	}
	result := &entity.ScheduleDetailsView{
		ID:             schedule.ID,
		Title:          schedule.Title,
		OrgID:          schedule.OrgID,
		StartAt:        schedule.StartAt,
		EndAt:          schedule.EndAt,
		IsAllDay:       schedule.IsAllDay,
		ClassType:      schedule.ClassType,
		DueAt:          schedule.DueAt,
		Description:    schedule.Description,
		Version:        schedule.ScheduleVersion,
		IsRepeat:       schedule.RepeatID != "",
		Status:         schedule.Status,
		RealTimeStatus: *realTimeData,
		IsHomeFun:      schedule.IsHomeFun,
		IsHidden:       schedule.IsHidden,
	}

	// get role type
	roleType, err := GetScheduleRelationModel().GetRelationTypeByScheduleID(ctx, operator, schedule.ID)
	if err != nil {
		log.Error(ctx, "get relation type error", log.Any("op", operator), log.Any("schedule", schedule), log.Err(err))
		return nil, err
	}
	result.RoleType = roleType

	// verify is exist feedback
	existFeedback, err := GetScheduleFeedbackModel().ExistByScheduleID(ctx, operator, schedule.ID)
	if err != nil {
		log.Error(ctx, "exist by schedule id error", log.Any("op", operator), log.Any("schedule", schedule), log.Err(err))
		return nil, err
	}
	result.ExistFeedback = existFeedback

	if schedule.Attachment != "" {
		var attachment entity.ScheduleShortInfo
		err := json.Unmarshal([]byte(schedule.Attachment), &attachment)
		if err != nil {
			log.Error(ctx, "Unmarshal schedule.Attachment error", log.Err(err), log.String("schedule.Attachment", schedule.Attachment))
			return nil, err
		}
		result.Attachment = attachment
	}

	if schedule.RepeatJson != "" {
		var repeat entity.RepeatOptions
		err := json.Unmarshal([]byte(schedule.RepeatJson), &repeat)
		if err != nil {
			log.Error(ctx, "Unmarshal schedule.RepeatJson error", log.Err(err), log.String("schedule.RepeatJson", schedule.RepeatJson))
			return nil, err
		}
		result.Repeat = repeat
	}

	classID, err := GetScheduleRelationModel().GetClassRosterID(ctx, operator, schedule.ID)
	if err != nil {
		return nil, err
	}

	if classID != "" {
		classes, err := s.getClassInfoMapByClassIDs(ctx, operator, []string{classID})
		if err != nil {
			log.Error(ctx, "get class error", log.Err(err), log.String("classID", classID))
			return nil, err
		}
		if item, ok := classes[classID]; ok {
			result.Class = item
		}
	}

	if schedule.ProgramID != "" {
		programMap, err := s.getProgramsByIDs(ctx, operator, []string{schedule.ProgramID})
		if err != nil {
			log.Error(ctx, "get program info error", log.Err(err), log.String("ProgramID", schedule.ProgramID))
			return nil, err
		}
		if item, ok := programMap[schedule.ProgramID]; ok {
			result.Program = item
		}
	}

	result.Subjects, err = GetScheduleRelationModel().GetSubjects(ctx, operator, schedule.ID)
	if err != nil {
		return nil, err
	}

	if schedule.LessonPlanID != "" {
		result.LessonPlan, err = s.getLessonPlanWithMaterial(ctx, operator, schedule.LessonPlanID)
		if err != nil {
			log.Error(ctx, "get lesson plan with material error", log.Err(err), log.Any("schedule", schedule))
			return nil, err
		}
	}
	result.Status = result.Status.GetScheduleStatus(entity.ScheduleStatusInput{
		EndAt:     result.EndAt,
		DueAt:     result.DueAt,
		ClassType: result.ClassType,
	})
	return result, nil
}

func (s *scheduleModel) processUsersAccessible(ctx context.Context, operator *entity.Operator, data *entity.ScheduleDetailsView) (*entity.ScheduleDetailsView, error) {
	scheduleRelations, err := GetScheduleRelationModel().GetUsersByScheduleID(ctx, operator, data.ID)
	if err != nil {
		return nil, err
	}
	classRosterUserMap := make(map[string]*entity.ScheduleUserInput)
	classRosterUserIDs := make([]string, 0)
	partUserInput := make([]*entity.ScheduleUserInput, 0)
	for _, item := range scheduleRelations {
		switch item.RelationType {
		case entity.ScheduleRelationTypeClassRosterTeacher, entity.ScheduleRelationTypeClassRosterStudent:
			classRosterUserMap[item.RelationID] = &entity.ScheduleUserInput{
				ID:   item.RelationID,
				Type: item.RelationType,
			}
			classRosterUserIDs = append(classRosterUserIDs, item.RelationID)
		case entity.ScheduleRelationTypeParticipantTeacher, entity.ScheduleRelationTypeParticipantStudent:
			partUserInput = append(partUserInput, &entity.ScheduleUserInput{
				ID:   item.RelationID,
				Type: item.RelationType,
			})
		}
	}
	accessibleUserList := make([]*entity.ScheduleAccessibleUserView, 0)
	if data.Class != nil {
		isClassAccessible, err := s.AccessibleClass(ctx, operator, data.Class.ID)
		if err != nil {
			log.Error(ctx, "GetByID:AccessibleClassRosterUser error",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("data", data),
			)
			return nil, err
		}
		data.Class.Enable = isClassAccessible

		userInfos, err := external.GetUserServiceProvider().BatchGet(ctx, operator, classRosterUserIDs)
		if err != nil {
			return nil, err
		}
		for _, item := range userInfos {
			if !item.Valid {
				continue
			}
			if u, ok := classRosterUserMap[item.ID]; ok {
				temp := &entity.ScheduleAccessibleUserView{
					ID:     item.ID,
					Name:   item.Name,
					Type:   u.Type,
					Enable: isClassAccessible,
				}
				accessibleUserList = append(accessibleUserList, temp)
			}
		}
	}
	accessiblePart, err := s.AccessibleParticipantUser(ctx, operator, partUserInput)
	if err != nil {
		log.Error(ctx, "GetByID:AccessibleParticipantUser error",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("userInput", partUserInput),
		)
		return nil, err
	}
	accessibleUserList = append(accessibleUserList, accessiblePart...)
	for _, item := range accessibleUserList {
		temp := &entity.ScheduleAccessibleUserView{
			ID:     item.ID,
			Name:   item.Name,
			Type:   item.Type,
			Enable: item.Enable,
		}
		switch item.Type {
		case entity.ScheduleRelationTypeParticipantTeacher:
			data.ParticipantsTeachers = append(data.ParticipantsTeachers, temp)
		case entity.ScheduleRelationTypeParticipantStudent:
			data.ParticipantsStudents = append(data.ParticipantsStudents, temp)
		case entity.ScheduleRelationTypeClassRosterTeacher:
			data.ClassRosterTeachers = append(data.ClassRosterTeachers, temp)
		case entity.ScheduleRelationTypeClassRosterStudent:
			data.ClassRosterStudents = append(data.ClassRosterStudents, temp)
		}
	}
	data.Teachers = append(data.Teachers, data.ClassRosterTeachers...)
	data.Teachers = append(data.Teachers, data.ParticipantsTeachers...)
	return data, nil
}

func (s *scheduleModel) getByIDFormCache(ctx context.Context, operator *entity.Operator, id string) (*entity.ScheduleDetailsView, error) {
	cacheData, err := da.GetScheduleRedisDA().SearchToScheduleDetails(ctx, operator.OrgID, operator.UserID, id)
	if err != nil {
		return nil, err
	}
	cacheData.Status = cacheData.Status.GetScheduleStatus(entity.ScheduleStatusInput{
		EndAt:     cacheData.EndAt,
		DueAt:     cacheData.DueAt,
		ClassType: cacheData.ClassType,
	})
	if cacheData.LessonPlan != nil {
		lessonPlanAuthed, err := s.getLessonPlanAuthed(ctx, operator, id, cacheData.LessonPlan.ID)
		if err != nil {
			log.Error(ctx, "GetByID using cache:GetScheduleRealTimeStatus error",
				log.Err(err),
				log.Any("operator", operator),
			)
			return nil, err
		}
		cacheData.RealTimeStatus = *lessonPlanAuthed
	}
	return cacheData, nil
}

func (s *scheduleModel) GetByID(ctx context.Context, operator *entity.Operator, id string) (*entity.ScheduleDetailsView, error) {
	cacheData, err := s.getByIDFormCache(ctx, operator, id)
	if err == nil {
		log.Debug(ctx, "GetByID:using cache",
			log.Any("op", operator),
			log.Any("id", id),
		)
		return cacheData, nil
	}

	schedule, err := s.getByIDFormDB(ctx, operator, id)
	if err != nil {
		return nil, err
	}

	result, err := s.processSingleSchedule(ctx, operator, schedule)
	if err != nil {
		return nil, err
	}

	result, err = s.processUsersAccessible(ctx, operator, result)
	if err != nil {
		return nil, err
	}

	da.GetScheduleRedisDA().Add(ctx, operator.OrgID, &da.ScheduleCacheCondition{
		UserID:     operator.UserID,
		ScheduleID: id,
	}, result)

	return result, nil
}

func (s *scheduleModel) AccessibleClass(ctx context.Context, operator *entity.Operator, classID string) (bool, error) {
	permissionMap, err := GetSchedulePermissionModel().HasScheduleOrgPermissions(ctx, operator, []external.PermissionName{
		external.ScheduleCreateEvent,
		external.ScheduleCreateMySchoolEvent,
		external.ScheduleCreateMyEvent,
	})
	if err == constant.ErrForbidden {
		log.Info(ctx, "no permission to edit class", log.String("classID", classID), log.Any("operator", operator), log.Any("permissionMap", permissionMap))
		return false, nil
	}

	if err != nil {
		return false, err
	}
	classIDs := []string{classID}

	if permissionMap[external.ScheduleCreateEvent] {
		return true, nil
	}
	if permissionMap[external.ScheduleCreateMySchoolEvent] {
		operatorSchoolList, err := external.GetSchoolServiceProvider().GetByPermission(ctx, operator, external.ScheduleCreateMySchoolEvent)
		if err != nil {
			log.Error(ctx, "GetSchoolServiceProvider.GetByPermission error", log.Err(err))
			return false, err
		}
		opSchoolMap := make(map[string]bool)
		for _, item := range operatorSchoolList {
			opSchoolMap[item.ID] = true
		}

		userClassSchoolMap, err := external.GetSchoolServiceProvider().GetByClasses(ctx, operator, classIDs)
		if err != nil {
			log.Error(ctx, "GetSchoolServiceProvider.GetByClasses error", log.Err(err), log.Err(err), log.Strings("classIDs", classIDs))
			return false, err
		}
		flag := false
		classSchools := userClassSchoolMap[classID]
		for _, item := range classSchools {
			if opSchoolMap[item.ID] {
				flag = true
				break
			}
		}
		return flag, nil
	}
	if permissionMap[external.ScheduleCreateMyEvent] {
		opClasses, err := external.GetClassServiceProvider().GetByUserID(ctx, operator, operator.UserID)
		if err != nil {
			return false, err
		}
		flag := false
		for _, item := range opClasses {
			if item.ID == classID {
				flag = true
				break
			}
		}
		return flag, nil
	}
	return false, nil
}

func (s *scheduleModel) AccessibleParticipantUser(ctx context.Context, operator *entity.Operator, users []*entity.ScheduleUserInput) ([]*entity.ScheduleAccessibleUserView, error) {
	result := make([]*entity.ScheduleAccessibleUserView, 0)
	if len(users) <= 0 {
		return result, nil
	}
	userIDs := make([]string, len(users))
	usersMap := make(map[string]*entity.ScheduleUserInput)
	for i, item := range users {
		userIDs[i] = item.ID
		usersMap[item.ID] = item
	}

	userInfoList, err := external.GetUserServiceProvider().BatchGet(ctx, operator, userIDs)
	if err != nil {
		log.Error(ctx, "OperatorAccessibleUser:GetUserServiceProvider.BatchGet error",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("userIDs", userIDs),
		)
		return nil, err
	}
	for _, item := range userInfoList {
		if !item.Valid {
			log.Info(ctx, "user is valid", log.Any("user", item), log.Any("operator", operator))
			return nil, constant.ErrInvalidArgs
		}
		if user, ok := usersMap[item.ID]; ok {
			info := &entity.ScheduleAccessibleUserView{
				ID:     item.ID,
				Name:   item.Name,
				Type:   user.Type,
				Enable: false,
			}
			result = append(result, info)
		}
	}
	permissionMap, err := GetSchedulePermissionModel().HasScheduleOrgPermissions(ctx, operator, []external.PermissionName{
		external.ScheduleCreateEvent,
		external.ScheduleCreateMySchoolEvent,
		external.ScheduleCreateMyEvent,
	})
	if err == constant.ErrForbidden {
		log.Info(ctx, "no permission to edit participant user", log.Any("operator", operator), log.Any("permissionMap", permissionMap))
		return result, nil
	}

	userSchoolMap, err := external.GetSchoolServiceProvider().GetByUsers(ctx, operator, operator.OrgID, userIDs)
	if err != nil {
		return nil, err
	}

	// org permission
	if permissionMap[external.ScheduleCreateEvent] {
		for _, userInfo := range result {
			userInfo.Enable = true
		}
		return result, nil
	}
	if permissionMap[external.ScheduleCreateMySchoolEvent] {
		schoolList, err := external.GetSchoolServiceProvider().GetByPermission(ctx, operator, external.ScheduleCreateMySchoolEvent)
		if err != nil {
			log.Error(ctx, "GetSchoolServiceProvider.GetByPermission error", log.Err(err))
			return nil, err
		}
		operatorSchoolMap := make(map[string]bool)
		for _, item := range schoolList {
			operatorSchoolMap[item.ID] = true
		}
		for _, userInfo := range result {
			userSchoolInfos := userSchoolMap[userInfo.ID]
			for _, schoolInfo := range userSchoolInfos {
				if operatorSchoolMap[schoolInfo.ID] {
					userInfo.Enable = true
					break
				}
			}
		}
		return result, nil
	}
	if permissionMap[external.ScheduleCreateMyEvent] {
		return result, nil
	}
	return result, nil
}

func (s *scheduleModel) GetTeacherByName(ctx context.Context, operator *entity.Operator, orgID, name string) ([]*external.Teacher, error) {
	teacherService := external.GetTeacherServiceProvider()
	teachers, err := teacherService.Query(ctx, operator, orgID, name)
	if err != nil {
		log.Error(ctx, "querySchedule:query teacher info error", log.Err(err), log.String("name", name))
		return nil, err
	}

	return teachers, nil
}

func (s *scheduleModel) ExistScheduleByLessonPlanID(ctx context.Context, lessonPlanID string) (bool, error) {
	if strings.TrimSpace(lessonPlanID) == "" {
		log.Info(ctx, "lessonPlanID is empty", log.String("lessonPlanID", lessonPlanID))
		return false, errors.New("lessonPlanID is empty")
	}
	lessonPlanPastIDs, err := GetContentModel().GetPastContentIDByID(ctx, dbo.MustGetDB(ctx), lessonPlanID)
	if err != nil {
		logger.Error(ctx, "ExistScheduleByLessonPlanID:GetContentModel.GetPastContentIDByID error",
			log.Err(err),
			log.String("lessonPlanID", lessonPlanID),
		)
		return false, err
	}
	condition := da.ScheduleCondition{
		LessonPlanIDs: entity.NullStrings{
			Strings: lessonPlanPastIDs,
			Valid:   true,
		},
		EndAtGe: sql.NullInt64{
			Int64: time.Now().Unix(),
			Valid: true,
		},
	}
	count, err := da.GetScheduleDA().Count(ctx, condition, &entity.Schedule{})
	if err != nil {
		log.Error(ctx, "get schedule count by condition error", log.Err(err), log.Any("condition", condition))
		return false, err
	}

	return count > 0, nil
}

func (s *scheduleModel) ExistScheduleByID(ctx context.Context, id string) (bool, error) {
	condition := &da.ScheduleCondition{
		IDs: entity.NullStrings{
			Strings: []string{id},
			Valid:   true,
		},
	}
	count, err := da.GetScheduleDA().Count(ctx, condition, &entity.Schedule{})
	if err != nil {
		log.Error(ctx, "get schedule count by condition error", log.Err(err), log.Any("condition", condition))
		return false, err
	}

	return count > 0, nil
}

func (s *scheduleModel) GetPlainByID(ctx context.Context, id string) (*entity.SchedulePlain, error) {
	var schedule = new(entity.Schedule)
	err := da.GetScheduleDA().Get(ctx, id, schedule)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "GetPlainByID not found", log.Err(err), log.String("id", id))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetPlainByID error", log.Err(err), log.String("id", id))
		return nil, err
	}
	if schedule.DeleteAt != 0 {
		log.Error(ctx, "GetPlainByID deleted", log.Err(err), log.Any("schedule", schedule))
		return nil, constant.ErrRecordNotFound
	}
	result := new(entity.SchedulePlain)
	result.Schedule = schedule

	return result, nil
}

func (s *scheduleModel) verifyData(ctx context.Context, operator *entity.Operator, v *entity.ScheduleVerify) error {
	// class
	classService := external.GetClassServiceProvider()
	classInfos, err := classService.BatchGet(ctx, operator, []string{v.ClassID})
	if err != nil {
		log.Error(ctx, "verifyData:GetClassServiceProvider BatchGet error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}
	for _, item := range classInfos {
		if item == nil {
			log.Error(ctx, "verifyData:GetClassServiceProvider class info not found", log.Any("ScheduleVerify", v))
			return constant.ErrRecordNotFound
		}
	}

	if v.ClassType == entity.ScheduleClassTypeTask {
		return nil
	}
	// subject
	if len(v.SubjectIDs) != 0 {
		_, err = external.GetSubjectServiceProvider().BatchGet(ctx, operator, v.SubjectIDs)
		if err != nil {
			log.Error(ctx, "verifyData:GetSubjectServiceProvider BatchGet error", log.Err(err), log.Any("ScheduleVerify", v))
			return err
		}
	}

	// program
	if v.ProgramID == "" {
		log.Info(ctx, "programID is required", log.Any("op", operator), log.Any("input", v))
		return constant.ErrInvalidArgs
	}
	programIDs := []string{v.ProgramID}
	_, err = external.GetProgramServiceProvider().BatchGet(ctx, operator, programIDs)
	if err != nil {
		log.Error(ctx, "verifyData:GetProgramServiceProvider BatchGet error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}
	if v.ClassType == entity.ScheduleClassTypeHomework && v.IsHomeFun {
		return nil
	}
	// verify lessPlan type
	lessonPlanInfo, err := GetContentModel().GetContentNameByID(ctx, dbo.MustGetDB(ctx), v.LessonPlanID)
	if err != nil {
		log.Error(ctx, "verifyData:get lessPlan info error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}
	if lessonPlanInfo.ContentType != entity.ContentTypePlan {
		log.Error(ctx, "verifyData:content type is not lesson", log.Any("lessonPlanInfo", lessonPlanInfo), log.Any("ScheduleVerify", v))
		return constant.ErrInvalidArgs
	}
	// verify lessPlan is valid
	_, err = s.VerifyLessonPlanAuthed(ctx, operator, v.LessonPlanID)
	return err
}

func (s *scheduleModel) VerifyLessonPlanAuthed(ctx context.Context, operator *entity.Operator, lessonPlanID string) (bool, error) {
	// verify lessPlan is valid
	contentMap, err := GetAuthedContentRecordsModel().GetContentAuthByIDList(ctx, []string{lessonPlanID}, operator)
	if err != nil {
		log.Error(ctx, "GetAuthedContentRecordsModel.GetContentAuthByIDList error", log.Err(err),
			log.String("lessonPlanID", lessonPlanID),
			log.Any("operator", operator),
		)
		return false, err
	}
	item, ok := contentMap[lessonPlanID]
	if !ok {
		log.Error(ctx, "lesson plan not found", log.Any("operator", operator), log.String("lessonPlanID", lessonPlanID))
		return false, ErrScheduleLessonPlanUnAuthed
	}
	if item == entity.ContentUnauthed {
		log.Info(ctx, "lesson plan unAuthed", log.Any("operator", operator),
			log.Any("lessonInfo", item),
			log.String("lessonPlanID", lessonPlanID),
		)
		return false, ErrScheduleLessonPlanUnAuthed
	}

	return true, nil
}

func (s *scheduleModel) UpdateScheduleStatus(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string, status entity.ScheduleStatus) error {
	var schedule = new(entity.Schedule)
	err := da.GetScheduleDA().GetTx(ctx, tx, id, schedule)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "UpdateScheduleStatus: get schedule by id failed, schedule not found", log.Err(err), log.String("id", id))
		return constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "UpdateScheduleStatus: get schedule by id failed",
			log.Err(err),
			log.String("id", id),
		)
		return err
	}
	if schedule.DeleteAt != 0 {
		log.Error(ctx, "UpdateScheduleStatus: get schedule by id failed, schedule not found", log.String("id", id))
		return constant.ErrRecordNotFound
	}

	schedule.Status = status
	_, err = da.GetScheduleDA().UpdateTx(ctx, tx, schedule)
	if err != nil {
		log.Error(ctx, "UpdateScheduleStatus: update schedule status error",
			log.String("id", id),
			log.Any("schedule", schedule),
			log.Err(err),
		)
		return err
	}
	//err = da.GetScheduleRedisDA().Clean(ctx, operator, []string{id})
	//if err != nil {
	//	log.Info(ctx, "UpdateScheduleStatus:GetScheduleRedisDA.Clean error", log.Err(err))
	//}
	return nil
}

func (s *scheduleModel) GetLessonPlanByCondition(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, condition *da.ScheduleCondition) ([]*entity.ScheduleShortInfo, error) {
	lessonPlanIDs, err := da.GetScheduleDA().GetLessonPlanIDsByCondition(ctx, tx, condition)
	if err != nil {
		logger.Error(ctx, "GetLessonPlanByCondition:get lessonPlanIDs error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("operator", operator),
		)
		return nil, err
	}
	latestIDs, err := GetContentModel().GetLatestContentIDByIDList(ctx, tx, lessonPlanIDs)
	if err != nil {
		logger.Error(ctx, "GetLessonPlanByCondition:get latest lessonPlanIDs error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("operator", operator),
			log.Strings("lessonPlanIDs", lessonPlanIDs),
			log.Strings("latestIDs", latestIDs),
		)
		return nil, err
	}
	latestIDs = utils.SliceDeduplication(latestIDs)
	lessonPlanInfos, err := GetContentModel().GetContentNameByIDList(ctx, tx, latestIDs)
	if err != nil {
		logger.Error(ctx, "GetLessonPlanByCondition:get lessonPlan info error",
			log.Err(err),
			log.Strings("lessonPlanIDs", lessonPlanIDs),
			log.Strings("latestIDs", latestIDs),
			log.Any("condition", condition),
			log.Any("operator", operator),
		)
	}
	result := make([]*entity.ScheduleShortInfo, len(lessonPlanInfos))
	for i, item := range lessonPlanInfos {
		result[i] = &entity.ScheduleShortInfo{
			ID:   item.ID,
			Name: item.Name,
		}
	}
	return result, nil
}

func (s *scheduleModel) GetScheduleIDsByCondition(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, condition *entity.ScheduleIDsCondition) ([]string, error) {
	lessonPlanPastIDs, err := GetContentModel().GetPastContentIDByID(ctx, tx, condition.LessonPlanID)
	if err != nil {
		logger.Error(ctx, "GetScheduleIDsByCondition:get past lessonPlan id error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("operator", operator),
		)
		return nil, err
	}
	daCondition := &da.ScheduleCondition{
		RelationID: sql.NullString{
			String: condition.ClassID,
			Valid:  true,
		},
		LessonPlanIDs: entity.NullStrings{
			Strings: lessonPlanPastIDs,
			Valid:   true,
		},
		StartAtLt: sql.NullInt64{
			Int64: condition.StartAt,
			Valid: true,
		},
	}
	var scheduleList []*entity.Schedule
	err = da.GetScheduleDA().Query(ctx, daCondition, &scheduleList)
	if err != nil {
		log.Error(ctx, "schedule query error", log.Err(err), log.Any("daCondition", daCondition))
		return nil, err
	}
	var result = make([]string, len(scheduleList))
	for i, item := range scheduleList {
		result[i] = item.ID
	}
	return result, nil
}

func (s *scheduleModel) GetScheduleIDsByOrgID(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, orgID string) ([]string, error) {
	condition := &da.ScheduleCondition{
		OrgID: sql.NullString{
			String: orgID,
			Valid:  true,
		},
	}
	var scheduleList []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "GetScheduleIDsByOrgID:GetScheduleDA.Query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	var result = make([]string, len(scheduleList))
	for i, item := range scheduleList {
		result[i] = item.ID
	}
	return result, nil
}

func (s *scheduleModel) StartScheduleRepeat(ctx context.Context, template *entity.Schedule, options *entity.RepeatOptions, location *time.Location) ([]*entity.Schedule, error) {
	if options == nil || !options.Type.Valid() {
		return []*entity.Schedule{template}, nil
	}

	cfg := NewRepeatConfig(options, location)
	plan, err := NewRepeatCyclePlan(ctx, template.StartAt, template.EndAt, cfg)
	if err != nil {
		log.Error(ctx, "StartScheduleRepeat:NewRepeatCyclePlan error", log.Err(err), log.Any("template", template), log.Any("cfg", cfg))
		return nil, err
	}
	endRule, err := NewEndRepeatCycleRule(options)
	if err != nil {
		log.Error(ctx, "StartScheduleRepeat:NewEndRepeatCycleRule error", log.Err(err), log.Any("template", template), log.Any("options", options))
		return nil, err
	}
	planResult, err := plan.GenerateTimeByEndRule(endRule)
	if err != nil {
		log.Error(ctx, "StartScheduleRepeat:GenerateTimeByEndRule error", log.Err(err),
			log.Any("template", template),
			log.Any("plan", plan),
			log.Any("endRule", endRule),
		)
		return nil, err
	}
	result := make([]*entity.Schedule, len(planResult))
	for i, item := range planResult {
		temp := template.Clone()
		temp.StartAt = item.Start
		temp.EndAt = item.End
		temp.ID = utils.NewID()
		result[i] = &temp
	}
	return result, nil
}

func (s *scheduleModel) QueryScheduledDates(ctx context.Context, op *entity.Operator, condition *da.ScheduleCondition, loc *time.Location) ([]string, error) {
	cacheData, err := da.GetScheduleRedisDA().SearchToStrings(ctx, op.OrgID, condition)
	if err == nil && len(cacheData) > 0 {
		log.Debug(ctx, "Query:using cache",
			log.Any("condition", condition),
			log.Any("cacheData", cacheData),
		)
		return cacheData, nil
	}
	var scheduleList []*entity.Schedule
	err = da.GetScheduleDA().Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "GetHasScheduleDate:GetScheduleDA.Query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	dateList := make([]string, 0)
	for _, item := range scheduleList {
		if item.ClassType == entity.ScheduleClassTypeHomework && item.DueAt <= 0 {
			continue
		}
		betweenTimes := utils.DateBetweenTimeAndFormat(item.StartAt, item.EndAt, loc)
		dateList = append(dateList, betweenTimes...)
	}
	result := utils.SliceDeduplication(dateList)

	err = da.GetScheduleRedisDA().Add(ctx, op.OrgID, &da.ScheduleCacheCondition{Condition: condition}, result)
	if err != nil {
		log.Info(ctx, "QueryScheduledDates:GetScheduleRedisDA.AddDates error", log.Err(err))
	}

	return result, nil
}

func (s *scheduleModel) getLessonPlanAuthed(ctx context.Context, op *entity.Operator, scheduleID string, lessonPlanID string) (*entity.ScheduleRealTimeView, error) {
	result := new(entity.ScheduleRealTimeView)
	result.ID = scheduleID
	if lessonPlanID == "" {
		return result, nil
	}

	// lesson plan real time info
	result.LessonPlanIsAuth, _ = s.VerifyLessonPlanAuthed(ctx, op, lessonPlanID)

	return result, nil
}

func (s *scheduleModel) GetByIDs(ctx context.Context, op *entity.Operator, ids []string) ([]*entity.SchedulePlain, error) {
	var scheduleList []*entity.Schedule
	condition := &da.ScheduleCondition{
		IDs: entity.NullStrings{
			Strings: ids,
			Valid:   true,
		},
	}
	err := da.GetScheduleDA().Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "get by ids error", log.Strings("ids", ids))
		return nil, err
	}
	result := make([]*entity.SchedulePlain, len(scheduleList))
	for i, item := range scheduleList {
		result[i] = &entity.SchedulePlain{Schedule: item}
	}
	return result, nil
}

func (s *scheduleModel) getRelationCondition(ctx context.Context, op *entity.Operator) (*da.ScheduleCondition, error) {
	permissionMap, err := GetSchedulePermissionModel().HasScheduleOrgPermissions(ctx, op, []external.PermissionName{
		external.ScheduleViewOrgCalendar,
		external.ScheduleViewSchoolCalendar,
		external.ScheduleViewMyCalendar,
	})
	if err != nil {
		return nil, err
	}

	condition := &da.ScheduleCondition{
		OrgID: sql.NullString{
			String: op.OrgID,
			Valid:  true,
		},
	}

	if permissionMap[external.ScheduleViewOrgCalendar] {
		return condition, nil
	}
	if permissionMap[external.ScheduleViewSchoolCalendar] {
		schoolList, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, external.ScheduleViewSchoolCalendar)
		if err != nil {
			log.Error(ctx, "GetSchoolServiceProvider.GetByPermission error",
				log.Err(err),
				log.Any("op", op),
				log.String("permission", external.ScheduleViewSchoolCalendar.String()),
			)
			return nil, constant.ErrInternalServer
		}
		relationIDs := make([]string, len(schoolList))
		for i, item := range schoolList {
			relationIDs[i] = item.ID
		}

		condition.RelationIDs = entity.NullStrings{
			Strings: relationIDs,
			Valid:   true,
		}
		return condition, nil
	}
	if permissionMap[external.ScheduleViewMyCalendar] {
		condition.RelationIDs = entity.NullStrings{
			Strings: []string{op.UserID},
			Valid:   true,
		}
		return condition, nil
	}
	return condition, nil
}

func (s *scheduleModel) getProgramCondition(ctx context.Context, op *entity.Operator) (*da.ScheduleCondition, error) {
	return s.getRelationCondition(ctx, op)
}

func (s *scheduleModel) GetPrograms(ctx context.Context, op *entity.Operator) ([]*entity.ScheduleShortInfo, error) {
	condition, err := s.getProgramCondition(ctx, op)
	if err != nil {
		return nil, err
	}
	dbProgramIDs, err := da.GetScheduleDA().GetPrograms(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		log.Error(ctx, "get program ids from db error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	amsPrograms, err := external.GetProgramServiceProvider().GetByOrganization(ctx, op)
	if err != nil {
		log.Error(ctx, "get program from ams error", log.Err(err), log.Any("op", op))
		return nil, err
	}

	amsProgramMap := make(map[string]*external.Program, len(amsPrograms))
	for _, amsProgram := range amsPrograms {
		amsProgramMap[amsProgram.ID] = amsProgram
	}

	result := make([]*entity.ScheduleShortInfo, 0, len(amsPrograms))
	for _, sID := range dbProgramIDs {
		if program, ok := amsProgramMap[sID]; ok {
			result = append(result, &entity.ScheduleShortInfo{
				ID:   program.ID,
				Name: program.Name,
			})
		}
	}

	return result, nil
}

func (s *scheduleModel) getSubjectCondition(ctx context.Context, op *entity.Operator, programID string) (*da.ScheduleCondition, error) {
	condition, err := s.getRelationCondition(ctx, op)
	if err != nil {
		return nil, err
	}
	condition.ProgramIDs = entity.NullStrings{
		Strings: []string{programID},
		Valid:   true,
	}
	return condition, nil
}

func (s *scheduleModel) GetSubjects(ctx context.Context, op *entity.Operator, programID string) ([]*entity.ScheduleShortInfo, error) {
	condition, err := s.getSubjectCondition(ctx, op, programID)
	if err != nil {
		return nil, err
	}
	dbSubjectIDs, err := da.GetScheduleDA().GetSubjects(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		log.Error(ctx, "get subject ids from db error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	amsSubjects, err := external.GetSubjectServiceProvider().GetByProgram(ctx, op, programID)
	if err != nil {
		log.Error(ctx, "get subject ids by program id from ams error", log.Err(err), log.Any("op", op), log.Any("programID", programID))
		return nil, err
	}

	amsSubjectMap := make(map[string]*external.Subject, len(amsSubjects))
	for _, amsSubject := range amsSubjects {
		amsSubjectMap[amsSubject.ID] = amsSubject
	}

	result := make([]*entity.ScheduleShortInfo, 0, len(amsSubjects))
	for _, sID := range dbSubjectIDs {
		if subject, ok := amsSubjectMap[sID]; ok {
			result = append(result, &entity.ScheduleShortInfo{
				ID:   subject.ID,
				Name: subject.Name,
			})
		}
	}
	return result, nil
}

func (s *scheduleModel) getClassTypesCondition(ctx context.Context, op *entity.Operator) (*da.ScheduleCondition, error) {
	return s.getRelationCondition(ctx, op)
}

func (s *scheduleModel) GetClassTypes(ctx context.Context, op *entity.Operator) ([]*entity.ScheduleShortInfo, error) {
	condition, err := s.getClassTypesCondition(ctx, op)
	if err != nil {
		return nil, err
	}
	classTypes, err := da.GetScheduleDA().GetClassTypes(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.ScheduleShortInfo, len(classTypes))
	for i, item := range classTypes {
		name := entity.ScheduleClassType(item)

		result[i] = &entity.ScheduleShortInfo{
			ID:   item,
			Name: name.ToLabel().String(),
		}
	}

	return result, nil
}

func (s *scheduleModel) GetRosterClassNotStartScheduleIDs(ctx context.Context, rosterClassID string, userIDs []string) ([]string, error) {
	condition := da.NewNotStartScheduleCondition(rosterClassID, userIDs)

	var scheduleList []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "schedule query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	var result = make([]string, len(scheduleList))
	for i, item := range scheduleList {
		result[i] = item.ID
	}

	return result, nil
}

func (s *scheduleModel) GetScheduleViewByID(ctx context.Context, op *entity.Operator, id string) (*entity.SchedulePopup, error) {
	schedule, err := s.getByIDFormDB(ctx, op, id)
	if err != nil {
		log.Error(ctx, "get by id from db error", log.Any("op", op), log.String("id", id))
		return nil, err
	}

	result := &entity.SchedulePopup{
		ID:        schedule.ID,
		Title:     schedule.Title,
		StartAt:   schedule.StartAt,
		EndAt:     schedule.EndAt,
		ClassType: schedule.ClassType,
		DueAt:     schedule.DueAt,
		Status:    schedule.Status,
		IsHomeFun: schedule.IsHomeFun,
		IsHidden:  schedule.IsHidden,
		RoomID:    schedule.ID,
	}

	// get role type
	roleType, err := GetScheduleRelationModel().GetRelationTypeByScheduleID(ctx, op, schedule.ID)
	if err != nil {
		log.Error(ctx, "get relation type error", log.Any("op", op), log.Any("schedule", schedule), log.Err(err))
		return nil, err
	}
	result.RoleType = roleType

	// verify is exist feedback
	existFeedback, err := GetScheduleFeedbackModel().ExistByScheduleID(ctx, op, schedule.ID)
	if err != nil {
		log.Error(ctx, "exist by schedule id error", log.Any("op", op), log.Any("schedule", schedule), log.Err(err))
		return nil, err
	}
	result.ExistFeedback = existFeedback

	if schedule.Attachment != "" {
		var attachment entity.ScheduleShortInfo
		err := json.Unmarshal([]byte(schedule.Attachment), &attachment)
		if err != nil {
			log.Error(ctx, "Unmarshal schedule.Attachment error", log.Err(err), log.String("schedule.Attachment", schedule.Attachment))
			return nil, err
		}
		result.Attachment = attachment
	}

	classID, err := GetScheduleRelationModel().GetClassRosterID(ctx, op, schedule.ID)
	if err != nil {
		return nil, err
	}

	if classID != "" {
		classes, err := s.getClassInfoMapByClassIDs(ctx, op, []string{classID})
		if err != nil {
			log.Error(ctx, "get class error", log.Err(err), log.String("classID", classID))
			return nil, err
		}
		if item, ok := classes[classID]; ok {
			result.Class = &entity.ScheduleShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}

	if schedule.LessonPlanID != "" {
		result.LessonPlan, err = s.getLessonPlanWithMaterial(ctx, op, schedule.LessonPlanID)
		if err != nil {
			log.Error(ctx, "get lesson plan with material error", log.Err(err), log.Any("schedule", schedule))
			return nil, err
		}
	}
	result.Status = result.Status.GetScheduleStatus(entity.ScheduleStatusInput{
		EndAt:     result.EndAt,
		DueAt:     result.DueAt,
		ClassType: result.ClassType,
	})

	users, err := GetScheduleRelationModel().GetUsers(ctx, op, schedule.ID)
	if err != nil {
		return nil, err
	}
	result.Teachers = users.Teachers
	result.Students = users.Students

	return result, nil
}

var (
	_scheduleOnce  sync.Once
	_scheduleModel IScheduleModel
)

func GetScheduleModel() IScheduleModel {
	_scheduleOnce.Do(func() {
		_scheduleModel = &scheduleModel{}
	})
	return _scheduleModel
}
