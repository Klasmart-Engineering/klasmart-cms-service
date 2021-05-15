package model

import (
	"context"
	"database/sql"
	"errors"
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

type IContentAssessmentModel interface {
	List(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.ListH5PAssessmentsArgs) (*entity.ListH5PAssessmentsResult, error)
	GetDetail(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, id string) (*entity.GetH5PAssessmentDetailResult, error)
	Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.UpdateH5PAssessmentArgs)
	AddClassAndLive(ctx context.Context, operator *entity.Operator, args entity.AddAssessmentArgs) (string, error)
	AddStudy(ctx context.Context, operator *entity.Operator, args entity.AddAssessmentArgs) error
}

var (
	contentAssessmentModelInstance     IContentAssessmentModel
	contentAssessmentModelInstanceOnce = sync.Once{}
)

func GetContentAssessmentModel() IContentAssessmentModel {
	contentAssessmentModelInstanceOnce.Do(func() {
		contentAssessmentModelInstance = &contentAssessmentModel{}
	})
	return contentAssessmentModelInstance
}

type contentAssessmentModel struct{}

func (m *contentAssessmentModel) List(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.ListH5PAssessmentsArgs) (*entity.ListH5PAssessmentsResult, error) {
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

func (m *contentAssessmentModel) GetDetail(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, id string) (*entity.GetH5PAssessmentDetailResult, error) {
	panic("implement me")
}

func (m *contentAssessmentModel) Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.UpdateH5PAssessmentArgs) {
	panic("implement me")
}

func (m *contentAssessmentModel) AddClassAndLive(ctx context.Context, operator *entity.Operator, args entity.AddAssessmentArgs) (string, error) {
	panic("implement me")
}

func (m *contentAssessmentModel) AddStudy(ctx context.Context, operator *entity.Operator, args entity.AddAssessmentArgs) error {
	panic("implement me")
}
