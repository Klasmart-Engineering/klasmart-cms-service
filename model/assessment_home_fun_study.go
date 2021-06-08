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
		ID:               study.ID,
		ScheduleID:       study.ScheduleID,
		Title:            study.Title,
		TeacherIDs:       study.TeacherIDs,
		TeacherNames:     teacherNames,
		StudentID:        study.StudentID,
		StudentName:      studentName,
		Status:           study.Status,
		DueAt:            study.DueAt,
		CompleteAt:       study.CompleteAt,
		AssessFeedbackID: study.AssessFeedbackID,
		AssessScore:      study.AssessScore,
		AssessComment:    study.AssessComment,
	}

	if result.AssessScore == 0 {
		result.AssessScore = entity.HomeFunStudyAssessScoreAverage
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

	result := entity.ListHomeFunStudiesResult{Total: total}
	for _, item := range items {
		var teacherNames []string
		for _, id := range item.TeacherIDs {
			teacherNames = append(teacherNames, teacherNamesMap[id])
		}
		studentName := studentNamesMap[item.StudentID]
		result.Items = append(result.Items, &entity.ListHomeFunStudiesResultItem{
			ID:               item.ID,
			Title:            item.Title,
			TeacherNames:     teacherNames,
			StudentName:      studentName,
			Status:           item.Status,
			DueAt:            item.DueAt,
			LatestFeedbackAt: item.LatestFeedbackAt,
			AssessScore:      item.AssessScore,
			CompleteAt:       item.CompleteAt,
			ScheduleID:       item.ScheduleID,
		})
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
