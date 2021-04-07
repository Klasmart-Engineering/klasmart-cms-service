package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IOutcomeSetDA interface {
	CreateSet(ctx context.Context, tx *dbo.DBContext, outcomeSet *entity.Set) error
	UpdateOutcomeSet(ctx context.Context, tx *dbo.DBContext, outcomeSet *entity.Set) error
	SearchSet(ctx context.Context, tx *dbo.DBContext, condition *SetCondition) (int, []*entity.Set, error)
	SearchOutcomeSet(ctx context.Context, tx *dbo.DBContext, condition *OutcomeSetCondition) (int, []*entity.OutcomeSet, error)
	BulkBindOutcomeSet(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcomeSets []*entity.OutcomeSet) error
	BindOutcomeSet(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcomeSets []*entity.OutcomeSet) error
	DeleteBoundOutcomeSet(ctx context.Context, tx *dbo.DBContext, outcomeID string) error
	SearchOutcomeBySetName(ctx context.Context, op *entity.Operator, name string) ([]*entity.OutcomeSet, error)
	SearchSetsByOutcome(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string) (map[string][]*entity.Set, error)
	IsSetExist(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, name string) (bool, error)
}

var outcomeSetDA *SetSqlDA

var _outcomeSetOnce sync.Once

func GetOutcomeSetDA() IOutcomeSetDA {
	_outcomeSetOnce.Do(func() {
		outcomeSetDA = new(SetSqlDA)
	})
	return outcomeSetDA
}
