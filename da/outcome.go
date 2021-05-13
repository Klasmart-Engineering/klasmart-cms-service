package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IOutcomeDA interface {
	CreateOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) error
	UpdateOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) error
	DeleteOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) error

	GetOutcomeByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Outcome, error)
	GetOutcomeBySourceID(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, sourceID string) (*entity.Outcome, error)
	SearchOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, condition *OutcomeCondition) (int, []*entity.Outcome, error)

	UpdateLatestHead(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, oldHeader, newHeader string) error
}

var _outcomeOnce sync.Once

var outcomeDA *OutcomeSQLDA

func GetOutcomeDA() IOutcomeDA {
	_outcomeOnce.Do(func() {
		outcomeDA = new(OutcomeSQLDA)
	})
	return outcomeDA
}
