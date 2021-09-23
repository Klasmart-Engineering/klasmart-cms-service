package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IReportDA interface {
	DataAccessor
	ITeacherLoadAssessment
	ITeacherLoadLesson
}
type ReportDA struct {
	BaseDA
}

func (r *ReportDA) GetTeacherLoadReportOfAssignment(ctx context.Context, op *entity.Operator, req *entity.TeacherLoadAssignmentRequest) (res []*entity.TeacherLoadAssignmentResponse, err error) {
	panic("implement me")
}

var _reportDA *ReportDA
var _reportDAOnce sync.Once

func GetReportDA() IReportDA {
	_reportDAOnce.Do(func() {
		_reportDA = new(ReportDA)
	})
	return _reportDA
}

type ITeacherLoadAssessment interface {
	GetTeacherLoadReportOfAssignment(ctx context.Context, op *entity.Operator, req *entity.TeacherLoadAssignmentRequest) (res []*entity.TeacherLoadAssignmentResponse, err error)
}
type ITeacherLoadLesson interface {
}
