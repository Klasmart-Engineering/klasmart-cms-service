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
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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
	Add(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input []*entity.AddStudyInput) ([]string, error)
	Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args *entity.UpdateAssessmentArgs) error
	Delete(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) error
}

type studyAssessmentModel struct {
	assessmentBase
}

func (m *studyAssessmentModel) GetDetail(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, id string) (*entity.AssessmentDetail, error) {
	return m.assessmentBase.getDetail(ctx, tx, operator, id)
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
		assessments []*entity.Assessment
		cond        = da.QueryAssessmentConditions{
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
	log.Debug(ctx, "List: print query cond", log.Any("cond", cond))
	total, err := da.GetAssessmentDA().PageTx(ctx, tx, &cond, &assessments)
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
	if views, err = m.toViews(ctx, tx, operator, assessments, entity.ConvertToViewsOptions{
		CheckedStudents:  sql.NullBool{Bool: true, Valid: true},
		EnableProgram:    true,
		EnableSubjects:   true,
		EnableTeachers:   true,
		EnableStudents:   true,
		EnableClass:      true,
		EnableLessonPlan: true,
	}); err != nil {
		log.Error(ctx, "List: GetAssessmentUtils().toViews: get failed",
			log.Err(err),
			log.Any("assessments", assessments),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// get scores
	var roomIDs []string
	for _, v := range views {
		roomIDs = append(roomIDs, v.RoomID)
	}
	roomMap, err := m.batchGetRoomScoreMap(ctx, operator, roomIDs, false)
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
			CompleteRate:  m.getRoomCompleteRate(roomMap[v.RoomID], v),
			RemainingTime: remainingTime,
			CompleteAt:    v.CompleteTime,
			ScheduleID:    v.ScheduleID,
			CreateAt:      v.CreateAt,
		}
		result.Items = append(result.Items, &newItem)
	}

	return &result, nil
}

func (m *studyAssessmentModel) BatchCheckAnyoneAttempted(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, roomIDs []string) (map[string]bool, error) {
	if len(roomIDs) == 0 {
		return map[string]bool{}, nil
	}
	roomMap, err := m.batchGetRoomScoreMap(ctx, operator, roomIDs, false)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(roomIDs))
	for _, id := range roomIDs {
		if v := roomMap[id]; v != nil {
			result[id] = v.AnyoneAttempted
		} else {
			result[id] = false
		}
	}
	return result, nil
}

func (m *studyAssessmentModel) Add(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input []*entity.AddStudyInput) ([]string, error) {
	log.Debug(ctx, "add studies args", log.Any("input", input), log.Any("operator", operator))

	// check if assessment already exits
	scheduleIDs := make([]string, 0, len(input))
	for _, item := range input {
		scheduleIDs = append(scheduleIDs, item.ScheduleID)
	}
	count, err := da.GetAssessmentDA().CountTx(ctx, tx, &da.QueryAssessmentConditions{
		ClassTypes: entity.NullScheduleClassTypes{
			Value: []entity.ScheduleClassType{entity.ScheduleClassTypeHomework},
			Valid: true,
		},
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}, entity.Assessment{})
	if err != nil {
		log.Error(ctx, "Add: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Strings("schedule_id", scheduleIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	if count > 0 {
		log.Info(ctx, "Add: assessment already exists",
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("operator", operator),
		)
		return nil, nil
	}

	// get class name map
	var classIDs []string
	for _, item := range input {
		classIDs = append(classIDs, item.ClassID)
	}
	classNameMap, err := external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "Add: external.GetClassServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("class_ids", classIDs),
			log.Any("schedule_ids", scheduleIDs),
		)
		return nil, err
	}

	// get contents
	var lessonPlanIDs []string
	for _, item := range input {
		lessonPlanIDs = append(lessonPlanIDs, item.LessonPlanID)
	}
	lessonPlanMap, err := m.batchGetLatestLessonPlanMap(ctx, tx, operator, lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "Add: GetAssessmentUtils().batchGetLatestLessonPlanMap: get failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
		)
		return nil, err
	}

	// add assessment
	newAssessments := make([]*entity.Assessment, 0, len(scheduleIDs))
	now := time.Now().Unix()
	for _, item := range input {
		className := classNameMap[item.ClassID]
		if className == "" {
			className = constant.AssessmentNoClass
		}
		newAssessments = append(newAssessments, &entity.Assessment{
			ID:         utils.NewID(),
			ScheduleID: item.ScheduleID,
			Title:      m.generateTitle(className, item.ScheduleTitle),
			Status:     entity.AssessmentStatusInProgress,
			CreateAt:   now,
			UpdateAt:   now,
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
	scheduleIDToAssessmentIDMap := make(map[string]string, len(newAssessments))
	for _, item := range newAssessments {
		scheduleIDToAssessmentIDMap[item.ScheduleID] = item.ID
	}
	var attendances []*entity.AssessmentAttendance
	for _, item := range input {
		for _, attendance := range item.Attendances {
			newAttendance := entity.AssessmentAttendance{
				ID:           utils.NewID(),
				AssessmentID: scheduleIDToAssessmentIDMap[item.ScheduleID],
				AttendanceID: attendance.RelationID,
				Checked:      true,
			}
			switch attendance.RelationType {
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
			attendances = append(attendances, &newAttendance)
		}
	}
	if err := da.GetAssessmentAttendanceDA().BatchInsert(ctx, tx, attendances); err != nil {
		log.Error(ctx, "add studies: batch insert attendance failed",
			log.Err(err),
			log.Any("attendances", attendances),
			log.Any("input", input),
		)
		return nil, err
	}

	// add contents
	var (
		scheduleMap        = make(map[string]*entity.AddStudyInput, len(input))
		assessmentContents []*entity.AssessmentContent
	)
	for _, item := range input {
		scheduleMap[item.ScheduleID] = item
	}
	assessmentContentKeys := map[[2]string]bool{}
	for _, a := range newAssessments {
		schedule := scheduleMap[a.ScheduleID]
		if schedule == nil {
			log.Error(ctx, "add study assessment: not found schedule by id", log.Any("input", input))
			return nil, ErrAssessmentNotFoundSchedule
		}
		lp := lessonPlanMap[schedule.LessonPlanID]
		assessmentContents = append(assessmentContents, &entity.AssessmentContent{
			ID:           utils.NewID(),
			AssessmentID: a.ID,
			ContentID:    lp.ID,
			ContentName:  lp.Name,
			ContentType:  entity.ContentTypePlan,
			Checked:      true,
		})
		for _, lm := range lp.Materials {
			key := [2]string{a.ID, lm.ID}
			if assessmentContentKeys[key] {
				continue
			}
			assessmentContentKeys[key] = true
			assessmentContents = append(assessmentContents, &entity.AssessmentContent{
				ID:           utils.NewID(),
				AssessmentID: a.ID,
				ContentID:    lm.ID,
				ContentName:  lm.Name,
				ContentType:  entity.ContentTypeMaterial,
				Checked:      true,
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
	for _, a := range newAssessments {
		newAssessmentIDs = append(newAssessmentIDs, a.ID)
	}

	return newAssessmentIDs, nil
}

func (m *studyAssessmentModel) Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args *entity.UpdateAssessmentArgs) error {
	return m.assessmentBase.update(ctx, tx, operator, args)
}

func (m *studyAssessmentModel) Delete(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) error {
	if len(scheduleIDs) == 0 {
		return nil
	}
	var assessments []entity.Assessment
	if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
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
	}, &assessments); err != nil {
		log.Error(ctx, "DeleteStudy: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("operator", operator),
		)
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
