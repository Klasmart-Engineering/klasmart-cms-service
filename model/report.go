package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"sync"
)

type IReportModel interface {
	ListStudentsReport(ctx context.Context, tx *dbo.DBContext, cmd entity.ListStudentsReportCommand) (*entity.StudentsReport, error)
	GetStudentDetailReport(ctx context.Context, tx *dbo.DBContext, cmd entity.GetStudentDetailReportCommand) (*entity.StudentDetailReport, error)
}

var (
	reportModelInstance IReportModel
	reportModelOnce     = sync.Once{}
)

func GetReportModel() IReportModel {
	reportModelOnce.Do(func() {
		reportModelInstance = &reportModel{}
	})
	return reportModelInstance
}

type reportModel struct{}

func (r *reportModel) ListStudentsReport(ctx context.Context, tx *dbo.DBContext, cmd entity.ListStudentsReportCommand) (*entity.StudentsReport, error) {
	var assessmentIDs []string
	{
		scheduleIDs, err := r.getScheduleIDs(ctx, tx, cmd.TeacherID, cmd.ClassID, cmd.LessonPlanID)
		if err != nil {
			log.Error(ctx, "list students report: get schedule ids failed",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		assessments, err := da.GetAssessmentDA().BatchGetAssessmentsByScheduleIDs(ctx, tx, scheduleIDs)
		if err != nil {
			log.Error(ctx, "list students report: batch get assessment failed by schedule ids",
				log.Err(err),
				log.Any("cmd", "cmd"),
			)
			return nil, err
		}
		for _, assessment := range assessments {
			assessmentIDs = append(assessmentIDs, assessment.ID)
		}
	}
	var students []*external.Student
	{
		var err error
		students, err = external.GetClassServiceProvider().GetStudents(ctx, cmd.ClassID)
		if err != nil {
			log.Error(ctx, "list students report: get students",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
	}
	var (
		attendanceIDs         []string
		attendanceIDExistsMap = map[string]bool{}
	)
	{
		var err error
		assessmentAttendances, err := da.GetAssessmentAttendanceDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs)
		if err != nil {
			log.Error(ctx, "list students report: batch get attendance ids failed by assessment ids",
				log.Err(err),
				log.Any("assessment_ids", assessmentIDs),
				log.Any("cmd", "cmd"),
			)
			return nil, err
		}
		for _, item := range assessmentAttendances {
			attendanceIDs = append(attendanceIDs, item.AttendanceID)
		}
		attendanceIDs = r.uniqueStrings(attendanceIDs)
		for _, item := range attendanceIDs {
			attendanceIDExistsMap[item] = true
		}
	}
	var achievedAssessmentOutcomeMap = map[string][]entity.AssessmentOutcome{}
	{
		outcomeAttendances, err := da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs)
		if err != nil {
			log.Error(ctx, "list students report: batch get failed",
				log.Err(err),
				log.Any("assessment_ids", assessmentIDs),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		var (
			attendanceAssessmentOutcomeKeyMap = map[string]*entity.AssessmentOutcomeKey{}
			assessmentOutcomeKeys             []entity.AssessmentOutcomeKey
		)
		for _, item := range outcomeAttendances {
			key := entity.AssessmentOutcomeKey{
				AssessmentID: item.AssessmentID,
				OutcomeID:    item.OutcomeID,
			}
			attendanceAssessmentOutcomeKeyMap[item.AttendanceID] = &key
			assessmentOutcomeKeys = append(assessmentOutcomeKeys, key)
		}

		assessmentOutcomeMap, err := da.GetAssessmentOutcomeDA().BatchGetMapByKeys(ctx, tx, assessmentOutcomeKeys)
		if err != nil {
			log.Error(ctx, "list student report: batch get failed by assessment ids",
				log.Err(err),
				log.Any("cmd", "cmd"),
			)
			return nil, err
		}

		for attendanceID, assessmentOutcomeKey := range attendanceAssessmentOutcomeKeyMap {
			if assessmentOutcomeKey == nil {
				continue
			}
			assessmentOutcome := assessmentOutcomeMap[*assessmentOutcomeKey]
			if assessmentOutcome == nil {
				continue
			}
			achievedAssessmentOutcomeMap[attendanceID] = append(achievedAssessmentOutcomeMap[attendanceID], *assessmentOutcome)
		}
	}

	//excludeAchievedAssessmentOutcomeMap := map[string][]entity.AssessmentOutcome{}
	//{
	//	outcomeAttendances, err := da.().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs)
	//	if err != nil {
	//		log.Error(ctx, "list students report: batch get failed",
	//			log.Err(err),
	//			log.Any("assessment_ids", assessmentIDs),
	//			log.Any("cmd", cmd),
	//		)
	//		return nil, err
	//	}
	//	var (
	//		attendanceAssessmentOutcomeKeyMap = map[string]*entity.AssessmentOutcomeKey{}
	//		assessmentOutcomeKeys             []entity.AssessmentOutcomeKey
	//	)
	//	for _, item := range outcomeAttendances {
	//		key := entity.AssessmentOutcomeKey{
	//			AssessmentID: item.AssessmentID,
	//			OutcomeID:    item.OutcomeID,
	//		}
	//		attendanceAssessmentOutcomeKeyMap[item.AttendanceID] = &key
	//		assessmentOutcomeKeys = append(assessmentOutcomeKeys, key)
	//	}
	//
	//	assessmentOutcomeMap, err := da.GetAssessmentOutcomeDA().BatchGetMapByKeys(ctx, tx, assessmentOutcomeKeys)
	//	if err != nil {
	//		log.Error(ctx, "list student report: batch get failed by assessment ids",
	//			log.Err(err),
	//			log.Any("cmd", "cmd"),
	//		)
	//		return nil, err
	//	}
	//
	//	for attendanceID, assessmentOutcomeKey := range attendanceAssessmentOutcomeKeyMap {
	//		if assessmentOutcomeKey == nil {
	//			continue
	//		}
	//		assessmentOutcome := assessmentOutcomeMap[*assessmentOutcomeKey]
	//		if assessmentOutcome == nil {
	//			continue
	//		}
	//		achievedAssessmentOutcomeMap[attendanceID] = append(achievedAssessmentOutcomeMap[attendanceID], *assessmentOutcome)
	//	}
	//}

	var result entity.StudentsReport
	for _, student := range students {
		newItem := entity.StudentReportItem{StudentName: student.Name}

		if !attendanceIDExistsMap[student.ID] {
			result.Items = append(result.Items, newItem)
			continue
		}

		newItem.AllAchievedCount = len(achievedAssessmentOutcomeMap[student.ID])

		result.Items = append(result.Items, newItem)
	}

	panic("not implemented")
}

func (r *reportModel) GetStudentDetailReport(ctx context.Context, tx *dbo.DBContext, cmd entity.GetStudentDetailReportCommand) (*entity.StudentDetailReport, error) {
	//outcomes, err := GetOutcomeModel().GetLearningOutcomesByIDs(ctx, tx, outcomeIDs, cmd.Operator)
	//if err != nil {
	//	log.Error(ctx, "list students report: get learning outcomes failed by ids",
	//		log.Err(err),
	//		log.Strings("outcome_ids", outcomeIDs),
	//		log.Any("cmd", "cmd"),
	//	)
	//	return nil, err
	//}
	//for _, item := range outcomes {
	//	outcomeMap[item.ID] = item
	//}
	panic("implement me")
}

func (r *reportModel) getScheduleIDs(ctx context.Context, tx *dbo.DBContext, classID string, teacherID string, lessonPlanID string) ([]string, error) {
	items, err := GetScheduleModel().Query(ctx, &da.ScheduleCondition{
		TeacherID: sql.NullString{
			String: teacherID,
			Valid:  true,
		},
		LessonPlanID: sql.NullString{
			String: lessonPlanID,
			Valid:  true,
		},
		ClassID: sql.NullString{
			String: classID,
			Valid:  true,
		},
		Status: sql.NullString{
			String: string(entity.ScheduleStatusClosed),
			Valid:  true,
		},
	})
	if err != nil {
		return nil, err
	}
	var result []string
	for _, item := range items {
		result = append(result, item.ID)
	}
	return result, nil
}

func (r *reportModel) uniqueStrings(input []string) []string {
	m := map[string]bool{}
	for _, item := range input {
		m[item] = true
	}
	var result []string
	for k := range m {
		result = append(result, k)
	}
	return result
}
