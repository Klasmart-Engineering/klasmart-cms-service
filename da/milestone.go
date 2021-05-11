package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IMilestoneDA interface {
	Create(ctx context.Context, tx *dbo.DBContext, milestone *entity.Milestone) error
	Update(ctx context.Context, tx *dbo.DBContext, milestone *entity.Milestone) error
	Search(ctx context.Context, tx *dbo.DBContext, condition *MilestoneCondition) (int, []*entity.Milestone, error)
	GetByID(ctx context.Context, tx *dbo.DBContext, ID string) (*entity.Milestone, error)

	BatchPublish(ctx context.Context, tx *dbo.DBContext, publishIDs []string) error
	BatchHide(ctx context.Context, tx *dbo.DBContext, hideIDs []string) error
	BatchUnLock(ctx context.Context, tx *dbo.DBContext, unLockIDs []string) error
	BatchUpdateLatest(ctx context.Context, tx *dbo.DBContext, ancestorLatest map[string]string) error
	BatchDelete(ctx context.Context, tx *dbo.DBContext, milestoneIDs []string) error

	UnbindOutcomes(ctx context.Context, tx *dbo.DBContext, outcomeAncestors []string) error

	FindGap(ctx context.Context, num int) (int, error)
}

var milestoneDA *MilestoneSQLDA

var _milestoneOnce sync.Once

func GetMilestoneDA() IMilestoneDA {
	_milestoneOnce.Do(func() {
		milestoneDA = new(MilestoneSQLDA)
	})
	return milestoneDA
}

var milestoneOutcomeDA *MilestoneOutcomeSQLDA
var _milestoneOutcomeOnce sync.Once

func GetMilestoneOutcomeDA() *MilestoneOutcomeSQLDA {
	_milestoneOutcomeOnce.Do(func() {
		milestoneOutcomeDA = new(MilestoneOutcomeSQLDA)
	})
	return milestoneOutcomeDA
}
