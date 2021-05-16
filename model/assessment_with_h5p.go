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
	Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.UpdateH5PAssessmentArgs)
	AddClassAndLive(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AddAssessmentArgs) (string, error)
	AddStudy(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleID string) (string, error)
	DeleteStudy(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleID string) error
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
		CheckedStudents:       sql.NullBool{Bool: true, Valid: true},
		EnableProgram:         true,
		EnableSubjects:        true,
		EnableTeachers:        true,
		EnableStudents:        true,
		EnableClass:           true,
		EnableLessonPlan:      true,
		EnableLessonMaterials: true,
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
	panic("implement me")
}

func (m *h5pAssessmentModel) Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.UpdateH5PAssessmentArgs) {
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

func (m *h5pAssessmentModel) AddStudy(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleID string) (string, error) {
	log.Debug(ctx, "AddClassAndLive: add assessment args", log.String("schedule_id", scheduleID), log.Any("operator", operator))

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
			Strings: []string{scheduleID},
			Valid:   true,
		},
	}, &assessments); err != nil {
		log.Error(ctx, "AddClassAndLive: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.String("schedule_id", scheduleID),
			log.Any("operator", operator),
		)
		return "", err
	}
	if len(assessments) > 0 {
		log.Info(ctx, "AddClassAndLive: assessment already exists",
			log.String("schedule_id", scheduleID),
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
	if schedule, err = GetScheduleModel().GetPlainByID(ctx, scheduleID); err != nil {
		log.Error(ctx, "AddClassAndLive: GetScheduleModel().GetPlainByID: get failed",
			log.Err(err),
			log.String("schedule_id", scheduleID),
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
			log.String("string", scheduleID),
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
			log.String("lesson_plan_id", schedule.LessonPlanID),
			log.Any("schedule", schedule),
			log.Any("operator", operator),
		)
	} else {
		contents = append(contents, lessonPlan)
		if materials, err = GetContentModel().GetContentSubContentsByID(ctx, dbo.MustGetDB(ctx), lessonPlan.ID, operator); err != nil {
			log.Warn(ctx, "AddClassAndLive: GetContentModel().GetContentSubContentsByID: get materials failed",
				log.Err(err),
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
			ID:         newAssessmentID,
			ScheduleID: scheduleID,
			Type:       entity.AssessmentTypeClassAndLiveH5P,
			Status:     entity.AssessmentStatusInProgress,
		}
	)

	if classNameMap, err = external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, []string{schedule.ClassID}); err != nil {
		log.Error(ctx, "Add: external.GetClassServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("class_ids", []string{schedule.ClassID}),
			log.Any("schedule_id", scheduleID),
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
			log.Any("schedule_id", scheduleID),
			log.Any("new_item", newAssessment),
		)
		return "", err
	}

	// add attendances
	var finalAttendanceIDs []string
	users, err := GetScheduleRelationModel().GetUsersByScheduleID(ctx, operator, scheduleID)
	if err != nil {
		return "", err
	}
	for _, u := range users {
		finalAttendanceIDs = append(finalAttendanceIDs, u.RelationID)
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
			log.Any("schedule_id", scheduleID),
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
			log.Any("schedule_id", scheduleID),
			log.Any("operator", operator),
		)
		return "", err
	}

	return newAssessmentID, nil
}

func (m *h5pAssessmentModel) DeleteStudy(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleID string) error {
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
			Strings: []string{scheduleID},
			Valid:   true,
		},
	}, &assessments); err != nil {
		log.Error(ctx, "DeleteStudy: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Any("schedule_id", scheduleID),
			log.Any("operator", operator),
		)
		return err
	}
	for _, a := range assessments {
		if err := da.GetAssessmentDA().SoftDelete(ctx, tx, a.ID); err != nil {
			log.Error(ctx, "DeleteStudy: da.GetAssessmentDA().SoftDelete",
				log.Err(err),
				log.String("schedule_id", scheduleID),
			)
			return err
		}
	}
	return nil
}
