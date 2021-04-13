package model

import (
	"context"
	"errors"
	"fmt"
	"sort"
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

type IAssessmentModel interface {
	GetPlain(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string) (*entity.Assessment, error)
	Get(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string) (*entity.AssessmentDetail, error)
	List(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsArgs) (*entity.ListAssessmentsResult, error)
	Add(ctx context.Context, operator *entity.Operator, args entity.AddAssessmentArgs) (string, error)
	Update(ctx context.Context, operator *entity.Operator, args entity.UpdateAssessmentArgs) error
}

var (
	assessmentModelInstance     IAssessmentModel
	assessmentModelInstanceOnce = sync.Once{}
)

func GetAssessmentModel() IAssessmentModel {
	assessmentModelInstanceOnce.Do(func() {
		assessmentModelInstance = &assessmentModel{}
	})
	return assessmentModelInstance
}

type assessmentModel struct{}

func (m *assessmentModel) GetPlain(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string) (*entity.Assessment, error) {
	assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "GetPlain: da.GetAssessmentDA().GetExcludeSoftDeleted: get assessment failed",
			log.Err(err),
			log.String("id", "id"),
			log.Any("operator", operator),
		)
		return nil, err
	}
	return assessment, nil
}

func (m *assessmentModel) Get(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string) (*entity.AssessmentDetail, error) {
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
	if views, err = m.convertToAssessmentViews(ctx, tx, operator, []*entity.Assessment{assessment}, nil); err != nil {
		log.Error(ctx, "Get: m.convertToAssessmentViews: get failed",
			log.Err(err),
			log.String("assessment_id", id),
			log.Any("operator", operator),
		)
		return nil, err
	}
	view = views[0]

	// fill partial result
	result := entity.AssessmentDetail{
		ID:           assessment.ID,
		Title:        assessment.Title,
		Status:       assessment.Status,
		CompleteTime: assessment.CompleteTime,
		Teachers:     view.Teachers,
		Students:     view.Students,
		Program:      view.Program,
		Subject:      view.Subject,
		ClassEndTime: assessment.ClassEndTime,
		ClassLength:  assessment.ClassLength,
	}

	// fill outcome attendances
	var assessmentOutcomes []*entity.AssessmentOutcome
	if err := da.GetAssessmentOutcomeDA().QueryTx(ctx, tx, &da.QueryAssessmentOutcomeConditions{
		AssessmentIDs: []string{id},
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
			if o.Checked {
				result.NumberOfOutcomes++
			}
		}
		if outcomes, err = GetOutcomeModel().GetLearningOutcomesByIDs(ctx, operator, tx, outcomeIDs); err != nil {
			log.Error(ctx, "Get: GetOutcomeModel().GetLearningOutcomesByIDs: get failed",
				log.Err(err),
				log.Strings("outcome_ids", outcomeIDs),
				log.String("assessment_id", id),
				log.Any("operator", operator),
			)
			return nil, err
		}
		sort.Sort(outcomesSortByAssumedAndName(outcomes))

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
			newOutcomeAttendances := entity.OutcomeAttendances{
				OutcomeID:     o.ID,
				OutcomeName:   o.Name,
				Assumed:       o.Assumed,
				Skip:          assessmentOutcomeMap[o.ID].Skip,
				NoneAchieved:  assessmentOutcomeMap[o.ID].NoneAchieved,
				AttendanceIDs: outcomeAttendanceIDsMap[o.ID],
				Checked:       assessmentOutcomeMap[o.ID].Checked,
			}
			result.OutcomeAttendances = append(result.OutcomeAttendances, &newOutcomeAttendances)
		}
	}

	// fill lesson plan and lesson materials
	var (
		plan                     *entity.AssessmentContent
		materials                []*entity.AssessmentContent
		contentIDs               []string
		currentContentOutcomeMap map[string][]string
	)
	if plan, err = da.GetAssessmentContentDA().GetPlan(ctx, tx, id); err != nil {
		log.Error(ctx, "Get: da.GetAssessmentContentDA().GetPlan: get failed",
			log.Err(err),
			log.String("assessment_id", id),
		)
	} else {
		contentIDs = append(contentIDs, plan.ID)
	}
	if materials, err = da.GetAssessmentContentDA().GetMaterials(ctx, tx, id); err != nil {
		log.Error(ctx, "Get: da.GetAssessmentContentDA().GetMaterials: get failed",
			log.Err(err),
			log.String("assessment_id", id),
		)
	} else {
		for _, m := range materials {
			contentIDs = append(contentIDs, m.ID)
		}
	}
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
		if plan != nil {
			result.Plan = entity.AssessmentContentView{
				ID:         plan.ContentID,
				Name:       plan.ContentName,
				Checked:    true,
				OutcomeIDs: currentContentOutcomeMap[plan.ID],
			}
		}
		for _, m := range materials {
			result.Materials = append(result.Materials, &entity.AssessmentContentView{
				ID:         m.ContentID,
				Name:       m.ContentName,
				Comment:    m.ContentComment,
				Checked:    m.Checked,
				OutcomeIDs: currentContentOutcomeMap[m.ID],
			})
		}
	}

	// fill room id and class name from schedule
	schedule, err := GetScheduleModel().GetPlainByID(ctx, assessment.ScheduleID)
	if err != nil {
		log.Error(ctx, "Get: GetScheduleModel().GetByID: get failed",
			log.Err(err),
			log.String("schedule_id", assessment.ScheduleID),
			log.Any("operator", operator),
		)
		return nil, err
	}
	result.RoomID = schedule.RoomID
	classNameMap, err := external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, []string{schedule.ClassID})
	if err != nil {
		log.Error(ctx, "Get: external.GetClassServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.String("class_id", schedule.ClassID),
			log.Any("schedule", schedule),
			log.Any("operator", operator),
		)
		return nil, err
	}
	result.Class = entity.AssessmentClass{
		ID:   schedule.ClassID,
		Name: classNameMap[schedule.ClassID],
	}

	return &result, nil
}

func (m *assessmentModel) List(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsArgs) (*entity.ListAssessmentsResult, error) {
	// check permission
	var (
		checker = NewAssessmentPermissionChecker(operator)
		err     error
	)
	if err = checker.SearchAllPermissions(ctx); err != nil {
		log.Error(ctx, "List: checker.SearchAllPermissions: search failed",
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, err
	}
	if !checker.CheckStatus(args.Status) {
		return nil, constant.ErrForbidden
	}

	// get assessment list
	var (
		assessments []*entity.Assessment
		cond        = da.QueryAssessmentConditions{
			OrgID:                   &operator.OrgID,
			Status:                  args.Status,
			AllowTeacherIDs:         checker.AllowTeacherIDs(),
			TeacherIDAndStatusPairs: checker.AllowPairs(),
			ClassType:               args.ClassType,
			OrderBy:                 args.OrderBy,
			Page:                    args.Page,
			PageSize:                args.PageSize,
		}
		teachers    []*external.Teacher
		scheduleIDs []string
	)
	if args.TeacherName != nil {
		if teachers, err = external.GetTeacherServiceProvider().Query(ctx, operator, operator.OrgID, *args.TeacherName); err != nil {
			log.Error(ctx, "List: external.GetTeacherServiceProvider().Query: query failed",
				log.Err(err),
				log.String("org_id", operator.OrgID),
				log.String("teacher_name", *args.TeacherName),
				log.Any("args", args),
				log.Any("operator", operator),
			)
			return nil, err
		}
		log.Debug(ctx, "List: external.GetTeacherServiceProvider().Query: query success",
			log.String("org_id", operator.OrgID),
			log.String("teacher_name", *args.TeacherName),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		if len(teachers) > 0 {
			for _, item := range teachers {
				cond.TeacherIDs = append(cond.TeacherIDs, item.ID)
			}
		} else {
			cond.TeacherIDs = []string{}
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
	if scheduleIDs == nil {
		cond.ScheduleIDs = []string{}
	} else {
		cond.ScheduleIDs = append(cond.ScheduleIDs, scheduleIDs...)
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
	if views, err = m.convertToAssessmentViews(ctx, tx, operator, assessments, da.RefBool(true)); err != nil {
		log.Error(ctx, "List: m.convertToAssessmentViews: get failed",
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
			Subject:      v.Subject,
			Teachers:     v.Teachers,
			ClassEndTime: v.ClassEndTime,
			CompleteTime: v.CompleteTime,
			Status:       v.Status,
		}
		result.Items = append(result.Items, &newItem)
	}

	return &result, nil
}

func (m *assessmentModel) convertToAssessmentViews(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessments []*entity.Assessment, checkedStudents *bool) ([]*entity.AssessmentView, error) {
	//
	var (
		err           error
		assessmentIDs []string
		subjectIDs    []string
		programIDs    []string
	)
	for _, a := range assessments {
		assessmentIDs = append(assessmentIDs, a.ID)
		if a.SubjectID != "" {
			subjectIDs = append(subjectIDs, a.SubjectID)
		}
		if a.ProgramID != "" {
			programIDs = append(programIDs, a.ProgramID)
		}
	}

	// fill program
	programNameMap, err := external.GetProgramServiceProvider().BatchGetNameMap(ctx, operator, programIDs)
	if err != nil {
		log.Error(ctx, "convertToAssessmentViews: external.GetProgramServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("assessment_ids", assessmentIDs),
			log.Strings("program_ids", programIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// fill subject
	subjectNameMap, err := external.GetSubjectServiceProvider().BatchGetNameMap(ctx, operator, subjectIDs)
	if err != nil {
		log.Error(ctx, "convertToAssessmentViews: external.GetSubjectServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("assessment_ids", assessmentIDs),
			log.Strings("subject_ids", subjectIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// fill students and teachers
	var (
		assessmentAttendances    []*entity.AssessmentAttendance
		assessmentAttendancesMap = map[string]*entity.AssessmentAttendance{}
		studentIDs               []string
		studentNameMap           map[string]string
		assessmentStudentsMap    = map[string][]*entity.AssessmentAttendance{}
		teacherIDs               []string
		teacherNameMap           map[string]string
		assessmentTeachersMap    = map[string][]*entity.AssessmentAttendance{}
	)
	if err := da.GetAssessmentAttendanceDA().QueryTx(ctx, tx, &da.QueryAssessmentAttendanceConditions{
		AssessmentIDs: assessmentIDs,
		Checked:       checkedStudents,
	}, &assessmentAttendances); err != nil {
		log.Error(ctx, "convertToAssessmentViews: da.GetAssessmentAttendanceDA().QueryTx: query failed",
			log.Err(err),
			log.Strings("assessment_ids", assessmentIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	for _, a := range assessmentAttendances {
		assessmentAttendancesMap[a.AttendanceID] = a
		switch a.Role {
		case entity.AssessmentAttendanceRoleStudent:
			studentIDs = append(studentIDs, a.AttendanceID)
			assessmentStudentsMap[a.AssessmentID] = append(assessmentStudentsMap[a.AssessmentID], a)
		case entity.AssessmentAttendanceRoleTeacher:
			teacherIDs = append(teacherIDs, a.AttendanceID)
			assessmentTeachersMap[a.AssessmentID] = append(assessmentTeachersMap[a.AssessmentID], a)
		}
	}
	if teacherNameMap, err = external.GetTeacherServiceProvider().BatchGetNameMap(ctx, operator, teacherIDs); err != nil {
		log.Error(ctx, "convertToAssessmentViews: external.GetTeacherServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("teacher_ids", teacherIDs),
			log.Strings("assessment_ids", assessmentIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	if studentNameMap, err = external.GetStudentServiceProvider().BatchGetNameMap(ctx, operator, studentIDs); err != nil {
		log.Error(ctx, "convertToAssessmentViews: external.GetStudentServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("student_ids", studentIDs),
			log.Strings("assessment_ids", assessmentIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}

	var result []*entity.AssessmentView
	for _, a := range assessments {
		v := entity.AssessmentView{
			Assessment: a,
			Program: entity.AssessmentProgram{
				ID:   a.ProgramID,
				Name: programNameMap[a.ProgramID],
			},
			Subject: entity.AssessmentSubject{
				ID:   a.SubjectID,
				Name: subjectNameMap[a.SubjectID],
			},
		}
		for _, t := range assessmentTeachersMap[a.ID] {
			v.Teachers = append(v.Teachers, &entity.AssessmentTeacher{
				ID:   t.AttendanceID,
				Name: teacherNameMap[t.AttendanceID],
			})
		}
		for _, s := range assessmentStudentsMap[a.ID] {
			v.Students = append(v.Students, &entity.AssessmentStudent{
				ID:      s.AttendanceID,
				Name:    studentNameMap[s.AttendanceID],
				Checked: s.Checked,
			})
		}
		result = append(result, &v)
	}

	return result, nil
}

func (m *assessmentModel) Add(ctx context.Context, operator *entity.Operator, args entity.AddAssessmentArgs) (string, error) {
	// check if assessment already exits
	var assessments []entity.Assessment
	if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
		OrgID:       &operator.OrgID,
		ScheduleIDs: []string{args.ScheduleID},
	}, &assessments); err != nil {
		log.Error(ctx, "Add: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}
	if len(assessments) > 0 {
		log.Info(ctx, "Add: assessment already exists",
			log.Any("args", args),
			log.Any("assessments", assessments),
			log.Any("operator", operator),
		)
		return "", nil
	}

	// get schedule and check class type
	var (
		schedule *entity.SchedulePlain
		err      error
	)
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

	// get contents
	var (
		latestContent   *entity.ContentInfoWithDetails
		materialIDs     []string
		materials       []*entity.SubContentsWithName
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
		if outcomes, err = GetOutcomeModel().GetLearningOutcomesByIDs(ctx, operator, dbo.MustGetDB(ctx), outcomeIDs); err != nil {
			log.Error(ctx, "Add: GetOutcomeModel().GetLearningOutcomesByIDs: get failed",
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

	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// add assessment
		var (
			err           error
			now           = time.Now().Unix()
			classNameMap  map[string]string
			newAssessment = entity.Assessment{
				ID:           newAssessmentID,
				ScheduleID:   args.ScheduleID,
				ProgramID:    schedule.ProgramID,
				SubjectID:    schedule.SubjectID,
				ClassLength:  args.ClassLength,
				ClassEndTime: args.ClassEndTime,
				CreateAt:     now,
				UpdateAt:     now,
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
			return err
		}
		newAssessment.Title = m.generateTitle(newAssessment.ClassEndTime, classNameMap[schedule.ClassID], schedule.Title)
		if _, err := da.GetAssessmentDA().InsertTx(ctx, tx, &newAssessment); err != nil {
			log.Error(ctx, "add assessment: add failed",
				log.Err(err),
				log.Any("args", args),
				log.Any("new_item", newAssessment),
			)
			return err
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
				return err
			}
			for _, u := range users {
				finalAttendanceIDs = append(finalAttendanceIDs, u.RelationID)
			}
		default:
			finalAttendanceIDs = args.AttendanceIDs
		}
		if scheduleRelations, err = GetScheduleRelationModel().GetByRelationIDs(ctx, operator, finalAttendanceIDs); err != nil {
			log.Error(ctx, "addAssessmentAttendances: GetScheduleRelationModel().GetByRelationIDs: get failed",
				log.Err(err),
				log.Any("attendance_ids", finalAttendanceIDs),
				log.String("assessment_id", newAssessmentID),
				log.Any("operator", operator),
			)
			return err
		}
		if err = m.addAssessmentAttendances(ctx, tx, operator, newAssessmentID, scheduleRelations); err != nil {
			log.Error(ctx, "Add: m.addAssessmentAttendances: add failed",
				log.Err(err),
				log.String("assessment_id", newAssessmentID),
				log.Strings("attendance_ids", finalAttendanceIDs),
				log.Any("args", args),
				log.Any("operator", operator),
			)
			return err
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
			return err
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
			return err
		}

		// add assessment contents map
		if err = m.addAssessmentContents(ctx, tx, operator, newAssessmentID, contents); err != nil {
			log.Error(ctx, "Add: m.addAssessmentContents: add failed",
				log.Err(err),
				log.String("assessment_id", newAssessmentID),
				log.Any("contents", contents),
				log.Any("args", args),
				log.Any("operator", operator),
			)
			return err
		}

		return nil
	}); err != nil {
		log.Error(ctx, "Add: tx failed",
			log.Err(err),
			log.String("assessment_id", newAssessmentID),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}

	return newAssessmentID, nil
}

func (m *assessmentModel) addAssessmentContents(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessmentID string, contents []*entity.ContentInfoWithDetails) error {
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
		log.Error(ctx, "addAssessmentContents: da.GetAssessmentContentDA().BatchInsert: batch insert failed",
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
		log.Error(ctx, "addAssessmentContents: da.GetAssessmentContentOutcomeDA().BatchInsert: batch insert failed",
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

func (m *assessmentModel) addAssessmentAttendances(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessmentID string, scheduleRelations []*entity.ScheduleRelation) error {
	if len(scheduleRelations) == 0 {
		return nil
	}
	var (
		err                   error
		assessmentAttendances []*entity.AssessmentAttendance
	)
	for _, relation := range scheduleRelations {
		newAttendance := entity.AssessmentAttendance{
			ID:           utils.NewID(),
			AssessmentID: assessmentID,
			AttendanceID: relation.RelationID,
			Checked:      true,
		}
		switch relation.RelationType {
		case entity.ScheduleRelationTypeClassRosterStudent:
			newAttendance.Origin = entity.AssessmentAttendanceOriginClassRoaster
			newAttendance.Role = entity.AssessmentAttendanceRoleStudent
		case entity.ScheduleRelationTypeClassRosterTeacher:
			newAttendance.Origin = entity.AssessmentAttendanceOriginClassRoaster
			newAttendance.Role = entity.AssessmentAttendanceRoleTeacher
		case entity.ScheduleRelationTypeParticipantStudent:
			newAttendance.Origin = entity.AssessmentAttendanceOriginParticipants
			newAttendance.Role = entity.AssessmentAttendanceRoleStudent
		case entity.ScheduleRelationTypeParticipantTeacher:
			newAttendance.Origin = entity.AssessmentAttendanceOriginParticipants
			newAttendance.Role = entity.AssessmentAttendanceRoleTeacher
		default:
			continue
		}
		assessmentAttendances = append(assessmentAttendances, &newAttendance)
	}
	if err = da.GetAssessmentAttendanceDA().BatchInsert(ctx, tx, assessmentAttendances); err != nil {
		log.Error(ctx, "addAssessmentAttendances: da.GetAssessmentAttendanceDA().BatchInsert: batch insert failed",
			log.Err(err),
			log.Any("assessment_attendances", assessmentAttendances),
			log.String("assessment_id", assessmentID),
			log.Any("scheduleRelations", scheduleRelations),
			log.Any("operator", operator),
		)
		return err
	}
	return nil
}

func (m *assessmentModel) addAssessmentOutcomes(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessmentID string, outcomes []*entity.Outcome) error {
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

func (m *assessmentModel) addOutcomeAttendances(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessmentID string, outcomes []*entity.Outcome, scheduleRelations []*entity.ScheduleRelation) error {
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

func (m *assessmentModel) Update(ctx context.Context, operator *entity.Operator, args entity.UpdateAssessmentArgs) error {
	// validate
	if !args.Action.Valid() {
		log.Error(ctx, "update assessment: invalid action", log.Any("args", args))
		return constant.ErrInvalidArgs
	}
	if args.OutcomeAttendances != nil {
		for _, item := range *args.OutcomeAttendances {
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
		log.Error(ctx, "Update: NewAssessmentPermissionChecker(operator).HasP439: check permission 439 failed",
			log.Any("args", args),
			log.Any("operator", operator),
		)
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
		return errors.New("update assessment: assessment has completed, not allow update")
	}

	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// update assessment attendance check property
		if args.AttendanceIDs != nil {
			if err := da.GetAssessmentAttendanceDA().UncheckByAssessmentID(ctx, tx, args.ID); err != nil {
				log.Error(ctx, "update assessment: uncheck assessment attendance failed",
					log.Err(err),
					log.Any("args", args),
				)
				return err
			}
			if err := da.GetAssessmentAttendanceDA().BatchCheckByAssessmentIDAndAttendanceIDs(ctx, tx, args.ID, *args.AttendanceIDs); err != nil {
				log.Error(ctx, "update assessment: check assessment attendance failed",
					log.Err(err),
					log.Any("args", args),
				)
				return err
			}
		}

		if args.OutcomeAttendances != nil {
			// update assessment outcomes map
			if err := da.GetAssessmentOutcomeDA().UncheckByAssessmentID(ctx, tx, args.ID); err != nil {
				log.Error(ctx, "Update: da.GetAssessmentOutcomeDA().UncheckByAssessmentID: uncheck assessment outcome failed by assessment id",
					log.Err(err),
					log.Any("args", args),
					log.String("id", args.ID),
				)
				return err
			}
			for _, oa := range *args.OutcomeAttendances {
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
			for _, oa := range *args.OutcomeAttendances {
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
		for _, ma := range args.Materials {
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

func (m *assessmentModel) generateTitle(classEndTime int64, className string, lessonName string) string {
	return fmt.Sprintf("%s-%s-%s", time.Unix(classEndTime, 0).Format("20060102"), className, lessonName)
}

func (m *assessmentModel) getAssessmentContentOutcomeMap(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string, contentIDs []string) (map[string]map[string][]string, error) {
	var assessmentContentOutcomes []*entity.AssessmentContentOutcome
	cond := da.QueryAssessmentContentOutcomeConditions{
		AssessmentIDs: assessmentIDs,
		ContentIDs:    contentIDs,
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
			result[co.AssessmentID] = map[string][]string{}
		} else {
			result[co.AssessmentID][co.ContentID] = append(result[co.AssessmentID][co.ContentID], co.OutcomeID)
		}
	}
	return result, nil
}

type outcomesSortByAssumedAndName []*entity.Outcome

func (s outcomesSortByAssumedAndName) Len() int {
	return len(s)
}

func (s outcomesSortByAssumedAndName) Less(i, j int) bool {
	if s[i].Assumed && !s[j].Assumed {
		return true
	} else if !s[i].Assumed && s[j].Assumed {
		return false
	} else {
		return s[i].Name < s[j].Name
	}
}

func (s outcomesSortByAssumedAndName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// endregion

// region permission checker

type AssessmentPermissionChecker struct {
	operator              *entity.Operator
	allowStatusComplete   bool
	allowStatusInProgress bool
	allowTeacherIDs       []string
	allowPairs            []*entity.AssessmentTeacherIDAndStatusPair
}

func NewAssessmentPermissionChecker(operator *entity.Operator) *AssessmentPermissionChecker {
	return &AssessmentPermissionChecker{operator: operator}
}

func (c *AssessmentPermissionChecker) SearchAllPermissions(ctx context.Context) error {
	if err := c.SearchOrgPermissions(ctx); err != nil {
		log.Error(ctx, "SearchAllPermissions: SearchOrgPermissions: check org permission failed",
			log.Err(err),
			log.Any("c", c),
		)
		return err
	}

	if err := c.SearchSchoolPermissions(ctx); err != nil {
		log.Error(ctx, "SearchAllPermissions: SearchOrgPermissions: check school permission failed",
			log.Err(err),
			log.Any("c", c),
		)
		return err
	}

	if err := c.SearchSelfPermissions(ctx); err != nil {
		log.Error(ctx, "SearchAllPermissions: SearchSelfPermissions: check self permission failed",
			log.Err(err),
			log.Any("c", c),
		)
		return err
	}

	return nil
}

func (c *AssessmentPermissionChecker) SearchOrgPermissions(ctx context.Context) error {
	hasP424, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.operator, external.AssessmentViewOrgCompletedAssessments424)
	if err != nil {
		log.Error(ctx, "SearchOrgPermissions: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 424 failed",
			log.Err(err),
			log.Any("c", c),
		)
		return err
	}

	hasP425, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.operator, external.AssessmentViewOrgInProgressAssessments425)
	if err != nil {
		log.Error(ctx, "SearchOrgPermissions: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 425 failed",
			log.Err(err),
			log.Any("c", c),
		)
		return err
	}

	if hasP424 || hasP425 {
		teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, c.operator, c.operator.OrgID)
		if err != nil {
			log.Error(ctx, "SearchOrgPermissions: external.GetTeacherServiceProvider().GetByOrganization: get teachers failed",
				log.Err(err),
				log.Any("c", c),
			)
			return err
		}
		var teacherIDs []string
		for _, teacher := range teachers {
			teacherIDs = append(teacherIDs, teacher.ID)
		}
		c.allowTeacherIDs = append(c.allowTeacherIDs, teacherIDs...)

		if hasP424 {
			c.allowStatusComplete = true
			for _, teacherID := range teacherIDs {
				c.allowPairs = append(c.allowPairs, &entity.AssessmentTeacherIDAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusComplete,
				})
			}
		}

		if hasP425 {
			c.allowStatusInProgress = true
			for _, teacherID := range teacherIDs {
				c.allowPairs = append(c.allowPairs, &entity.AssessmentTeacherIDAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusInProgress,
				})
			}
		}
	}

	return nil
}

func (c *AssessmentPermissionChecker) SearchSchoolPermissions(ctx context.Context) error {
	hasP426, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.operator, external.AssessmentViewSchoolCompletedAssessments426)
	if err != nil {
		log.Error(ctx, "SearchSchoolPermissions: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 426 failed",
			log.Err(err),
			log.Any("c", c),
		)
		return err
	}

	hasP427, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.operator, external.AssessmentViewSchoolInProgressAssessments427)
	if err != nil {
		log.Error(ctx, "SearchSchoolPermissions: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 427 failed",
			log.Err(err),
			log.Any("c", c),
		)
		return err
	}

	if hasP426 || hasP427 {
		schools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, c.operator)
		if err != nil {
			log.Error(ctx, "SearchSchoolPermissions: external.GetSchoolServiceProvider().GetByOperator: get schools failed",
				log.Err(err),
				log.Any("c", c),
			)
			return err
		}
		var schoolIDs []string
		for _, school := range schools {
			schoolIDs = append(schoolIDs, school.ID)
		}
		schoolID2TeachersMap, err := external.GetTeacherServiceProvider().GetBySchools(ctx, c.operator, schoolIDs)
		if err != nil {
			log.Error(ctx, "SearchSchoolPermissions: external.GetTeacherServiceProvider().GetBySchools: get teachers failed",
				log.Err(err),
				log.Any("c", c),
				log.Any("school_ids", schoolIDs),
			)
			return err
		}

		var teacherIDs []string
		for _, teachers := range schoolID2TeachersMap {
			for _, teacher := range teachers {
				teacherIDs = append(teacherIDs, teacher.ID)
			}
		}
		c.allowTeacherIDs = append(c.allowTeacherIDs, teacherIDs...)

		if hasP426 {
			c.allowStatusComplete = true
			for _, teacherID := range c.allowTeacherIDs {
				c.allowPairs = append(c.allowPairs, &entity.AssessmentTeacherIDAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusComplete,
				})
			}
		}

		if hasP427 {
			c.allowStatusInProgress = true
			for _, teacherID := range teacherIDs {
				c.allowPairs = append(c.allowPairs, &entity.AssessmentTeacherIDAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusInProgress,
				})
			}
		}
	}

	return nil
}

func (c *AssessmentPermissionChecker) SearchSelfPermissions(ctx context.Context) error {
	hasP414, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.operator, external.AssessmentViewCompletedAssessments414)
	if err != nil {
		log.Error(ctx, "SearchSelfPermissions: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 414 failed",
			log.Err(err),
			log.Any("c", c),
		)
		return err
	}
	hasP415, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.operator, external.AssessmentViewInProgressAssessments415)
	if err != nil {
		log.Error(ctx, "SearchSelfPermissions: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 415 failed",
			log.Err(err),
			log.Any("c", c),
		)
		return err
	}

	if hasP414 || hasP415 {
		if hasP414 {
			c.allowStatusComplete = true
			c.allowPairs = append(c.allowPairs, &entity.AssessmentTeacherIDAndStatusPair{
				TeacherID: c.operator.UserID,
				Status:    entity.AssessmentStatusComplete,
			})
		}
		if hasP415 {
			c.allowStatusInProgress = true
			c.allowPairs = append(c.allowPairs, &entity.AssessmentTeacherIDAndStatusPair{
				TeacherID: c.operator.UserID,
				Status:    entity.AssessmentStatusInProgress,
			})
		}
		c.allowTeacherIDs = append(c.allowTeacherIDs, c.operator.UserID)
	}

	return nil
}

func (c *AssessmentPermissionChecker) AllowTeacherIDs() []string {
	return c.allowTeacherIDs
}

func (c *AssessmentPermissionChecker) AllowPairs() []*entity.AssessmentTeacherIDAndStatusPair {
	var result []*entity.AssessmentTeacherIDAndStatusPair

	m := map[string]*struct {
		allowStatusInProgress bool
		allowStatusComplete   bool
	}{}
	for _, pair := range c.allowPairs {
		if m[pair.TeacherID] == nil {
			m[pair.TeacherID] = &struct {
				allowStatusInProgress bool
				allowStatusComplete   bool
			}{}
		}
		switch pair.Status {
		case entity.AssessmentStatusComplete:
			m[pair.TeacherID].allowStatusComplete = true
		case entity.AssessmentStatusInProgress:
			m[pair.TeacherID].allowStatusInProgress = true
		}
	}

	for teacherID, statuses := range m {
		if statuses.allowStatusInProgress && statuses.allowStatusComplete {
			continue
		}
		if !statuses.allowStatusInProgress && !statuses.allowStatusComplete {
			continue
		}
		if statuses.allowStatusComplete {
			result = append(result, &entity.AssessmentTeacherIDAndStatusPair{
				TeacherID: teacherID,
				Status:    entity.AssessmentStatusComplete,
			})
		}
		if statuses.allowStatusInProgress {
			result = append(result, &entity.AssessmentTeacherIDAndStatusPair{
				TeacherID: teacherID,
				Status:    entity.AssessmentStatusInProgress,
			})
		}
	}

	return result
}

func (c *AssessmentPermissionChecker) AllowStatuses() []entity.AssessmentStatus {
	var result []entity.AssessmentStatus
	if c.allowStatusInProgress {
		result = append(result, entity.AssessmentStatusInProgress)
	}
	if c.allowStatusComplete {
		result = append(result, entity.AssessmentStatusComplete)
	}
	return result
}

func (c *AssessmentPermissionChecker) CheckTeacherIDs(ids []string) bool {
	allow := c.AllowTeacherIDs()
	for _, id := range ids {
		for _, a := range allow {
			if a == id {
				return true
			}
		}
	}
	return false
}

func (c *AssessmentPermissionChecker) CheckStatus(s *entity.AssessmentStatus) bool {
	if s == nil {
		return true
	}
	if !s.Valid() {
		return true
	}
	allow := c.AllowStatuses()
	for _, a := range allow {
		if a == *s {
			return true
		}
	}
	return false
}

func (c *AssessmentPermissionChecker) HasP439(ctx context.Context) (bool, error) {
	hasP439, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, c.operator, external.AssessmentEditInProgressAssessment439)
	if err != nil {
		log.Error(ctx, "HasP439: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 439 failed",
			log.Err(err),
			log.Any("operator", c.operator),
			log.Any("c", c),
		)
		return false, err
	}
	return hasP439, nil
}

// endregion
