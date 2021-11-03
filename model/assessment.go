package model

import (
	"context"
	"encoding/json"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

var (
	assessmentModelInstance     IAssessmentModel
	assessmentModelInstanceOnce = sync.Once{}
)

var (
	ErrInvalidOrderByValue   = errors.New("invalid order by value")
	ErrInvalidAssessmentType = errors.New("invalid assessment type")
)

func GetAssessmentModel() IAssessmentModel {
	assessmentModelInstanceOnce.Do(func() {
		assessmentModelInstance = &assessmentModel{
			AmsServices:           external.GetAmsServices(),
			ScheduleModel:         GetScheduleModel(),
			ScheduleRelationModel: GetScheduleRelationModel(),
		}
	})
	return assessmentModelInstance
}

type IAssessmentModel interface {
	Query(ctx context.Context, operator *entity.Operator, conditions *da.QueryAssessmentConditions) ([]*entity.Assessment, error)
	Summary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsSummaryArgs) (*entity.AssessmentsSummary, error)
	StudentQuery(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, conditions *entity.StudentQueryAssessmentConditions) (int, []*entity.StudentAssessment, error)

	ScheduleEndClassCallback(ctx context.Context, operator *entity.Operator, args *entity.AddClassAndLiveAssessmentArgs) error

	PrepareAddInput(ctx context.Context, operator *entity.Operator, input []*entity.AssessmentAddInput) (*entity.BatchAddAssessmentSuperArgs, error)
	BatchAdd(ctx context.Context, operator *entity.Operator, input *entity.BatchAddAssessmentSuperArgs) error
	BatchAddTx(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input *entity.BatchAddAssessmentSuperArgs) error
}

type assessmentModel struct {
	assessmentBase
	AmsServices           external.AmsServices
	ScheduleModel         IScheduleModel
	ScheduleRelationModel IScheduleRelationModel
}

func (m *assessmentModel) ScheduleEndClassCallback(ctx context.Context, operator *entity.Operator, args *entity.AddClassAndLiveAssessmentArgs) error {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixScheduleID, args.ScheduleID)
	if err != nil {
		log.Error(ctx, "add class and live assessment: lock fail",
			log.Err(err),
			log.Any("args", args),
		)
		return err
	}
	locker.Lock()
	defer locker.Unlock()

	superArgs, err := GetAssessmentModel().PrepareAddInput(ctx, operator, []*entity.AssessmentAddInput{
		&entity.AssessmentAddInput{
			ScheduleID:   args.ScheduleID,
			ClassLength:  args.ClassLength,
			ClassEndTime: args.ClassEndTime,
			Attendances:  args.AttendanceIDs,
		}})
	if err != nil {
		return err
	}

	err = GetAssessmentModel().BatchAdd(ctx, operator, superArgs)
	if err != nil {
		return err
	}
	return nil
}

func (m *assessmentModel) PrepareAddInput(ctx context.Context, operator *entity.Operator, input []*entity.AssessmentAddInput) (*entity.BatchAddAssessmentSuperArgs, error) {
	scheduleIDs := make([]string, len(input))
	inputMap := make(map[string]*entity.AssessmentAddInput)
	for i, item := range input {
		scheduleIDs[i] = item.ScheduleID
		inputMap[item.ScheduleID] = item
	}

	// get schedules
	schedules, err := m.ScheduleModel.GetVariableDataByIDs(ctx, operator, scheduleIDs, &entity.ScheduleInclude{
		ClassRosterClass: true,
	})
	if err != nil {
		return nil, err
	}

	// get user
	scheduleUserRelationMap, err := m.ScheduleRelationModel.GetRelationMap(ctx, operator, scheduleIDs, []entity.ScheduleRelationType{
		entity.ScheduleRelationTypeParticipantTeacher,
		entity.ScheduleRelationTypeParticipantStudent,
		entity.ScheduleRelationTypeClassRosterTeacher,
		entity.ScheduleRelationTypeClassRosterStudent,
	})

	assessmentArgs := make([]*entity.AddAssessmentArgs, 0, len(schedules))

	for _, item := range schedules {
		assessmentType, err := entity.GetAssessmentTypeByScheduleType(ctx, entity.GetAssessmentTypeInput{
			ScheduleType: item.ClassType,
			IsHomeFun:    item.IsHomeFun,
		})
		if err != nil {
			return nil, err
		}

		inputMapItem, ok := inputMap[item.ID]
		if !ok {
			return nil, constant.ErrInvalidArgs
		}

		// generate assessment title
		var className string
		if item.ClassRosterClass != nil {
			className = item.ClassRosterClass.Name
		}
		title, err := assessmentType.Title(ctx, entity.GenerateAssessmentTitleInput{
			ClassName:    className,
			ScheduleName: item.Schedule.Title,
			ClassEndTime: inputMapItem.ClassEndTime,
		})
		if err != nil {
			return nil, err
		}

		scheduleUsers, ok := scheduleUserRelationMap[item.ID]
		if !ok {
			return nil, constant.ErrInvalidArgs
		}

		// processing class,live,study type
		assessmentArgsItem := &entity.AddAssessmentArgs{
			Title:         title,
			ScheduleID:    item.ID,
			ScheduleTitle: item.Title,
			LessonPlanID:  item.LessonPlanID,
			ClassID:       item.ClassRosterClass.ID,
			ClassLength:   inputMapItem.ClassLength,
			ClassEndTime:  inputMapItem.ClassEndTime,
		}

		if assessmentType == entity.AssessmentTypeClass || assessmentType == entity.AssessmentTypeStudy {
			assessmentArgsItem.Attendances = scheduleUsers
		} else if assessmentType == entity.AssessmentTypeLive {
			userRelationMap := make(map[string]*entity.ScheduleRelation)
			for _, userItem := range scheduleUsers {
				userRelationMap[userItem.RelationID] = userItem
			}

			for _, userID := range inputMapItem.Attendances {
				if relationItem, ok := userRelationMap[userID]; ok {
					assessmentArgsItem.Attendances = append(assessmentArgsItem.Attendances, relationItem)
				}
			}
		}
		assessmentArgs = append(assessmentArgs, assessmentArgsItem)
	}

	superArgs, err := m.assessmentBase.prepareBatchAddSuperArgs(ctx, dbo.MustGetDB(ctx), operator, assessmentArgs)
	if err != nil {
		log.Error(ctx, "prepare add assessment args: prepare batch add super args failed",
			log.Err(err),
			log.Any("assessmentArgs", assessmentArgs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	return superArgs, nil
}

func (m *assessmentModel) BatchAdd(ctx context.Context, operator *entity.Operator, input *entity.BatchAddAssessmentSuperArgs) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		return m.BatchAddTx(ctx, tx, operator, input)
	})
}
func (m *assessmentModel) BatchAddTx(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input *entity.BatchAddAssessmentSuperArgs) error {
	_, err := m.assessmentBase.batchAdd(ctx, tx, operator, input)
	if err != nil {
		return err
	}
	return nil
}

func (m *assessmentModel) Query(ctx context.Context, operator *entity.Operator, conditions *da.QueryAssessmentConditions) ([]*entity.Assessment, error) {
	r, err := da.GetAssessmentDA().Query(ctx, conditions)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (m *assessmentModel) Summary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsSummaryArgs) (*entity.AssessmentsSummary, error) {
	// check permission
	var (
		checker = NewAssessmentPermissionChecker(operator)
		err     error
	)
	if err = checker.SearchAllPermissions(ctx); err != nil {
		return nil, err
	}
	if args.Status.Valid && !checker.CheckStatus(args.Status.Value) {
		log.Error(ctx, "summary: check status failed",
			log.Any("args", args),
			log.Any("checker", checker),
		)
		return nil, constant.ErrForbidden
	}

	// get assessment list
	var (
		cond = da.QueryAssessmentConditions{
			OrgID: entity.NullString{
				String: operator.OrgID,
				Valid:  true,
			},
			Status: args.Status,
			AllowTeacherIDs: entity.NullStrings{
				Strings: checker.AllowTeacherIDs(),
				Valid:   true,
			},
			AllowTeacherIDAndStatusPairs: entity.NullAssessmentAllowTeacherIDAndStatusPairs{
				Values: checker.allowPairs,
				Valid:  len(checker.allowPairs) > 0,
			},
		}
		teachers []*external.Teacher
	)
	if args.TeacherName.Valid {
		if teachers, err = m.AmsServices.TeacherService.Query(ctx, operator, operator.OrgID, args.TeacherName.String); err != nil {
			log.Error(ctx, "List: external.GetTeacherServiceProvider().Query: query failed",
				log.Err(err),
				log.String("org_id", operator.OrgID),
				log.String("teacher_name", args.TeacherName.String),
				log.Any("args", args),
				log.Any("operator", operator),
			)
			return nil, err
		}
		log.Debug(ctx, "summary: query teachers success",
			log.String("org_id", operator.OrgID),
			log.String("teacher_name", args.TeacherName.String),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		if len(teachers) > 0 {
			cond.TeacherIDs.Valid = true
			for _, item := range teachers {
				cond.TeacherIDs.Strings = append(cond.TeacherIDs.Strings, item.ID)
			}
		} else {
			cond.TeacherIDs.Valid = false
		}
	}

	assessments, err := da.GetAssessmentDA().Query(ctx, &cond)
	if err != nil {
		log.Error(ctx, "summary: query assessments failed",
			log.Err(err),
			log.Any("cond", cond),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}
	if len(assessments) == 0 {
		return nil, nil
	}

	r := entity.AssessmentsSummary{}
	for _, a := range assessments {
		switch a.Status {
		case entity.AssessmentStatusComplete:
			r.Complete++
		case entity.AssessmentStatusInProgress:
			r.InProgress++
		}
	}

	// merge home fun study summary
	r2, err := GetHomeFunStudyModel().Summary(ctx, tx, operator, args)
	if err != nil {
		log.Error(ctx, "summary: get home fun study summary",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, err
	}
	r.InProgress += r2.InProgress
	r.Complete += r2.Complete

	return &r, nil
}

func (m *assessmentModel) StudentQuery(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, condition *entity.StudentQueryAssessmentConditions) (int, []*entity.StudentAssessment, error) {
	assessmentType := entity.AssessmentType(condition.ClassType)
	if !assessmentType.Valid() {
		log.Warn(ctx, "invalid assessment type",
			log.String("assessmentType", assessmentType.String()))
		return 0, nil, ErrInvalidAssessmentType
	}

	scheduleClassType := assessmentType.ToScheduleClassType()
	if scheduleClassType.IsHomeFun {
		//Query assessments
		return m.studentsHomeFunStudyQuery(ctx, operator, tx, scheduleClassType, condition)
	}
	//Query home fun study
	return m.studentsAssessmentQuery(ctx, operator, tx, scheduleClassType, condition)
}

func (m *assessmentModel) studentsAssessmentQuery(ctx context.Context,
	operator *entity.Operator, tx *dbo.DBContext, scheduleClassType entity.AssessmentScheduleType,
	condition *entity.StudentQueryAssessmentConditions) (int, []*entity.StudentAssessment, error) {
	studentIDs := entity.NullStrings{
		Strings: []string{condition.StudentID},
		Valid:   true,
	}
	scheduleIDs := entity.NullStrings{
		Strings: condition.ScheduleIDs,
		Valid:   len(condition.ScheduleIDs) > 0,
	}
	orgID := entity.NullString{
		String: condition.OrgID,
		Valid:  true,
	}
	classType := entity.NullString{
		String: scheduleClassType.ClassType.String(),
		Valid:  scheduleClassType.ClassType != "",
	}
	orderBy := entity.NullAssessmentsOrderBy{
		Value: entity.AssessmentOrderBy(condition.OrderBy),
		Valid: condition.OrderBy != "",
	}
	if orderBy.Valid && !orderBy.Value.Valid() {
		log.Error(ctx, "orderBy invalid",
			log.Any("orderBy", orderBy),
			log.Any("condition", condition),
		)
		return 0, nil, ErrInvalidOrderByValue
	}

	status := entity.NullAssessmentStatus{
		Value: entity.AssessmentStatus(condition.Status),
		Valid: condition.Status != "",
	}

	teacherID := entity.NullStrings{
		Strings: []string{condition.TeacherID},
		Valid:   condition.TeacherID != "",
	}
	createBetween := entity.NullTimeRange{Valid: false}
	if condition.CreatedStartAt > 0 && condition.CreatedEndAt > 0 {
		createBetween.StartAt = condition.CreatedStartAt
		createBetween.EndAt = condition.CreatedEndAt
		createBetween.Valid = true
	}

	updateBetween := entity.NullTimeRange{Valid: false}
	if condition.UpdateStartAt > 0 && condition.UpdateEndAt > 0 {
		updateBetween.StartAt = condition.UpdateStartAt
		updateBetween.EndAt = condition.UpdateEndAt
		updateBetween.Valid = true
	}

	completeBetween := entity.NullTimeRange{Valid: false}
	if condition.CompleteStartAt > 0 && condition.CompleteEndAt > 0 {
		completeBetween.StartAt = condition.CompleteStartAt
		completeBetween.EndAt = condition.CompleteEndAt
		completeBetween.Valid = true
	}
	ids := entity.NullStrings{
		Strings: []string{condition.ID},
		Valid:   condition.ID != "",
	}

	var r []*entity.Assessment

	conditions := &da.QueryAssessmentConditions{
		IDs:                          ids,
		OrgID:                        orgID,
		Status:                       status,
		StudentIDs:                   studentIDs,
		TeacherIDs:                   teacherID,
		ScheduleIDs:                  scheduleIDs,
		AllowTeacherIDAndStatusPairs: entity.NullAssessmentAllowTeacherIDAndStatusPairs{},
		CreatedBetween:               createBetween,
		UpdateBetween:                updateBetween,
		CompleteBetween:              completeBetween,
		ClassType:                    classType,
		OrderBy:                      orderBy,
		Pager:                        utils.GetDboPager(condition.Page, condition.PageSize),
	}
	total, r, err := da.GetAssessmentDA().Page(ctx, conditions)
	if err != nil {
		log.Error(ctx, "StudentQuery:GetAssessmentDA.QueryTx failed",
			log.Err(err),
			log.Any("condition", condition),
		)
		return 0, nil, err
	}
	res, err := m.assessmentsToStudentAssessments(ctx, operator, tx, condition.StudentID, r)
	if err != nil {
		log.Error(ctx, "StudentQuery:assessmentsToStudentAssessments failed",
			log.Err(err),
			log.Any("assessment", r),
		)
		return 0, nil, err
	}
	return total, res, nil
}

func (m *assessmentModel) assessmentsToStudentAssessments(ctx context.Context,
	operator *entity.Operator, tx *dbo.DBContext, studentID string, r []*entity.Assessment) ([]*entity.StudentAssessment, error) {
	res := make([]*entity.StudentAssessment, len(r))
	ids := make([]string, len(r))
	scheduleIDs := make([]string, len(r))
	for i := range r {
		res[i] = &entity.StudentAssessment{
			ID:         r[i].ID,
			Title:      r[i].Title,
			Status:     string(r[i].Status),
			CreateAt:   r[i].CreateAt,
			CompleteAt: r[i].CompleteTime,
			UpdateAt:   r[i].UpdateAt,
			ScheduleID: r[i].ScheduleID,
			StudentID:  studentID,
			IsHomeFun:  false,
		}
		ids[i] = r[i].ID
		scheduleIDs[i] = r[i].ScheduleID
	}
	err := m.fillStudentAssessments(ctx, operator, tx, studentID, res)
	if err != nil {
		log.Error(ctx, "fillStudentAssessments failed",
			log.Err(err),
			log.Any("res", res),
		)
		return nil, err
	}

	return res, nil
}

func (m *assessmentModel) studentsHomeFunStudyQuery(ctx context.Context,
	operator *entity.Operator, tx *dbo.DBContext, scheduleClassType entity.AssessmentScheduleType,
	condition *entity.StudentQueryAssessmentConditions) (int, []*entity.StudentAssessment, error) {
	studentIDs := entity.NullStrings{
		Strings: []string{condition.StudentID},
		Valid:   true,
	}
	classType := entity.NullString{
		String: scheduleClassType.ClassType.String(),
		Valid:  scheduleClassType.ClassType != "",
	}
	teacherIDs := utils.NullSQLJSONStringArray{
		Values: utils.SQLJSONStringArray([]string{condition.TeacherID}),
		Valid:  condition.TeacherID != "",
	}
	scheduleIDs := entity.NullStrings{
		Strings: condition.ScheduleIDs,
		Valid:   len(condition.ScheduleIDs) > 0,
	}
	orgID := entity.NullString{
		String: condition.OrgID,
		Valid:  true,
	}
	orderBy := entity.NullListHomeFunStudiesOrderBy{
		Value: entity.ListHomeFunStudiesOrderBy(condition.OrderBy),
		Valid: condition.OrderBy != "",
	}
	if orderBy.Valid && !orderBy.Value.Valid() {
		log.Error(ctx, "orderBy invalid",
			log.Any("orderBy", orderBy),
			log.Any("condition", condition),
		)
		return 0, nil, ErrInvalidOrderByValue
	}

	status := entity.NullAssessmentStatus{
		Value: entity.AssessmentStatus(condition.Status),
		Valid: condition.Status != "",
	}

	createBetween := entity.NullTimeRange{Valid: false}
	if condition.CreatedStartAt > 0 && condition.CreatedEndAt > 0 {
		createBetween.StartAt = condition.CreatedStartAt
		createBetween.EndAt = condition.CreatedEndAt
		createBetween.Valid = true
	}

	updateBetween := entity.NullTimeRange{Valid: false}
	if condition.UpdateStartAt > 0 && condition.UpdateEndAt > 0 {
		updateBetween.StartAt = condition.UpdateStartAt
		updateBetween.EndAt = condition.UpdateEndAt
		updateBetween.Valid = true
	}

	completeBetween := entity.NullTimeRange{Valid: false}
	if condition.CompleteStartAt > 0 && condition.CompleteEndAt > 0 {
		completeBetween.StartAt = condition.CompleteStartAt
		completeBetween.EndAt = condition.CompleteEndAt
		completeBetween.Valid = true
	}
	ids := entity.NullStrings{
		Strings: []string{condition.ID},
		Valid:   condition.ID != "",
	}

	//Query home fun study
	var r []*entity.HomeFunStudy

	conditions := &da.QueryHomeFunStudyCondition{
		IDs:             ids,
		OrgID:           orgID,
		Status:          status,
		StudentIDs:      studentIDs,
		TeacherIDs:      teacherIDs,
		ScheduleIDs:     scheduleIDs,
		CreatedBetween:  createBetween,
		UpdateBetween:   updateBetween,
		CompleteBetween: completeBetween,
		ClassType:       classType,
		OrderBy:         orderBy,
		Pager:           utils.GetDboPager(condition.Page, condition.PageSize),
	}

	total, err := da.GetHomeFunStudyDA().PageTx(ctx, tx, conditions, &r)
	if err != nil {
		log.Error(ctx, "StudentQuery:GetHomeFunStudyDA.QueryTx failed",
			log.Err(err),
			log.Any("condition", condition),
		)
		return 0, nil, err
	}
	res, err := m.homeFunStudyToStudentAssessments(ctx, operator, tx, condition.StudentID, r)
	if err != nil {
		log.Error(ctx, "StudentQuery:assessmentsToStudentAssessments failed",
			log.Err(err),
			log.Any("assessment", r),
		)
		return 0, nil, err
	}
	return total, res, nil
}

func (m *assessmentModel) homeFunStudyToStudentAssessments(ctx context.Context,
	operator *entity.Operator, tx *dbo.DBContext, studentID string, r []*entity.HomeFunStudy) ([]*entity.StudentAssessment, error) {
	res := make([]*entity.StudentAssessment, len(r))
	for i := range r {
		teacherIDs := []string(r[i].TeacherIDs)
		res[i] = &entity.StudentAssessment{
			ID:         r[i].ID,
			Title:      r[i].Title,
			Status:     string(r[i].Status),
			CreateAt:   r[i].CreateAt,
			UpdateAt:   r[i].UpdateAt,
			CompleteAt: r[i].CompleteAt,
			CompleteBy: r[i].CompleteBy,
			ScheduleID: r[i].ScheduleID,
			Comment:    r[i].AssessComment,
			Score:      int(r[i].AssessScore),
			FeedbackID: r[i].LatestFeedbackID,
			StudentID:  studentID,
			TeacherIDs: teacherIDs,
			IsHomeFun:  true,
		}
	}
	err := m.fillStudentAssessments(ctx, operator, tx, studentID, res)
	if err != nil {
		log.Error(ctx, "fillStudentAssessments failed",
			log.Err(err),
			log.Any("res", res),
		)
		return nil, err
	}
	return res, nil
}

func (m *assessmentModel) fillStudentAssessments(ctx context.Context,
	operator *entity.Operator,
	tx *dbo.DBContext,
	studentID string,
	assessments []*entity.StudentAssessment) error {
	//collect schedule ID & assessment ID
	collectedIDs := m.collectAssessmentsRelatedIDs(ctx, assessments)

	//query schedules
	schedulesMap, err := m.querySchedulesMap(ctx, collectedIDs.ScheduleIDs)
	if err != nil {
		log.Error(ctx, "querySchedulesMap failed",
			log.Err(err),
			log.Strings("scheduleIDs", collectedIDs.ScheduleIDs),
		)
		return err
	}

	noneHomeFunScheduleIDs := make([]string, 0, len(schedulesMap))
	for _, item := range schedulesMap {
		if item.ClassType == entity.ScheduleClassTypeHomework && item.IsHomeFun {
			continue
		}
		noneHomeFunScheduleIDs = append(noneHomeFunScheduleIDs, item.ID)
	}

	//query Assessment Comments
	commentMap, err := m.queryAssessmentComments(ctx, operator, noneHomeFunScheduleIDs, studentID)
	if err != nil {
		log.Error(ctx, "queryAssessmentComments failed",
			log.Err(err),
			log.Strings("collectedIDs.ScheduleIDs", collectedIDs.ScheduleIDs),
			log.String("studentID", studentID),
		)
		return err
	}

	//query teachers info in assessments
	teacherAssessmentsMap, teacherInfoMap, err := m.queryTeacherMap(ctx,
		operator,
		tx,
		assessments,
		collectedIDs.AllAssessmentIDs)
	if err != nil {
		log.Error(ctx, "GetTeacherServiceProvider.BatchGetNameMap failed",
			log.Err(err),
			log.Strings("assessmentIDs", collectedIDs.AllAssessmentIDs),
		)
		return err
	}

	//query feedbacks
	feedbackMap, err := m.queryFeedbackInfo(ctx, operator, collectedIDs.FeedbackIDs)
	if err != nil {
		log.Error(ctx, "queryFeedbackInfo failed",
			log.Err(err),
			log.Strings("feedbackIDs", collectedIDs.FeedbackIDs),
		)
		return err
	}

	//fill assessments
	err = m.buildStudentAssessments(ctx,
		assessments,
		schedulesMap,
		teacherInfoMap,
		teacherAssessmentsMap,
		feedbackMap,
		commentMap)
	if err != nil {
		log.Error(ctx, "buildStudentAssessments failed",
			log.Err(err),
			log.Any("assessments", assessments),
			log.Any("schedulesMap", schedulesMap),
			log.Any("teacherInfoMap", teacherInfoMap),
			log.Any("teacherAssessmentsMap", teacherAssessmentsMap),
			log.Any("feedbackMap", feedbackMap),
		)
		return err
	}

	return nil
}

func (m *assessmentModel) isCommentNil(ctx context.Context,
	assessment *entity.StudentAssessment,
	scheduleCommentMap map[string]map[string]string,
	teacherID string) bool {
	//home fun comment is in assessment comment
	if assessment.IsHomeFun {
		return assessment.Comment == ""
	}

	//query comment for teacher
	_, exists := scheduleCommentMap[assessment.ScheduleID][teacherID]
	return !exists
}

func (m *assessmentModel) buildStudentAssessments(ctx context.Context,
	assessments []*entity.StudentAssessment,
	schedulesMap map[string]*entity.Schedule,
	teacherInfoMap map[string]*external.NullableUser,
	teacherAssessmentsMap map[string][]string,
	feedbackMap map[string][]*entity.FeedbackAssignmentView,
	scheduleCommentMap map[string]map[string]string) error {

	for i := range assessments {
		//build schedule
		schedule := schedulesMap[assessments[i].ScheduleID]
		if schedule != nil {
			scheduleAttachment := new(entity.StudentAssessmentAttachment)
			err := json.Unmarshal([]byte(schedule.Attachment), scheduleAttachment)
			if err != nil {
				log.Error(ctx, "Unmarshal schedule attachment failed",
					log.Err(err),
					log.Any("schedule", schedule),
				)
				return err
			}
			assessments[i].Schedule = &entity.StudentAssessmentSchedule{
				ID:         schedule.ID,
				Title:      schedule.Title,
				Type:       schedule.ClassType.String(),
				Attachment: scheduleAttachment,
			}
		}

		//build teacher
		assessmentTeacherIDs := teacherAssessmentsMap[assessments[i].ID]
		assessments[i].TeacherComments = make([]*entity.StudentAssessmentTeacher, 0, len(assessmentTeacherIDs))
		for j := range assessmentTeacherIDs {
			teacherID := assessmentTeacherIDs[j]
			if m.isCommentNil(ctx, assessments[i], scheduleCommentMap, teacherID) {
				continue
			}
			teacherComment := &entity.StudentAssessmentTeacher{
				Teacher: &entity.StudentAssessmentTeacherInfo{
					ID: teacherID,
				},
			}
			teacherInfo := teacherInfoMap[assessmentTeacherIDs[j]]
			if teacherInfo != nil && teacherInfo.Valid {
				teacherComment.Teacher.GivenName = teacherInfo.GivenName
				teacherComment.Teacher.FamilyName = teacherInfo.FamilyName
				teacherComment.Teacher.Avatar = teacherInfo.Avatar
			}

			//home fun comment is in assessment comment
			if assessments[i].IsHomeFun {
				teacherComment.Comment = assessments[i].Comment
			} else {
				//query comment for teacher
				comment, _ := scheduleCommentMap[assessments[i].ScheduleID][teacherID]
				teacherComment.Comment = comment
			}

			assessments[i].TeacherComments = append(assessments[i].TeacherComments, teacherComment)
		}

		//build student attachments
		if assessments[i].FeedbackID != "" {
			assessments[i].FeedbackAttachments = make([]entity.StudentAssessmentAttachment, len(feedbackMap[assessments[i].FeedbackID]))
			for j := range feedbackMap[assessments[i].FeedbackID] {
				feedbackInfo := feedbackMap[assessments[i].FeedbackID][j]
				assessments[i].FeedbackAttachments[j] = entity.StudentAssessmentAttachment{
					ID:   feedbackInfo.AttachmentID,
					Name: feedbackInfo.AttachmentName,
				}
			}
		}
	}
	return nil
}

func (m *assessmentModel) queryAssessmentComments(ctx context.Context, operator *entity.Operator, scheduleIDs []string, studentID string) (map[string]map[string]string, error) {
	commentMap, err := getAssessmentH5P().batchGetRoomCommentObjectMap(ctx, operator, scheduleIDs)
	if err != nil {
		log.Error(ctx, "getAssessmentH5p.batchGetRoomCommentMap failed",
			log.Err(err),
			log.Strings("scheduleIDs", scheduleIDs),
		)
		return nil, err
	}
	comments := make(map[string]map[string]string)
	for i := range scheduleIDs {
		if commentMap[scheduleIDs[i]] != nil {
			studentComments := commentMap[scheduleIDs[i]][studentID]
			comments[scheduleIDs[i]] = make(map[string]string)
			for j := range studentComments {
				if studentComments[j] == nil {
					continue
				}
				log.Debug(ctx, "test info",
					log.Any("comments", comments),
					log.Any("scheduleID", scheduleIDs[i]),
					log.Any("studentComment", studentComments[j]))
				comments[scheduleIDs[i]][studentComments[j].TeacherID] = studentComments[j].Comment
			}
		}
	}
	return comments, nil
}

func (m *assessmentModel) collectAssessmentsRelatedIDs(ctx context.Context, assessments []*entity.StudentAssessment) entity.StudentCollectRelatedIDs {
	scheduleIDs := make([]string, len(assessments))
	allAssessmentIDs := make([]string, len(assessments))
	normalAssessmentIDs := make([]string, 0)
	feedbackIDs := make([]string, 0)

	for i := range assessments {
		scheduleIDs[i] = assessments[i].ScheduleID
		allAssessmentIDs[i] = assessments[i].ID
		if assessments[i].FeedbackID != "" {
			feedbackIDs = append(feedbackIDs, assessments[i].FeedbackID)
		}
		if !assessments[i].IsHomeFun {
			normalAssessmentIDs = append(normalAssessmentIDs, assessments[i].ID)
		}
	}

	//deduplicate ids
	scheduleIDs = utils.SliceDeduplication(scheduleIDs)
	allAssessmentIDs = utils.SliceDeduplication(allAssessmentIDs)
	return entity.StudentCollectRelatedIDs{
		ScheduleIDs:      scheduleIDs,
		AllAssessmentIDs: allAssessmentIDs,
		FeedbackIDs:      feedbackIDs,
		AssessmentsIDs:   normalAssessmentIDs,
	}
}

func (m *assessmentModel) queryFeedbackInfo(ctx context.Context, operator *entity.Operator, feedbackIDs []string) (map[string][]*entity.FeedbackAssignmentView, error) {
	//query feedbacks
	var err error
	feedbackMap := make(map[string][]*entity.FeedbackAssignmentView)
	if len(feedbackIDs) > 0 {
		feedbackMap, err = GetFeedbackAssignmentModel().QueryMap(ctx, operator, &da.FeedbackAssignmentCondition{
			FeedBackIDs: entity.NullStrings{
				Strings: feedbackIDs,
				Valid:   true,
			},
		})
		if err != nil {
			log.Error(ctx, "GetTeacherServiceProvider.BatchGetNameMap failed",
				log.Err(err),
				log.Strings("feedbackIDs", feedbackIDs),
			)
			return nil, err
		}
	}
	return feedbackMap, nil
}

func (m *assessmentModel) queryTeacherMap(ctx context.Context,
	operator *entity.Operator,
	tx *dbo.DBContext,
	assessments []*entity.StudentAssessment,
	assessmentIDs []string) (map[string][]string, map[string]*external.NullableUser, error) {
	//query teachers in assessments
	teacherAssessmentsMap := make(map[string][]string)
	attendances := make([]*entity.AssessmentAttendance, 0)
	err := da.GetAssessmentAttendanceDA().QueryTx(ctx, tx, &da.QueryAssessmentAttendanceConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		Role: entity.NullAssessmentAttendanceRole{
			Value: entity.AssessmentAttendanceRoleTeacher,
			Valid: true,
		},
	}, &attendances)
	if err != nil {
		log.Error(ctx, "GetAssessmentAttendanceDA.QueryTx failed",
			log.Err(err),
			log.Strings("assessmentIDs", assessmentIDs),
		)
		return nil, nil, err
	}

	//collect teacher
	teacherIDs := make([]string, 0)
	for i := range attendances {
		teacherAssessmentsMap[attendances[i].AssessmentID] = append(teacherAssessmentsMap[attendances[i].AssessmentID], attendances[i].AttendanceID)
		teacherIDs = append(teacherIDs, attendances[i].AttendanceID)
	}

	//collect home fun study teachers
	for i := range assessments {
		if assessments[i].IsHomeFun && assessments[i].CompleteBy != "" {
			teacherAssessmentsMap[assessments[i].ID] = append(teacherAssessmentsMap[assessments[i].ID], assessments[i].CompleteBy)
			teacherIDs = append(teacherIDs, assessments[i].CompleteBy)
		} else if !assessments[i].IsHomeFun {
			teacherAssessmentsMap[assessments[i].ID] = append(teacherAssessmentsMap[assessments[i].ID], assessments[i].TeacherIDs...)
		}
		teacherIDs = append(teacherIDs, assessments[i].TeacherIDs...)
	}

	//query teacher info
	teacherInfoMap, err := m.AmsServices.UserService.BatchGetMap(ctx, operator, teacherIDs)
	if err != nil {
		log.Error(ctx, "GetTeacherServiceProvider.BatchGetNameMap failed",
			log.Err(err),
			log.Strings("teacherIDs", teacherIDs),
		)
		return nil, nil, err
	}
	return teacherAssessmentsMap, teacherInfoMap, nil
}

func (m *assessmentModel) querySchedulesMap(ctx context.Context, scheduleIDs []string) (map[string]*entity.Schedule, error) {
	schedules, err := GetScheduleModel().QueryUnsafe(ctx, &entity.ScheduleQueryCondition{IDs: entity.NullStrings{
		Strings: scheduleIDs,
		Valid:   true,
	}})
	if err != nil {
		log.Error(ctx, "GetScheduleModel.QueryUnsafe failed",
			log.Err(err),
			log.Strings("scheduleIDs", scheduleIDs),
		)
		return nil, err
	}
	schedulesMap := make(map[string]*entity.Schedule)
	for i := range schedules {
		schedulesMap[schedules[i].ID] = schedules[i]
	}
	return schedulesMap, nil
}
