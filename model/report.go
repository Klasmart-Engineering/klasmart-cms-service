package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IReportModel interface {
	GetReportLessPlanInfo(ctx context.Context, tx *dbo.DBContext, teacherID string, classID string, operator entity.Operator) ([]*entity.ReportLessonPlanInfo, error)
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

func (r *reportModel) GetReportLessPlanInfo(ctx context.Context, tx *dbo.DBContext, teacherID string, classID string, operator entity.Operator) ([]*entity.ReportLessonPlanInfo, error) {
	lessonPlanIDs, err := da.GetScheduleDA().GetLessonPlanIDsByTeacherAndClass(ctx, tx, teacherID, classID)
	if err != nil {
		logger.Error(ctx, "GetLessPlanInfo:get lessonPlanIDs error",
			log.Err(err),
			log.String("teacherID", teacherID),
			log.String("classID", classID),
			log.Any("operator", operator),
		)
		return nil, err
	}
	lessonPlanInfos, err := GetContentModel().GetContentNameByIDList(ctx, tx, lessonPlanIDs)
	if err != nil {
		logger.Error(ctx, "GetLessPlanInfo:get lessonPlan info error",
			log.Err(err),
			log.String("teacherID", teacherID),
			log.String("classID", classID),
			log.Strings("lessonPlanIDs", lessonPlanIDs),
			log.Any("operator", operator),
		)
	}
	result := make([]*entity.ReportLessonPlanInfo, len(lessonPlanInfos))
	for i, item := range lessonPlanInfos {
		result[i] = &entity.ReportLessonPlanInfo{
			ID:   item.ID,
			Name: item.Name,
		}
	}
	return result, nil
}

func (r *reportModel) ListStudentsReport(ctx context.Context, tx *dbo.DBContext, cmd entity.ListStudentsReportCommand) (*entity.StudentsReport, error) {
	// TODO: lessonPlanID -> scheduleIDs
	var scheduleIDs []string
	assessments, err := da.GetAssessmentDA().BatchGetAssessmentsByScheduleIDs(ctx, tx, scheduleIDs)
	if err != nil {
		log.Error(ctx, "list student report: batch get assessment failed by schedule ids",
			log.Err(err),
			log.Any("cmd", "cmd"),
		)
		return nil, err
	}
	var assessmentIDs []string
	for _, assessment := range assessments {
		assessmentIDs = append(assessmentIDs, assessment.ID)
	}
	assessmentOutcomes, err := da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "list student report: batch get failed by assessment ids",
			log.Err(err),
			log.Any("cmd", "cmd"),
		)
		return nil, err
	}
	var outcomeIDs []string
	for _, item := range assessmentOutcomes {
		outcomeIDs = append(outcomeIDs, item.OutcomeID)
	}
	outcomeCond := &entity.OutcomeCondition{IDs: outcomeIDs}
	_, outcomes, err := GetOutcomeModel().SearchLearningOutcome(ctx, tx, outcomeCond, cmd.Operator)
	if err != nil {
		log.Error(ctx, "list student report: search learning outcome failed",
			log.Err(err),
			log.Any("outcome_cond", outcomeCond),
			log.Any("cmd", "cmd"),
		)
		return nil, err
	}
	outcomeNameMap := map[string]string{}
	for _, outcome := range outcomes {
		outcomeNameMap[outcome.ID] = outcome.Name
	}
	attendanceIDs, err := da.GetAssessmentAttendanceDA().BatchGetAttendanceIDsByAssessmentIDs(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "list student report: batch get attendance ids failed by assessment ids",
			log.Err(err),
			log.Any("assessment_ids", assessmentIDs),
			log.Any("cmd", "cmd"),
		)
		return nil, err
	}
	_ = attendanceIDs
	panic("implement me")
}

func (r *reportModel) GetStudentDetailReport(ctx context.Context, tx *dbo.DBContext, cmd entity.GetStudentDetailReportCommand) (*entity.StudentDetailReport, error) {
	panic("implement me")
}
