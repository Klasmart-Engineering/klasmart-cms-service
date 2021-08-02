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
		assessmentModelInstance = &assessmentModel{}
	})
	return assessmentModelInstance
}

type IAssessmentModel interface {
	Query(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, conditions *da.QueryAssessmentConditions) ([]*entity.Assessment, error)
	Summary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsSummaryArgs) (*entity.AssessmentsSummary, error)

	StudentQuery(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, conditions *entity.StudentQueryAssessmentConditions) (int, []*entity.StudentAssessment, error)
}

type assessmentModel struct {
	assessmentBase
}

func (m *assessmentModel) Query(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, conditions *da.QueryAssessmentConditions) ([]*entity.Assessment, error) {
	var r []*entity.Assessment
	if err := da.GetAssessmentDA().QueryTx(ctx, tx, conditions, &r); err != nil {
		log.Error(ctx, "query assessments failed",
			log.Err(err),
			log.Any("conditions", conditions),
			log.Any("operator", operator),
		)
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
		assessments []*entity.Assessment
		cond        = da.QueryAssessmentConditions{
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
		if teachers, err = external.GetTeacherServiceProvider().Query(ctx, operator, operator.OrgID, args.TeacherName.String); err != nil {
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

	if err := da.GetAssessmentDA().QueryTx(ctx, tx, &cond, &assessments); err != nil {
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
	if condition.Page < 1 {
		condition.Page = 1
	}
	if condition.PageSize < 1 {
		condition.PageSize = 10
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
		completeBetween.EndAt = condition.CompleteStartAt
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
		AllowTeacherIDAndStatusPairs: entity.NullAssessmentAllowTeacherIDAndStatusPairs{},
		CreatedBetween:               createBetween,
		UpdateBetween:                updateBetween,
		CompleteBetween:              completeBetween,
		ClassType:                    classType,
		OrderBy:                      orderBy,
		Pager: dbo.Pager{
			Page:     condition.Page,
			PageSize: condition.PageSize,
		},
	}
	total, err := da.GetAssessmentDA().PageTx(ctx, tx, conditions, &r)
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
		completeBetween.EndAt = condition.CompleteStartAt
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
		CreatedBetween:  createBetween,
		UpdateBetween:   updateBetween,
		CompleteBetween: completeBetween,
		ClassType:       classType,
		OrderBy:         orderBy,
		Pager: dbo.Pager{
			Page:     condition.Page,
			PageSize: condition.PageSize,
		},
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
		comments := make([]string, 0)
		if r[i].AssessComment != "" {
			comments = append(comments, r[i].AssessComment)
		}
		teacherIDs := []string(r[i].TeacherIDs)
		res[i] = &entity.StudentAssessment{
			ID:         r[i].ID,
			Title:      r[i].Title,
			Status:     string(r[i].Status),
			CreateAt:   r[i].CreateAt,
			UpdateAt:   r[i].UpdateAt,
			CompleteAt: r[i].CompleteAt,
			ScheduleID: r[i].ScheduleID,
			Comment:    comments,
			Score:      int(r[i].AssessScore),
			FeedbackID: r[i].AssessFeedbackID,
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

	//query Assessment Comments
	commentMap, err := m.queryAssessmentComments(ctx, operator, collectedIDs.ScheduleIDs, studentID)
	if err != nil {
		log.Error(ctx, "queryAssessmentComments failed",
			log.Err(err),
			log.Strings("collectedIDs.ScheduleIDs", collectedIDs.ScheduleIDs),
			log.String("studentID", studentID),
		)
		return err
	}

	//query teachers info in assessments
	teacherAssessmentsMap, teacherInfoMap, err := m.queryTeacherMap(ctx, operator, tx, assessments, collectedIDs.AllAssessmentIDs)
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

func (m *assessmentModel) buildStudentAssessments(ctx context.Context,
	assessments []*entity.StudentAssessment,
	schedulesMap map[string]*entity.Schedule,
	teacherInfoMap map[string]*external.NullableUser,
	teacherAssessmentsMap map[string][]string,
	feedbackMap map[string][]*entity.FeedbackAssignmentView,
	scheduleCommentMap map[string][]string) error {

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
		assessments[i].Teachers = make([]*entity.StudentAssessmentTeacher, len(assessmentTeacherIDs))
		for j := range assessmentTeacherIDs {
			assessments[i].Teachers[j] = &entity.StudentAssessmentTeacher{
				ID: assessmentTeacherIDs[j],
			}
			teacherInfo := teacherInfoMap[assessmentTeacherIDs[j]]
			if teacherInfo != nil && teacherInfo.Valid {
				assessments[i].Teachers[j].GivenName = teacherInfo.GivenName
				assessments[i].Teachers[j].FamilyName = teacherInfo.FamilyName
			}
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

		if !assessments[i].IsHomeFun {
			assessments[i].Comment = scheduleCommentMap[assessments[i].ScheduleID]
		}
	}
	return nil
}

func (m *assessmentModel) queryAssessmentComments(ctx context.Context, operator *entity.Operator, scheduleIDs []string, studentID string) (map[string][]string, error) {
	commentMap, err := getAssessmentH5P().batchGetRoomCommentMap(ctx, operator, scheduleIDs)
	if err != nil {
		log.Error(ctx, "getAssessmentH5p.batchGetRoomCommentMap failed",
			log.Err(err),
			log.Strings("scheduleIDs", scheduleIDs),
		)
		return nil, err
	}
	comments := make(map[string][]string)
	for i := range scheduleIDs {
		if commentMap[scheduleIDs[i]] != nil {
			comments[scheduleIDs[i]] = commentMap[scheduleIDs[i]][studentID]
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
		teacherAssessmentsMap[assessments[i].ID] = append(teacherAssessmentsMap[assessments[i].ID], assessments[i].TeacherIDs...)
		teacherIDs = append(teacherIDs, assessments[i].TeacherIDs...)
	}

	//query teacher info
	teacherInfoMap, err := external.GetUserServiceProvider().BatchGetMap(ctx, operator, teacherIDs)
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
