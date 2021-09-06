package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (m *reportModel) AddStudentUsageRecordTx(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, record *entity.StudentUsageRecord) (err error) {

	//
	panic("not implement")
}
