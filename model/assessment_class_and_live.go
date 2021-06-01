package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sort"
	"sync"
	"time"
)

var (
	ErrNotFoundAttendance     = errors.New("not found attendance")
	ErrAssessmentHasCompleted = errors.New("assessment has completed")

	classAndLiveAssessmentModelInstance     IClassAndLiveAssessmentModel
	classAndLiveAssessmentModelInstanceOnce = sync.Once{}
)

func GetClassAndLiveAssessmentModel() IClassAndLiveAssessmentModel {
	classAndLiveAssessmentModelInstanceOnce.Do(func() {
		classAndLiveAssessmentModelInstance = &classAndLiveAssessmentModel{}
	})
	return classAndLiveAssessmentModelInstance
}

type IClassAndLiveAssessmentModel interface {
	GetDetail(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string) (*entity.AssessmentDetail, error)
	Summary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsSummaryArgs) (*entity.AssessmentsSummary, error)
	List(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsArgs) (*entity.ListAssessmentsResult, error)
	Add(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AddAssessmentArgs) (string, error)
	Update(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.UpdateAssessmentArgs) error
}

type classAndLiveAssessmentModel struct{}

func (m *classAndLiveAssessmentModel) GetDetail(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string) (*entity.AssessmentDetail, error) {
	assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "Get: da.GetAssessmentDA().GetExcludeSoftDeleted: get failed",
			log.Err(err),
			log.String("assessment_id", id),
		)
		return nil, err
	}

	// convert to assessment view
	var (
		views []*entity.AssessmentView
		view  *entity.AssessmentView
	)
	if views, err = GetAssessmentUtils().ToViews(ctx, tx, operator, []*entity.Assessment{assessment}, entity.ConvertToViewsOptions{
		EnableProgram:    true,
		EnableSubjects:   true,
		EnableTeachers:   true,
		EnableStudents:   true,
		EnableClass:      true,
		EnableLessonPlan: true,
	}); err != nil {
		log.Error(ctx, "Get: GetAssessmentUtils().ToViews: get failed",
			log.Err(err),
			log.String("assessment_id", id),
			log.Any("operator", operator),
		)
		return nil, err
	}
	view = views[0]

	// fill result
	result := entity.AssessmentDetail{
		ID:           assessment.ID,
		ScheduleID:   view.ScheduleID,
		RoomID:       view.RoomID,
		Title:        assessment.Title,
		Class:        view.Class,
		Status:       assessment.Status,
		CompleteTime: assessment.CompleteTime,
		Teachers:     view.Teachers,
		Students:     view.Students,
		Program:      view.Program,
		Subjects:     view.Subjects,
		ClassEndTime: assessment.ClassEndTime,
		ClassLength:  assessment.ClassLength,
		DueAt:        view.Schedule.DueAt,
	}

	// fill lesson plan and lesson materials
	var contentIDs []string
	lp, err := da.GetAssessmentContentDA().GetLessonPlan(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "get assessment detail: get lesson plan failed",
			log.Err(err),
			log.String("assessment_id", id),
		)
		return nil, err
	} else {
		contentIDs = append(contentIDs, lp.ContentID)
	}
	lms, err := da.GetAssessmentContentDA().GetLessonMaterials(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "Get: da.GetAssessmentContentDA().GetLessonMaterials: get failed",
			log.Err(err),
			log.String("assessment_id", id),
		)
		return nil, err
	} else {
		for _, lm := range lms {
			contentIDs = append(contentIDs, lm.ContentID)
		}
	}
	var currentContentOutcomeMap map[string][]string
	if len(contentIDs) > 0 {
		assessmentContentOutcomeMap, err := m.getAssessmentContentOutcomeMap(ctx, tx, []string{id}, contentIDs)
		if err != nil {
			log.Error(ctx, "Get: m.getAssessmentContentOutcomeMap: get failed",
				log.Err(err),
				log.String("assessment_id", id),
				log.Strings("content_ids", contentIDs),
			)
			return nil, err
		}
		currentContentOutcomeMap = assessmentContentOutcomeMap[id]
		if lp != nil {
			result.LessonPlan = entity.AssessmentDetailContent{
				ID:         lp.ContentID,
				Name:       lp.ContentName,
				Checked:    true,
				OutcomeIDs: currentContentOutcomeMap[lp.ContentID],
			}
		}
		for _, m := range lms {
			result.LessonMaterials = append(result.LessonMaterials, &entity.AssessmentDetailContent{
				ID:         m.ContentID,
				Name:       m.ContentName,
				Comment:    m.ContentComment,
				Checked:    m.Checked,
				OutcomeIDs: currentContentOutcomeMap[m.ContentID],
			})
		}
	}

	// fill outcomes
	var assessmentOutcomes []*entity.AssessmentOutcome
	if err := da.GetAssessmentOutcomeDA().QueryTx(ctx, tx, &da.QueryAssessmentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: []string{id},
			Valid:   true,
		},
	}, &assessmentOutcomes); err != nil {
		log.Error(ctx, "Get: da.GetAssessmentOutcomeDA().GetListByAssessmentID: get list failed",
			log.Err(err),
			log.String("id", id),
		)
		return nil, err
	}
	if len(assessmentOutcomes) > 0 {
		var (
			assessmentOutcomeMap = make(map[string]entity.AssessmentOutcome, len(assessmentOutcomes))
			outcomeIDs           = make([]string, 0, len(assessmentOutcomes))
			outcomes             = make([]*entity.Outcome, 0, len(assessmentOutcomes))
		)
		for _, o := range assessmentOutcomes {
			assessmentOutcomeMap[o.OutcomeID] = *o
			outcomeIDs = append(outcomeIDs, o.OutcomeID)
		}
		if outcomes, err = GetOutcomeModel().GetByIDs(ctx, operator, tx, outcomeIDs); err != nil {
			log.Error(ctx, "Get: GetOutcomeModel().GetByIDs: get failed",
				log.Err(err),
				log.Strings("outcome_ids", outcomeIDs),
				log.String("assessment_id", id),
				log.Any("operator", operator),
			)
			return nil, err
		}
		sort.Sort(OutcomesOrderByAssumedAndName(outcomes))

		var (
			outcomeAttendances      = make([]*entity.OutcomeAttendance, 0, len(outcomeIDs))
			outcomeAttendanceIDsMap = make(map[string][]string, len(outcomeIDs))
		)
		outcomeAttendances, err := da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDAndOutcomeIDs(ctx, tx, id, outcomeIDs)
		if err != nil {
			log.Error(ctx, "Get: da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDAndOutcomeIDs: batch get failed",
				log.Err(err),
				log.Strings("outcome_ids", outcomeIDs),
				log.String("assessment_id", id),
				log.Any("operator", operator),
			)
			return nil, err
		}
		for _, item := range outcomeAttendances {
			outcomeAttendanceIDsMap[item.OutcomeID] = append(outcomeAttendanceIDsMap[item.OutcomeID], item.AttendanceID)
		}
		for _, o := range outcomes {
			newOutcome := entity.AssessmentDetailOutcome{
				OutcomeID:     o.ID,
				OutcomeName:   o.Name,
				Assumed:       o.Assumed,
				Skip:          assessmentOutcomeMap[o.ID].Skip,
				NoneAchieved:  assessmentOutcomeMap[o.ID].NoneAchieved,
				AttendanceIDs: outcomeAttendanceIDsMap[o.ID],
				Checked:       assessmentOutcomeMap[o.ID].Checked,
			}
			result.Outcomes = append(result.Outcomes, &newOutcome)
		}
	}

	// fill remaining time
	result.RemainingTime = GetAssessmentUtils().CalcRemainingTime(view.Schedule.DueAt, view.CreateAt)

	// fill student view items
	result.StudentViewItems, err = GetAssessmentUtils().GetH5PStudentViewItems(ctx, operator, view)
	if err != nil {
		log.Error(ctx, "get assessment detail: get student view items failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("view", view),
		)
		return nil, err
	}

	return &result, nil
}

func (m *classAndLiveAssessmentModel) Summary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsSummaryArgs) (*entity.AssessmentsSummary, error) {
	// check permission
	var (
		checker = NewAssessmentPermissionChecker(operator)
		err     error
	)
	if err = checker.SearchAllPermissions(ctx); err != nil {
		return nil, err
	}
	if args.Status.Valid && !checker.CheckStatus(args.Status.Value) {
		log.Error(ctx, "get outcome summary: check status failed",
			log.Any("args", args),
			log.Any("checker", checker),
		)
		return nil, constant.ErrForbidden
	}

	// get assessment list
	var (
		assessments []*entity.Assessment
		cond        = da.QueryAssessmentConditions{
			ClassTypes: entity.NullScheduleClassTypes{
				Value: []entity.ScheduleClassType{entity.ScheduleClassTypeOnlineClass, entity.ScheduleClassTypeOfflineClass},
				Valid: false,
			},
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
		teachers    []*external.Teacher
		scheduleIDs []string
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
		log.Debug(ctx, "List: external.GetTeacherServiceProvider().Query: query success",
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
	if scheduleIDs, err = GetScheduleModel().GetScheduleIDsByOrgID(ctx, tx, operator, operator.OrgID); err != nil {
		log.Error(ctx, "List: GetScheduleModel().GetScheduleIDsByOrgID: get failed",
			log.Err(err),
			log.String("org_id", operator.OrgID),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}
	cond.ScheduleIDs = entity.NullStrings{
		Strings: scheduleIDs,
		Valid:   true,
	}

	if err := da.GetAssessmentDA().QueryTx(ctx, tx, &cond, &assessments); err != nil {
		log.Error(ctx, "List: da.GetAssessmentDA().QueryTx: query failed",
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

	return &r, nil
}

func (m *classAndLiveAssessmentModel) List(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsArgs) (*entity.ListAssessmentsResult, error) {
	// check permission
	var (
		checker = NewAssessmentPermissionChecker(operator)
		err     error
	)
	if err = checker.SearchAllPermissions(ctx); err != nil {
		return nil, err
	}
	if !checker.CheckStatus(args.Status.Value) {
		log.Error(ctx, "list outcome assessments: check status failed",
			log.Any("args", args),
			log.Any("checker", checker),
		)
		return nil, constant.ErrForbidden
	}

	// get assessment list
	var (
		assessments []*entity.Assessment
		cond        = da.QueryAssessmentConditions{
			ClassTypes: entity.NullScheduleClassTypes{
				Value: []entity.ScheduleClassType{entity.ScheduleClassTypeOnlineClass, entity.ScheduleClassTypeOfflineClass},
				Valid: true,
			},
			OrgID: entity.NullString{
				String: operator.OrgID,
				Valid:  true,
			},
			Status: args.Status,
			AllowTeacherIDs: entity.NullStrings{
				Strings: checker.allowTeacherIDs,
				Valid:   true,
			},
			AllowTeacherIDAndStatusPairs: entity.NullAssessmentAllowTeacherIDAndStatusPairs{
				Values: checker.AllowPairs(),
				Valid:  len(checker.AllowPairs()) > 0,
			},
			OrderBy: args.OrderBy,
			Pager:   args.Pager,
		}
		teachers    []*external.Teacher
		scheduleIDs []string
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
		log.Debug(ctx, "List: external.GetTeacherServiceProvider().Query: query success",
			log.String("org_id", operator.OrgID),
			log.String("teacher_name", args.TeacherName.String),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		cond.TeacherIDs.Valid = true
		for _, item := range teachers {
			cond.TeacherIDs.Strings = append(cond.TeacherIDs.Strings, item.ID)
		}
	}
	if scheduleIDs, err = GetScheduleModel().GetScheduleIDsByOrgID(ctx, tx, operator, operator.OrgID); err != nil {
		log.Error(ctx, "List: GetScheduleModel().GetScheduleIDsByOrgID: get failed",
			log.Err(err),
			log.String("org_id", operator.OrgID),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}
	cond.ScheduleIDs = entity.NullStrings{
		Strings: scheduleIDs,
		Valid:   true,
	}
	if err := da.GetAssessmentDA().QueryTx(ctx, tx, &cond, &assessments); err != nil {
		log.Error(ctx, "List: da.GetAssessmentDA().QueryTx: query failed",
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

	// get assessment list total
	var total int
	if total, err = da.GetAssessmentDA().CountTx(ctx, tx, &cond, &entity.Assessment{}); err != nil {
		log.Error(ctx, "List: da.GetAssessmentDA().CountTx: count failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("cond", cond),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// convert to assessment view
	var views []*entity.AssessmentView
	if views, err = GetAssessmentUtils().ToViews(ctx, tx, operator, assessments, entity.ConvertToViewsOptions{
		CheckedStudents: sql.NullBool{Bool: true, Valid: true},
		EnableProgram:   true,
		EnableSubjects:  true,
		EnableTeachers:  true,
	}); err != nil {
		log.Error(ctx, "List: GetAssessmentUtils().ToViews: get failed",
			log.Err(err),
			log.Any("assessments", assessments),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// construct result
	var result = entity.ListAssessmentsResult{Total: total}
	for _, v := range views {
		newItem := entity.AssessmentItem{
			ID:           v.ID,
			Title:        v.Title,
			Program:      v.Program,
			Subjects:     v.Subjects,
			Teachers:     v.Teachers,
			ClassEndTime: v.ClassEndTime,
			CompleteTime: v.CompleteTime,
			Status:       v.Status,
		}
		result.Items = append(result.Items, &newItem)
	}

	return &result, nil
}

func (m *classAndLiveAssessmentModel) Add(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AddAssessmentArgs) (string, error) {
	log.Debug(ctx, "add assessment args", log.Any("args", args), log.Any("operator", operator))

	// clean data
	args.AttendanceIDs = utils.SliceDeduplicationExcludeEmpty(args.AttendanceIDs)

	// use distributed lock
	lockKey := fmt.Sprintf("%s_%s", entity.ScheduleClassTypeOnlineClass, entity.ScheduleClassTypeOfflineClass)
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixAssessmentAddLock, args.ScheduleID, lockKey)
	if err != nil {
		log.Error(ctx, "add outcome assessment",
			log.Err(err),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()

	// check if assessment already exits
	count, err := da.GetAssessmentDA().CountTx(ctx, tx, &da.QueryAssessmentConditions{
		ClassTypes: entity.NullScheduleClassTypes{
			Value: []entity.ScheduleClassType{entity.ScheduleClassTypeOnlineClass, entity.ScheduleClassTypeOfflineClass},
			Valid: true,
		},
		ScheduleIDs: entity.NullStrings{
			Strings: []string{args.ScheduleID},
			Valid:   true,
		},
	}, entity.Assessment{})
	if err != nil {
		log.Error(ctx, "Add: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}
	if count > 0 {
		log.Info(ctx, "Add: assessment already exists",
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", nil
	}

	// get schedule and check class type
	var schedule *entity.SchedulePlain
	if schedule, err = GetScheduleModel().GetPlainByID(ctx, args.ScheduleID); err != nil {
		log.Error(ctx, "Add: GetScheduleModel().GetPlainByID: get failed",
			log.Err(err),
			log.Any("schedule_id", args.ScheduleID),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		switch err {
		case constant.ErrRecordNotFound, dbo.ErrRecordNotFound:
			return "", constant.ErrInvalidArgs
		default:
			return "", err
		}
	}
	if schedule.ClassType == entity.ScheduleClassTypeHomework || schedule.ClassType == entity.ScheduleClassTypeTask {
		log.Info(ctx, "Add: invalid schedule class type",
			log.String("class_type", string(schedule.ClassType)),
			log.Any("schedule", schedule),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", nil
	}

	// fix: permission
	operator.OrgID = schedule.OrgID

	// get contents
	var (
		latestContent   *entity.ContentInfoWithDetails
		materialIDs     []string
		materials       []*SubContentsWithName
		materialDetails []*entity.ContentInfoWithDetails
		contents        []*entity.ContentInfoWithDetails
	)
	if latestContent, err = GetContentModel().GetVisibleContentByID(ctx, dbo.MustGetDB(ctx), schedule.LessonPlanID, operator); err != nil {
		log.Warn(ctx, "Add: GetContentModel().GetVisibleContentByID: get latest content failed",
			log.Err(err),
			log.Any("args", args),
			log.String("lesson_plan_id", schedule.LessonPlanID),
			log.Any("schedule", schedule),
			log.Any("operator", operator),
		)
	} else {
		contents = append(contents, latestContent)
		if materials, err = GetContentModel().GetContentSubContentsByID(ctx, dbo.MustGetDB(ctx), latestContent.ID, operator); err != nil {
			log.Warn(ctx, "Add: GetContentModel().GetContentSubContentsByID: get materials failed",
				log.Err(err),
				log.Any("args", args),
				log.String("latest_lesson_plan_id", latestContent.ID),
				log.Any("latest_content", latestContent),
				log.Any("operator", operator),
				log.Any("schedule", schedule),
			)
		} else {
			for _, m := range materials {
				materialIDs = append(materialIDs, m.ID)
			}
			materialIDs = utils.SliceDeduplicationExcludeEmpty(materialIDs)
			if materialDetails, err = GetContentModel().GetContentByIDList(ctx, dbo.MustGetDB(ctx), materialIDs, operator); err != nil {
				log.Warn(ctx, "Add: GetContentModel().GetContentByIDList: get contents failed",
					log.Err(err),
					log.Strings("material_ids", materialIDs),
					log.Any("latest_content", latestContent),
					log.Any("schedule", schedule),
					log.Any("args", args),
					log.Any("operator", operator),
				)
			} else {
				contents = append(contents, materialDetails...)
			}
		}
	}

	// get outcomes
	var (
		outcomeIDs []string
		outcomes   []*entity.Outcome
	)
	for _, c := range contents {
		outcomeIDs = append(outcomeIDs, c.Outcomes...)
	}
	if len(outcomeIDs) > 0 {
		outcomeIDs = utils.SliceDeduplication(outcomeIDs)
		if outcomes, err = GetOutcomeModel().GetByIDs(ctx, operator, dbo.MustGetDB(ctx), outcomeIDs); err != nil {
			log.Error(ctx, "Add: GetOutcomeModel().GetByIDs: get failed",
				log.Err(err),
				log.Strings("outcome_ids", outcomeIDs),
				log.Any("args", args),
				log.Any("operator", operator),
			)
			return "", err
		}
	}

	// generate new assessment id
	var newAssessmentID = utils.NewID()

	// add assessment
	var (
		now           = time.Now().Unix()
		classNameMap  map[string]string
		newAssessment = entity.Assessment{
			ID:           newAssessmentID,
			ScheduleID:   args.ScheduleID,
			CreateAt:     now,
			UpdateAt:     now,
			ClassLength:  args.ClassLength,
			ClassEndTime: args.ClassEndTime,
		}
	)
	if len(outcomeIDs) == 0 {
		newAssessment.Status = entity.AssessmentStatusComplete
		newAssessment.CompleteTime = now
	} else {
		newAssessment.Status = entity.AssessmentStatusInProgress
	}
	if classNameMap, err = external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, []string{schedule.ClassID}); err != nil {
		log.Error(ctx, "Add: external.GetClassServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("class_ids", []string{schedule.ClassID}),
			log.Any("args", args),
		)
		return "", err
	}
	newAssessment.Title = m.generateTitle(newAssessment.ClassEndTime, classNameMap[schedule.ClassID], schedule.Title)
	// insert assessment
	if _, err := da.GetAssessmentDA().InsertTx(ctx, tx, &newAssessment); err != nil {
		log.Error(ctx, "add assessment: add failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("new_item", newAssessment),
		)
		return "", err
	}

	// add assessment attendances map
	var (
		finalAttendanceIDs []string
		scheduleRelations  []*entity.ScheduleRelation
	)
	switch schedule.ClassType {
	case entity.ScheduleClassTypeOfflineClass:
		users, err := GetScheduleRelationModel().GetUsersByScheduleID(ctx, operator, args.ScheduleID)
		if err != nil {
			return "", err
		}
		for _, u := range users {
			finalAttendanceIDs = append(finalAttendanceIDs, u.RelationID)
		}
	default:
		finalAttendanceIDs = args.AttendanceIDs
	}

	cond := &da.ScheduleRelationCondition{
		ScheduleID: sql.NullString{
			String: schedule.ID,
			Valid:  true,
		},
		RelationIDs: entity.NullStrings{
			Strings: finalAttendanceIDs,
			Valid:   true,
		},
	}
	if scheduleRelations, err = GetScheduleRelationModel().Query(ctx, operator, cond); err != nil {
		log.Error(ctx, "addAssessmentAttendances: GetScheduleRelationModel().GetByRelationIDs: get failed",
			log.Err(err),
			log.Any("attendance_ids", finalAttendanceIDs),
			log.String("assessment_id", newAssessmentID),
			log.Any("operator", operator),
			log.Any("condition", cond),
		)
		return "", err
	}
	if len(scheduleRelations) == 0 {
		log.Error(ctx, "Add: GetScheduleRelationModel().Query: not found any schedule relations",
			log.Err(err),
			log.Any("attendance_ids", finalAttendanceIDs),
			log.String("assessment_id", newAssessmentID),
			log.Any("operator", operator),
			log.Any("condition", cond),
		)
		return "", ErrNotFoundAttendance
	}
	if err = GetAssessmentUtils().AddAttendances(ctx, tx, operator, entity.AddAttendancesInput{
		AssessmentID:      newAssessmentID,
		ScheduleRelations: scheduleRelations,
	}); err != nil {
		log.Error(ctx, "Add: m.addAssessmentAttendances: add failed",
			log.Err(err),
			log.String("assessment_id", newAssessmentID),
			log.Strings("attendance_ids", finalAttendanceIDs),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}

	// add assessment outcomes map
	if err = m.addAssessmentOutcomes(ctx, tx, operator, newAssessmentID, outcomes); err != nil {
		log.Error(ctx, "Add: m.addAssessmentOutcomes: add failed",
			log.Err(err),
			log.String("assessment_id", newAssessmentID),
			log.Any("outcomes", outcomes),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}

	// add outcome attendances map
	if err = m.addOutcomeAttendances(ctx, tx, operator, newAssessmentID, outcomes, scheduleRelations); err != nil {
		log.Error(ctx, "Add: m.addOutcomeAttendances: add failed",
			log.Err(err),
			log.String("assessment_id", newAssessmentID),
			log.Any("outcomes", outcomes),
			log.Any("schedule_relations", scheduleRelations),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}

	// add assessment contents map
	if err = m.addAssessmentContentsAndOutcomes(ctx, tx, operator, newAssessmentID, contents); err != nil {
		log.Error(ctx, "Add: m.addAssessmentContentsAndOutcomes: add failed",
			log.Err(err),
			log.String("assessment_id", newAssessmentID),
			log.Any("contents", contents),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}

	return newAssessmentID, nil
}

func (m *classAndLiveAssessmentModel) Update(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.UpdateAssessmentArgs) error {
	// validate
	if !args.Action.Valid() {
		log.Error(ctx, "update assessment: invalid action", log.Any("args", args))
		return constant.ErrInvalidArgs
	}
	if args.Outcomes != nil {
		for _, item := range args.Outcomes {
			if item.Skip && item.NoneAchieved {
				log.Error(ctx, "update assessment: check skip and none achieved combination", log.Any("args", args))
				return constant.ErrInvalidArgs
			}
			if (item.Skip || item.NoneAchieved) && len(item.AttendanceIDs) > 0 {
				log.Error(ctx, "update assessment: check skip and none achieved combination with attendance ids", log.Any("args", args))
				return constant.ErrInvalidArgs
			}
		}
	}

	assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, dbo.MustGetDB(ctx), args.ID)
	if err != nil {
		log.Error(ctx, "update assessment: get assessment exclude soft deleted failed",
			log.Err(err),
			log.Any("args", args),
		)
		return err
	}

	// permission check
	hasP439, err := NewAssessmentPermissionChecker(operator).HasP439(ctx)
	if err != nil {
		return err
	}
	if !hasP439 {
		log.Error(ctx, "update assessment: not have permission 439",
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return constant.ErrForbidden
	}
	teacherIDs, err := da.GetAssessmentAttendanceDA().GetTeacherIDsByAssessmentID(ctx, dbo.MustGetDB(ctx), args.ID)
	if err != nil {
		log.Error(ctx, "Update: da.GetAssessmentAttendanceDA().GetTeacherIDsByAssessmentID: get failed",
			log.String("assessment_id", args.ID),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return err
	}
	hasOperator := false
	for _, tid := range teacherIDs {
		if tid == operator.UserID {
			hasOperator = true
			break
		}
	}
	if !hasOperator {
		log.Error(ctx, "update assessment: not find my assessment",
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return constant.ErrForbidden
	}
	if assessment.Status == entity.AssessmentStatusComplete {
		log.Info(ctx, "update assessment: assessment has completed, not allow update",
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return ErrAssessmentHasCompleted
	}

	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// update assessment students check property
		if args.StudentIDs != nil {
			if err := da.GetAssessmentAttendanceDA().UncheckStudents(ctx, tx, args.ID); err != nil {
				log.Error(ctx, "update: da.GetAssessmentAttendanceDA().UncheckStudents: uncheck failed",
					log.Err(err),
					log.Any("args", args),
				)
				return err
			}
			if args.StudentIDs != nil && len(args.StudentIDs) > 0 {
				if err := da.GetAssessmentAttendanceDA().BatchCheck(ctx, tx, args.ID, args.StudentIDs); err != nil {
					log.Error(ctx, "update: da.GetAssessmentAttendanceDA().BatchCheck: check failed",
						log.Err(err),
						log.Any("args", args),
					)
					return err
				}
			}
		}

		if args.Outcomes != nil {
			// update assessment outcomes map
			if err := da.GetAssessmentOutcomeDA().UncheckByAssessmentID(ctx, tx, args.ID); err != nil {
				log.Error(ctx, "Update: da.GetAssessmentOutcomeDA().UncheckStudents: uncheck assessment outcome failed by assessment id",
					log.Err(err),
					log.Any("args", args),
					log.String("id", args.ID),
				)
				return err
			}
			for _, oa := range args.Outcomes {
				newAssessmentOutcome := entity.AssessmentOutcome{
					AssessmentID: args.ID,
					OutcomeID:    oa.OutcomeID,
					Skip:         oa.Skip,
					NoneAchieved: oa.NoneAchieved,
					Checked:      true,
				}
				if err := da.GetAssessmentOutcomeDA().UpdateByAssessmentIDAndOutcomeID(ctx, tx, newAssessmentOutcome); err != nil {
					log.Error(ctx, "update assessment: batch update assessment outcome failed",
						log.Err(err),
						log.Any("new_assessment_outcome", newAssessmentOutcome),
						log.Any("args", args),
						log.String("assessment_id", args.ID),
					)
					return err
				}
			}
			// update outcome attendances map
			var (
				outcomeIDs         []string
				outcomeAttendances []*entity.OutcomeAttendance
			)
			for _, oa := range args.Outcomes {
				outcomeIDs = append(outcomeIDs, oa.OutcomeID)
				if oa.Skip {
					continue
				}
				for _, attendanceID := range oa.AttendanceIDs {
					outcomeAttendances = append(outcomeAttendances, &entity.OutcomeAttendance{
						ID:           utils.NewID(),
						AssessmentID: args.ID,
						OutcomeID:    oa.OutcomeID,
						AttendanceID: attendanceID,
					})
				}
			}
			if err := da.GetOutcomeAttendanceDA().BatchDeleteByAssessmentIDAndOutcomeIDs(ctx, tx, args.ID, outcomeIDs); err != nil {
				log.Error(ctx, "update assessment: batch delete outcome attendance map failed by outcome ids",
					log.Err(err),
					log.Strings("outcome_ids", outcomeIDs),
					log.Any("args", args),
				)
				return err
			}
			if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, outcomeAttendances); err != nil {
				log.Error(ctx, "update assessment: batch insert outcome attendance map failed",
					log.Err(err),
					log.Any("outcome_attendances", outcomeAttendances),
					log.Any("args", args),
				)
				return err
			}
		}

		/// update assessment contents map
		for _, ma := range args.LessonMaterials {
			updateArgs := da.UpdatePartialAssessmentContentArgs{
				AssessmentID:   args.ID,
				ContentID:      ma.ID,
				ContentComment: ma.Comment,
				Checked:        ma.Checked,
			}
			if err = da.GetAssessmentContentDA().UpdatePartial(ctx, tx, updateArgs); err != nil {
				log.Error(ctx, "Update: da.GetAssessmentContentDA().UpdatePartial: update failed",
					log.Err(err),
					log.Any("args", args),
					log.Any("update_args", updateArgs),
					log.Any("operator", operator),
				)
				return err
			}
		}

		// check and update status
		if args.Action == entity.UpdateAssessmentActionComplete {
			if err := da.GetAssessmentDA().UpdateStatus(ctx, tx, args.ID, entity.AssessmentStatusComplete); err != nil {
				log.Error(ctx, "Update: da.GetAssessmentDA().UpdateStatus: update failed",
					log.Err(err),
					log.Any("args", args),
					log.Any("operator", operator),
				)
				return err
			}
		}

		return nil
	}); err != nil {
		log.Error(ctx, "Update: tx failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return err
	}

	return nil
}

// region utils

func (m *classAndLiveAssessmentModel) generateTitle(classEndTime int64, className string, scheduleTitle string) string {
	if className == "" {
		className = constant.AssessmentNoClass
	}
	return fmt.Sprintf("%s-%s-%s", time.Unix(classEndTime, 0).Format("20060102"), className, scheduleTitle)
}

func (m *classAndLiveAssessmentModel) getAssessmentContentOutcomeMap(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string, contentIDs []string) (map[string]map[string][]string, error) {
	var assessmentContentOutcomes []*entity.AssessmentContentOutcome
	cond := da.QueryAssessmentContentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		ContentIDs: entity.NullStrings{
			Strings: contentIDs,
			Valid:   true,
		},
	}
	if err := da.GetAssessmentContentOutcomeDA().QueryTx(ctx, tx, &cond, &assessmentContentOutcomes); err != nil {
		log.Error(ctx, "getAssessmentContentOutcomeMap: da.GetAssessmentContentOutcomeDA().QueryTx: get failed",
			log.Err(err),
			log.Any("cond", cond),
			log.Strings("assessment_ids", assessmentIDs),
			log.Strings("content_ids", contentIDs),
		)
		return nil, err
	}
	result := map[string]map[string][]string{}
	for _, co := range assessmentContentOutcomes {
		if result[co.AssessmentID] == nil {
			result[co.AssessmentID] = map[string][]string{co.ContentID: {co.OutcomeID}}
		} else {
			result[co.AssessmentID][co.ContentID] = append(result[co.AssessmentID][co.ContentID], co.OutcomeID)
		}
	}
	return result, nil
}

func (m *classAndLiveAssessmentModel) getLessonMaterialToOutcomeNamesMap(ctx context.Context, tx *dbo.DBContext) {

}

func (m *classAndLiveAssessmentModel) addAssessmentContentsAndOutcomes(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessmentID string, contents []*entity.ContentInfoWithDetails) error {
	if len(contents) == 0 {
		return nil
	}
	var (
		assessmentContents        []*entity.AssessmentContent
		assessmentContentOutcomes []*entity.AssessmentContentOutcome
	)
	for _, c := range contents {
		assessmentContents = append(assessmentContents, &entity.AssessmentContent{
			ID:           utils.NewID(),
			AssessmentID: assessmentID,
			ContentID:    c.ID,
			ContentName:  c.Name,
			ContentType:  c.ContentType,
			Checked:      true,
		})
		for _, oid := range c.Outcomes {
			assessmentContentOutcomes = append(assessmentContentOutcomes, &entity.AssessmentContentOutcome{
				ID:           utils.NewID(),
				AssessmentID: assessmentID,
				ContentID:    c.ID,
				OutcomeID:    oid,
			})
		}
	}
	if len(assessmentContents) == 0 {
		return nil
	}
	if err := da.GetAssessmentContentDA().BatchInsert(ctx, tx, assessmentContents); err != nil {
		log.Error(ctx, "addAssessmentContentsAndOutcomes: da.GetAssessmentContentDA().BatchInsert: batch insert failed",
			log.Err(err),
			log.Any("assessment_contents", assessmentContents),
			log.String("assessment_id", assessmentID),
			log.Any("contents", contents),
			log.Any("operator", operator),
		)
		return err
	}
	if len(assessmentContentOutcomes) == 0 {
		return nil
	}
	if err := da.GetAssessmentContentOutcomeDA().BatchInsert(ctx, tx, assessmentContentOutcomes); err != nil {
		log.Error(ctx, "addAssessmentContentsAndOutcomes: da.GetAssessmentContentOutcomeDA().BatchInsert: batch insert failed",
			log.Err(err),
			log.Any("assessment_content_outcomes", assessmentContentOutcomes),
			log.String("assessment_id", assessmentID),
			log.Any("contents", contents),
			log.Any("operator", operator),
		)
		return err
	}

	return nil
}

func (m *classAndLiveAssessmentModel) addAssessmentOutcomes(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessmentID string, outcomes []*entity.Outcome) error {
	if len(outcomes) == 0 {
		return nil
	}
	var assessmentOutcomes []*entity.AssessmentOutcome
	for _, outcome := range outcomes {
		assessmentOutcomes = append(assessmentOutcomes, &entity.AssessmentOutcome{
			ID:           utils.NewID(),
			AssessmentID: assessmentID,
			OutcomeID:    outcome.ID,
			Skip:         false,
			NoneAchieved: !outcome.Assumed,
			Checked:      true,
		})
	}
	if err := da.GetAssessmentOutcomeDA().BatchInsert(ctx, tx, assessmentOutcomes); err != nil {
		log.Error(ctx, "addAssessmentOutcomes: da.GetAssessmentOutcomeDA().BatchInsert: batch insert failed",
			log.Err(err),
			log.Any("assessment_outcomes", assessmentOutcomes),
			log.String("assessment_id", assessmentID),
			log.Any("outcomes", outcomes),
			log.Any("operator", operator),
		)
		return err
	}
	return nil
}

func (m *classAndLiveAssessmentModel) addOutcomeAttendances(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessmentID string, outcomes []*entity.Outcome, scheduleRelations []*entity.ScheduleRelation) error {
	if len(outcomes) == 0 || len(scheduleRelations) == 0 {
		return nil
	}
	var (
		studentIDs         []string
		outcomeAttendances []*entity.OutcomeAttendance
	)
	for _, r := range scheduleRelations {
		if r.RelationType == entity.ScheduleRelationTypeClassRosterStudent ||
			r.RelationType == entity.ScheduleRelationTypeParticipantStudent {
			studentIDs = append(studentIDs, r.RelationID)
		}
	}
	for _, outcome := range outcomes {
		if !outcome.Assumed {
			continue
		}
		for _, sid := range studentIDs {
			outcomeAttendances = append(outcomeAttendances, &entity.OutcomeAttendance{
				ID:           utils.NewID(),
				AssessmentID: assessmentID,
				OutcomeID:    outcome.ID,
				AttendanceID: sid,
			})
		}
	}
	if len(outcomeAttendances) == 0 {
		return nil
	}
	if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, outcomeAttendances); err != nil {
		log.Error(ctx, "addOutcomeAttendances: da.GetOutcomeAttendanceDA().BatchInsert: batch insert failed",
			log.Err(err),
			log.Any("outcomeAttendances", outcomeAttendances),
			log.String("assessment_id", assessmentID),
			log.Any("schedule_relations", scheduleRelations),
			log.Any("operator", operator),
		)
		return err
	}
	return nil
}

type OutcomesOrderByAssumedAndName []*entity.Outcome

func (s OutcomesOrderByAssumedAndName) Len() int {
	return len(s)
}

func (s OutcomesOrderByAssumedAndName) Less(i, j int) bool {
	if s[i].Assumed && !s[j].Assumed {
		return true
	} else if !s[i].Assumed && s[j].Assumed {
		return false
	} else {
		return s[i].Name < s[j].Name
	}
}

func (s OutcomesOrderByAssumedAndName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type AssessmentAttendanceOrderByOrigin []*entity.AssessmentAttendance

func (a AssessmentAttendanceOrderByOrigin) Len() int {
	return len(a)
}

func (a AssessmentAttendanceOrderByOrigin) Less(i, j int) bool {
	if a[i].Origin == entity.AssessmentAttendanceOriginParticipants &&
		a[j].Origin == entity.AssessmentAttendanceOriginClassRoaster {
		return false
	}
	return true
}

func (a AssessmentAttendanceOrderByOrigin) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// endregion
