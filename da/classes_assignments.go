package da

import (
	"context"
	"sync"

	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type IClassesAssignmentsDA interface {
	BatchInsertTx(ctx context.Context, tx *dbo.DBContext, records []*entity.ClassesAssignmentsRecords) error
	BatchUpdateFinishAndEnd(ctx context.Context, tx *dbo.DBContext, scheduleID string, attendedIDs []string, endTime int64) error
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
