package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type OutcomeRelationSQLDA struct {
	dbo.BaseDA
}

func (OutcomeRelationSQLDA) TableName() string {
	return entity.OutcomeRelationTable
}

func (OutcomeRelationSQLDA) MasterType() entity.RelationType {
	return entity.OutcomeType
}

func (mas OutcomeRelationSQLDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, masterIDs []string) error {
	return GetRelationDA().DeleteTx(ctx, tx, mas.TableName(), mas.MasterType(), masterIDs)
}

var outcomeRelationDA *OutcomeRelationSQLDA
var _outcomeRelationDAOnce sync.Once

func GetOutcomeRelationDA() *OutcomeRelationSQLDA {
	_outcomeRelationDAOnce.Do(func() {
		outcomeRelationDA = new(OutcomeRelationSQLDA)
	})
	return outcomeRelationDA
}
