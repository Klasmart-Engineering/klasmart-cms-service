package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
