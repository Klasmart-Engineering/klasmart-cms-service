package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (m *reportModel) GetTeacherLoadReportOfAssignment(ctx context.Context, op *entity.Operator, req *entity.TeacherLoadAssignmentRequest) (res []*entity.TeacherLoadAssignmentResponse, err error) {
	panic("implement me")
}
