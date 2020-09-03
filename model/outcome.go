package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IOutcomeModel interface {
	CreateLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) error
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

func (o OutcomeModel) CreateLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) error {
	outcome.ID = utils.NewID()
	outcome.AncestorID = outcome.ID
	outcome.AuthorID = operator.UserID
	outcome.AuthorName = "GetName" // TODO
	err := da.GetOutcomeDA().CreateOutcome(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "CreateLearningOutcome failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.Any("outcome", outcome))
	}
	return err
}

func (o OutcomeModel) UpdateLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) error {
	data, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcome.ID)
	if err == dbo.ErrRecordNotFound {
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "UpdateLearningOutcome: GetOutcomeByID failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.Any("data", outcome))
		return err
	}
	data.Update(outcome)
	err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, data)
	if err != nil {
		log.Error(ctx, "UpdateLearningOutcome: UpdateOutcome failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.Any("data", outcome))
		return err
	}
	return nil
}

func (o OutcomeModel) DeleteLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) error {
	err := da.GetOutcomeDA().DeleteOutcome(ctx, tx, outcomeID)
	if err != nil {
		log.Error(ctx, "CreateLearningOutcome failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.String("outcome_id", outcomeID))
		return err
	}
	return nil
}

func (o OutcomeModel) SearchLearningOutcome(ctx context.Context, tx *dbo.DBContext, condition *da.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error) {
	panic("implement me")
}

func (o OutcomeModel) LockLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) (string, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixContentLock)
	if err != nil {
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()
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
