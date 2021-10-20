package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

var (
	ErrAssessmentNotFoundSchedule = errors.New("assessment: not found schedule")

	studyAssessmentModelInstance     IStudyAssessmentModel
	studyAssessmentModelInstanceOnce = sync.Once{}
)

func GetStudyAssessmentModel() IStudyAssessmentModel {
	studyAssessmentModelInstanceOnce.Do(func() {
		studyAssessmentModelInstance = &studyAssessmentModel{}
	})
	return studyAssessmentModelInstance
}

type IStudyAssessmentModel interface {
	GetDetail(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, id string) (*entity.AssessmentDetail, error)
	List(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args *entity.ListStudyAssessmentsArgs) (*entity.ListStudyAssessmentsResult, error)
	BatchCheckAnyoneAttempted(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, roomIDs []string) (map[string]bool, error)
	PrepareAddArgs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args []*entity.AddAssessmentArgs) (*entity.BatchAddAssessmentSuperArgs, error)
	Add(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input *entity.BatchAddAssessmentSuperArgs) ([]string, error)
	Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args *entity.UpdateAssessmentArgs) error
	Delete(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) error
}

type studyAssessmentModel struct {
	assessmentBase
}

func (m *studyAssessmentModel) GetDetail(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, id string) (*entity.AssessmentDetail, error) {
	return m.assessmentBase.getDetail(ctx, operator, id)
}

func (m *studyAssessmentModel) List(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args *entity.ListStudyAssessmentsArgs) (*entity.ListStudyAssessmentsResult, error) {
	// check args
	if len(args.ClassTypes) == 0 {
		errMsg := "list h5p assessments: require assessment type"
		log.Error(ctx, errMsg, log.Any("args", args), log.Any("operator", operator))
		return nil, constant.ErrInvalidArgs
	}

	// check permission
	var (
		checker = NewAssessmentPermissionChecker(operator)
		err     error
	)
	if err = checker.SearchAllPermissions(ctx); err != nil {
		log.Error(ctx, "List: checker.SearchAllPermissions: search failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, err
	}
	if !checker.CheckStatus(args.Status.Value) {
		log.Error(ctx, "list h5p assessments: check status failed",
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, constant.ErrForbidden
	}

	// get assessment list
	var (
		cond = da.QueryAssessmentConditions{
			ClassTypes: entity.NullScheduleClassTypes{
				Value: args.ClassTypes,
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
		case entity.ListStudyAssessmentsQueryTypeTeacherName:
			cond.TeacherIDs.Valid = true
			teachers, err := external.GetTeacherServiceProvider().Query(ctx, operator, operator.OrgID, args.Query)
			if err != nil {
				return nil, err
			}
			for _, item := range teachers {
				cond.TeacherIDs.Strings = append(cond.TeacherIDs.Strings, item.ID)
			}
		}
	}

	total, assessments, err := da.GetAssessmentDA().Page(ctx, &cond)
	if err != nil {
		log.Error(ctx, "List: da.GetAssessmentDA().QueryTx: query failed",
			log.Err(err),
			log.Any("cond", cond),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// convert to assessment view
	var views []*entity.AssessmentView
	if views, err = m.toViews(ctx, operator, assessments, entity.ConvertToViewsOptions{
		CheckedStudents:  sql.NullBool{Bool: true, Valid: true},
		EnableTeachers:   true,
		EnableClass:      true,
		EnableLessonPlan: true,
	}); err != nil {
		return nil, err
	}

	// get scores
	var roomIDs []string
	for _, v := range views {
		roomIDs = append(roomIDs, v.RoomID)
	}
	roomMap, err := getAssessmentH5P().batchGetRoomMap(ctx, operator, roomIDs)
	if err != nil {
		log.Error(ctx, "list h5p assessments: get room user scores map failed",
			log.Err(err),
			log.Strings("room_ids", roomIDs),
		)
		return nil, err
	}

	// construct result
	var result = entity.ListStudyAssessmentsResult{Total: total}
	for _, v := range views {
		teacherNames := make([]string, 0, len(v.Teachers))
		for _, t := range v.Teachers {
			teacherNames = append(teacherNames, t.Name)
		}
		var remainingTime int64
		if v.Schedule.DueAt != 0 {
			remainingTime = v.Schedule.DueAt - time.Now().Unix()
		} else {
			remainingTime = time.Unix(v.CreateAt, 0).Add(config.Get().Assessment.DefaultRemainingTime).Unix() - time.Now().Unix()
		}
		if remainingTime < 0 {
			remainingTime = 0
		}

		newItem := entity.ListStudyAssessmentsResultItem{
			ID:            v.ID,
			Title:         v.Title,
			TeacherNames:  teacherNames,
			ClassName:     v.Class.Name,
			DueAt:         v.Schedule.DueAt,
			CompleteRate:  getAssessmentH5P().calcRoomCompleteRate(ctx, roomMap[v.RoomID], v),
			RemainingTime: remainingTime,
			CompleteAt:    v.CompleteTime,
			ScheduleID:    v.ScheduleID,
			CreateAt:      v.CreateAt,
			LessonPlan:    v.LessonPlan,
		}
		result.Items = append(result.Items, &newItem)
	}

	return &result, nil
}

func (m *studyAssessmentModel) BatchCheckAnyoneAttempted(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, roomIDs []string) (map[string]bool, error) {
	if len(roomIDs) == 0 {
		return map[string]bool{}, nil
	}
	roomMap, err := getAssessmentH5P().batchGetRoomMap(ctx, operator, roomIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(roomIDs))
	for _, id := range roomIDs {
		if r := roomMap[id]; r != nil {
			result[id] = getAssessmentH5P().isAnyoneAttempted(r)
		} else {
			result[id] = false
		}
	}
	return result, nil
}

func (m *studyAssessmentModel) PrepareAddArgs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args []*entity.AddAssessmentArgs) (*entity.BatchAddAssessmentSuperArgs, error) {
	classIDs := make([]string, 0, len(args))
	for _, item := range args {
		classIDs = append(classIDs, item.ClassID)
	}
	classNameMap, err := external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "add studies: batch get class name map failed",
			log.Err(err),
			log.Strings("class_ids", classIDs),
		)
		return nil, err
	}
	for _, item := range args {
		className := classNameMap[item.ClassID]
		if className == "" {
			className = constant.AssessmentNoClass
		}
		item.Title = m.generateTitle(className, item.ScheduleTitle)
	}

	superArgs, err := m.assessmentBase.prepareBatchAddSuperArgs(ctx, tx, operator, args)
	if err != nil {
		log.Error(ctx, "prepare add assessment args: prepare batch add super args failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}

	return superArgs, nil
}

func (m *studyAssessmentModel) Add(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.BatchAddAssessmentSuperArgs) ([]string, error) {
	log.Debug(ctx, "add studies args", log.Any("args", args), log.Any("operator", operator))

	return m.assessmentBase.batchAdd(ctx, tx, operator, args)
}

func (m *studyAssessmentModel) Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args *entity.UpdateAssessmentArgs) error {
	return m.assessmentBase.update(ctx, tx, operator, args)
}

func (m *studyAssessmentModel) Delete(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) error {
	if len(scheduleIDs) == 0 {
		return nil
	}

	assessments, err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
		ClassTypes: entity.NullScheduleClassTypes{
			Value: []entity.ScheduleClassType{entity.ScheduleClassTypeHomework},
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
	})
	if err != nil {
		return err
	}

	assessmentIDs := make([]string, 0, len(assessments))
	for _, a := range assessments {
		assessmentIDs = append(assessmentIDs, a.ID)
	}
	if err := da.GetAssessmentDA().BatchSoftDelete(ctx, tx, assessmentIDs); err != nil {
		log.Error(ctx, "DeleteStudy: da.GetAssessmentDA().BatchSoftDelete",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return err
	}
	return nil
}

// region utils

func (m *studyAssessmentModel) generateTitle(className, lessonName string) string {
	return fmt.Sprintf("%s-%s", className, lessonName)
}

// endregion
