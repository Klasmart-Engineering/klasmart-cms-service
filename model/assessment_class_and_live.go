package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
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
	Update(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.UpdateAssessmentArgs) error
}

type classAndLiveAssessmentModel struct {
	assessmentBase
}

func (m *classAndLiveAssessmentModel) GetDetail(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string) (*entity.AssessmentDetail, error) {
	return m.assessmentBase.getDetail(ctx, operator, id)
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
		log.Info(ctx, "list outcome assessments: check status failed",
			log.Any("args", args),
			log.Any("checker", checker),
		)
		return nil, constant.ErrForbidden
	}

	// get assessment list
	var (
		cond = da.QueryAssessmentConditions{
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
		teachers []*external.Teacher
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
	assessments, err := da.GetAssessmentDA().Query(ctx, &cond)
	if err != nil {
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
	if total, err = da.GetAssessmentDA().CountTx(ctx, tx, &cond); err != nil {
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
	if views, err = m.toViews(ctx, operator, assessments, entity.ConvertToViewsOptions{
		CheckedStudents:  sql.NullBool{Bool: true, Valid: true},
		EnableProgram:    true,
		EnableSubjects:   true,
		EnableTeachers:   true,
		EnableLessonPlan: true,
	}); err != nil {
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
			LessonPlan:   v.LessonPlan,
		}
		result.Items = append(result.Items, &newItem)
	}

	return &result, nil
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
