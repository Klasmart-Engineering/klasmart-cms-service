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
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

var (
	homeFunStudyModelInstance     IHomeFunStudyModel
	homeFunStudyModelInstanceOnce = sync.Once{}
)

func GetHomeFunStudyModel() IHomeFunStudyModel {
	homeFunStudyModelInstanceOnce.Do(func() {
		homeFunStudyModelInstance = &homeFunStudyModel{}
	})
	return homeFunStudyModelInstance
}

var ErrHomeFunStudyHasCompleted = errors.New("home fun study has completed")

type IHomeFunStudyModel interface {
	List(ctx context.Context, operator *entity.Operator, args entity.ListHomeFunStudiesArgs) (*entity.ListHomeFunStudiesResult, error)
	Get(ctx context.Context, operator *entity.Operator, id string) (*entity.GetHomeFunStudyResult, error)
	GetPlain(ctx context.Context, operator *entity.Operator, id string) (*entity.HomeFunStudy, error)
	GetByScheduleIDAndStudentID(ctx context.Context, operator *entity.Operator, scheduleID string, studentID string) (*entity.HomeFunStudy, error)
	Save(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.SaveHomeFunStudyArgs) error
	Assess(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AssessHomeFunStudyArgs) error
}

type homeFunStudyModel struct{}

func (m *homeFunStudyModel) List(ctx context.Context, operator *entity.Operator, args entity.ListHomeFunStudiesArgs) (*entity.ListHomeFunStudiesResult, error) {
	checker := NewAssessmentPermissionChecker(operator)
	if err := checker.SearchAllPermissions(ctx); err != nil {
		log.Error(ctx, "List: checker.SearchAllPermissions: checker failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, err
	}
	if !checker.CheckStatus(args.Status) {
		return nil, constant.ErrForbidden
	}
	cond := da.QueryHomeFunStudyCondition{
		Status:          args.Status,
		OrderBy:         args.OrderBy,
		AllowTeacherIDs: checker.AllowTeacherIDs(),
		AllowPairs:      checker.AllowPairs(),
		Page:            args.Page,
		PageSize:        args.PageSize,
	}
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
	for _, t := range teachers {
		cond.TeacherIDs = append(cond.TeacherIDs, t.ID)
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
	for _, s := range students {
		cond.StudentIDs = append(cond.StudentIDs, s.ID)
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
	studentNamesMap, err := m.getStudentNamesMap(ctx, operator, studentIDs)
	if err != nil {
		log.Error(ctx, "m.getStudentNamesMap: get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
			log.Any("all_student_ids", studentIDs),
		)
		return nil, err
	}
	teacherNamesMap, err := m.getTeacherNamesMap(ctx, operator, teacherIDs)
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

func (m *homeFunStudyModel) Get(ctx context.Context, operator *entity.Operator, id string) (*entity.GetHomeFunStudyResult, error) {
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
	if !checker.CheckStatus(&study.Status) {
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

	subjects, err := external.GetSubjectServiceProvider().BatchGet(ctx, operator, []string{study.SubjectID})
	if err != nil {
		log.Error(ctx, "Get: external.GetSubjectServiceProvider().BatchGet: get subject failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("subject_id", study.SubjectID),
		)
		return nil, err
	}
	if len(subjects) == 0 {
		log.Error(ctx, "Get: external.GetSubCategoryServiceProvider().BatchGet: not found subject",
			log.Err(err),
			log.Any("operator", operator),
			log.String("subject_id", study.SubjectID),
		)
		return nil, err
	}
	subject := subjects[0]

	result := &entity.GetHomeFunStudyResult{
		ID:               study.ID,
		ScheduleID:       study.ScheduleID,
		Title:            study.Title,
		TeacherIDs:       study.TeacherIDs,
		TeacherNames:     teacherNames,
		StudentID:        study.StudentID,
		StudentName:      studentName,
		SubjectName:      subject.Name,
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

func (m *homeFunStudyModel) GetPlain(ctx context.Context, operator *entity.Operator, id string) (*entity.HomeFunStudy, error) {
	var study entity.HomeFunStudy
	if err := da.GetHomeFunStudyDA().Get(ctx, id, &study); err != nil {
		log.Error(ctx, "GetPlain: da.GetHomeFunStudyDA().Get: get failed",
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

func (m *homeFunStudyModel) GetByScheduleIDAndStudentID(ctx context.Context, operator *entity.Operator, scheduleID string, studentID string) (*entity.HomeFunStudy, error) {
	cond := da.QueryHomeFunStudyCondition{
		OrgID:      &operator.OrgID,
		ScheduleID: &scheduleID,
		StudentIDs: []string{studentID},
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

func (m *homeFunStudyModel) Save(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.SaveHomeFunStudyArgs) error {
	cond := da.QueryHomeFunStudyCondition{
		ScheduleID: &args.ScheduleID,
		StudentIDs: []string{args.StudentID},
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
		title = m.EncodeTitle(&classes[0].Name, args.LessonName)
	} else {
		title = m.EncodeTitle(nil, args.LessonName)
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
			SubjectID:        args.SubjectID,
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
		if err == sql.ErrNoRows {
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
		return ErrHomeFunStudyHasCompleted
	}
	if args.Action == entity.UpdateHomeFunStudyActionComplete {
		study.Status = entity.AssessmentStatusComplete
	}
	study.AssessFeedbackID = args.AssessFeedbackID
	study.AssessScore = args.AssessScore
	study.AssessComment = args.AssessComment
	study.CompleteAt = time.Now().Unix()
	if err := da.GetHomeFunStudyDA().SaveTx(ctx, tx, &study); err != nil {
		log.Error(ctx, "da.GetHomeFunStudyDA().SaveTx: save failed",
			log.Err(err),
			log.Any("study", study))
		return err
	}
	return nil
}

func (m *homeFunStudyModel) EncodeTitle(className *string, lessonName string) string {
	if className == nil {
		tmp := constant.HomeFunStudyNoClass
		className = &tmp
	}
	return fmt.Sprintf("%s-%s", *className, lessonName)
}

func (m *homeFunStudyModel) getTeacherNamesMap(ctx context.Context, operator *entity.Operator, teacherIDs []string) (map[string]string, error) {
	namesMap := map[string]string{}
	teachers, err := external.GetTeacherServiceProvider().BatchGet(ctx, operator, teacherIDs)
	if err != nil {
		log.Error(ctx, "external.GetTeacherServiceProvider().BatchGet: batch get failed",
			log.Err(err),
			log.Strings("teacher_ids", teacherIDs),
		)
		return nil, err
	}
	for _, t := range teachers {
		namesMap[t.ID] = t.Name
	}
	return namesMap, nil
}

func (m *homeFunStudyModel) getStudentNamesMap(ctx context.Context, operator *entity.Operator, studentIDs []string) (map[string]string, error) {
	students, err := external.GetStudentServiceProvider().BatchGet(ctx, operator, studentIDs)
	if err != nil {
		log.Error(ctx, "external.GetStudentServiceProvider().BatchGet: batch get failed",
			log.Err(err),
			log.Strings("student_ids", studentIDs),
		)
		return nil, err
	}
	namesMap := map[string]string{}
	for _, s := range students {
		namesMap[s.ID] = s.Name
	}
	return namesMap, nil
}
