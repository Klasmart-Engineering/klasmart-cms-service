package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IOutcomeDA interface {
	CreateOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome) error
	UpdateOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome) error
	DeleteOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome) error

	GetOutcomeByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Outcome, error)
	SearchOutcome(ctx context.Context, tx *dbo.DBContext, condition *OutcomeCondition) (int, []*entity.Outcome, error)

	UpdateLatestHead(ctx context.Context, tx *dbo.DBContext, oldHeader, newHeader string) error
}

var _outcomeOnce sync.Once

var outcomeDA *OutcomeSqlDA

func GetOutcomeDA() IOutcomeDA {
	_outcomeOnce.Do(func() {
		outcomeDA = new(OutcomeSqlDA)
	})
	return outcomeDA
}
