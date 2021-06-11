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
	List(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.QueryAssessmentsArgs) (*entity.ListAssessmentsResult, error)
	Add(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.AddClassAndLiveAssessmentArgs) (string, error)
	Update(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.UpdateAssessmentArgs) error
}

type classAndLiveAssessmentModel struct {
	assessmentBase
}

func (m *classAndLiveAssessmentModel) GetDetail(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string) (*entity.AssessmentDetail, error) {
	return m.assessmentBase.getDetail(ctx, tx, operator, id)
}

func (m *classAndLiveAssessmentModel) List(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.QueryAssessmentsArgs) (*entity.ListAssessmentsResult, error) {
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
	if args.ClassType.Valid {
		cond.ClassTypes = entity.NullScheduleClassTypes{
			Value: []entity.ScheduleClassType{args.ClassType.Value},
			Valid: true,
		}
	} else {
		cond.ClassTypes = entity.NullScheduleClassTypes{
			Value: []entity.ScheduleClassType{entity.ScheduleClassTypeOnlineClass, entity.ScheduleClassTypeOfflineClass},
			Valid: true,
		}
	}
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
	if views, err = m.toViews(ctx, tx, operator, assessments, entity.ConvertToViewsOptions{
		CheckedStudents: sql.NullBool{Bool: true, Valid: true},
		EnableProgram:   true,
		EnableSubjects:  true,
		EnableTeachers:  true,
	}); err != nil {
		log.Error(ctx, "List: GetAssessmentUtils().toViews: get failed",
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

func (m *classAndLiveAssessmentModel) Add(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.AddClassAndLiveAssessmentArgs) (string, error) {
	log.Debug(ctx, "add class and live assessment: print args", log.Any("args", args), log.Any("operator", operator))

	// clean data
	args.AttendanceIDs = utils.SliceDeduplicationExcludeEmpty(args.AttendanceIDs)

	// get schedule
	schedule, err := GetScheduleModel().GetPlainByID(ctx, args.ScheduleID)
	if err != nil {
		log.Error(ctx, "add class and live assessment: get schedule failed",
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

	// check class type
	if schedule.ClassType != entity.ScheduleClassTypeOnlineClass && schedule.ClassType != entity.ScheduleClassTypeOfflineClass {
		log.Info(ctx, "add class and live assessment: invalid schedule class type",
			log.String("class_type", string(schedule.ClassType)),
			log.Any("schedule", schedule),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", nil
	}

	// fix empty org id
	operator.OrgID = schedule.OrgID

	// generate assessment title
	classNameMap, err := external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, []string{schedule.ClassID})
	if err != nil {
		log.Error(ctx, "Add: external.GetClassServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("class_ids", []string{schedule.ClassID}),
			log.Any("args", args),
		)
		return "", err
	}
	assessmentTitle := m.generateTitle(args.ClassEndTime, classNameMap[schedule.ClassID], schedule.Title)

	// get attendances
	var finalAttendanceIDs []string
	switch schedule.ClassType {
	case entity.ScheduleClassTypeOfflineClass:
		users, err := GetScheduleRelationModel().GetUsersByScheduleID(ctx, operator, args.ScheduleID)
		if err != nil {
			log.Error(ctx, "add class and live assessments: get users by schedule id failed",
				log.Err(err),
				log.Any("args", args),
			)
			return "", err
		}
		for _, u := range users {
			finalAttendanceIDs = append(finalAttendanceIDs, u.RelationID)
		}
	default:
		finalAttendanceIDs = args.AttendanceIDs
	}
	scheduleRelationCond := &da.ScheduleRelationCondition{
		ScheduleID: sql.NullString{
			String: schedule.ID,
			Valid:  true,
		},
		RelationIDs: entity.NullStrings{
			Strings: args.AttendanceIDs,
			Valid:   true,
		},
	}
	scheduleRelations, err := GetScheduleRelationModel().Query(ctx, operator, scheduleRelationCond)
	if err != nil {
		log.Error(ctx, "add class and live assessments: query schedule relations failed",
			log.Err(err),
			log.Any("attendance_ids", finalAttendanceIDs),
			log.Any("operator", operator),
			log.Any("condition", scheduleRelationCond),
		)
		return "", err
	}
	if len(scheduleRelations) == 0 {
		log.Error(ctx, "add class and live assessments: not found schedule relations",
			log.Err(err),
			log.Any("attendance_ids", finalAttendanceIDs),
			log.Any("operator", operator),
			log.Any("condition", scheduleRelationCond),
		)
		return "", ErrNotFoundAttendance
	}
	ids, err := m.assessmentBase.batchAdd(ctx, tx, operator, []*entity.AddAssessmentArgs{{
		Title:         assessmentTitle,
		ScheduleID:    args.ScheduleID,
		ScheduleTitle: schedule.Title,
		LessonPlanID:  schedule.LessonPlanID,
		ClassID:       schedule.ClassID,
		ClassLength:   args.ClassLength,
		ClassEndTime:  args.ClassEndTime,
		Attendances:   scheduleRelations,
	}})
	if err != nil {
		return "", err
	}
	if len(ids) > 0 {
		return ids[0], nil
	}

	return "", nil
}

func (m *classAndLiveAssessmentModel) Update(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.UpdateAssessmentArgs) error {
	return m.assessmentBase.update(ctx, tx, operator, args)
}

// region utils

func (m *classAndLiveAssessmentModel) generateTitle(classEndTime int64, className string, scheduleTitle string) string {
	if className == "" {
		className = constant.AssessmentNoClass
	}
	return fmt.Sprintf("%s-%s-%s", time.Unix(classEndTime, 0).Format("20060102"), className, scheduleTitle)
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
