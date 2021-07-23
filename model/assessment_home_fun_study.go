package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrHomeFunStudyHasCompleted   = errors.New("home fun study has completed")
	ErrHomeFunStudyHasNewFeedback = errors.New("home fun study has new feedback")

	homeFunStudyModelInstance     IHomeFunStudyModel
	homeFunStudyModelInstanceOnce = sync.Once{}
)

func GetHomeFunStudyModel() IHomeFunStudyModel {
	homeFunStudyModelInstanceOnce.Do(func() {
		homeFunStudyModelInstance = &homeFunStudyModel{}
	})
	return homeFunStudyModelInstance
}

type IHomeFunStudyModel interface {
	Query(ctx context.Context, operator *entity.Operator, conditions *da.QueryHomeFunStudyCondition, result interface{}) error
	Get(ctx context.Context, operator *entity.Operator, id string) (*entity.HomeFunStudy, error)
	GetDetail(ctx context.Context, operator *entity.Operator, id string) (*entity.GetHomeFunStudyResult, error)
	GetByScheduleIDAndStudentID(ctx context.Context, operator *entity.Operator, scheduleID string, studentID string) (*entity.HomeFunStudy, error)
	Summary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsSummaryArgs) (*entity.AssessmentsSummary, error)
	List(ctx context.Context, operator *entity.Operator, args *entity.ListHomeFunStudiesArgs) (*entity.ListHomeFunStudiesResult, error)
	Save(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.SaveHomeFunStudyArgs) error
	Assess(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AssessHomeFunStudyArgs) error
}

type homeFunStudyModel struct {
	assessmentBase
}

func (m *homeFunStudyModel) Query(ctx context.Context, operator *entity.Operator, conditions *da.QueryHomeFunStudyCondition, result interface{}) error {
	if err := da.GetHomeFunStudyDA().Query(ctx, conditions, result); err != nil {
		log.Error(ctx, "query home fun study: query failed",
			log.Err(err),
			log.Any("conditions", conditions),
		)
		return err
	}
	return nil
}

func (m *homeFunStudyModel) Get(ctx context.Context, operator *entity.Operator, id string) (*entity.HomeFunStudy, error) {
	var study entity.HomeFunStudy
	if err := da.GetHomeFunStudyDA().Get(ctx, id, &study); err != nil {
		log.Error(ctx, "Get: da.GetHomeFunStudyDA().Get: get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("id", id),
		)
		if err == dbo.ErrRecordNotFound {
			return nil, constant.ErrRecordNotFound
		}
		return nil, err
	}
	return &study, nil
}

func (m *homeFunStudyModel) GetDetail(ctx context.Context, operator *entity.Operator, id string) (*entity.GetHomeFunStudyResult, error) {
	var study entity.HomeFunStudy
	if err := da.GetHomeFunStudyDA().Get(ctx, id, &study); err != nil {
		log.Error(ctx, "da.GetHomeFunStudyDA().Get: get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("id", id),
		)
		if err == sql.ErrNoRows {
			return nil, constant.ErrRecordNotFound
		}
		return nil, err
	}

	checker := NewAssessmentPermissionChecker(operator)
	if err := checker.SearchAllPermissions(ctx); err != nil {
		log.Error(ctx, "Get: checker.SearchAllPermissions: checker failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("id", id),
		)
		return nil, err
	}
	if study.Status.Valid() && !checker.CheckStatus(study.Status) {
		log.Error(ctx, "Get: checker.CheckStatus: status not allowed",
			log.Any("operator", operator),
			log.String("id", id),
			log.String("status", string(study.Status)),
			log.Any("study", study),
		)
		return nil, constant.ErrForbidden
	}
	if !checker.CheckTeacherIDs(study.TeacherIDs) {
		log.Error(ctx, "Get: checker.CheckTeacherIDs: teacher ids not allowed",
			log.Any("operator", operator),
			log.String("id", id),
			log.Strings("teacher_ids", study.TeacherIDs),
			log.Any("study", study),
		)
		return nil, constant.ErrForbidden
	}

	var (
		studentName  string
		teacherNames []string
	)
	student, err := external.GetStudentServiceProvider().Get(ctx, operator, study.StudentID)
	if err != nil {
		log.Error(ctx, "external.GetStudentServiceProvider().Get: get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("id", id),
		)
		return nil, err
	}
	studentName = student.Name
	teachers, err := external.GetTeacherServiceProvider().BatchGet(ctx, operator, study.TeacherIDs)
	if err != nil {
		log.Error(ctx, "external.GetTeacherServiceProvider().BatchGet: batch get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("teacher_ids", study.TeacherIDs),
		)
		return nil, err
	}
	for _, t := range teachers {
		teacherNames = append(teacherNames, t.Name)
	}

	result := &entity.GetHomeFunStudyResult{
		ID:         study.ID,
		ScheduleID: study.ScheduleID,
		Title:      study.Title,
		TeacherIDs: study.TeacherIDs, TeacherNames: teacherNames, StudentID: study.StudentID, StudentName: studentName, Status: study.Status,
		DueAt:            study.DueAt,
		CompleteAt:       study.CompleteAt,
		AssessFeedbackID: study.AssessFeedbackID,
		AssessScore:      study.AssessScore,
		AssessComment:    study.AssessComment,
	}

	if result.AssessScore == 0 {
		result.AssessScore = entity.HomeFunStudyAssessScoreAverage
	}

	// fill outcomes

	// batch get assessment outcomes
	var assessmentOutcomes []*entity.AssessmentOutcome
	if err := da.GetAssessmentOutcomeDA().Query(ctx, &da.QueryAssessmentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{Strings: []string{id}, Valid: true},
		Checked:       entity.NullBool{Bool: true, Valid: true},
	}, &assessmentOutcomes); err != nil {
		log.Error(ctx, "assess home fun study: query assessment outcome failed",
			log.Err(err),
			log.String("id", id),
		)
		return nil, err
	}

	// batch get outcome attendance map
	outcomeIDs := make([]string, 0, len(assessmentOutcomes))
	for _, ao := range assessmentOutcomes {
		outcomeIDs = append(outcomeIDs, ao.OutcomeID)
	}
	outcomeAttendances, err := da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDAndOutcomeIDs(ctx, dbo.MustGetDB(ctx), id, outcomeIDs)
	if err != nil {
		log.Error(ctx, "assess home fun study: query outcome attendance failed",
			log.Err(err),
			log.String("id", id),
		)
		return nil, err
	}
	outcomeAttendanceMap := make(map[string][]*entity.OutcomeAttendance, len(outcomeAttendances))
	for _, oa := range outcomeAttendances {
		outcomeAttendanceMap[oa.OutcomeID] = append(outcomeAttendanceMap[oa.OutcomeID], oa)
	}

	// batch get outcome map
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, operator, dbo.MustGetDB(ctx), outcomeIDs)
	outcomeMap := make(map[string]*entity.Outcome, len(outcomes))
	for _, o := range outcomes {
		outcomeMap[o.ID] = o
	}
	for _, ao := range assessmentOutcomes {
		var outcomeName string
		var assumed bool
		outcome := outcomeMap[ao.OutcomeID]
		if outcome != nil {
			outcomeName = outcome.Name
			assumed = outcome.Assumed
		}
		var status entity.HomeFunStudyOutcomeStatus
		if ao.Skip {
			status = entity.HomeFunStudyOutcomeStatusNotAttempted
		} else {
			if len(outcomeAttendanceMap[ao.OutcomeID]) > 0 {
				status = entity.HomeFunStudyOutcomeStatusAchieved
			} else {
				status = entity.HomeFunStudyOutcomeStatusNotAchieved
			}
		}
		result.Outcomes = append(result.Outcomes, &entity.HomeFunStudyOutcome{
			OutcomeID:   ao.OutcomeID,
			OutcomeName: outcomeName,
			Assumed:     assumed,
			Status:      status,
		})
	}

	return result, nil
}

func (m *homeFunStudyModel) GetByScheduleIDAndStudentID(ctx context.Context, operator *entity.Operator, scheduleID string, studentID string) (*entity.HomeFunStudy, error) {
	cond := da.QueryHomeFunStudyCondition{
		OrgID: entity.NullString{
			String: operator.OrgID,
			Valid:  true,
		},
		ScheduleID: entity.NullString{
			String: scheduleID,
			Valid:  true,
		},
		StudentIDs: entity.NullStrings{
			Strings: []string{studentID},
			Valid:   true,
		},
		Pager: dbo.Pager{
			Page:     1,
			PageSize: 1,
		},
	}
	var studies []*entity.HomeFunStudy
	if err := da.GetHomeFunStudyDA().Query(ctx, &cond, &studies); err != nil {
		log.Error(ctx, "GetByScheduleIDAndStudentID: da.GetHomeFunStudyDA().Query: query home fun study",
			log.Err(err),
			log.Any("cond", cond),
		)
		return nil, err
	}
	if len(studies) == 0 {
		log.Error(ctx, "GetByScheduleIDAndStudentID: not found home fun study",
			log.String("schedule_id", scheduleID),
			log.String("student_id", studentID),
		)
		return nil, constant.ErrRecordNotFound
	}
	return studies[0], nil
}

func (m *homeFunStudyModel) Summary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsSummaryArgs) (*entity.AssessmentsSummary, error) {
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
		studies []*entity.HomeFunStudy
		cond    = da.QueryHomeFunStudyCondition{
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
				Values: checker.AllowPairs(),
				Valid:  len(checker.AllowPairs()) > 0,
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
				cond.TeacherIDs.Values = append(cond.TeacherIDs.Values, item.ID)
			}
		} else {
			cond.TeacherIDs.Valid = false
		}
	}

	if err := da.GetHomeFunStudyDA().QueryTx(ctx, tx, &cond, &studies); err != nil {
		log.Error(ctx, "summary: query studies failed",
			log.Err(err),
			log.Any("cond", cond),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}
	if len(studies) == 0 {
		return &entity.AssessmentsSummary{}, nil
	}

	r := entity.AssessmentsSummary{}
	for _, a := range studies {
		switch a.Status {
		case entity.AssessmentStatusComplete:
			r.Complete++
		case entity.AssessmentStatusInProgress:
			r.InProgress++
		}
	}

	log.Debug(ctx, "home fun study summary",
		log.Any("studies", studies),
		log.Any("result", r),
		log.Any("args", args),
		log.Any("operator", operator),
	)

	return &r, nil
}

func (m *homeFunStudyModel) List(ctx context.Context, operator *entity.Operator, args *entity.ListHomeFunStudiesArgs) (*entity.ListHomeFunStudiesResult, error) {
	checker := NewAssessmentPermissionChecker(operator)
	if err := checker.SearchAllPermissions(ctx); err != nil {
		log.Error(ctx, "List: checker.SearchAllPermissions: checker failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, err
	}
	if args.Status.Valid && !checker.CheckStatus(args.Status.Value) {
		log.Error(ctx, "list home fun study assessment",
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, constant.ErrForbidden
	}
	cond := da.QueryHomeFunStudyCondition{
		OrgID: entity.NullString{
			String: operator.OrgID,
			Valid:  true,
		},
		Status:  args.Status,
		OrderBy: args.OrderBy,
		AllowTeacherIDs: entity.NullStrings{
			Strings: checker.allowTeacherIDs,
			Valid:   true,
		},
		AllowTeacherIDAndStatusPairs: entity.NullAssessmentAllowTeacherIDAndStatusPairs{
			Values: checker.AllowPairs(),
			Valid:  len(checker.AllowPairs()) > 0,
		},
		Pager: args.Pager,
	}
	if args.Query != "" {
		teachers, err := external.GetTeacherServiceProvider().Query(ctx, operator, operator.OrgID, args.Query)
		if err != nil {
			log.Error(ctx, "external.GetTeacherServiceProvider().Query: query failed",
				log.Err(err),
				log.Any("operator", operator),
				log.String("org_id", operator.OrgID),
				log.Any("args", args),
			)
			return nil, err
		}
		cond.TeacherIDs.Valid = true
		for _, t := range teachers {
			cond.TeacherIDs.Values = append(cond.TeacherIDs.Values, t.ID)
		}
	}
	students, err := external.GetStudentServiceProvider().Query(ctx, operator, operator.OrgID, args.Query)
	if err != nil {
		log.Error(ctx, "external.GetStudentServiceProvider().Query: query failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("org_id", operator.OrgID),
			log.Any("args", args),
		)
		return nil, err
	}
	if len(students) > 0 {
		cond.StudentIDs.Valid = true
		for _, s := range students {
			cond.StudentIDs.Strings = append(cond.StudentIDs.Strings, s.ID)
		}
	}

	var (
		items []entity.HomeFunStudy
		total int
	)
	if total, err = da.GetHomeFunStudyDA().Page(ctx, &cond, &items); err != nil {
		log.Error(ctx, "da.GetHomeFunStudyDA().Query: query failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("org_id", operator.OrgID),
			log.Any("args", args),
		)
		return nil, err
	}

	var (
		studentIDs []string
		teacherIDs []string
	)
	for _, item := range items {
		studentIDs = append(studentIDs, item.StudentID)
		teacherIDs = append(teacherIDs, item.TeacherIDs...)
	}
	studentNamesMap, err := external.GetStudentServiceProvider().BatchGetNameMap(ctx, operator, studentIDs)
	if err != nil {
		log.Error(ctx, "m.getStudentNamesMap: get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
			log.Any("all_student_ids", studentIDs),
		)
		return nil, err
	}
	teacherNamesMap, err := external.GetTeacherServiceProvider().BatchGetNameMap(ctx, operator, teacherIDs)
	if err != nil {
		log.Error(ctx, "m.getTeacherNamesMap: get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
			log.Any("all_teacher_ids", teacherIDs),
		)
		return nil, err
	}

	// get lesson plans
	scheduleIDs := make([]string, 0, len(items))
	for _, item := range items {
		scheduleIDs = append(scheduleIDs, item.ScheduleID)
	}
	schedules, err := GetScheduleModel().GetVariableDataByIDs(ctx, operator, scheduleIDs, nil)
	if err != nil {
		log.Error(ctx, "get home fun study list: get schedules failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}
	lessonPlanIDs := make([]string, 0, len(schedules))
	lessonPlanIDToScheduleIDMap := make(map[string]string, len(schedules))
	for _, s := range schedules {
		lessonPlanIDs = append(lessonPlanIDs, s.LessonPlanID)
		lessonPlanIDToScheduleIDMap[s.LessonPlanID] = s.ID
	}
	lessonPlans, err := GetContentModel().GetRawContentByIDList(ctx, dbo.MustGetDB(ctx), lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "get home fun study list: get contents failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}
	scheduleIDToLessonPlanMap := make(map[string]*entity.Content, len(lessonPlans))
	for _, lp := range lessonPlans {
		scheduleID := lessonPlanIDToScheduleIDMap[lp.ID]
		if scheduleID == "" {
			continue
		}
		scheduleIDToLessonPlanMap[scheduleID] = lp
	}

	result := entity.ListHomeFunStudiesResult{Total: total}
	for _, item := range items {
		var teacherNames []string
		for _, id := range item.TeacherIDs {
			teacherNames = append(teacherNames, teacherNamesMap[id])
		}
		resultItem := entity.ListHomeFunStudiesResultItem{
			ID:               item.ID,
			Title:            item.Title,
			TeacherNames:     teacherNames,
			StudentName:      studentNamesMap[item.StudentID],
			Status:           item.Status,
			DueAt:            item.DueAt,
			LatestFeedbackAt: item.LatestFeedbackAt,
			AssessScore:      item.AssessScore,
			CompleteAt:       item.CompleteAt,
			ScheduleID:       item.ScheduleID,
		}
		if lp := scheduleIDToLessonPlanMap[item.ScheduleID]; lp != nil {
			resultItem.LessonPlan = &entity.AssessmentLessonPlan{
				ID:   lp.ID,
				Name: lp.Name,
			}
		}
		result.Items = append(result.Items, &resultItem)
	}

	return &result, nil
}

func (m *homeFunStudyModel) Save(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.SaveHomeFunStudyArgs) error {
	cond := da.QueryHomeFunStudyCondition{
		ScheduleID: entity.NullString{
			String: args.ScheduleID,
			Valid:  true,
		},
		StudentIDs: entity.NullStrings{
			Strings: []string{args.StudentID},
			Valid:   true,
		},
	}
	var studies []*entity.HomeFunStudy
	if err := da.GetHomeFunStudyDA().QueryTx(ctx, tx, &cond, &studies); err != nil {
		log.Error(ctx, "da.GetHomeFunStudyDA().CountTx: count failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
			log.Any("cond", cond),
		)
		return err
	}
	var study *entity.HomeFunStudy
	if len(studies) > 0 {
		study = studies[0]
		if study.Status == entity.AssessmentStatusComplete {
			return ErrHomeFunStudyHasCompleted
		}
	}

	var title string
	if len(args.ClassID) > 0 {
		classes, err := external.GetClassServiceProvider().BatchGet(ctx, operator, []string{args.ClassID})
		if err != nil {
			log.Error(ctx, "external.GetClassServiceProvider().BatchGet: batch get class failed",
				log.Err(err),
				log.Any("operator", operator),
				log.String("class_id", args.ClassID),
				log.Any("args", args),
			)
			return err
		}
		if len(classes) == 0 {
			log.Error(ctx, "Save: not found class",
				log.Any("operator", operator),
				log.String("class_id", args.ClassID),
				log.Any("args", args),
			)
			return errors.New("not found class")
		}
		title = m.generateTitle(&classes[0].Name, args.LessonName)
	} else {
		title = m.generateTitle(nil, args.LessonName)
	}

	if study != nil {
		study.Title = title
		study.TeacherIDs = args.TeacherIDs
		study.DueAt = args.DueAt
		study.LatestFeedbackID = args.LatestFeedbackID
		study.LatestFeedbackAt = args.LatestFeedbackAt
		study.UpdateAt = time.Now().Unix()
	} else {
		now := time.Now().Unix()
		study = &entity.HomeFunStudy{
			ID:               utils.NewID(),
			ScheduleID:       args.ScheduleID,
			Title:            title,
			TeacherIDs:       args.TeacherIDs,
			StudentID:        args.StudentID,
			Status:           entity.AssessmentStatusInProgress,
			DueAt:            args.DueAt,
			LatestFeedbackID: args.LatestFeedbackID,
			LatestFeedbackAt: args.LatestFeedbackAt,
			CreateAt:         now,
			UpdateAt:         now,
		}
	}
	if err := da.GetHomeFunStudyDA().SaveTx(ctx, tx, study); err != nil {
		log.Error(ctx, "da.GetHomeFunStudyDA().SaveTx: save failed",
			log.Err(err),
			log.Any("study", *study))
		return err
	}

	// get all related outcome ids
	outcomeIDs, err := GetScheduleModel().GetLearningOutcomeIDs(ctx, operator, args.ScheduleID)
	if err != nil {
		log.Error(ctx, "save home fun study: get outcome ids failed",
			log.Err(err),
			log.Any("args", args),
		)
		return err
	}

	// batch insert assessment outcomes
	var insertingAssessmentOutcomes []*entity.AssessmentOutcome
	for _, oid := range outcomeIDs {
		insertingAssessmentOutcomes = append(insertingAssessmentOutcomes, &entity.AssessmentOutcome{
			ID:           utils.NewID(),
			AssessmentID: study.ID,
			OutcomeID:    oid,
			Checked:      true,
		})
	}
	if err := da.GetAssessmentOutcomeDA().BatchInsert(ctx, tx, insertingAssessmentOutcomes); err != nil {
		log.Error(ctx, "save home fun study: batch insert assessment outcomes failed",
			log.Err(err),
			log.Any("inserting_assessment_outcomes", insertingAssessmentOutcomes),
			log.Any("args", args),
		)
		return err
	}

	// batch get all outcomes
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, operator, tx, outcomeIDs)
	if err != nil {
		log.Error(ctx, "save home fun study: batch get outcomes failed",
			log.Err(err),
			log.Any("outcome_ids", outcomeIDs),
			log.Any("args", args),
		)
		return err
	}
	outcomeMap := make(map[string]*entity.Outcome, len(outcomes))
	for _, o := range outcomes {
		outcomeMap[o.ID] = o
	}

	// add outcome attendance
	var insertingOutcomeAttendances []*entity.OutcomeAttendance
	for _, o := range outcomes {
		if !o.Assumed {
			continue
		}
		insertingOutcomeAttendances = append(insertingOutcomeAttendances, &entity.OutcomeAttendance{
			ID:           utils.NewID(),
			AssessmentID: study.ID,
			OutcomeID:    o.ID,
			AttendanceID: study.StudentID,
		})
	}
	if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, insertingOutcomeAttendances); err != nil {
		log.Error(ctx, "save home fun study: batch insert outcome attendance failed",
			log.Err(err),
			log.Any("inserting_outcome_attendance", insertingOutcomeAttendances),
			log.Any("args", args),
		)
		return err
	}

	return nil
}

func (m *homeFunStudyModel) Assess(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AssessHomeFunStudyArgs) error {
	checker := NewAssessmentPermissionChecker(operator)
	ok, err := checker.HasP439(ctx)
	if err != nil {
		log.Error(ctx, "Assess: NewAssessmentPermissionChecker: check assess permission",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return err
	}
	if !ok {
		log.Error(ctx, "Assess: no permission",
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return constant.ErrForbidden
	}
	if err := checker.SearchAllPermissions(ctx); err != nil {
		log.Error(ctx, "Assess: checker.SearchAllPermissions: search permissions failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return err
	}

	var study entity.HomeFunStudy
	if err := da.GetHomeFunStudyDA().GetTx(ctx, tx, args.ID, &study); err != nil {
		log.Error(ctx, "da.GetHomeFunStudyDA().GetTx: get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
			log.String("id", args.ID),
		)
		if err == dbo.ErrRecordNotFound {
			return constant.ErrRecordNotFound
		}
		return err
	}
	if !checker.CheckTeacherIDs(study.TeacherIDs) {
		log.Error(ctx, "Assess: checker.CheckTeacherIDs: check teacher ids failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("teacher_ids", study.TeacherIDs),
			log.Any("args", args),
		)
		return err
	}
	if study.Status == entity.AssessmentStatusComplete {
		log.Error(ctx, "Assess: study has completed",
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return ErrHomeFunStudyHasCompleted
	}

	study.AssessFeedbackID = args.AssessFeedbackID
	study.AssessScore = args.AssessScore
	study.AssessComment = args.AssessComment
	if err := da.GetHomeFunStudyDA().SaveTx(ctx, tx, &study); err != nil {
		log.Error(ctx, "da.GetHomeFunStudyDA().SaveTx: save failed",
			log.Err(err),
			log.Any("study", study))
		return err
	}

	newest, err := GetScheduleFeedbackModel().GetNewest(ctx, operator, study.StudentID, study.ScheduleID)
	if err != nil {
		log.Error(ctx, "Assess: GetScheduleFeedbackModel().GetNewest: get newest feedback failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
			log.Any("study", study),
		)
		return err
	}
	if args.AssessFeedbackID != newest.ScheduleFeedback.ID {
		log.Error(ctx, "Assess: study has new feedback",
			log.Any("operator", operator),
			log.Any("args", args),
			log.Any("study", study),
		)
		return ErrHomeFunStudyHasNewFeedback
	}

	// loop update outcomes status
	for _, o := range args.Outcomes {
		e := entity.AssessmentOutcome{
			AssessmentID: args.ID,
			OutcomeID:    o.OutcomeID,
			Skip:         o.Status == entity.HomeFunStudyOutcomeStatusNotAttempted,
		}
		if err := da.GetAssessmentOutcomeDA().UpdateByAssessmentIDAndOutcomeID(ctx, tx, &e); err != nil {
			log.Error(ctx, "assess home fun study: update assessment outcome failed",
				log.Err(err),
				log.Any("e", e),
				log.Any("args", args),
			)
			return err
		}
	}

	// batch update outcome attendance list, use full update mode.
	var deletingOutcomeIDs []string
	var insertingOutcomeAttendances []*entity.OutcomeAttendance
	for _, o := range args.Outcomes {
		deletingOutcomeIDs = append(deletingOutcomeIDs, o.OutcomeID)
		insertingOutcomeAttendances = append(insertingOutcomeAttendances, &entity.OutcomeAttendance{
			ID:           utils.NewID(),
			AssessmentID: args.ID,
			OutcomeID:    o.OutcomeID,
			AttendanceID: study.StudentID,
		})
	}
	if err := da.GetOutcomeAttendanceDA().BatchDeleteByAssessmentIDAndOutcomeIDs(ctx, tx, args.ID, deletingOutcomeIDs); err != nil {
		log.Error(ctx, "assess home fun study: batch delete outcome attendance failed",
			log.String("assessment_id", args.ID),
			log.Strings("delete_outcome_ids", deletingOutcomeIDs),
			log.Any("args", args),
			log.Err(err),
		)
		return err
	}
	if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, insertingOutcomeAttendances); err != nil {
		log.Error(ctx, "assess home fun study: batch insert outcome attendance failed",
			log.Any("inserting_outcome_attendances", insertingOutcomeAttendances),
			log.Any("args", args),
			log.Err(err),
		)
		return err
	}

	if args.Action == entity.UpdateHomeFunStudyActionComplete {
		study.Status = entity.AssessmentStatusComplete
		study.CompleteAt = time.Now().Unix()
		if err := da.GetHomeFunStudyDA().SaveTx(ctx, tx, &study); err != nil {
			log.Error(ctx, "da.GetHomeFunStudyDA().SaveTx: save failed",
				log.Err(err),
				log.Any("study", study))
			return err
		}
	}

	return nil
}

// region utils

func (m *homeFunStudyModel) generateTitle(className *string, lessonName string) string {
	if className == nil {
		tmp := constant.AssessmentNoClass
		className = &tmp
	}
	return fmt.Sprintf("%s-%s", *className, lessonName)
}

// endregion
