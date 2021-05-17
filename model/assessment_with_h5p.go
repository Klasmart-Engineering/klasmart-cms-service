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

type IH5PAssessmentModel interface {
	List(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.ListH5PAssessmentsArgs) (*entity.ListH5PAssessmentsResult, error)
	GetDetail(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, id string) (*entity.GetH5PAssessmentDetailResult, error)
	Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.UpdateH5PAssessmentArgs) error
	AddClassAndLive(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AddAssessmentArgs) (string, error)
	DeleteStudies(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) error
	AddStudies(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) ([]string, error)
}

var (
	h5pAssessmentModelInstance     IH5PAssessmentModel
	h5pAssessmentModelInstanceOnce = sync.Once{}
)

func GetH5PAssessmentModel() IH5PAssessmentModel {
	h5pAssessmentModelInstanceOnce.Do(func() {
		h5pAssessmentModelInstance = &h5pAssessmentModel{}
	})
	return h5pAssessmentModelInstance
}

type h5pAssessmentModel struct{}

func (m *h5pAssessmentModel) List(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.ListH5PAssessmentsArgs) (*entity.ListH5PAssessmentsResult, error) {
	// check args
	if !args.Type.Valid() {
		err := errors.New("List: require assessment type")
		log.Error(ctx, "check args error", log.Err(err), log.Any("args", args), log.Any("operator", operator))
		return nil, err
	}

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
	if !checker.CheckStatus(args.Status.Value) {
		return nil, constant.ErrForbidden
	}

	// get assessment list
	var (
		assessments []*entity.Assessment
		cond        = da.QueryAssessmentConditions{
			Type: entity.NullAssessmentType{
				Value: args.Type,
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
	)

	if args.Query != "" {
		switch args.QueryType {
		case entity.ListH5PAssessmentsQueryTypeTeacherName:
			cond.TeacherIDs.Valid = true
			teachers, err := external.GetTeacherServiceProvider().Query(ctx, operator, operator.OrgID, args.Query)
			if err != nil {
				return nil, err
			}
			for _, item := range teachers {
				cond.TeacherIDs.Strings = append(cond.TeacherIDs.Strings, item.ID)
			}
		case entity.ListH5PAssessmentsQueryTypeClassName:
			cond.ClassIDs.Valid = true
			// TODO: Medivh: query classs by name
		default:
			cond.ClassIDsOrTeacherIDs.Valid = true
			teachers, err := external.GetTeacherServiceProvider().Query(ctx, operator, operator.OrgID, args.Query)
			if err != nil {
				return nil, err
			}
			for _, item := range teachers {
				cond.ClassIDsOrTeacherIDs.Value.TeacherIDs = append(cond.ClassIDsOrTeacherIDs.Value.TeacherIDs, item.ID)
			}
			// TODO: Medivh: query classs by name 2
		}
	}
	log.Debug(ctx, "List: print query cond", log.Any("cond", cond))
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
	if views, err = GetAssessmentModel().ConvertToViews(ctx, tx, operator, assessments, entity.ConvertToViewsOptions{
		CheckedStudents:  sql.NullBool{Bool: true, Valid: true},
		EnableProgram:    true,
		EnableSubjects:   true,
		EnableTeachers:   true,
		EnableStudents:   true,
		EnableClass:      true,
		EnableLessonPlan: true,
	}); err != nil {
		log.Error(ctx, "List: GetAssessmentModel().ConvertToViews: get failed",
			log.Err(err),
			log.Any("assessments", assessments),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// fill activities
	var allStudentIDs []string
	for _, v := range views {
		for _, s := range v.Students {
			if s.Checked {
				allStudentIDs = append(allStudentIDs, s.ID)
			}
		}
	}
	allStudentIDs = utils.SliceDeduplication(allStudentIDs)
	studentActivitiesMap, err := external.GetH5PServiceProvider().BatchGetMap(ctx, operator, allStudentIDs)
	if err != nil {
		return nil, err
	}

	// construct result
	var result = entity.ListH5PAssessmentsResult{Total: total}
	for _, v := range views {
		teacherNames := make([]string, 0, len(v.Teachers))
		for _, t := range v.Teachers {
			teacherNames = append(teacherNames, t.Name)
		}
		var (
			activityAllAttempted int
			activityAllCount     int
		)
		for _, s := range v.Students {
			aa := studentActivitiesMap[s.ID]
			for _, a := range aa {
				activityAllCount++
				if a.Attempted {
					activityAllAttempted++
				}
			}
		}
		var completeRate float64
		if activityAllCount != 0 {
			completeRate = float64(activityAllAttempted) / float64(activityAllCount)
		}
		var remainingTime int64
		if v.Schedule.DueAt != 0 {
			remainingTime = time.Now().Unix() - v.Schedule.DueAt
		} else {
			remainingTime = time.Now().Unix() - v.CreateAt
		}
		if remainingTime < 0 {
			remainingTime = 0
		}
		newItem := entity.ListH5PAssessmentsResultItem{
			ID:            v.ID,
			Title:         v.Title,
			TeacherNames:  teacherNames,
			ClassName:     v.Class.Name,
			DueAt:         v.Schedule.DueAt,
			CompleteRate:  completeRate,
			RemainingTime: remainingTime,
			CompleteAt:    v.CompleteTime,
			ScheduleID:    v.ScheduleID,
		}
		result.Items = append(result.Items, &newItem)
	}

	return &result, nil
}

func (m *h5pAssessmentModel) GetDetail(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, id string) (*entity.GetH5PAssessmentDetailResult, error) {
	assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "get detail: get assessment failed",
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
	if views, err = GetAssessmentModel().ConvertToViews(ctx, tx, operator, []*entity.Assessment{assessment}, entity.ConvertToViewsOptions{
		EnableProgram:  true,
		EnableSubjects: true,
		EnableTeachers: true,
		EnableStudents: true,
		EnableClass:    true,
	}); err != nil {
		log.Error(ctx, "Get: GetAssessmentModel().ConvertToViews: get failed",
			log.Err(err),
			log.String("assessment_id", id),
			log.Any("operator", operator),
		)
		return nil, err
	}
	view = views[0]

	// construct result
	result := entity.GetH5PAssessmentDetailResult{
		ID:               view.ID,
		Title:            view.Title,
		ClassName:        view.Class.Name,
		Teachers:         view.Teachers,
		Students:         view.Students,
		DueAt:            view.Schedule.DueAt,
		LessonPlan:       entity.H5PAssessmentLessonPlan{},
		LessonMaterials:  nil,
		CompleteRate:     0,
		CompleteAt:       view.CompleteTime,
		RemainingTime:    0,
		StudentViewItems: nil,
		ScheduleID:       view.ScheduleID,
		Status:           view.Status,
	}

	// remaining time
	if view.Schedule.DueAt != 0 {
		result.RemainingTime = time.Now().Unix() - view.Schedule.DueAt
	} else {
		result.RemainingTime = time.Now().Unix() - view.CreateAt
	}

	// fill lesson plan and lesson materials
	plan, err := da.GetAssessmentContentDA().GetPlan(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "Get: da.GetAssessmentContentDA().GetPlan: get failed",
			log.Err(err),
			log.String("assessment_id", id),
		)
	}
	result.LessonPlan = entity.H5PAssessmentLessonPlan{
		ID:      plan.ContentID,
		Name:    plan.ContentName,
		Comment: plan.ContentComment,
	}
	materials, err := da.GetAssessmentContentDA().GetMaterials(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "Get: da.GetAssessmentContentDA().GetMaterials: get failed",
			log.Err(err),
			log.String("assessment_id", id),
		)
	}
	for _, m := range materials {
		result.LessonMaterials = append(result.LessonMaterials, &entity.H5PAssessmentLessonMaterial{
			ID:      m.ContentID,
			Name:    m.ContentName,
			Comment: m.ContentComment,
			Checked: m.Checked,
		})
	}

	// get activities
	var allStudentIDs []string
	for _, v := range views {
		for _, s := range v.Students {
			if s.Checked {
				allStudentIDs = append(allStudentIDs, s.ID)
			}
		}
	}
	allStudentIDs = utils.SliceDeduplication(allStudentIDs)
	studentActivityMap, err := external.GetH5PServiceProvider().BatchGetMap(ctx, operator, allStudentIDs)
	if err != nil {
		return nil, err
	}

	// complete rate
	var (
		activityAllAttempted int
		activityAllCount     int
	)
	for _, s := range view.Students {
		aa := studentActivityMap[s.ID]
		for _, a := range aa {
			activityAllCount++
			if a.Attempted {
				activityAllAttempted++
			}
		}
	}
	if activityAllCount != 0 {
		result.CompleteRate = float64(activityAllAttempted) / float64(activityAllCount)
	}

	// get student comment
	var attendances []*entity.AssessmentAttendance
	if err := da.GetAssessmentAttendanceDA().Query(ctx, &da.QueryAssessmentAttendanceConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: []string{id},
			Valid:   true,
		},
	}, &attendances); err != nil {
		log.Error(ctx, "get h5p assessment detail: get failed",
			log.Err(err),
			log.String("id", id),
		)
	}
	attendanceCommentMap := make(map[string]string, len(attendances))

	// student view items
	for _, s := range view.Students {
		newItem := &entity.H5PAssessmentStudentViewItem{
			StudentID:       s.ID,
			StudentName:     s.Name,
			Comment:         attendanceCommentMap[s.ID],
			LessonMaterials: nil,
		}
		activityMap := studentActivityMap[s.ID]
		for _, lm := range view.LessonMaterials {
			var activity *external.H5PStudentScore
			if activityMap != nil {
				activity = activityMap[lm.Source]
			}
			if activity == nil {
				activity = &external.H5PStudentScore{}
			}
			newItem.LessonMaterials = append(newItem.LessonMaterials, &entity.H5PAssessmentStudentViewLessonMaterial{
				LessonMaterialID:   lm.ID,
				LessonMaterialName: lm.Name,
				LessonMaterialType: string(activity.ActivityType),
				Answer:             activity.Answer,
				MaxScore:           activity.MaxPossibleScore,
				AchievedScore:      activity.AchievedScore,
			})
		}
		result.StudentViewItems = append(result.StudentViewItems, newItem)
	}

	return &result, nil
}

func (m *h5pAssessmentModel) Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.UpdateH5PAssessmentArgs) error {
	panic("implement me")
}

func (m *h5pAssessmentModel) AddClassAndLive(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AddAssessmentArgs) (string, error) {
	log.Debug(ctx, "AddClassAndLive: add assessment args", log.Any("args", args), log.Any("operator", operator))

	// clean data
	args.AttendanceIDs = utils.SliceDeduplicationExcludeEmpty(args.AttendanceIDs)

	// check if assessment already exits
	var assessments []entity.Assessment
	if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
		Type: entity.NullAssessmentType{
			Value: entity.AssessmentTypeClassAndLiveH5P,
			Valid: true,
		},
		OrgID: entity.NullString{
			String: operator.OrgID,
			Valid:  true,
		},
		ScheduleIDs: entity.NullStrings{
			Strings: []string{args.ScheduleID},
			Valid:   true,
		},
	}, &assessments); err != nil {
		log.Error(ctx, "AddClassAndLive: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}
	if len(assessments) > 0 {
		log.Info(ctx, "AddClassAndLive: assessment already exists",
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
		log.Error(ctx, "AddClassAndLive: GetScheduleModel().GetPlainByID: get failed",
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
		log.Info(ctx, "AddClassAndLive: invalid schedule class type",
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
		lessonPlan      *entity.ContentInfoWithDetails
		materialIDs     []string
		materials       []*SubContentsWithName
		materialDetails []*entity.ContentInfoWithDetails
		contents        []*entity.ContentInfoWithDetails
	)
	if lessonPlan, err = GetContentModel().GetVisibleContentByID(ctx, dbo.MustGetDB(ctx), schedule.LessonPlanID, operator); err != nil {
		log.Warn(ctx, "AddClassAndLive: GetContentModel().GetVisibleContentByID: get latest content failed",
			log.Err(err),
			log.Any("args", args),
			log.String("lesson_plan_id", schedule.LessonPlanID),
			log.Any("schedule", schedule),
			log.Any("operator", operator),
		)
	} else {
		contents = append(contents, lessonPlan)
		if materials, err = GetContentModel().GetContentSubContentsByID(ctx, dbo.MustGetDB(ctx), lessonPlan.ID, operator); err != nil {
			log.Warn(ctx, "AddClassAndLive: GetContentModel().GetContentSubContentsByID: get materials failed",
				log.Err(err),
				log.Any("args", args),
				log.String("latest_lesson_plan_id", lessonPlan.ID),
				log.Any("latest_content", lessonPlan),
				log.Any("operator", operator),
				log.Any("schedule", schedule),
			)
		} else {
			for _, m := range materials {
				materialIDs = append(materialIDs, m.ID)
			}
			materialIDs = utils.SliceDeduplicationExcludeEmpty(materialIDs)
			if materialDetails, err = GetContentModel().GetContentByIDList(ctx, dbo.MustGetDB(ctx), materialIDs, operator); err != nil {
				log.Warn(ctx, "AddClassAndLive: GetContentModel().GetContentByIDList: get contents failed",
					log.Err(err),
					log.Strings("material_ids", materialIDs),
					log.Any("latest_content", lessonPlan),
					log.Any("schedule", schedule),
					log.Any("args", args),
					log.Any("operator", operator),
				)
			} else {
				contents = append(contents, materialDetails...)
			}
		}
	}

	// generate new assessment id
	var newAssessmentID = utils.NewID()

	// add assessment
	var (
		classNameMap  map[string]string
		newAssessment = entity.Assessment{
			ID:           newAssessmentID,
			ScheduleID:   args.ScheduleID,
			Type:         entity.AssessmentTypeClassAndLiveH5P,
			ClassLength:  args.ClassLength,
			ClassEndTime: args.ClassEndTime,
			Status:       entity.AssessmentStatusInProgress,
		}
	)

	if classNameMap, err = external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, []string{schedule.ClassID}); err != nil {
		log.Error(ctx, "Add: external.GetClassServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("class_ids", []string{schedule.ClassID}),
			log.Any("args", args),
		)
		return "", err
	}
	className := classNameMap[schedule.ClassID]
	if className == "" {
		className = constant.AssessmentNoClass
	}
	newAssessment.Title = fmt.Sprintf("%s-%s", className, lessonPlan.Name)
	if _, err := da.GetAssessmentDA().InsertTx(ctx, tx, &newAssessment); err != nil {
		log.Error(ctx, "add assessment: add failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("new_item", newAssessment),
		)
		return "", err
	}

	// add attendances
	var finalAttendanceIDs []string
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
	if err = GetAssessmentModel().AddAttendances(ctx, tx, operator, entity.AddAttendancesInput{
		AssessmentID:  newAssessmentID,
		ScheduleID:    schedule.ID,
		AttendanceIDs: finalAttendanceIDs,
	}); err != nil {
		log.Error(ctx, "Add: GetAssessmentModel().AddAttendances: add failed",
			log.Err(err),
			log.String("assessment_id", newAssessmentID),
			log.Strings("attendance_ids", finalAttendanceIDs),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}

	// add contents
	if err = GetAssessmentModel().AddContents(ctx, tx, operator, newAssessmentID, contents); err != nil {
		log.Error(ctx, "Add: GetAssessmentModel().AddContents: add failed",
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

func (m *h5pAssessmentModel) AddStudies(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) ([]string, error) {
	log.Debug(ctx, "AddStudies: add assessment args", log.Strings("schedule_ids", scheduleIDs), log.Any("operator", operator))

	// check if assessment already exits
	var assessments []*entity.Assessment
	if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
		Type: entity.NullAssessmentType{
			Value: entity.AssessmentTypeClassAndLiveH5P,
			Valid: true,
		},
		OrgID: entity.NullString{
			String: operator.OrgID,
			Valid:  true,
		},
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}, &assessments); err != nil {
		log.Error(ctx, "AddStudies: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Strings("schedule_id", scheduleIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	if len(assessments) > 0 {
		log.Info(ctx, "AddStudies: assessment already exists",
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("assessments", assessments),
			log.Any("operator", operator),
		)
		return nil, nil
	}

	// get schedules
	var (
		schedules []*entity.ScheduleVariable
		err       error
	)
	if schedules, err = GetScheduleModel().GetVariableDataByIDs(ctx, operator, scheduleIDs, nil); err != nil {
		log.Error(ctx, "AddStudies: GetScheduleModel().GetPlainByID: get failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	if len(schedules) == 0 {
		return nil, nil
	}

	// fix: permission
	operator.OrgID = schedules[0].OrgID

	// get class name map
	var (
		classIDs     []string
		classNameMap map[string]string
	)
	for _, s := range schedules {
		classIDs = append(classIDs, s.ClassID)
	}
	if classNameMap, err = external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, classIDs); err != nil {
		log.Error(ctx, "AddStudies: external.GetClassServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("class_ids", classIDs),
			log.Any("schedule_ids", scheduleIDs),
		)
		return nil, err
	}

	// get contents
	var (
		lessonPlanIDs []string
		lessonPlanMap map[string]*entity.AssessmentExternalLessonPlan
	)
	for _, s := range schedules {
		lessonPlanIDs = append(lessonPlanIDs, s.LessonPlanID)
	}
	if lessonPlanMap, err = GetAssessmentModel().BatchGetLatestLessonPlanMap(ctx, tx, operator, lessonPlanIDs); err != nil {
		log.Error(ctx, "AddStudies: GetAssessmentModel().BatchGetLatestLessonPlanMap: get failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
		)
		return nil, err
	}

	// add assessment
	newAssessments := make([]*entity.Assessment, 0, len(scheduleIDs))
	for _, s := range schedules {
		className := classNameMap[s.ClassID]
		if className == "" {
			className = constant.AssessmentNoClass
		}
		newAssessments = append(newAssessments, &entity.Assessment{
			ID:         utils.NewID(),
			Title:      fmt.Sprintf("%s-%s", className, lessonPlanMap[s.LessonPlanID].Name),
			ScheduleID: s.ID,
			Type:       entity.AssessmentTypeClassAndLiveH5P,
			Status:     entity.AssessmentStatusInProgress,
		})
	}

	if err := da.GetAssessmentDA().BatchInsert(ctx, tx, newAssessments); err != nil {
		log.Error(ctx, "add studies: add failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("new_assessments", newAssessments),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// add attendances
	relations, err := GetScheduleRelationModel().Query(ctx, operator, &da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
		RelationTypes: entity.NullStrings{
			Strings: []string{
				entity.ScheduleRelationTypeClassRosterTeacher.String(),
				entity.ScheduleRelationTypeClassRosterStudent.String(),
				entity.ScheduleRelationTypeParticipantTeacher.String(),
				entity.ScheduleRelationTypeParticipantStudent.String(),
			},
			Valid: true,
		},
	})
	if err != nil {
		log.Error(ctx, "add studies: query schedule relation failed", log.Err(err), log.Strings("schedule_ids", scheduleIDs))
		return nil, err
	}
	scheduleIDToAttendanceIDsMap := map[string][]string{}
	for _, r := range relations {
		scheduleIDToAttendanceIDsMap[r.ScheduleID] = append(scheduleIDToAttendanceIDsMap[r.ScheduleID], r.RelationID)
	}
	input := entity.BatchAddAttendancesInput{}
	for _, a := range assessments {
		input.Items = append(input.Items, &entity.BatchAddAttendancesInputItem{
			AssessmentID:  a.ID,
			ScheduleID:    a.ScheduleID,
			AttendanceIDs: scheduleIDToAttendanceIDsMap[a.ScheduleID],
		})
	}
	if err = GetAssessmentModel().BatchAddAttendances(ctx, tx, operator, input); err != nil {
		log.Error(ctx, "Add: GetAssessmentModel().AddAttendances: add failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// add contents
	var (
		scheduleMap        = make(map[string]*entity.ScheduleVariable, len(schedules))
		assessmentContents []*entity.AssessmentContent
	)
	for _, s := range schedules {
		scheduleMap[s.ID] = s
	}
	for _, a := range assessments {
		lp := lessonPlanMap[scheduleMap[a.ScheduleID].LessonPlanID]
		assessmentContents = append(assessmentContents, &entity.AssessmentContent{
			ID:           utils.NewID(),
			AssessmentID: a.ID,
			ContentID:    lp.ID,
			ContentName:  lp.Name,
			ContentType:  entity.ContentTypePlan,
			Checked:      true,
		})
		for _, lm := range lp.Materials {
			assessmentContents = append(assessmentContents, &entity.AssessmentContent{
				ID:            utils.NewID(),
				AssessmentID:  a.ID,
				ContentID:     lm.ID,
				ContentName:   lm.Name,
				ContentType:   entity.ContentTypeMaterial,
				ContentSource: lm.Source,
				Checked:       true,
			})
		}
	}
	if err := da.GetAssessmentContentDA().BatchInsert(ctx, tx, assessmentContents); err != nil {
		log.Error(ctx, "addAssessmentContentsAndOutcomes: da.GetAssessmentContentDA().BatchInsert: batch insert failed",
			log.Err(err),
			log.Any("schedule_ids", scheduleIDs),
			log.Any("assessment_contents", assessmentContents),
			log.Any("operator", operator),
		)
		return nil, err
	}

	var newAssessmentIDs []string
	for _, a := range assessments {
		newAssessmentIDs = append(newAssessmentIDs, a.ID)
	}

	return newAssessmentIDs, nil
}

func (m *h5pAssessmentModel) DeleteStudies(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) error {
	if len(scheduleIDs) == 0 {
		return nil
	}
	var assessments []entity.Assessment
	if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
		Type: entity.NullAssessmentType{
			Value: entity.AssessmentTypeClassAndLiveH5P,
			Valid: true,
		},
		OrgID: entity.NullString{
			String: operator.OrgID,
			Valid:  true,
		},
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}, &assessments); err != nil {
		log.Error(ctx, "DeleteStudy: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("operator", operator),
		)
		return err
	}
	for _, a := range assessments {
		if err := da.GetAssessmentDA().SoftDelete(ctx, tx, a.ID); err != nil {
			log.Error(ctx, "DeleteStudy: da.GetAssessmentDA().SoftDelete",
				log.Err(err),
				log.Strings("schedule_ids", scheduleIDs),
			)
			return err
		}
	}
	return nil
}
