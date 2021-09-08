package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ClassesAssignmentsSQLDA struct {
	dbo.BaseDA
}

func (ClassesAssignmentsSQLDA) BatchInsertTx(ctx context.Context, tx *dbo.DBContext) {
	panic("implement me")
}

func (ClassesAssignmentsSQLDA) QueryTx(ctx context.Context, tx *dbo.DBContext) ([]*entity.ClassesAssignmentsRecords, error) {
	panic("implement me")
}

func (ClassesAssignmentsSQLDA) CountAttendedTx(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string, attendanceType string) (map[string]int, error) {
	panic("implement me")
}
