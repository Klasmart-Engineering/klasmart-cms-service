package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mq"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrScheduleEditMissTime         = errors.New("editable time has expired")
	ErrScheduleLessonPlanUnAuthed   = errors.New("schedule content data unAuthed")
	ErrScheduleEditMissTimeForDueAt = errors.New("editable time has expired for due at")
	ErrScheduleAlreadyHidden        = errors.New("schedule already hidden")
	ErrScheduleAlreadyFeedback      = errors.New("students already submitted feedback")
	ErrScheduleStudyAlreadyProgress = errors.New("students already started")
)

type IScheduleModel interface {
	// schedule operation
	Add(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleAddView) ([]*entity.Schedule, error)
	Update(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleUpdateView) ([]*entity.Schedule, error)
	Delete(ctx context.Context, op *entity.Operator, id string, editType entity.ScheduleEditType) error

	QueryByCondition(ctx context.Context, operator *entity.Operator, condition *da.ScheduleCondition, loc *time.Location) ([]*entity.ScheduleListView, error)
	QueryScheduledDatesByCondition(ctx context.Context, operator *entity.Operator, condition *da.ScheduleCondition, loc *time.Location) ([]string, error)
	Page(ctx context.Context, operator *entity.Operator, condition *da.ScheduleCondition) (int, []*entity.ScheduleSearchView, error)

	// Excluding deleted
	GetByID(ctx context.Context, operator *entity.Operator, id string) (*entity.ScheduleDetailsView, error)
	GetScheduleViewByID(ctx context.Context, op *entity.Operator, id string) (*entity.ScheduleViewDetail, error)

	ConflictDetection(ctx context.Context, op *entity.Operator, input *entity.ScheduleConflictInput) (*entity.ScheduleConflictView, error)

	ExistScheduleByLessonPlanID(ctx context.Context, lessonPlanID string) (bool, error)
	ExistScheduleByID(ctx context.Context, id string) (bool, error)
	GetPlainByID(ctx context.Context, id string) (*entity.SchedulePlain, error)
	UpdateScheduleStatus(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string, status entity.ScheduleStatus) error

	// todo queryScheduleIDs
	GetScheduleIDsByCondition(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, condition *entity.ScheduleIDsCondition) ([]string, error)
	GetScheduleIDsByOrgID(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, orgID string) ([]string, error)
	// todo queryScheduleIDs
	GetRosterClassNotStartScheduleIDs(ctx context.Context, rosterClassID string, userIDs []string) ([]string, error)
	// move to other
	VerifyLessonPlanAuthed(ctx context.Context, operator *entity.Operator, lessonPlanID string) (bool, error)
	UpdateScheduleShowOption(ctx context.Context, op *entity.Operator, scheduleID string, option entity.ScheduleShowOption) (string, error)
	// schedule_filter
	GetPrograms(ctx context.Context, op *entity.Operator) ([]*entity.ScheduleShortInfo, error)
	GetSubjects(ctx context.Context, op *entity.Operator, programID string) ([]*entity.ScheduleShortInfo, error)
	GetLearningOutcomeIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string][]string, error)

	GetSubjectsBySubjectIDs(ctx context.Context, op *entity.Operator, subjectIDs []string) (map[string]*entity.ScheduleShortInfo, error)
	GetVariableDataByIDs(ctx context.Context, op *entity.Operator, ids []string, include *entity.ScheduleInclude) ([]*entity.ScheduleVariable, error)
	GetTeachingLoad(ctx context.Context, input *entity.ScheduleTeachingLoadInput) ([]*entity.ScheduleTeachingLoadView, error)
	//prepareScheduleTimeViewCondition(ctx context.Context, query *entity.ScheduleTimeViewQuery, op *entity.Operator, loc *time.Location) (*da.ScheduleCondition, error)
	// query with user permission
	Query(ctx context.Context, query *entity.ScheduleTimeViewQuery, op *entity.Operator, loc *time.Location) ([]*entity.ScheduleListView, error)
	QueryScheduledDates(ctx context.Context, query *entity.ScheduleTimeViewQuery, op *entity.Operator, loc *time.Location) ([]string, error)

	// without permission check, internal function call
	QueryUnsafe(ctx context.Context, condition *entity.ScheduleQueryCondition) ([]*entity.Schedule, error)
	QueryScheduleTimeView(ctx context.Context, query *entity.ScheduleTimeViewListRequest, op *entity.Operator, loc *time.Location) (int, []*entity.ScheduleTimeView, error)

	QueryByConditionInternal(ctx context.Context, condition *da.ScheduleCondition) (int, []*entity.ScheduleSimplified, error)
}

type scheduleModel struct {
	scheduleDA         da.IScheduleDA
	scheduleRelationDA da.IScheduleRelationDA

	userService    external.UserServiceProvider
	schoolService  external.SchoolServiceProvider
	classService   external.ClassServiceProvider
	programService external.ProgramServiceProvider
	subjectService external.SubjectServiceProvider
	teacherService external.TeacherServiceProvider
}

func (s *scheduleModel) Add(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleAddView) ([]*entity.Schedule, error) {
	// todo move to api
	viewData.SubjectIDs = utils.SliceDeduplicationExcludeEmpty(viewData.SubjectIDs)
	// verify data
	err := s.verifyData(ctx, op, &entity.ScheduleVerifyInput{
		ClassID:      viewData.ClassID,
		SubjectIDs:   viewData.SubjectIDs,
		ProgramID:    viewData.ProgramID,
		LessonPlanID: viewData.LessonPlanID,
		ClassType:    viewData.ClassType,
		IsHomeFun:    viewData.IsHomeFun,
		OutcomeIDs:   viewData.OutcomeIDs,
	})
	if err != nil {
		log.Error(ctx, "add schedule: verify data error",
			log.Err(err),
			log.Any("viewData", viewData))
		return nil, err
	}
	if viewData.ClassType == entity.ScheduleClassTypeTask {
		viewData.LessonPlanID = ""
		viewData.ProgramID = ""
		viewData.SubjectIDs = nil
	}

	schedule, err := viewData.ToSchedule(ctx)

	relationInput := &entity.ScheduleRelationInput{
		ClassRosterClassID:     viewData.ClassID,
		ClassRosterTeacherIDs:  viewData.ClassRosterTeacherIDs,
		ClassRosterStudentIDs:  viewData.ClassRosterStudentIDs,
		ParticipantsTeacherIDs: viewData.ParticipantsTeacherIDs,
		ParticipantsStudentIDs: viewData.ParticipantsStudentIDs,
		SubjectIDs:             viewData.SubjectIDs,
	}

	// homefun study can bind learning outcome
	if viewData.ClassType == entity.ScheduleClassTypeHomework && viewData.IsHomeFun {
		relationInput.LearningOutcomeIDs = viewData.OutcomeIDs
	}

	relations, err := s.prepareScheduleRelationAddData(ctx, op, relationInput)
	if err != nil {
		log.Error(ctx, "prepareScheduleRelationAddData error", log.Err(err), log.Any("op", op), log.Any("relationInput", relationInput))
		return nil, err
	}

	scheduleList, allRelations, batchAddAssessmentSuperArgs, err := s.prepareScheduleAddData(ctx, op, schedule, &viewData.Repeat, viewData.Location, relations)
	if err != nil {
		log.Error(ctx, "prepareScheduleAddData: error",
			log.Err(err),
			log.Any("schedule", schedule),
			log.Any("option", &viewData.Repeat),
			log.Any("location", viewData.Location),
			log.Any("relations", relations))
		return nil, err
	}

	result, err := dbo.GetTransResult(ctx, func(ctx context.Context, tx *dbo.DBContext) (interface{}, error) {
		result, err := s.addSchedule(ctx, tx, op, scheduleList, allRelations, batchAddAssessmentSuperArgs)
		if err != nil {
			log.Error(ctx, "add schedule: error",
				log.Err(err),
				log.Any("scheduleList", scheduleList),
				log.Any("allRelations", allRelations),
				log.Any("batchAddAssessmentSuperArgs", batchAddAssessmentSuperArgs))
			return nil, err
		}
		return result, nil
	})

	err = da.GetScheduleRedisDA().Clean(ctx, op.OrgID)
	if err != nil {
		log.Warn(ctx, "clean schedule cache error", log.String("orgID", op.OrgID), log.Err(err))
	}

	go removeResourceMetadata(ctx, viewData.Attachment.ID)

	return result.([]*entity.Schedule), nil
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
	accessiblePartUser, err := s.accessibleParticipantUser(ctx, op, partUsers)
	if err != nil {
		log.Error(ctx, "s.accessibleParticipantUser error", log.Err(err), log.Any("input", input), log.Any("op", op))
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
			{
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
	scheduleRelations := make([]*entity.ScheduleRelation, 0)

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

	// learning outcome relation
	for _, outcomeID := range input.LearningOutcomeIDs {
		scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
			RelationID:   outcomeID,
			RelationType: entity.ScheduleRelationTypeLearningOutcome,
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

	partUserAccessible, err := s.accessibleParticipantUser(ctx, op, oldPartUsers)
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

func (s *scheduleModel) prepareScheduleAddData(ctx context.Context, op *entity.Operator, schedule *entity.Schedule, options *entity.RepeatOptions, location *time.Location, relations []*entity.ScheduleRelation) ([]*entity.Schedule, []*entity.ScheduleRelation, *entity.BatchAddAssessmentSuperArgs, error) {
	scheduleList, err := s.StartScheduleRepeat(ctx, schedule, options, location)
	if err != nil {
		log.Error(ctx, "schedule repeat error",
			log.Err(err),
			log.Any("schedule", schedule),
			log.Any("options", options),
			log.Any("location", location))
		return nil, nil, nil, err
	}
	if len(scheduleList) <= 0 {
		log.Error(ctx, "schedules prepareScheduleAddData error,schedules is empty",
			log.Any("schedule", schedule),
			log.Any("options", options))
		return nil, nil, nil, constant.ErrRecordNotFound
	}

	// add schedules relation
	allRelations := make([]*entity.ScheduleRelation, 0, len(scheduleList)*len(relations))
	userRelations := make(map[string][]*entity.ScheduleRelation, len(scheduleList))
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

			switch relation.RelationType {
			case entity.ScheduleRelationTypeClassRosterTeacher,
				entity.ScheduleRelationTypeClassRosterStudent,
				entity.ScheduleRelationTypeParticipantTeacher,
				entity.ScheduleRelationTypeParticipantStudent:
				userRelations[item.ID] = append(userRelations[item.ID], relation)
			}
		}
	}

	var batchAddAssessmentSuperArgs *entity.BatchAddAssessmentSuperArgs
	// add assessment
	if schedule.ClassType == entity.ScheduleClassTypeHomework && !schedule.IsHomeFun {
		studyInput := make([]*entity.AssessmentAddInputWhenCreateSchedule, len(scheduleList))

		for i, item := range scheduleList {
			assessmentType, err := entity.GetAssessmentTypeByScheduleType(ctx, entity.GetAssessmentTypeInput{
				ScheduleType: item.ClassType,
				IsHomeFun:    item.IsHomeFun,
			})
			if err != nil {
				return nil, nil, nil, err
			}

			studyInput[i] = &entity.AssessmentAddInputWhenCreateSchedule{
				ScheduleID:     item.ID,
				ScheduleTitle:  item.Title,
				LessonPlanID:   item.LessonPlanID,
				ClassID:        item.ClassID,
				AssessmentType: assessmentType,
				Attendances:    userRelations[item.ID],
			}
		}

		batchAddAssessmentSuperArgs, err = GetAssessmentModel().PrepareAddInputWhenCreateSchedule(ctx, op, studyInput)
		if err != nil {
			log.Error(ctx, "add schedule assessment error", log.Err(err), log.Any("studyInput", allRelations))
			return nil, nil, nil, err
		}
	}

	return scheduleList, allRelations, batchAddAssessmentSuperArgs, nil
}

func (s *scheduleModel) prepareScheduleUpdateData(ctx context.Context, op *entity.Operator, schedule *entity.Schedule, viewData *entity.ScheduleUpdateView) (*entity.Schedule, *entity.RepeatOptions, error) {
	newSchedule := *schedule
	// add schedule,update old schedule fields that need to be updated
	newSchedule.ID = utils.NewID()
	newSchedule.LessonPlanID = viewData.LessonPlanID
	newSchedule.ProgramID = viewData.ProgramID
	newSchedule.ClassID = viewData.ClassID
	newSchedule.StartAt = viewData.StartAt
	newSchedule.EndAt = viewData.EndAt
	newSchedule.Title = viewData.Title
	newSchedule.IsAllDay = viewData.IsAllDay
	newSchedule.Description = viewData.Description
	newSchedule.DueAt = viewData.DueAt
	newSchedule.ClassType = viewData.ClassType
	newSchedule.CreatedID = op.UserID
	newSchedule.CreatedAt = time.Now().Unix()
	newSchedule.UpdatedID = op.UserID
	newSchedule.UpdatedAt = time.Now().Unix()
	newSchedule.DeletedID = ""
	newSchedule.DeleteAt = 0
	newSchedule.IsHomeFun = viewData.IsHomeFun
	if viewData.ClassType != entity.ScheduleClassTypeHomework {
		newSchedule.IsHomeFun = false
	}
	// attachment
	b, err := json.Marshal(viewData.Attachment)
	if err != nil {
		log.Warn(ctx, "update schedule:marshal attachment error",
			log.Any("attachment", viewData.Attachment))
		return nil, nil, err
	}
	newSchedule.Attachment = string(b)

	// update repeat rule
	var repeatOptions *entity.RepeatOptions
	// if repeat selected, use repeat rule
	if viewData.IsRepeat {
		b, err := json.Marshal(viewData.Repeat)
		if err != nil {
			return nil, nil, err
		}
		newSchedule.RepeatJson = string(b)
		// if following selected, set repeat rule
		if viewData.EditType == entity.ScheduleEditWithFollowing {
			repeatOptions = &viewData.Repeat
		}
		if newSchedule.RepeatID == "" {
			newSchedule.RepeatID = utils.NewID()
		}
	} else {
		// if repeat not selected,but need to update follow schedule, use old schedule repeat rule
		if viewData.EditType == entity.ScheduleEditWithFollowing {
			var repeat = new(entity.RepeatOptions)
			if err := json.Unmarshal([]byte(newSchedule.RepeatJson), repeat); err != nil {
				log.Error(ctx, "update schedule:unmarshal schedule repeatJson error",
					log.Err(err),
					log.Any("viewData", viewData),
					log.Any("schedule", newSchedule),
				)
				return nil, nil, err
			}
			repeatOptions = repeat
		}
	}

	return &newSchedule, repeatOptions, nil
}

// finished
func (s *scheduleModel) addSchedule(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, scheduleList []*entity.Schedule, scheduleRelations []*entity.ScheduleRelation, batchAddAssessmentSuperArgs *entity.BatchAddAssessmentSuperArgs) ([]*entity.Schedule, error) {
	// insert into `schedules` table
	result, err := s.scheduleDA.InsertInBatchesTx(ctx, tx, scheduleList, len(scheduleList))
	if err != nil {
		log.Error(ctx, "s.scheduleDA.InsertInBatchesTx error",
			log.Err(err),
			log.Any("scheduleList", scheduleList))
		return nil, err
	}

	// insert into `schedules_relations` table
	_, err = s.scheduleRelationDA.InsertInBatchesTx(ctx, tx, scheduleRelations, len(scheduleRelations))
	if err != nil {
		log.Error(ctx, "s.scheduleRelationDA.InsertInBatchesTx error",
			log.Err(err),
			log.Any("scheduleRelations", scheduleRelations))
		return nil, err
	}

	// TODO add assessment
	if batchAddAssessmentSuperArgs != nil {
		err = GetAssessmentModel().BatchAddTx(ctx, tx, op, batchAddAssessmentSuperArgs)
		if err != nil {
			log.Error(ctx, "add schedule assessment error",
				log.Err(err),
				log.Any("batchAddAssessmentSuperArgs", batchAddAssessmentSuperArgs))
			return nil, err
		}
	}

	return result.([]*entity.Schedule), nil
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
	if schedule.ClassType == entity.ScheduleClassTypeHomework {
		if schedule.IsHomeFun {
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
		} else {
			existAssessment, err := GetStudyAssessmentModel().BatchCheckAnyoneAttempted(ctx, dbo.MustGetDB(ctx), op, []string{schedule.ID})
			if err != nil {
				log.Error(ctx, "judgment anyone attempt error", log.Err(err), log.String("scheduleID", schedule.ID))
				return nil, err
			}
			if existAssessment[schedule.ID] {
				log.Info(ctx, "The schedule has already been attended", log.Any("scheduleID", schedule.ID))
				return nil, ErrScheduleStudyAlreadyProgress
			}
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

func (s *scheduleModel) Update(ctx context.Context, operator *entity.Operator, viewData *entity.ScheduleUpdateView) ([]*entity.Schedule, error) {
	schedule, err := s.checkScheduleStatus(ctx, operator, viewData.ID)
	if err != nil {
		log.Error(ctx, "update schedule: get schedule by id error",
			log.Any("viewData", viewData),
			log.Err(err),
		)
		return nil, err
	}
	viewData.SubjectIDs = utils.SliceDeduplicationExcludeEmpty(viewData.SubjectIDs)
	// verify data
	err = s.verifyData(ctx, operator, &entity.ScheduleVerifyInput{
		ClassID:      viewData.ClassID,
		SubjectIDs:   viewData.SubjectIDs,
		ProgramID:    viewData.ProgramID,
		LessonPlanID: viewData.LessonPlanID,
		ClassType:    viewData.ClassType,
		IsHomeFun:    viewData.IsHomeFun,
		OutcomeIDs:   viewData.OutcomeIDs,
	})
	if err != nil {
		log.Error(ctx, "update schedule: verify data error",
			log.Err(err),
			log.Any("viewData", viewData))
		return nil, err
	}

	if viewData.ClassType == entity.ScheduleClassTypeTask {
		viewData.LessonPlanID = ""
		viewData.ProgramID = ""
		viewData.SubjectIDs = nil
	}

	// update schedule
	relationInput := &entity.ScheduleRelationInput{
		ScheduleID:             viewData.ID,
		ClassRosterClassID:     viewData.ClassID,
		ClassRosterTeacherIDs:  viewData.ClassRosterTeacherIDs,
		ClassRosterStudentIDs:  viewData.ClassRosterStudentIDs,
		ParticipantsTeacherIDs: viewData.ParticipantsTeacherIDs,
		ParticipantsStudentIDs: viewData.ParticipantsStudentIDs,
		SubjectIDs:             viewData.SubjectIDs,
	}

	// homefun study can bind learning outcome
	if viewData.ClassType == entity.ScheduleClassTypeHomework && viewData.IsHomeFun {
		relationInput.LearningOutcomeIDs = viewData.OutcomeIDs
	}

	relations, err := s.prepareScheduleRelationUpdateData(ctx, operator, relationInput)
	if err != nil {
		return nil, err
	}

	updateSchedule, repeatOptions, err := s.prepareScheduleUpdateData(ctx, operator, schedule, viewData)
	if err != nil {
		log.Error(ctx, "prepareScheduleUpdateData: error",
			log.Err(err),
			log.Any("schedule", schedule),
			log.Any("viewData", viewData))
		return nil, err
	}

	scheduleList, allRelations, batchAddAssessmentSuperArgs, err := s.prepareScheduleAddData(ctx, operator, updateSchedule, repeatOptions, viewData.Location, relations)
	if err != nil {
		log.Error(ctx, "prepareScheduleAddData: error",
			log.Err(err),
			log.Any("schedule", schedule),
			log.Any("option", &viewData.Repeat),
			log.Any("location", viewData.Location),
			log.Any("relations", relations))
		return nil, err
	}

	var result []*entity.Schedule
	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		var err error
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

		result, err = s.addSchedule(ctx, tx, operator, scheduleList, allRelations, batchAddAssessmentSuperArgs)
		if err != nil {
			log.Error(ctx, "s.addSchedule error",
				log.Err(err),
				log.Any("schedule", updateSchedule),
				log.Any("viewData", viewData),
			)
			return err
		}
		return nil
	}); err != nil {
		log.Error(ctx, "update schedule: tx failed", log.Err(err))
		return nil, err
	}
	err = da.GetScheduleRedisDA().Clean(ctx, operator.OrgID)
	if err != nil {
		log.Warn(ctx, "clean schedule cache error", log.String("orgID", operator.OrgID), log.Err(err))
	}

	go removeResourceMetadata(ctx, viewData.Attachment.ID)
	return result, nil
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
	var scheduleIDs []string

	if editType == entity.ScheduleEditOnlyCurrent ||
		(editType == entity.ScheduleEditWithFollowing && schedule.RepeatID == "") {

		scheduleIDs = append(scheduleIDs, schedule.ID)

	} else if editType == entity.ScheduleEditWithFollowing {
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

		scheduleIDs = make([]string, len(scheduleList))
		for i, item := range scheduleList {
			scheduleIDs[i] = item.ID
		}
	}
	if len(scheduleIDs) <= 0 {
		log.Info(ctx, "no need to delete", log.Any("schedule", schedule), log.Any("editType", editType))
		return nil
	}

	// delete schedule relation error
	err := da.GetScheduleRelationDA().Delete(ctx, tx, scheduleIDs)
	if err != nil {
		log.Error(ctx, "delete schedule relation error",
			log.Err(err),
			log.Any("op", op),
		)
		return err
	}

	// delete schedule assessment relation error
	err = GetStudyAssessmentModel().Delete(ctx, tx, op, scheduleIDs)
	if err != nil {
		log.Error(ctx, "delete schedule assessment relation error",
			log.Err(err),
			log.Any("op", op),
		)
		return err
	}

	return nil
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
		log.Error(ctx, "da.GetScheduleDA().Page error",
			log.Err(err),
			log.Any("condition", condition))
		return 0, nil, err
	}

	// in schedule
	var classIDs []string
	var programIDs []string
	var lessonPlanIDs []string
	// in schedule_relation
	var subjectIDs []string
	var teacherIDs []string

	result := make([]*entity.ScheduleSearchView, len(scheduleList))
	resultMap := make(map[string]*entity.ScheduleSearchView, len(scheduleList))
	scheduleIDs := make([]string, len(scheduleList))
	for i, schedule := range scheduleList {
		scheduleIDs[i] = schedule.ID
		result[i] = &entity.ScheduleSearchView{
			ID:      schedule.ID,
			StartAt: schedule.StartAt,
			EndAt:   schedule.EndAt,
			DueAt:   schedule.DueAt,
			Title:   schedule.Title,
		}

		resultMap[schedule.ID] = result[i]

		classIDs = append(classIDs, schedule.ClassID)
		programIDs = append(programIDs, schedule.ProgramID)
		lessonPlanIDs = append(lessonPlanIDs, schedule.LessonPlanID)
	}

	// query schedule_relation, include teachers, students and subjects(one-to-many relationship)
	scheduleRelations, err := GetScheduleRelationModel().Query(ctx, operator, &da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{Strings: scheduleIDs, Valid: true},
		RelationTypes: entity.NullStrings{
			Strings: []string{
				string(entity.ScheduleRelationTypeClassRosterTeacher),
				string(entity.ScheduleRelationTypeParticipantTeacher),
				string(entity.ScheduleRelationTypeClassRosterStudent),
				string(entity.ScheduleRelationTypeParticipantStudent),
				string(entity.ScheduleRelationTypeSubject),
			}, Valid: true},
	})

	for _, scheduleRelation := range scheduleRelations {
		switch scheduleRelation.RelationType {
		case entity.ScheduleRelationTypeClassRosterTeacher, entity.ScheduleRelationTypeParticipantTeacher:
			teacherIDs = append(teacherIDs, scheduleRelation.RelationID)
		case entity.ScheduleRelationTypeSubject:
			subjectIDs = append(subjectIDs, scheduleRelation.RelationID)
		}
	}

	classIDs = utils.SliceDeduplicationExcludeEmpty(classIDs)
	programIDs = utils.SliceDeduplicationExcludeEmpty(programIDs)
	lessonPlanIDs = utils.SliceDeduplicationExcludeEmpty(lessonPlanIDs)
	subjectIDs = utils.SliceDeduplicationExcludeEmpty(subjectIDs)
	teacherIDs = utils.SliceDeduplicationExcludeEmpty(teacherIDs)

	var classMap map[string]*external.NullableClass
	var programMap map[string]*external.Program
	var lessonPlanMap map[string]*entity.ScheduleShortInfo
	var subjectMap map[string]*external.Subject
	var teacherMap map[string]*external.NullableTeacher

	g := new(errgroup.Group)

	// get class info
	g.Go(func() error {
		classes, err := s.classService.BatchGetMap(ctx, operator, classIDs)
		if err != nil {
			log.Error(ctx, "s.classService.BatchGetMap error",
				log.Err(err),
				log.Strings("classIDs", classIDs))
			return err
		}
		classMap = classes
		return nil
	})

	// get program info
	g.Go(func() error {
		programs, err := s.programService.BatchGetMap(ctx, operator, programIDs)
		if err != nil {
			log.Error(ctx, "s.programService.BatchGetMap error",
				log.Err(err),
				log.Strings("programIDs", programIDs))
			return err
		}
		programMap = programs
		return nil
	})

	// get subject info
	g.Go(func() error {
		subjects, err := s.subjectService.BatchGetMap(ctx, operator, subjectIDs)
		if err != nil {
			log.Error(ctx, "s.subjectService.BatchGetMap error",
				log.Err(err),
				log.Strings("subjectIDs", subjectIDs))
			return err
		}
		subjectMap = subjects
		return nil
	})

	// get teacher info
	g.Go(func() error {
		teachers, err := s.teacherService.BatchGetMap(ctx, operator, teacherIDs)
		if err != nil {
			log.Error(ctx, "s.teacherService.BatchGetMap error",
				log.Err(err),
				log.Strings("teacherIDs", teacherIDs))
			return err
		}
		teacherMap = teachers
		return nil
	})

	// get lesson plan info
	g.Go(func() error {
		lessonPlans, err := s.getLessonPlanNameByIDs(ctx, dbo.MustGetDB(ctx), lessonPlanIDs)
		if err != nil {
			log.Error(ctx, "s.getLessonPlanNameByIDs error",
				log.Err(err),
				log.Any("lessonPlanIDs", lessonPlanIDs))
			return err
		}

		lessonPlanMap = lessonPlans
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Error(ctx, "get schedule basic info error",
			log.Err(err))
		return 0, nil, err
	}

	// fill schedule program, lesson plan, class
	for _, schedule := range scheduleList {
		if program, ok := programMap[schedule.ProgramID]; ok {
			resultMap[schedule.ID].Program = &entity.ScheduleShortInfo{
				ID:   program.ID,
				Name: program.Name,
			}
		}

		if lessonPlan, ok := lessonPlanMap[schedule.LessonPlanID]; ok {
			resultMap[schedule.ID].LessonPlan = &entity.ScheduleShortInfo{
				ID:   lessonPlan.ID,
				Name: lessonPlan.Name,
			}
		}

		if class, ok := classMap[schedule.ClassID]; ok && class.Valid {
			// TODO: why use entity.ScheduleAccessibleUserView, field `Type` and `Enable` never used
			resultMap[schedule.ID].Class = &entity.ScheduleAccessibleUserView{
				ID:   class.ID,
				Name: class.Name,
			}
		}
	}

	// fill schedule relation, subject, teacher, student count
	for _, scheduleRelation := range scheduleRelations {
		switch scheduleRelation.RelationType {
		case entity.ScheduleRelationTypeClassRosterTeacher, entity.ScheduleRelationTypeParticipantTeacher:
			if teacher, ok := teacherMap[scheduleRelation.RelationID]; ok {
				resultMap[scheduleRelation.ScheduleID].MemberTeachers = append(resultMap[scheduleRelation.ScheduleID].MemberTeachers, &entity.ScheduleShortInfo{
					ID:   teacher.ID,
					Name: teacher.Name,
				})
			}
		case entity.ScheduleRelationTypeClassRosterStudent, entity.ScheduleRelationTypeParticipantStudent:
			// TODO: if schedule_relation table exist dirty data, need to clean duplicate data
			resultMap[scheduleRelation.ScheduleID].StudentCount++
		case entity.ScheduleRelationTypeSubject:
			if subject, ok := subjectMap[scheduleRelation.RelationID]; ok {
				resultMap[scheduleRelation.ScheduleID].Subjects = append(resultMap[scheduleRelation.ScheduleID].Subjects, &entity.ScheduleShortInfo{
					ID:   subject.ID,
					Name: subject.Name,
				})
			}
		}
	}

	return total, result, nil
}

func (s *scheduleModel) queryByCache(ctx context.Context, op *entity.Operator, condition *da.ScheduleCondition) ([]*entity.ScheduleListView, error) {
	result, err := da.GetScheduleRedisDA().GetScheduleListView(ctx, op.OrgID, condition)
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

	studyScheduleIDs := make([]string, 0, len(scheduleList))
	scheduleIDs := make([]string, len(scheduleList))
	for i, item := range scheduleList {
		if item.ClassType == entity.ScheduleClassTypeHomework && !item.IsHomeFun {
			studyScheduleIDs = append(studyScheduleIDs, item.ID)
		}
		scheduleIDs[i] = item.ID
	}

	existAssessmentMap, err := GetStudyAssessmentModel().BatchCheckAnyoneAttempted(ctx, dbo.MustGetDB(ctx), op, studyScheduleIDs)
	if err != nil {
		log.Error(ctx, "judgment anyone attempt error", log.Err(err), log.Any("scheduleIDs", studyScheduleIDs))
		return nil, err
	}

	assessments, err := GetAssessmentModel().Query(ctx, op, &da.QueryAssessmentConditions{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   len(scheduleIDs) > 0,
		},
	})
	if err != nil {
		log.Error(ctx, "get assessment error", log.Err(err), log.Any("scheduleIDs", scheduleIDs))
		return nil, err
	}

	var homeFunStudyAssessments []*entity.HomeFunStudy
	err = GetHomeFunStudyModel().Query(ctx, op, &da.QueryHomeFunStudyCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   len(scheduleIDs) > 0,
		},
	}, &homeFunStudyAssessments)
	if err != nil {
		log.Error(ctx, "get homefun study assessment error",
			log.Err(err),
			log.Any("scheduleIDs", scheduleIDs))
		return nil, err
	}

	completeAssessmentMap := make(map[string]bool, len(assessments))
	for _, v := range assessments {
		if v.Status == entity.AssessmentStatusComplete {
			completeAssessmentMap[v.ScheduleID] = true
		}
	}

	completeHomefunStudyAssessmentMap := make(map[string]bool, len(homeFunStudyAssessments))
	for _, v := range homeFunStudyAssessments {
		if v.Status == entity.AssessmentStatusComplete {
			completeHomefunStudyAssessmentMap[v.ScheduleID] = true
		}
	}

	for _, item := range scheduleList {
		temp := &entity.ScheduleListView{
			ID:           item.ID,
			Title:        item.Title,
			StartAt:      item.StartAt,
			EndAt:        item.EndAt,
			IsRepeat:     item.RepeatID != "",
			LessonPlanID: item.LessonPlanID,
			Status:       item.Status,
			ClassID:      item.ClassID,
			ClassType:    item.ClassType,
			DueAt:        item.DueAt,
			IsHidden:     item.IsHidden,
			IsHomeFun:    item.IsHomeFun,
		}

		temp.ClassTypeLabel = entity.ScheduleShortInfo{
			ID:   item.ClassType.String(),
			Name: item.ClassType.ToLabel().String(),
		}

		temp.Status = temp.Status.GetScheduleStatus(entity.ScheduleStatusInput{
			EndAt:     temp.EndAt,
			DueAt:     temp.DueAt,
			ClassType: temp.ClassType,
		})

		if temp.ClassType == entity.ScheduleClassTypeHomework && temp.DueAt != 0 {
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
		temp.ExistAssessment = existAssessmentMap[item.ID]
		temp.CompleteAssessment = completeAssessmentMap[item.ID]
		if temp.ClassType == entity.ScheduleClassTypeHomework && temp.IsHomeFun {
			temp.CompleteAssessment = completeHomefunStudyAssessmentMap[item.ID]
		}

		result = append(result, temp)
	}

	return result, nil
}

func (s *scheduleModel) QueryByConditionInternal(ctx context.Context, condition *da.ScheduleCondition) (int, []*entity.ScheduleSimplified, error) {
	var scheduleList []*entity.Schedule
	total, err := da.GetScheduleDA().Page(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "schedule query error", log.Err(err), log.Any("condition", condition))
		return 0, nil, err
	}
	res := make([]*entity.ScheduleSimplified, len(scheduleList))
	for i := range scheduleList {
		res[i] = scheduleList[i].ToScheduleSimplified()
	}
	return total, res, nil
}

func (s *scheduleModel) QueryByCondition(ctx context.Context, op *entity.Operator, condition *da.ScheduleCondition, loc *time.Location) ([]*entity.ScheduleListView, error) {
	// cache
	// cacheData, err := s.queryByCache(ctx, op, condition)
	// if err == nil {
	// 	log.Info(ctx, "Query:using cache",
	// 		log.Any("op", op),
	// 		log.Any("condition", condition),
	// 	)
	// 	return cacheData, nil
	// }

	var scheduleList []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "da.GetScheduleDA().Query error",
			log.Err(err),
			log.Any("condition", condition))
		return nil, err
	}

	// cache
	// if err = da.GetScheduleRedisDA().Set(ctx, op.OrgID, &da.ScheduleCacheCondition{
	// 	Condition: condition,
	// 	DataType:  da.ScheduleListView,
	// }, result); err != nil {
	// 	log.Warn(ctx, "set cache error",
	// 		log.Err(err),
	// 		log.Any("condition", condition),
	// 		log.Any("data", result))
	// }

	return s.transformToScheduleListView(ctx, op, scheduleList, loc)
}

func (s *scheduleModel) getLessonPlanNameByIDs(ctx context.Context, tx *dbo.DBContext, lessonPlanIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	lessonPlans, err := GetContentModel().GetContentNameByIDList(ctx, tx, lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "GetContentModel().GetContentNameByIDList error",
			log.Err(err),
			log.Strings("lessonPlanIDs", lessonPlanIDs))
		return nil, err
	}

	lessonPlanMap := make(map[string]*entity.ScheduleShortInfo, len(lessonPlans))

	for _, lessonPlan := range lessonPlans {
		lessonPlanMap[lessonPlan.ID] = &entity.ScheduleShortInfo{
			ID:   lessonPlan.ID,
			Name: lessonPlan.Name,
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
		if err != nil && err != ErrScheduleLessonPlanUnAuthed {
			return nil, err
		}
		result.IsAuth = isAuth

		contentList, err := GetContentModel().GetContentSubContentsByID(ctx, dbo.MustGetDB(ctx), lessonPlanID, op, false)
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

func (s *scheduleModel) GetSubjectsBySubjectIDs(ctx context.Context, operator *entity.Operator, subjectIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
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

func (s *scheduleModel) GetByID(ctx context.Context, operator *entity.Operator, id string) (*entity.ScheduleDetailsView, error) {
	var schedule = new(entity.Schedule)
	err := s.scheduleDA.Get(ctx, id, schedule)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "schedule reocord not found",
			log.Err(err),
			log.String("scheduleID", id))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "s.scheduleDA.Get error",
			log.Err(err),
			log.String("scheduleID", id))
		return nil, err
	}

	if schedule.DeleteAt != 0 {
		log.Error(ctx, "schedule reocord has been deleted",
			log.Any("schedule", schedule))
		return nil, constant.ErrRecordNotFound
	}

	return s.transformToScheduleDetailsView(ctx, operator, schedule)
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

func (s *scheduleModel) accessibleParticipantUser(ctx context.Context, operator *entity.Operator, users []*entity.ScheduleUserInput) ([]*entity.ScheduleAccessibleUserView, error) {
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

func (s *scheduleModel) ExistScheduleByLessonPlanID(ctx context.Context, lessonPlanID string) (bool, error) {
	if strings.TrimSpace(lessonPlanID) == "" {
		log.Info(ctx, "lessonPlanID is empty", log.String("lessonPlanID", lessonPlanID))
		return false, errors.New("lessonPlanID is empty")
	}
	lessonPlanPastIDs, err := GetContentModel().GetPastContentIDByID(ctx, dbo.MustGetDB(ctx), lessonPlanID)
	if err != nil {
		log.Error(ctx, "ExistScheduleByLessonPlanID:GetContentModel.GetPastContentIDByID error",
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

func (s *scheduleModel) verifyData(ctx context.Context, operator *entity.Operator, v *entity.ScheduleVerifyInput) error {
	// class
	// classService := external.GetClassServiceProvider()
	classInfos, err := s.classService.BatchGet(ctx, operator, []string{v.ClassID})
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

	// verify learning outcome
	if len(v.OutcomeIDs) > 0 {
		_, outcomes, err := GetOutcomeModel().SearchWithoutRelation(ctx, operator, &entity.OutcomeCondition{
			IDs:           v.OutcomeIDs,
			PublishStatus: entity.OutcomeStatusPublished,
			Assumed:       -1,
		})
		if err != nil {
			log.Error(ctx, "verifyData: GetOutcomeModel().SearchWithoutRelation error",
				log.Err(err),
				log.Any("outcomeIDs", v.OutcomeIDs))
			return err
		}

		if len(outcomes) != len(v.OutcomeIDs) {
			log.Error(ctx, "verifyData: learning outcome not found",
				log.Any("outcomeIDs", v.OutcomeIDs),
				log.Any("outcomes", outcomes))
			return constant.ErrRecordNotFound
		}
	}

	return nil
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

func (s *scheduleModel) GetScheduleIDsByCondition(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, condition *entity.ScheduleIDsCondition) ([]string, error) {
	lessonPlanPastIDs, err := GetContentModel().GetPastContentIDByID(ctx, tx, condition.LessonPlanID)
	if err != nil {
		log.Error(ctx, "GetScheduleIDsByCondition:get past lessonPlan id error",
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

func (s *scheduleModel) QueryScheduledDatesByCondition(ctx context.Context, op *entity.Operator, condition *da.ScheduleCondition, loc *time.Location) ([]string, error) {
	cacheData, err := da.GetScheduleRedisDA().GetScheduledDates(ctx, op.OrgID, condition)
	if err == nil && len(cacheData) > 0 {
		log.Info(ctx, "Query:using cache",
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

	if err = da.GetScheduleRedisDA().Set(ctx, op.OrgID, &da.ScheduleCacheCondition{
		Condition: condition,
		DataType:  da.ScheduledDates,
	}, result); err != nil {
		log.Warn(ctx, "set cache error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("data", result))
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

func (s *scheduleModel) GetVariableDataByIDs(ctx context.Context, op *entity.Operator, ids []string, include *entity.ScheduleInclude) ([]*entity.ScheduleVariable, error) {
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

	result := make([]*entity.ScheduleVariable, len(scheduleList))
	for i, item := range scheduleList {
		plain := &entity.ScheduleVariable{
			RoomID:   item.ID,
			Schedule: item,
		}
		result[i] = plain
	}
	if include == nil {
		return result, nil
	}
	if include.Subject {
		scheduleSubjectMap, err := GetScheduleRelationModel().GetSubjectsByScheduleIDs(ctx, op, ids)
		if err != nil {
			return nil, err
		}
		for _, item := range result {
			item.Subjects = scheduleSubjectMap[item.ID]
		}
	}
	if include.ClassRosterClass {
		scheduleClassMap, err := GetScheduleRelationModel().GetClassRosterMap(ctx, op, ids)
		if err != nil {
			return nil, err
		}
		for _, item := range result {
			item.ClassRosterClass = scheduleClassMap[item.ID]
		}
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

func (s *scheduleModel) getSubjectCondition(ctx context.Context, op *entity.Operator, programID string) (*da.ScheduleRelationCondition, error) {
	condition, err := s.getRelationCondition(ctx, op)
	if err != nil {
		return nil, err
	}

	relationCondition := &da.ScheduleRelationCondition{
		ScheduleFilterSubject: &da.ScheduleFilterSubject{
			ProgramID: sql.NullString{
				String: programID,
				Valid:  true,
			},
			OrgID:       condition.OrgID,
			RelationIDs: condition.RelationIDs,
		},
	}
	return relationCondition, nil
}

func (s *scheduleModel) GetSubjects(ctx context.Context, op *entity.Operator, programID string) ([]*entity.ScheduleShortInfo, error) {
	condition, err := s.getSubjectCondition(ctx, op, programID)
	if err != nil {
		return nil, err
	}

	dbSubjectIDs, err := da.GetScheduleRelationDA().GetRelationIDsByCondition(ctx, dbo.MustGetDB(ctx), condition)
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

func (s *scheduleModel) GetLearningOutcomeIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string][]string, error) {
	var scheduleRelations []*entity.ScheduleRelation
	err := da.GetScheduleRelationDA().Query(ctx, &da.ScheduleRelationCondition{
		ScheduleIDs:  entity.NullStrings{Strings: scheduleIDs, Valid: true},
		RelationType: sql.NullString{String: string(entity.ScheduleRelationTypeLearningOutcome), Valid: true},
	}, &scheduleRelations)
	if err != nil {
		log.Error(ctx, "GetLearningOutcomeIDs error",
			log.Err(err),
			log.Any("op", op),
			log.Any("scheduleIDs", scheduleIDs))
		return nil, err
	}

	result := make(map[string][]string, len(scheduleIDs))
	for _, v := range scheduleIDs {
		result[v] = []string{}
	}

	for _, v := range scheduleRelations {
		result[v.ScheduleID] = append(result[v.ScheduleID], v.RelationID)
	}

	return result, nil
}

func (s *scheduleModel) GetScheduleViewByID(ctx context.Context, op *entity.Operator, id string) (*entity.ScheduleViewDetail, error) {
	var schedule = new(entity.Schedule)
	err := s.scheduleDA.Get(ctx, id, schedule)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "schedule reocord not found",
			log.Err(err),
			log.String("scheduleID", id))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "s.scheduleDA.Get error",
			log.Err(err),
			log.String("scheduleID", id))
		return nil, err
	}

	if schedule.DeleteAt != 0 {
		log.Error(ctx, "schedule reocord has been deleted",
			log.Any("schedule", schedule))
		return nil, constant.ErrRecordNotFound
	}

	return s.transformToScheduleViewDetail(ctx, op, schedule)
}

func (s *scheduleModel) GetTeachingLoad(ctx context.Context, input *entity.ScheduleTeachingLoadInput) ([]*entity.ScheduleTeachingLoadView, error) {
	condition := da.NewScheduleTeachLoadCondition(input)
	teachLoads, err := s.scheduleDA.GetTeachLoadByCondition(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		log.Error(ctx, "s.scheduleDA.GetTeachLoadByCondition error", log.Err(err), log.Any("input", input), log.Any("condition", condition))
		return nil, err
	}
	result := make([]*entity.ScheduleTeachingLoadView, 0, len(teachLoads))
	for _, loadItem := range teachLoads {
		resultItem := &entity.ScheduleTeachingLoadView{
			TeacherID: loadItem.TeacherID,
			ClassType: loadItem.ClassType,
			Durations: make([]*entity.ScheduleTeachingDuration, 0, len(input.TimeRanges)),
		}
		for i, duration := range loadItem.Durations {
			durationItem := new(entity.ScheduleTeachingDuration)
			durationItem.StartAt = input.TimeRanges[i].StartAt
			durationItem.EndAt = input.TimeRanges[i].EndAt
			durationItem.Duration = duration
			resultItem.Durations = append(resultItem.Durations, durationItem)
		}
		result = append(result, resultItem)
	}
	return result, nil
}

func (s *scheduleModel) PrepareScheduleTimeViewCondition(ctx context.Context, query *entity.ScheduleTimeViewQuery, op *entity.Operator, loc *time.Location) (*da.ScheduleCondition, error) {
	permissionMap, err := GetSchedulePermissionModel().HasScheduleOrgPermissions(ctx, op, []external.PermissionName{
		external.ScheduleViewOrgCalendar,
		external.ScheduleViewSchoolCalendar,
		external.ScheduleViewMyCalendar,
	})
	if err == constant.ErrForbidden {
		log.Info(ctx, "request info",
			log.Any("query", query),
			log.Any("op", op),
			log.Any("loc", loc),
		)
		return nil, constant.ErrForbidden
	}
	if err != nil {
		log.Info(ctx, "request info",
			log.Any("query", query),
			log.Any("op", op),
			log.Any("loc", loc),
		)
		return nil, constant.ErrInternalServer
	}

	viewType := query.ViewType
	condition := new(da.ScheduleCondition)
	if viewType != entity.ScheduleViewTypeFullView.String() {
		timeAt := query.TimeAt
		var (
			start int64
			end   int64
		)
		switch entity.ScheduleViewType(viewType) {
		case entity.ScheduleViewTypeDay:
			start = utils.TodayZeroByTimeStamp(timeAt, loc).Unix()
			end = utils.TodayEndByTimeStamp(timeAt, loc).Unix()
		case entity.ScheduleViewTypeWorkweek:
			start, end = utils.FindWorkWeekTimeRange(timeAt, loc)
		case entity.ScheduleViewTypeWeek:
			start, end = utils.FindWeekTimeRange(timeAt, loc)
		case entity.ScheduleViewTypeMonth:
			start, end = utils.FindMonthRange(timeAt, loc)
		case entity.ScheduleViewTypeYear:
			start = utils.StartOfYearByTimeStamp(timeAt, loc).Unix()
			end = utils.EndOfYearByTimeStamp(timeAt, loc).Unix()
		default:
			log.Info(ctx, "request info",
				log.Any("query", query),
				log.Any("op", op),
				log.Any("loc", loc),
			)
			return nil, constant.ErrInvalidArgs
		}
		startAndEndTimeViewRange := make([]sql.NullInt64, 2)
		startAndEndTimeViewRange[0] = sql.NullInt64{
			Valid: start >= 0,
			Int64: start,
		}
		startAndEndTimeViewRange[1] = sql.NullInt64{
			Valid: end >= 0,
			Int64: end,
		}
		condition.StartAndEndTimeViewRange = startAndEndTimeViewRange
	}

	if len(query.SubjectIDs) > 0 {
		condition.SubjectIDs = entity.NullStrings{
			Strings: query.SubjectIDs,
			Valid:   true,
		}
	}

	if len(query.ProgramIDs) > 0 {
		condition.ProgramIDs = entity.NullStrings{
			Strings: query.ProgramIDs,
			Valid:   true,
		}
	}

	if len(query.ClassTypes) > 0 {
		condition.ClassTypes = entity.NullStrings{
			Strings: query.ClassTypes,
			Valid:   true,
		}
	}

	if len(query.UserIDs) > 0 {
		condition.RelationUserIDs = entity.NullStrings{
			Strings: query.UserIDs,
			Valid:   true,
		}
	}

	if len(query.SchoolIDs) > 0 {
		condition.RelationSchoolIDs = entity.NullStrings{
			Strings: query.SchoolIDs,
			Valid:   true,
		}
	}

	if len(query.ClassIDs) > 0 {
		condition.RelationClassIDs = entity.NullStrings{
			Strings: query.ClassIDs,
			Valid:   true,
		}
	}

	condition.OrderBy = da.NewScheduleOrderBy(query.OrderBy)

	condition.OrgID = sql.NullString{
		String: op.OrgID,
		Valid:  true,
	}

	relationIDs := make([]string, 0)
	if permissionMap[external.ScheduleViewOrgCalendar] {
	} else if permissionMap[external.ScheduleViewSchoolCalendar] {
		schoolList, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, external.ScheduleViewSchoolCalendar)
		if err != nil {
			log.Error(ctx, "GetSchoolServiceProvider.GetByPermission error",
				log.Err(err),
				log.Any("op", op),
				log.String("permission", external.ScheduleViewSchoolCalendar.String()),
			)
			return nil, constant.ErrInternalServer
		}
		var relationIds []string
		for _, item := range schoolList {
			relationIds = append(relationIds, item.ID)
		}

		relationIDs = append(relationIDs, op.UserID)
		condition.RelationIDs = entity.NullStrings{
			Strings: relationIDs,
			Valid:   true,
		}
	} else if permissionMap[external.ScheduleViewMyCalendar] {
		condition.RelationID = sql.NullString{
			String: op.UserID,
			Valid:  true,
		}
	}
	condition.AnyTime = sql.NullBool{
		Bool:  query.Anytime,
		Valid: query.Anytime,
	}

	log.Debug(ctx, "condition info",
		log.String("viewType", viewType),
		log.Any("condition", condition),
	)
	return condition, nil
}

func (s *scheduleModel) Query(ctx context.Context, query *entity.ScheduleTimeViewQuery, op *entity.Operator, loc *time.Location) ([]*entity.ScheduleListView, error) {
	condition, err := s.PrepareScheduleTimeViewCondition(ctx, query, op, loc)
	if err != nil {
		return nil, err
	}

	var scheduleList []*entity.Schedule
	err = s.scheduleDA.Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "s.scheduleDA.Query error",
			log.Err(err),
			log.Any("condition", condition))
		return nil, err
	}

	return s.transformToScheduleListView(ctx, op, scheduleList, loc)
}

func (s *scheduleModel) QueryScheduledDates(ctx context.Context, query *entity.ScheduleTimeViewQuery, op *entity.Operator, loc *time.Location) ([]string, error) {
	condition, err := s.PrepareScheduleTimeViewCondition(ctx, query, op, loc)
	if err != nil {
		return nil, err
	}
	result, err := s.QueryScheduledDatesByCondition(ctx, op, condition, loc)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// without permission check, internal function call
func (s *scheduleModel) QueryUnsafe(ctx context.Context, condition *entity.ScheduleQueryCondition) ([]*entity.Schedule, error) {
	var scheduleList []*entity.Schedule
	daCondition := &da.ScheduleCondition{
		IDs:                condition.IDs,
		OrgID:              condition.OrgID,
		StartAtGe:          condition.StartAtGe,
		StartAtLt:          condition.StartAtLt,
		ClassTypes:         condition.ClassTypes,
		IsHomefun:          condition.IsHomefun,
		RelationSchoolIDs:  condition.RelationSchoolIDs,
		RelationClassIDs:   condition.RelationClassIDs,
		RelationTeacherIDs: condition.RelationTeacherIDs,
		RelationStudentIDs: condition.RelationStudentIDs,
		SubjectIDs:         condition.RelationSubjectIDs,
		DeleteAt:           condition.DeleteAt,
	}
	err := s.scheduleDA.Query(ctx, daCondition, &scheduleList)
	if err != nil {
		log.Error(ctx, "s.scheduleDA.Query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}

	return scheduleList, nil
}

func (s *scheduleModel) QueryScheduleTimeView(ctx context.Context, query *entity.ScheduleTimeViewListRequest, op *entity.Operator, loc *time.Location) (int, []*entity.ScheduleTimeView, error) {
	condition, err := s.PrepareScheduleTimeViewCondition(ctx, &entity.ScheduleTimeViewQuery{
		ViewType:       query.ViewType,
		TimeAt:         query.TimeAt,
		TimeZoneOffset: query.TimeZoneOffset,
		SchoolIDs:      query.SchoolIDs,
		TeacherIDs:     query.TeacherIDs,
		ClassIDs:       query.ClassIDs,
		SubjectIDs:     query.SubjectIDs,
		ProgramIDs:     query.ProgramIDs,
		ClassTypes:     query.ClassTypes,
		StartAtGe:      query.StartAtGe,
		EndAtLe:        query.EndAtLe,
		Anytime:        query.Anytime,
		OrderBy:        query.OrderBy,
	}, op, loc)
	if err != nil {
		return 0, nil, err
	}

	// StartAtGe query by start_at and due_at, required by APP team
	if query.StartAtGe >= 0 {
		// apply to StartAtGe and EndAtLe, union will include schedules that are only partially within the specified time frame, intersect will not
		if query.TimeBoundary == string(entity.UnionScheduleTimeBoundary) {
			condition.StartAtOrEndAtOrDueAtGe = sql.NullInt64{
				Int64: query.StartAtGe,
				Valid: true,
			}
		} else {
			condition.StartAtAndDueAtGe = sql.NullInt64{
				Int64: query.StartAtGe,
				Valid: true,
			}
		}
	}

	// EndAtLe query by end_at and due_at, required by APP team
	if query.EndAtLe >= 0 {
		if query.TimeBoundary == string(entity.UnionScheduleTimeBoundary) {
			condition.StartAtOrEndAtOrDueAtLe = sql.NullInt64{
				Int64: query.EndAtLe,
				Valid: true,
			}
		} else {
			condition.EndAtAndDueAtLe = sql.NullInt64{
				Int64: query.EndAtLe,
				Valid: true,
			}
		}
	}

	if query.DueAtEq >= 0 {
		condition.DueToEq = sql.NullInt64{
			Int64: query.DueAtEq,
			Valid: true,
		}
	}

	// pagination
	condition.Pager = utils.GetDboPagerFromInt(query.Page, query.PageSize)

	var scheduleList []*entity.Schedule
	total, err := da.GetScheduleDA().Page(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "da.GetScheduleDA().Page error",
			log.Err(err),
			log.Any("condition", condition))
		return 0, nil, err
	}

	result := make([]*entity.ScheduleTimeView, len(scheduleList))
	var homefunStudyScheduleIDs []string
	var notHomefunStudyScheduleIDs []string
	for i, v := range scheduleList {
		result[i] = &entity.ScheduleTimeView{
			ID:           v.ID,
			Title:        v.Title,
			StartAt:      v.StartAt,
			EndAt:        v.EndAt,
			DueAt:        v.DueAt,
			ClassType:    v.ClassType,
			Status:       v.Status,
			ClassID:      v.ClassID,
			IsHomeFun:    v.IsHomeFun,
			IsRepeat:     v.RepeatID != "",
			LessonPlanID: v.LessonPlanID,
			CreatedAt:    v.CreatedAt,
		}

		// handle schedule status
		result[i].Status = v.Status.GetScheduleStatus(entity.ScheduleStatusInput{
			EndAt:     v.EndAt,
			DueAt:     v.DueAt,
			ClassType: v.ClassType,
		})

		// handle homework schedule start and end time
		if v.ClassType == entity.ScheduleClassTypeHomework && v.DueAt != 0 {
			result[i].StartAt = utils.TodayZeroByTimeStamp(v.DueAt, loc).Unix()
			result[i].EndAt = utils.TodayEndByTimeStamp(v.DueAt, loc).Unix()
		}

		if v.ClassType == entity.ScheduleClassTypeHomework && v.IsHomeFun {
			homefunStudyScheduleIDs = append(homefunStudyScheduleIDs, v.ID)
		} else {
			notHomefunStudyScheduleIDs = append(notHomefunStudyScheduleIDs, v.ID)
		}
	}

	// query assessment_status, only for student user
	if query.WithAssessmentStatus {
		assessments, err := GetAssessmentModel().Query(ctx, op, &da.QueryAssessmentConditions{
			ScheduleIDs: entity.NullStrings{
				Strings: notHomefunStudyScheduleIDs,
				Valid:   true,
			},
		})
		if err != nil {
			log.Error(ctx, "GetAssessmentModel().Query error",
				log.Err(err),
				log.Any("notHomefunStudyScheduleIDs", notHomefunStudyScheduleIDs))
			return 0, nil, err
		}

		var homeFunStudyAssessments []*entity.HomeFunStudy
		err = GetHomeFunStudyModel().Query(ctx, op, &da.QueryHomeFunStudyCondition{
			ScheduleIDs: entity.NullStrings{
				Strings: homefunStudyScheduleIDs,
				Valid:   true,
			},
			StudentIDs: entity.NullStrings{
				Strings: []string{op.UserID},
				Valid:   true,
			},
		}, &homeFunStudyAssessments)
		if err != nil {
			log.Error(ctx, "GetHomeFunStudyModel().Query error",
				log.Err(err),
				log.Strings("homefunStudyScheduleIDs", homefunStudyScheduleIDs),
				log.String("studentID", op.UserID))
			return 0, nil, err
		}

		assessmentStatusMap := make(map[string]entity.AssessmentStatus)
		for _, assessment := range assessments {
			assessmentStatusMap[assessment.ScheduleID] = assessment.Status
		}

		for _, homefunStudyAssessment := range homeFunStudyAssessments {
			assessmentStatusMap[homefunStudyAssessment.ScheduleID] = homefunStudyAssessment.Status
		}

		for _, r := range result {
			r.AssessmentStatus = assessmentStatusMap[r.ID]
		}
	}

	return total, result, nil
}

// Schedule model interval function
func (s *scheduleModel) transformToScheduleDetailsView(ctx context.Context, operator *entity.Operator, schedule *entity.Schedule) (*entity.ScheduleDetailsView, error) {
	if schedule == nil {
		log.Debug(ctx, "schedule is nil")
		return nil, nil
	}

	scheduleDetailsView := &entity.ScheduleDetailsView{
		ID:          schedule.ID,
		Title:       schedule.Title,
		OrgID:       schedule.OrgID,
		StartAt:     schedule.StartAt,
		EndAt:       schedule.EndAt,
		IsAllDay:    schedule.IsAllDay,
		ClassType:   schedule.ClassType,
		DueAt:       schedule.DueAt,
		Description: schedule.Description,
		Version:     schedule.ScheduleVersion,
		IsRepeat:    schedule.RepeatID != "",
		Status:      schedule.Status,
		IsHomeFun:   schedule.IsHomeFun,
		IsHidden:    schedule.IsHidden,
		RoleType:    entity.ScheduleRoleTypeUnknown,
		ClassTypeLabel: entity.ScheduleShortInfo{
			ID:   schedule.ClassType.String(),
			Name: schedule.ClassType.ToLabel().String(),
		},
		OutcomeIDs: []string{},
	}

	// get schedule status, business logic not status in database
	scheduleDetailsView.Status = schedule.Status.GetScheduleStatus(entity.ScheduleStatusInput{
		EndAt:     schedule.EndAt,
		DueAt:     schedule.DueAt,
		ClassType: schedule.ClassType,
	})

	// get attachment
	if schedule.Attachment != "" {
		var attachment entity.ScheduleShortInfo
		err := json.Unmarshal([]byte(schedule.Attachment), &attachment)
		if err != nil {
			log.Error(ctx, "json.Unmarshal error",
				log.Err(err),
				log.String("schedule.Attachment", schedule.Attachment))
			return nil, err
		}
		scheduleDetailsView.Attachment = attachment
	}

	// get schedule repeat
	if schedule.RepeatJson != "" {
		var repeat entity.RepeatOptions
		err := json.Unmarshal([]byte(schedule.RepeatJson), &repeat)
		if err != nil {
			log.Error(ctx, "json.Unmarshal error",
				log.Err(err),
				log.String("schedule.RepeatJson", schedule.RepeatJson))
			return nil, err
		}
		scheduleDetailsView.Repeat = repeat
	}

	var scheduleRelations []*entity.ScheduleRelation
	err := s.scheduleRelationDA.Query(ctx, &da.ScheduleRelationCondition{
		ScheduleID: sql.NullString{
			String: schedule.ID,
			Valid:  true,
		},
	}, &scheduleRelations)

	var subjectIDs []string
	var classRoasterTeacherIDs []string
	var classRoasterStudentIDs []string
	var participantTeacherIDs []string
	var participantStudentIDs []string
	for _, scheduleRelation := range scheduleRelations {
		// get operator role type in the schedule
		if scheduleRelation.RelationID == operator.UserID {
			switch scheduleRelation.RelationType {
			case entity.ScheduleRelationTypeParticipantTeacher, entity.ScheduleRelationTypeClassRosterTeacher:
				scheduleDetailsView.RoleType = entity.ScheduleRoleTypeTeacher
			case entity.ScheduleRelationTypeParticipantStudent, entity.ScheduleRelationTypeClassRosterStudent:
				scheduleDetailsView.RoleType = entity.ScheduleRoleTypeStudent
			}
		}

		switch scheduleRelation.RelationType {
		// learning outcome relation, only for homefun homework
		case entity.ScheduleRelationTypeLearningOutcome:
			scheduleDetailsView.OutcomeIDs = append(scheduleDetailsView.OutcomeIDs, scheduleRelation.RelationID)
		case entity.ScheduleRelationTypeSubject:
			subjectIDs = append(subjectIDs, scheduleRelation.RelationID)
		case entity.ScheduleRelationTypeClassRosterTeacher:
			classRoasterTeacherIDs = append(classRoasterTeacherIDs, scheduleRelation.RelationID)
		case entity.ScheduleRelationTypeClassRosterStudent:
			classRoasterStudentIDs = append(classRoasterStudentIDs, scheduleRelation.RelationID)
		case entity.ScheduleRelationTypeParticipantTeacher:
			participantTeacherIDs = append(participantTeacherIDs, scheduleRelation.RelationID)
		case entity.ScheduleRelationTypeParticipantStudent:
			participantStudentIDs = append(participantStudentIDs, scheduleRelation.RelationID)
		}
	}

	g := new(errgroup.Group)

	var scheduleLessonPlan *entity.ScheduleLessonPlan
	var scheduleAccessibleUserView *entity.ScheduleAccessibleUserView

	var scheduleProgram *entity.ScheduleShortInfo
	var scheduleSubjects []*entity.ScheduleShortInfo
	var scheduleRealTimeStatus *entity.ScheduleRealTimeView
	var scheduleExistFeedback bool
	var scheduleExistAssessment bool
	var scheduleCompleteAssessment bool
	var userMap map[string]*external.NullableUser
	var schedulePermissionMap map[external.PermissionName]bool
	scheduleCanBeCreatedSchoolMap := make(map[string]bool)
	var schoolByClass []*external.School
	var classByOperator []*external.Class
	var participantUserSchoolMap map[string][]*external.School

	// get lesson plan
	if schedule.LessonPlanID != "" {
		g.Go(func() error {
			lessonPlan, err := s.getLessonPlanWithMaterial(ctx, operator, schedule.LessonPlanID)
			if err != nil {
				log.Error(ctx, "s.getLessonPlanWithMaterial error",
					log.Err(err),
					log.String("lessonPlanID", schedule.LessonPlanID))
				return err
			}

			scheduleLessonPlan = lessonPlan

			return nil
		})
	}

	// get class info
	if schedule.ClassID != "" {
		g.Go(func() error {
			classes, err := s.classService.BatchGet(ctx, operator, []string{schedule.ClassID})
			if err != nil {
				log.Error(ctx, "s.classService.BatchGet error",
					log.Err(err),
					log.String("classID", schedule.ClassID))
				return err
			}

			if len(classes) == 0 {
				log.Error(ctx, "class info not found", log.String("classID", schedule.ClassID))
				return constant.ErrRecordNotFound
			}

			scheduleAccessibleUserView = &entity.ScheduleAccessibleUserView{
				ID:   classes[0].ID,
				Name: classes[0].Name,
			}

			return nil
		})
	}

	// get program info
	if schedule.ProgramID != "" {
		g.Go(func() error {
			programs, err := s.programService.BatchGet(ctx, operator, []string{schedule.ProgramID})
			if err != nil {
				log.Error(ctx, "s.programService.BatchGet error",
					log.Err(err),
					log.String("programID", schedule.ProgramID))
				return err
			}

			if len(programs) == 0 {
				log.Error(ctx, "program info not found", log.String("programID", schedule.ProgramID))
				return constant.ErrRecordNotFound
			}

			scheduleProgram = &entity.ScheduleShortInfo{
				ID:   programs[0].ID,
				Name: programs[0].Name,
			}

			return nil
		})
	}

	// get subject info
	if len(subjectIDs) > 0 {
		g.Go(func() error {
			subjects, err := s.subjectService.BatchGet(ctx, operator, subjectIDs)
			if err != nil {
				log.Error(ctx, "s.subjectService.BatchGet error",
					log.Err(err),
					log.Strings("subjectIDs", subjectIDs))
				return err
			}

			for _, subject := range subjects {
				scheduleSubjects = append(scheduleSubjects, &entity.ScheduleShortInfo{
					ID:   subject.ID,
					Name: subject.Name,
				})
			}

			return nil
		})
	}

	// get lesson plan real time status
	if schedule.LessonPlanID != "" {
		g.Go(func() error {
			scheduleRealTimeView, err := s.getLessonPlanAuthed(ctx, operator, schedule.ID, schedule.LessonPlanID)
			if err != nil {
				log.Error(ctx, "s.getLessonPlanAuthed error",
					log.Err(err),
					log.String("ScheduleID", schedule.ID),
					log.String("LessonPlanID", schedule.LessonPlanID),
				)
				return err
			}
			scheduleRealTimeStatus = scheduleRealTimeView

			return nil
		})
	}

	// check if the schedule feedback exists
	g.Go(func() error {
		existFeedback, err := GetScheduleFeedbackModel().ExistByScheduleID(ctx, operator, schedule.ID)
		if err != nil {
			log.Error(ctx, "s.scheduleFeedbackModel.ExistByScheduleID error",
				log.Err(err),
				log.Any("op", operator),
				log.String("scheduleID", schedule.ID))
			return err
		}
		scheduleExistFeedback = existFeedback

		return nil
	})

	// check if the assessment exists, only not homefun homework
	if schedule.ClassType == entity.ScheduleClassTypeHomework && !schedule.IsHomeFun {
		g.Go(func() error {
			existAssessment, err := GetStudyAssessmentModel().BatchCheckAnyoneAttempted(ctx, dbo.MustGetDB(ctx), operator, []string{schedule.ID})
			if err != nil {
				log.Error(ctx, "s.studyAssessmentModel.BatchCheckAnyoneAttempted error",
					log.Err(err),
					log.String("scheduleID", schedule.ID))
				return err
			}
			scheduleExistAssessment = existAssessment[schedule.ID]

			return nil
		})
	}

	// get user info map
	userIDs := append(classRoasterTeacherIDs, classRoasterStudentIDs...)
	userIDs = append(userIDs, participantTeacherIDs...)
	userIDs = append(userIDs, participantStudentIDs...)
	if len(userIDs) > 0 {
		g.Go(func() error {
			users, err := s.userService.BatchGetMap(ctx, operator, userIDs)
			if err != nil {
				log.Error(ctx, "s.userService.BatchGetMap error",
					log.Err(err),
					log.Strings("userIDs", userIDs))
				return err
			}

			userMap = users

			return nil
		})
	}

	// check if the assessment completed, homefun homework
	if schedule.ClassType == entity.ScheduleClassTypeHomework && schedule.IsHomeFun {
		g.Go(func() error {
			var homeFunStudyAssessments []*entity.HomeFunStudy
			err = GetHomeFunStudyModel().Query(ctx, operator, &da.QueryHomeFunStudyCondition{
				ScheduleID: entity.NullString{
					String: schedule.ID,
					Valid:  true,
				},
			}, &homeFunStudyAssessments)
			if err != nil {
				log.Error(ctx, "s.homefunStudyModel.Query error",
					log.Err(err),
					log.Any("scheduleID", schedule.ID))
				return err
			}

			// if anyone student complete assessment, then schedule CompleteAssessment is true
			for _, homeFunStudyAssessment := range homeFunStudyAssessments {
				if homeFunStudyAssessment.Status == entity.AssessmentStatusComplete {
					scheduleCompleteAssessment = true
					break
				}
			}

			return nil
		})
	} else if schedule.ClassType != entity.ScheduleClassTypeTask {
		// check if the assessment completed, not homefun homework, no assessment for the Task
		g.Go(func() error {
			assessments, err := GetAssessmentModel().Query(ctx, operator, &da.QueryAssessmentConditions{
				ScheduleIDs: entity.NullStrings{
					Strings: []string{schedule.ID},
					Valid:   true,
				},
			})
			if err != nil {
				log.Error(ctx, "s.assessmentModel.Query error",
					log.Err(err),
					log.Any("scheduleID", schedule.ID))
				return err
			}

			for _, assessment := range assessments {
				if assessment.Status == entity.AssessmentStatusComplete {
					scheduleCompleteAssessment = true
					break
				}
			}

			return nil
		})
	}

	// check operator schedule permission
	g.Go(func() error {
		permissionMap, err := GetSchedulePermissionModel().HasScheduleOrgPermissions(ctx, operator, []external.PermissionName{
			external.ScheduleCreateEvent,
			external.ScheduleCreateMySchoolEvent,
			external.ScheduleCreateMyEvent,
		})
		if err != nil {
			log.Error(ctx, "GetSchedulePermissionModel().HasScheduleOrgPermissions error",
				log.Any("operator", operator))
			return nil
		}

		schedulePermissionMap = permissionMap

		return nil
	})

	// search for schools where the operator has permission to create schedule
	g.Go(func() error {
		schoolList, err := s.schoolService.GetByPermission(ctx, operator, external.ScheduleCreateMySchoolEvent)
		if err != nil {
			log.Error(ctx, "s.schoolService.GetByPermission error",
				log.Err(err))
			return err
		}

		for _, school := range schoolList {
			scheduleCanBeCreatedSchoolMap[school.ID] = true
		}

		return nil
	})

	// get isClassAccessible
	if schedule.ClassID != "" {
		// search for schools by class, for isClassAccessible
		g.Go(func() error {
			classSchoolMap, err := s.schoolService.GetByClasses(ctx, operator, []string{schedule.ClassID})
			if err != nil {
				log.Error(ctx, "s.schoolService.GetByClasses error",
					log.Err(err),
					log.String("classID", schedule.ClassID))
				return err
			}

			schoolByClass = classSchoolMap[schedule.ClassID]

			return nil
		})

		// search for classes by operator, for isClassAccessible
		g.Go(func() error {
			classes, err := s.classService.GetByUserID(ctx, operator, operator.UserID)
			if err != nil {
				log.Error(ctx, "s.classService.GetByUserID error",
					log.Err(err),
					log.Any("operator", operator))
				return err
			}

			classByOperator = classes

			return nil
		})
	}

	// check participant user is accessible
	participantUserIDs := append(participantStudentIDs, participantTeacherIDs...)
	if len(participantUserIDs) > 0 {
		// get participant user school map
		g.Go(func() error {
			userSchoolMap, err := s.schoolService.GetByUsers(ctx, operator, operator.OrgID, participantUserIDs)
			if err != nil {
				log.Error(ctx, "s.schoolService.GetByUsers error",
					log.Err(err),
					log.Strings("participantUserIDs", participantUserIDs))
				return err
			}

			participantUserSchoolMap = userSchoolMap
			return nil
		})

	}

	if err := g.Wait(); err != nil {
		log.Error(ctx, "transformToScheduleDetailsView error",
			log.Err(err))
		return nil, err
	}

	// fill to scheduleDetailsView
	scheduleDetailsView.LessonPlan = scheduleLessonPlan
	scheduleDetailsView.Class = scheduleAccessibleUserView
	scheduleDetailsView.Program = scheduleProgram
	scheduleDetailsView.Subjects = scheduleSubjects
	if scheduleRealTimeStatus != nil {
		scheduleDetailsView.RealTimeStatus = *scheduleRealTimeStatus
	}
	scheduleDetailsView.ExistFeedback = scheduleExistFeedback
	scheduleDetailsView.ExistAssessment = scheduleExistAssessment
	scheduleDetailsView.CompleteAssessment = scheduleCompleteAssessment

	// get isClassAccessible
	var isClassAccessible bool
	if scheduleDetailsView.Class != nil {
		if schedulePermissionMap[external.ScheduleCreateEvent] {
			isClassAccessible = true
		} else if schedulePermissionMap[external.ScheduleCreateMySchoolEvent] {
			for _, school := range schoolByClass {
				if scheduleCanBeCreatedSchoolMap[school.ID] {
					isClassAccessible = true
					break
				}
			}
		} else if schedulePermissionMap[external.ScheduleCreateMyEvent] {
			for _, class := range classByOperator {
				if class.ID == schedule.ClassID {
					isClassAccessible = true
					break
				}
			}
		}

		scheduleDetailsView.Class.Enable = isClassAccessible
	}

	// fill class_roaster_teachers
	for _, classclassRoasterTeacherID := range classRoasterTeacherIDs {
		if user, ok := userMap[classclassRoasterTeacherID]; ok && user.Valid {
			item := &entity.ScheduleAccessibleUserView{
				ID:     user.ID,
				Name:   user.Name,
				Type:   entity.ScheduleRelationTypeClassRosterTeacher,
				Enable: isClassAccessible,
			}
			scheduleDetailsView.ClassRosterTeachers = append(scheduleDetailsView.ClassRosterTeachers, item)
		}
	}

	// fill class_roaster_students
	for _, classclassRoasterStudentID := range classRoasterStudentIDs {
		if user, ok := userMap[classclassRoasterStudentID]; ok && user.Valid {
			item := &entity.ScheduleAccessibleUserView{
				ID:     user.ID,
				Name:   user.Name,
				Type:   entity.ScheduleRelationTypeClassRosterStudent,
				Enable: isClassAccessible,
			}
			scheduleDetailsView.ClassRosterStudents = append(scheduleDetailsView.ClassRosterStudents, item)
		}
	}

	// fill participants_teachers
	for _, participantTeacherID := range participantTeacherIDs {
		if user, ok := userMap[participantTeacherID]; ok && user.Valid {
			var enable bool
			if schedulePermissionMap[external.ScheduleCreateEvent] {
				enable = true
			} else if schedulePermissionMap[external.ScheduleCreateMySchoolEvent] {
				for _, school := range participantUserSchoolMap[user.ID] {
					if scheduleCanBeCreatedSchoolMap[school.ID] {
						enable = true
						break
					}
				}
			}
			item := &entity.ScheduleAccessibleUserView{
				ID:     user.ID,
				Name:   user.Name,
				Type:   entity.ScheduleRelationTypeParticipantTeacher,
				Enable: enable,
			}
			scheduleDetailsView.ParticipantsTeachers = append(scheduleDetailsView.ParticipantsTeachers, item)
		}
	}

	// fill participants_students
	for _, participantStudentID := range participantStudentIDs {
		if user, ok := userMap[participantStudentID]; ok && user.Valid {
			var enable bool
			if schedulePermissionMap[external.ScheduleCreateEvent] {
				enable = true
			} else if schedulePermissionMap[external.ScheduleCreateMySchoolEvent] {
				for _, school := range participantUserSchoolMap[user.ID] {
					if scheduleCanBeCreatedSchoolMap[school.ID] {
						enable = true
						break
					}
				}
			}
			item := &entity.ScheduleAccessibleUserView{
				ID:     user.ID,
				Name:   user.Name,
				Type:   entity.ScheduleRelationTypeParticipantStudent,
				Enable: enable,
			}
			scheduleDetailsView.ParticipantsStudents = append(scheduleDetailsView.ParticipantsStudents, item)
		}
	}

	scheduleDetailsView.Teachers = append(scheduleDetailsView.Teachers, scheduleDetailsView.ClassRosterTeachers...)
	scheduleDetailsView.Teachers = append(scheduleDetailsView.Teachers, scheduleDetailsView.ParticipantsTeachers...)

	return scheduleDetailsView, nil
}

func (s *scheduleModel) transformToScheduleViewDetail(ctx context.Context, operator *entity.Operator, schedule *entity.Schedule) (*entity.ScheduleViewDetail, error) {
	if schedule == nil {
		log.Debug(ctx, "schedule is nil")
		return nil, nil
	}

	classType := entity.ScheduleShortInfo{
		ID:   schedule.ClassType.String(),
		Name: schedule.ClassType.ToLabel().String(),
	}

	scheduleViewDetail := &entity.ScheduleViewDetail{
		ID:      schedule.ID,
		Title:   schedule.Title,
		StartAt: schedule.StartAt,
		EndAt:   schedule.EndAt,
		DueAt:   schedule.DueAt,
		// Duplicate fields
		ClassType:      classType,
		ClassTypeLabel: classType,
		Status:         schedule.Status,
		IsHomeFun:      schedule.IsHomeFun,
		IsHidden:       schedule.IsHidden,
		RoomID:         schedule.ID,
		IsRepeat:       schedule.RepeatID != "",
		LessonPlanID:   schedule.LessonPlanID,
		Description:    schedule.Description,
		// init empty slice
		OutcomeIDs: []string{},
		Teachers:   []*entity.ScheduleShortInfo{},
		Students:   []*entity.ScheduleShortInfo{},
	}

	// get schedule status, business logic not status in database
	scheduleViewDetail.Status = schedule.Status.GetScheduleStatus(entity.ScheduleStatusInput{
		EndAt:     schedule.EndAt,
		DueAt:     schedule.DueAt,
		ClassType: schedule.ClassType,
	})

	// get attachment
	if schedule.Attachment != "" {
		var attachment entity.ScheduleShortInfo
		err := json.Unmarshal([]byte(schedule.Attachment), &attachment)
		if err != nil {
			log.Error(ctx, "json.Unmarshal error",
				log.Err(err),
				log.String("schedule.Attachment", schedule.Attachment))
			return nil, err
		}
		scheduleViewDetail.Attachment = attachment
	}

	var scheduleRelations []*entity.ScheduleRelation
	err := s.scheduleRelationDA.Query(ctx, &da.ScheduleRelationCondition{
		ScheduleID: sql.NullString{
			String: schedule.ID,
			Valid:  true,
		},
	}, &scheduleRelations)

	var teacherIDs []string
	var studentIDs []string
	var userMap map[string]*external.NullableUser
	for _, scheduleRelation := range scheduleRelations {
		// get operator role type in the schedule
		if scheduleRelation.RelationID == operator.UserID {
			switch scheduleRelation.RelationType {
			case entity.ScheduleRelationTypeParticipantTeacher, entity.ScheduleRelationTypeClassRosterTeacher:
				scheduleViewDetail.RoleType = entity.ScheduleRoleTypeTeacher
			case entity.ScheduleRelationTypeParticipantStudent, entity.ScheduleRelationTypeClassRosterStudent:
				scheduleViewDetail.RoleType = entity.ScheduleRoleTypeStudent
			}
		}

		switch scheduleRelation.RelationType {
		// learning outcome relation, only for homefun homework
		case entity.ScheduleRelationTypeLearningOutcome:
			scheduleViewDetail.OutcomeIDs = append(scheduleViewDetail.OutcomeIDs, scheduleRelation.RelationID)
		case entity.ScheduleRelationTypeClassRosterTeacher, entity.ScheduleRelationTypeParticipantTeacher:
			teacherIDs = append(teacherIDs, scheduleRelation.RelationID)
		case entity.ScheduleRelationTypeClassRosterStudent, entity.ScheduleRelationTypeParticipantStudent:
			studentIDs = append(studentIDs, scheduleRelation.RelationID)
		}
	}

	g := new(errgroup.Group)

	var scheduleLessonPlan *entity.ScheduleLessonPlan
	var scheduleClass *entity.ScheduleShortInfo
	var scheduleExistFeedback bool
	var scheduleExistAssessment bool
	var scheduleCompleteAssessment bool

	// get lesson plan
	if schedule.LessonPlanID != "" {
		g.Go(func() error {
			lessonPlan, err := s.getLessonPlanWithMaterial(ctx, operator, schedule.LessonPlanID)
			if err != nil {
				log.Error(ctx, "s.getLessonPlanWithMaterial error",
					log.Err(err),
					log.String("lessonPlanID", schedule.LessonPlanID))
				return err
			}

			scheduleLessonPlan = lessonPlan

			return nil
		})
	}

	// get class info
	if schedule.ClassID != "" {
		g.Go(func() error {
			classes, err := s.classService.BatchGet(ctx, operator, []string{schedule.ClassID})
			if err != nil {
				log.Error(ctx, "s.classService.BatchGet error",
					log.Err(err),
					log.String("classID", schedule.ClassID))
				return err
			}

			if len(classes) == 0 {
				log.Error(ctx, "class info not found", log.String("classID", schedule.ClassID))
				return constant.ErrRecordNotFound
			}

			scheduleClass = &entity.ScheduleShortInfo{
				ID:   classes[0].ID,
				Name: classes[0].Name,
			}

			return nil
		})
	}

	// get user info map
	userIDs := append(teacherIDs, studentIDs...)
	if len(userIDs) > 0 {
		g.Go(func() error {
			users, err := s.userService.BatchGetMap(ctx, operator, userIDs)
			if err != nil {
				log.Error(ctx, "s.userService.BatchGetMap error",
					log.Err(err),
					log.Strings("userIDs", userIDs))
				return err
			}

			userMap = users

			return nil
		})
	}

	// check if the schedule feedback exists
	g.Go(func() error {
		existFeedback, err := GetScheduleFeedbackModel().ExistByScheduleID(ctx, operator, schedule.ID)
		if err != nil {
			log.Error(ctx, "s.scheduleFeedbackModel.ExistByScheduleID error",
				log.Err(err),
				log.Any("op", operator),
				log.String("scheduleID", schedule.ID))
			return err
		}
		scheduleExistFeedback = existFeedback

		return nil
	})

	// check if the assessment exists, only not homefun homework
	if schedule.ClassType == entity.ScheduleClassTypeHomework && !schedule.IsHomeFun {
		g.Go(func() error {
			existAssessment, err := GetStudyAssessmentModel().BatchCheckAnyoneAttempted(ctx, dbo.MustGetDB(ctx), operator, []string{schedule.ID})
			if err != nil {
				log.Error(ctx, "s.studyAssessmentModel.BatchCheckAnyoneAttempted error",
					log.Err(err),
					log.String("scheduleID", schedule.ID))
				return err
			}
			scheduleExistAssessment = existAssessment[scheduleViewDetail.ID]

			return nil
		})
	}

	// check if the assessment completed, homefun homework
	if schedule.ClassType == entity.ScheduleClassTypeHomework && schedule.IsHomeFun {
		g.Go(func() error {
			var homeFunStudyAssessments []*entity.HomeFunStudy
			err = GetHomeFunStudyModel().Query(ctx, operator, &da.QueryHomeFunStudyCondition{
				ScheduleID: entity.NullString{
					String: schedule.ID,
					Valid:  true,
				},
			}, &homeFunStudyAssessments)
			if err != nil {
				log.Error(ctx, "s.homefunStudyModel.Query error",
					log.Err(err),
					log.Any("scheduleID", schedule.ID))
				return err
			}

			for _, homeFunStudyAssessment := range homeFunStudyAssessments {
				if homeFunStudyAssessment.Status == entity.AssessmentStatusComplete {
					scheduleCompleteAssessment = true
					break
				}
			}

			return nil
		})
	} else if schedule.ClassType != entity.ScheduleClassTypeTask {
		// check if the assessment completed, not homefun homework, no assessment of the Task
		g.Go(func() error {
			assessments, err := GetAssessmentModel().Query(ctx, operator, &da.QueryAssessmentConditions{
				ScheduleIDs: entity.NullStrings{
					Strings: []string{schedule.ID},
					Valid:   true,
				},
			})
			if err != nil {
				log.Error(ctx, "s.assessmentModel.Query error",
					log.Err(err),
					log.Any("scheduleID", schedule.ID))
				return err
			}

			for _, assessment := range assessments {
				if assessment.Status == entity.AssessmentStatusComplete {
					scheduleCompleteAssessment = true
					break
				}
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		log.Error(ctx, "transformToScheduleViewDetail error",
			log.Err(err))
		return nil, err
	}

	// fill to scheduleViewDetail
	scheduleViewDetail.LessonPlan = scheduleLessonPlan
	scheduleViewDetail.Class = scheduleClass
	scheduleViewDetail.ExistFeedback = scheduleExistFeedback
	scheduleViewDetail.ExistAssessment = scheduleExistAssessment
	scheduleViewDetail.CompleteAssessment = scheduleCompleteAssessment

	for _, teacherID := range teacherIDs {
		if user, ok := userMap[teacherID]; ok && user.Valid {
			scheduleViewDetail.Teachers = append(scheduleViewDetail.Teachers, &entity.ScheduleShortInfo{
				ID:   user.ID,
				Name: user.Name,
			})
		} else {
			log.Warn(ctx, "teacher info not found", log.String("teacherID", teacherID))
		}
	}

	for _, studentID := range studentIDs {
		if user, ok := userMap[studentID]; ok && user.Valid {
			scheduleViewDetail.Students = append(scheduleViewDetail.Students, &entity.ScheduleShortInfo{
				ID:   user.ID,
				Name: user.Name,
			})
		} else {
			log.Warn(ctx, "student info not found", log.String("studentID", studentID))
		}
	}

	return scheduleViewDetail, nil
}

func (s *scheduleModel) transformToScheduleListView(ctx context.Context, operator *entity.Operator, scheduleList []*entity.Schedule, loc *time.Location) ([]*entity.ScheduleListView, error) {
	scheduleListView := make([]*entity.ScheduleListView, len(scheduleList))
	var homefunHomeworkIDs []string
	var notHomefunHomeworkIDs []string
	var withAssessmentScheduleIDs []string

	scheduleIDs := make([]string, len(scheduleList))
	for i, schedule := range scheduleList {
		if schedule.ClassType == entity.ScheduleClassTypeHomework {
			if schedule.IsHomeFun {
				homefunHomeworkIDs = append(homefunHomeworkIDs, schedule.ID)
			} else {
				notHomefunHomeworkIDs = append(notHomefunHomeworkIDs, schedule.ID)
				withAssessmentScheduleIDs = append(withAssessmentScheduleIDs, schedule.ID)
			}
		} else if schedule.ClassType != entity.ScheduleClassTypeTask {
			withAssessmentScheduleIDs = append(withAssessmentScheduleIDs, schedule.ID)
		}

		scheduleIDs[i] = schedule.ID
	}

	g := new(errgroup.Group)
	var scheduleExistAssessmentMap map[string]bool
	scheduleCompleteAssessmentMap := make(map[string]bool)
	var scheduleExistFeedbackMap map[string]bool
	scheduleOperatorRoleTypeMap := make(map[string]entity.ScheduleRoleType)

	// check if the assessment exists, only not homefun homework
	if len(notHomefunHomeworkIDs) > 0 {
		g.Go(func() error {
			existAssessment, err := GetStudyAssessmentModel().BatchCheckAnyoneAttempted(ctx, dbo.MustGetDB(ctx), operator, notHomefunHomeworkIDs)
			if err != nil {
				log.Error(ctx, "s.studyAssessmentModel.BatchCheckAnyoneAttempted error",
					log.Err(err),
					log.Strings("notHomefunHomeworkIDs", notHomefunHomeworkIDs))
				return err
			}
			scheduleExistAssessmentMap = existAssessment

			return nil
		})
	}

	// check if the assessment completed
	g.Go(func() error {
		if len(homefunHomeworkIDs) > 0 {
			var homeFunStudyAssessments []*entity.HomeFunStudy
			err := GetHomeFunStudyModel().Query(ctx, operator, &da.QueryHomeFunStudyCondition{
				ScheduleIDs: entity.NullStrings{
					Strings: homefunHomeworkIDs,
					Valid:   true,
				},
			}, &homeFunStudyAssessments)
			if err != nil {
				log.Error(ctx, "s.homefunStudyModel.Query error",
					log.Err(err),
					log.Any("homefunHomeworkIDs", homefunHomeworkIDs))
				return err
			}

			for _, homeFunStudyAssessment := range homeFunStudyAssessments {
				if homeFunStudyAssessment.Status == entity.AssessmentStatusComplete {
					scheduleCompleteAssessmentMap[homeFunStudyAssessment.ScheduleID] = true
				}
			}
		}

		if len(withAssessmentScheduleIDs) > 0 {
			assessments, err := GetAssessmentModel().Query(ctx, operator, &da.QueryAssessmentConditions{
				ScheduleIDs: entity.NullStrings{
					Strings: withAssessmentScheduleIDs,
					Valid:   true,
				},
			})
			if err != nil {
				log.Error(ctx, "s.assessmentModel.Query error",
					log.Err(err),
					log.Any("withAssessmentScheduleIDs", withAssessmentScheduleIDs))
				return err
			}

			for _, assessment := range assessments {
				if assessment.Status == entity.AssessmentStatusComplete {
					scheduleCompleteAssessmentMap[assessment.ScheduleID] = true
				}
			}
		}

		return nil
	})

	// check if the schedule feedback exists
	g.Go(func() error {
		existFeedback, err := GetScheduleFeedbackModel().ExistByScheduleIDs(ctx, operator, scheduleIDs)
		if err != nil {
			log.Error(ctx, "s.scheduleFeedbackModel.ExistByScheduleIDs error",
				log.Err(err),
				log.Any("op", operator),
				log.Strings("scheduleIDs", scheduleIDs))
			return err
		}
		scheduleExistFeedbackMap = existFeedback

		return nil
	})

	// get operator role type in the schedule
	g.Go(func() error {
		var scheduleRelations []*entity.ScheduleRelation
		err := s.scheduleRelationDA.Query(ctx, &da.ScheduleRelationCondition{
			ScheduleIDs: entity.NullStrings{
				Strings: scheduleIDs,
				Valid:   true,
			},
			RelationID: sql.NullString{
				String: operator.UserID,
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
		}, &scheduleRelations)
		if err != nil {
			log.Error(ctx, "s.scheduleRelationDA.Query error",
				log.Err(err),
				log.Strings("ScheduleIDs", scheduleIDs))
			return err
		}

		for _, scheduleRealtion := range scheduleRelations {
			switch scheduleRealtion.RelationType {
			case entity.ScheduleRelationTypeParticipantTeacher, entity.ScheduleRelationTypeClassRosterTeacher:
				scheduleOperatorRoleTypeMap[scheduleRealtion.ScheduleID] = entity.ScheduleRoleTypeTeacher
			case entity.ScheduleRelationTypeParticipantStudent, entity.ScheduleRelationTypeClassRosterStudent:
				scheduleOperatorRoleTypeMap[scheduleRealtion.ScheduleID] = entity.ScheduleRoleTypeStudent
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Error(ctx, "transformToScheduleListView error",
			log.Err(err))
		return nil, err
	}

	for i := range scheduleListView {
		schedule := scheduleList[i]
		item := &entity.ScheduleListView{
			ID:           schedule.ID,
			Title:        schedule.Title,
			StartAt:      schedule.StartAt,
			EndAt:        schedule.EndAt,
			IsRepeat:     schedule.RepeatID != "",
			LessonPlanID: schedule.LessonPlanID,
			ClassID:      schedule.ClassID,
			ClassType:    schedule.ClassType,
			DueAt:        schedule.DueAt,
			IsHidden:     schedule.IsHidden,
			IsHomeFun:    schedule.IsHomeFun,
			ClassTypeLabel: entity.ScheduleShortInfo{
				ID:   schedule.ClassType.String(),
				Name: schedule.ClassType.ToLabel().String(),
			},
			Status: schedule.Status.GetScheduleStatus(entity.ScheduleStatusInput{
				EndAt:     schedule.EndAt,
				DueAt:     schedule.DueAt,
				ClassType: schedule.ClassType,
			}),
			RoleType: entity.ScheduleRoleTypeUnknown,
		}

		// TODO: Perhaps this logic should be handed over to the frontend
		if schedule.ClassType == entity.ScheduleClassTypeHomework && schedule.DueAt != 0 {
			item.StartAt = utils.TodayZeroByTimeStamp(schedule.DueAt, loc).Unix()
			item.EndAt = utils.TodayEndByTimeStamp(schedule.DueAt, loc).Unix()
		}

		item.ExistFeedback = scheduleExistFeedbackMap[schedule.ID]
		item.ExistAssessment = scheduleExistAssessmentMap[schedule.ID]
		item.CompleteAssessment = scheduleCompleteAssessmentMap[schedule.ID]
		if scheduleOperatorRoleTypeMap, ok := scheduleOperatorRoleTypeMap[schedule.ID]; ok {
			item.RoleType = scheduleOperatorRoleTypeMap
		}

		scheduleListView[i] = item
	}

	return scheduleListView, nil
}

// model package interval function
func removeResourceMetadata(ctx context.Context, resourceID string) error {
	if resourceID == "" {
		return nil
	}

	var err error
	log.Debug(ctx, "start removeFileMetadata", log.String("resourceID", resourceID))
	defer log.Debug(ctx, "finish removeFileMetadata", log.String("resourceID", resourceID), log.Err(err))

	parts := strings.Split(resourceID, "-")
	if len(parts) != 2 {
		log.Error(ctx, "invalid resource id", log.String("resourceId", resourceID))
		return ErrInvalidResourceID
	}

	resourcePath := strings.Join(parts, "/")
	queue, err := mq.GetMQ(ctx)
	if err != nil {
		log.Error(ctx, "mq.GetMQ error",
			log.Err(err))
		return err
	}

	topic := entity.KFPSAttachment.Classify()
	err = queue.Publish(ctx, topic, resourcePath)
	if err != nil {
		log.Error(ctx, "queue.Publish error",
			log.String("topic", topic),
			log.String("message", resourcePath),
			log.Err(err))
		return err
	}

	return nil
}

var (
	_scheduleOnce  sync.Once
	_scheduleModel IScheduleModel
)

func GetScheduleModel() IScheduleModel {
	_scheduleOnce.Do(func() {
		_scheduleModel = &scheduleModel{
			scheduleDA:         da.GetScheduleDA(),
			scheduleRelationDA: da.GetScheduleRelationDA(),

			userService:    external.GetUserServiceProvider(),
			schoolService:  external.GetSchoolServiceProvider(),
			classService:   external.GetClassServiceProvider(),
			programService: external.GetProgramServiceProvider(),
			subjectService: external.GetSubjectServiceProvider(),
			teacherService: external.GetTeacherServiceProvider(),
		}
	})

	return _scheduleModel
}
