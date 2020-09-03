package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IOutcomeModel interface {
	CreateLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) (*entity.Outcome, error)
	UpdateLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) error
	DeleteLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) error
	SearchLearningOutcome(ctx context.Context, tx *dbo.DBContext, condition *da.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error)

	LockLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) (string, error)

	PublishLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, scope string, operator *entity.Operator) error
	BulkPubLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, scope string, operator *entity.Operator) error
	BulkDelLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) error

	SearchPrivateOutcomes(ctx context.Context, tx *dbo.DBContext, condition *da.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error)
	SearchPendingOutcomes(ctx context.Context, tx *dbo.DBContext, condition *da.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error)

	GetLearningOutcomesByIDs(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) ([]*entity.Outcome, error)
	GetLatestOutcomesByIDs(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) ([]*entity.Outcome, error)
}

type OutcomeModel struct {
}

func (o OutcomeModel) CreateLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) (*entity.Outcome, error) {
	panic("implement me")
}

func (o OutcomeModel) UpdateLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) error {
	panic("implement me")
}

func (o OutcomeModel) DeleteLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) error {
	panic("implement me")
}

func (o OutcomeModel) SearchLearningOutcome(ctx context.Context, tx *dbo.DBContext, condition *da.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error) {
	panic("implement me")
}

func (o OutcomeModel) LockLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) (string, error) {
	panic("implement me")
}

func (o OutcomeModel) PublishLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, scope string, operator *entity.Operator) error {
	panic("implement me")
}

func (o OutcomeModel) BulkPubLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, scope string, operator *entity.Operator) error {
	panic("implement me")
}

func (o OutcomeModel) BulkDelLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) error {
	panic("implement me")
}

func (o OutcomeModel) SearchPrivateOutcomes(ctx context.Context, tx *dbo.DBContext, condition *da.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error) {
	panic("implement me")
}

func (o OutcomeModel) SearchPendingOutcomes(ctx context.Context, tx *dbo.DBContext, condition *da.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error) {
	panic("implement me")
}

func (o OutcomeModel) GetLearningOutcomesByIDs(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) ([]*entity.Outcome, error) {
	panic("implement me")
}

func (o OutcomeModel) GetLatestOutcomesByIDs(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) ([]*entity.Outcome, error) {
	panic("implement me")
}
