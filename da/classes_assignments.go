package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IClassesAssignmentsDA interface {
	BatchInsertTx(ctx context.Context, tx *dbo.DBContext, record []*entity.ClassesAssignmentsRecords) error
	QueryTx(ctx context.Context, tx *dbo.DBContext, condition *ClassesAssignmentsCondition) ([]*entity.ClassesAssignmentsRecords, error)
	CountAttendedTx(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string, attendanceType string) (map[string]int, error)
}

var classesAssignmentsDA *ClassesAssignmentsSQLDA

var _classesAssignmentsOnce sync.Once

func GetClassesAssignmentsDA() IClassesAssignmentsDA {
	_classesAssignmentsOnce.Do(func() {
		classesAssignmentsDA = new(ClassesAssignmentsSQLDA)
	})
	return classesAssignmentsDA
}
