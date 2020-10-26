package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IReportModel interface {
	ListStudentReport(ctx context.Context, tx *dbo.DBContext, lessonPlanID string, operator entity.Operator) (entity.StudentReportList, error)
	GetStudentReportDetail(ctx context.Context, tx *dbo.DBContext, studentID string, lessonPlanID string, operator entity.Operator) (entity.StudentReportDetail, error)
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

func (r *reportModel) ListStudentReport(ctx context.Context, tx *dbo.DBContext, lessonPlanID string, operator entity.Operator) (entity.StudentReportList, error) {
	panic("implement me")
}

func (r *reportModel) GetStudentReportDetail(ctx context.Context, tx *dbo.DBContext, studentID string, lessonPlanID string, operator entity.Operator) (entity.StudentReportDetail, error) {
	panic("implement me")
}
