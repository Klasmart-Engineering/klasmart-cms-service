package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
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
	Save(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.SaveHomeFunStudyArgs) error
	Assess(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AssessHomeFunStudyArgs) error
}

type homeFunStudyModel struct{}

func (m *homeFunStudyModel) List(ctx context.Context, operator *entity.Operator, args entity.ListHomeFunStudiesArgs) (*entity.ListHomeFunStudiesResult, error) {
	filter := NewAssessmentTeacherAndStatusFilter(operator)
	allowed, err := filter.FilterAll(ctx)
	if err != nil {
		log.Error(ctx, "List: filter.FilterAll: filter failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, err
	}
	if !allowed {
		return nil, constant.ErrForbidden
	}
	if !filter.CheckStatus(args.Status) {
		return nil, constant.ErrForbidden
	}
	cond := da.QueryHomeFunStudyCondition{
		Status:          args.Status,
		OrderBy:         args.OrderBy,
		AllowTeacherIDs: filter.AllowTeacherIDs(),
		AllowPairs:      filter.AllowPairs(),
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
		allStudentIDs []string
		allTeacherIDs []string
	)
	for _, item := range items {
		allStudentIDs = append(allStudentIDs, item.StudentID)
		teacherIDs, err := m.DecodeTeacherIDs(ctx, item.TeacherIDs)
		if err != nil {
			log.Error(ctx, "m.DecodeTeacherIDs: json decode failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("args", args),
				log.Any("teacher_ids", item.TeacherIDs),
				log.Any("items", items),
			)
			return nil, err
		}
		allTeacherIDs = append(allTeacherIDs, teacherIDs...)
	}
	studentNamesMap, err := m.getStudentNamesMap(ctx, operator, allStudentIDs)
	if err != nil {
		log.Error(ctx, "m.getStudentNamesMap: get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
			log.Any("all_student_ids", allStudentIDs),
		)
		return nil, err
	}
	teacherNamesMap, err := m.getTeacherNamesMap(ctx, operator, allTeacherIDs)
	if err != nil {
		log.Error(ctx, "m.getTeacherNamesMap: get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
			log.Any("all_teacher_ids", allTeacherIDs),
		)
		return nil, err
	}

	result := entity.ListHomeFunStudiesResult{Total: total}
	for _, item := range items {
		var teacherNames []string
		teacherIDs, err := m.DecodeTeacherIDs(ctx, item.TeacherIDs)
		if err != nil {
			log.Error(ctx, "m.DecodeTeacherIDs: json decode failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("args", args),
				log.Any("teacher_ids", item.TeacherIDs),
				log.Any("items", items),
			)
			return nil, err
		}
		for _, id := range teacherIDs {
			teacherNames = append(teacherNames, teacherNamesMap[id])
		}
		studentName := studentNamesMap[item.StudentID]
		result.Items = append(result.Items, &entity.ListHomeFunStudiesResultItem{
			ID:               item.ID,
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
	var item entity.HomeFunStudy
	if err := da.GetHomeFunStudyDA().Get(ctx, id, &item); err != nil {
		log.Error(ctx, "da.GetHomeFunStudyDA().Get: get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("id", id),
		)
		return nil, err
	}

	var (
		studentName  string
		teacherNames []string
	)
	student, err := external.GetStudentServiceProvider().Get(ctx, operator, id)
	if err != nil {
		log.Error(ctx, "external.GetStudentServiceProvider().Get: get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("id", id),
		)
		return nil, err
	}
	studentName = student.Name
	teacherIDs, err := m.DecodeTeacherIDs(ctx, item.TeacherIDs)
	if err != nil {
		log.Error(ctx, "m.DecodeTeacherIDs: decode json failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("teacherIDs", item.TeacherIDs),
			log.Any("item", item),
		)
		return nil, err
	}
	teachers, err := external.GetTeacherServiceProvider().BatchGet(ctx, operator, teacherIDs)
	if err != nil {
		log.Error(ctx, "external.GetTeacherServiceProvider().BatchGet: batch get failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("teacherIDs", item.TeacherIDs),
		)
		return nil, err
	}
	for _, t := range teachers {
		teacherNames = append(teacherNames, t.Name)
	}

	return &entity.GetHomeFunStudyResult{
		ID:               item.ID,
		Title:            item.Title,
		TeacherNames:     teacherNames,
		StudentName:      studentName,
		DueAt:            item.DueAt,
		CompleteAt:       item.CompleteAt,
		AssessFeedbackID: item.AssessFeedbackID,
		AssessScore:      item.AssessScore,
		AssessComment:    item.AssessComment,
	}, nil
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

	jsonTeacherIDs, err := m.EncodeTeacherIDs(ctx, args.TeacherIDs)
	if err != nil {
		log.Error(ctx, "external.GetClassServiceProvider().BatchGet: batch get class failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("teacher_ids", args.TeacherIDs),
			log.Any("args", args),
		)
		return err
	}

	if study != nil {
		study.Title = title
		study.TeacherIDs = jsonTeacherIDs
		study.DueAt = args.DueAt
		study.LatestFeedbackID = args.LatestFeedbackID
		study.LatestFeedbackAt = args.LatestFeedbackAt
		study.UpdateAt = time.Now().Unix()
	} else {
		now := time.Now().Unix()
		study = &entity.HomeFunStudy{
			ScheduleID:       args.ScheduleID,
			Title:            title,
			TeacherIDs:       jsonTeacherIDs,
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
	if err := m.hasAssessPermission(ctx, operator); err != nil {
		log.Error(ctx, "m.hasAssessPermission: check assess permission",
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
		return err
	}
	if study.Status == entity.AssessmentStatusComplete {
		return ErrHomeFunStudyHasCompleted
	}
	switch args.Action {
	case entity.UpdateHomeFunStudyActionComplete:
		study.Status = entity.AssessmentStatusComplete
		fallthrough
	case entity.UpdateHomeFunStudyActionSave:
		study.AssessFeedbackID = args.AssessFeedbackID
		study.AssessScore = args.AssessScore
		study.AssessComment = args.AssessComment
	}
	if err := da.GetHomeFunStudyDA().SaveTx(ctx, tx, study); err != nil {
		log.Error(ctx, "da.GetHomeFunStudyDA().SaveTx: save failed",
			log.Err(err),
			log.Any("study", study))
		return err
	}
	return nil
}

func (m *homeFunStudyModel) DecodeTeacherIDs(ctx context.Context, s string) ([]string, error) {
	if s == "[]" {
		return nil, nil
	}
	var result []string
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		log.Error(ctx, "json.Unmarshal: decode teacher ids failed",
			log.Err(err),
			log.String("teacher_ids", s),
		)
		return nil, err
	}
	return result, nil
}

func (m *homeFunStudyModel) EncodeTeacherIDs(ctx context.Context, teacherIDs []string) (string, error) {
	if len(teacherIDs) == 0 {
		return "[]", nil
	}
	bs, err := json.Marshal(teacherIDs)
	if err != nil {
		log.Error(ctx, "EncodeTeacherIDs: json marshal error",
			log.Err(err),
			log.Strings("teacher_ids", teacherIDs),
		)
		return "", err
	}
	return string(bs), nil
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

func (m *homeFunStudyModel) hasAssessPermission(ctx context.Context, operator *entity.Operator) error {
	if operator.Role == entity.RoleTeacher {
		return nil
	}
	hasP439, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.AssessmentEditInProgressAssessment439)
	if err != nil {
		log.Error(ctx, "external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 439 failed",
			log.Any("operator", operator),
		)
		return err
	}
	if !hasP439 {
		log.Error(ctx, "hasAssessPermission: not have permission 439",
			log.Any("operator", operator),
		)
		return constant.ErrForbidden
	}
	return nil
}
