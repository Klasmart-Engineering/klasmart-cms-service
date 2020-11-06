package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IReportDA interface {
	ListStudentsReport(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) (*entity.StudentsReport, error)
	GetStudentDetailReport(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) (*entity.StudentDetailReport, error)
}

func GetReportDA() IReportDA {
	return &reportDA{}
}

type reportDA struct{}

func (r *reportDA) ListStudentsReport(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) (*entity.StudentsReport, error) {
	panic("implement me")
}

func (r *reportDA) GetStudentDetailReport(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) (*entity.StudentDetailReport, error) {
	panic("implement me")
}
