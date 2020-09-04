package model

import (
	"context"
	"database/sql"
	"github.com/jinzhu/gorm"
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

func (ocm OutcomeModel) CreateLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) (err error) {
	// outcome get value from api lay, this lay add some information
	outcome.ID = utils.NewID()
	outcome.AncestorID = outcome.ID
	outcome.AuthorName, err = ocm.getAuthorNameByID(ctx, outcome.AuthorID)
	if err != nil {
		log.Error(ctx, "CreateLearningOutcome: getAuthorNameByID failed",
			log.String("op", outcome.AuthorID),
			log.Any("outcome", outcome))
		return
	}
	outcome.OrganizationID, _, err = ocm.getRootOrganizationByAuthorID(ctx, operator.UserID)
	if err != nil {
		log.Error(ctx, "CreateLearningOutcome: getRootOrganizationByAuthorID failed",
			log.String("op", operator.UserID),
			log.Any("outcome", outcome))
		return
	}
	outcome.Shortcode, err = ocm.getShortCode(ctx)
	if err != nil {
		log.Error(ctx, "CreateLearningOutcome: getShortCode failed",
			log.String("op", operator.UserID),
			log.Any("outcome", outcome))
		return
	}
	err = da.GetOutcomeDA().CreateOutcome(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "CreateLearningOutcome: CreateOutcome failed",
			log.String("op", operator.UserID),
			log.Any("outcome", outcome))
	}
	return err
}

func (ocm OutcomeModel) UpdateLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) error {
	data, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcome.ID)
	if err == dbo.ErrRecordNotFound {
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "UpdateLearningOutcome: GetOutcomeByID failed",
			log.String("op", operator.UserID),
			log.Any("data", outcome))
		return err
	}
	data.Update(outcome)
	err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, data)
	if err != nil {
		log.Error(ctx, "UpdateLearningOutcome: UpdateOutcome failed",
			log.String("op", operator.UserID),
			log.Any("data", outcome))
		return err
	}
	return nil
}

func (ocm OutcomeModel) DeleteLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			log.Error(ctx, "DeleteLearningOutcome: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = ocm.deleteOutcome(ctx, tx, outcome)
		if err != nil {
			log.Error(ctx, "DeleteLearningOutcome: deleteOutcome failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
		}
		return err
	})

	return err
}

func (ocm OutcomeModel) SearchLearningOutcome(ctx context.Context, tx *dbo.DBContext, condition *da.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error) {
	condition.PublishStatus = dbo.NullStrings{Strings: []string{entity.ContentStatusPublished}, Valid: true}
	total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "SearchLearningOutcome: DeleteOutcome failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	return total, outcomes, nil
}

func (ocm OutcomeModel) LockLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) (string, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixContentLock)
	if err != nil {
		log.Error(ctx, "LockLearningOutcome: NewLock failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.String("outcome_id", outcomeID))
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()
	outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
	if err == dbo.ErrRecordNotFound {
		return "", ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "LockLearningOutcome: GetOutcomeByID failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.String("outcome_id", outcomeID))
		return "", err
	}

	err = ocm.lockOutcome(ctx, tx, outcome, operator)
	if err != nil {
		return "", err
	}
	newVersion := outcome.Clone()
	err = da.GetOutcomeDA().CreateOutcome(ctx, tx, &newVersion)
	if err != nil {
		log.Error(ctx, "LockLearningOutcome: CreateOutcome failed",
			log.String("op", operator.UserID),
			log.String("outcome_id", outcomeID),
			log.Any("outcome", newVersion))
		return "", err
	}
	return newVersion.ID, nil
}

func (ocm OutcomeModel) PublishLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, scope string, operator *entity.Operator) error {
	outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
	if err == dbo.ErrRecordNotFound {
		err = ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "PublishLearningOutcome: GetOutcomeByID failed",
			log.String("op", operator.UserID),
			log.String("outcome_id", outcomeID))
		return err
	}
	err = outcome.SetStatus(entity.ContentStatusPublished)
	if err != nil {
		log.Error(ctx, "PublishLearningOutcome: SetStatus failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.Any("outcome", outcome))
		return ErrInvalidContentStatusToPublish
	}
	err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "PublishLearningOutcome: UpdateOutcome failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.Any("outcome", outcome))
		return err
	}
	return nil
}

func (ocm OutcomeModel) BulkPubLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, scope string, operator *entity.Operator) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		condition := da.OutcomeCondition{
			IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		}
		total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, &condition)
		if err != nil {
			log.Error(ctx, "BulkPubLearningOutcome: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Int("total", total),
				log.Strings("outcome_id", outcomeIDs))
			return err
		}
		if total == 0 {
			log.Warn(ctx, "BulkPubLearningOutcome: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Int("total", total),
				log.Strings("outcome_id", outcomeIDs))
			return ErrResourceNotFound
		}
		for _, o := range outcomes {
			err = o.SetStatus(entity.ContentStatusPublished)
			if err != nil {
				log.Error(ctx, "BulkPubLearningOutcome: SetStatus failed",
					log.String("op", operator.UserID),
					log.Any("outcome", o))
				return ErrInvalidContentStatusToPublish
			}
			err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, o)
			if err != nil {
				log.Error(ctx, "BulkPubLearningOutcome: UpdateOutcome failed",
					log.String("op", operator.UserID),
					log.Any("outcome", o))
				return err
			}
		}
		return nil
	})
	return err
}

func (ocm OutcomeModel) BulkDelLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		condition := da.OutcomeCondition{
			IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		}
		total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, &condition)
		if err != nil {
			log.Error(ctx, "BulkDelLearningOutcome: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Int("total", total),
				log.Strings("outcome_ids", outcomeIDs))
			return err
		}

		for _, o := range outcomes {
			err = ocm.deleteOutcome(ctx, tx, o)
			if err != nil {
				log.Error(ctx, "BulkDelLearningOutcome: DeleteOutcome failed",
					log.String("op", operator.UserID),
					log.String("outcome_id", o.ID))
				return err
			}
		}
		return nil
	})
	return err
}

func (ocm OutcomeModel) SearchPrivateOutcomes(ctx context.Context, tx *dbo.DBContext, condition *da.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error) {
	condition.PublishStatus = dbo.NullStrings{
		Strings: []string{entity.ContentStatusDraft, entity.ContentStatusPending, entity.ContentStatusRejected},
		Valid:   true,
	}
	condition.AuthorID = sql.NullString{String: user.UserID, Valid: true}
	total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "BulkDelLearningOutcome: DeleteOutcome failed",
			log.Err(err),
			log.String("op", user.UserID),
			log.Any("outcome", ocm))
		return 0, nil, err
	}
	return total, outcomes, nil
}

func (ocm OutcomeModel) SearchPendingOutcomes(ctx context.Context, tx *dbo.DBContext, condition *da.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error) {
	condition.PublishStatus = dbo.NullStrings{
		Strings: []string{entity.ContentStatusDraft, entity.ContentStatusPending},
		Valid:   true,
	}
	condition.PublishScope = sql.NullString{String: "user's top org id", Valid: true}
	total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "SearchPendingOutcomes: SearchOutcome failed",
			log.Err(err),
			log.String("op", user.UserID),
			log.Any("outcome", ocm))
		return 0, nil, err
	}
	return total, outcomes, nil
}

func (ocm OutcomeModel) GetLearningOutcomesByIDs(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) ([]*entity.Outcome, error) {
	condition := da.OutcomeCondition{
		IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
	}
	_, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, &condition)
	if err != nil {
		log.Error(ctx, "GetLearningOutcomesByIDs: SearchOutcome failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.Any("outcome", ocm))
		return nil, err
	}
	return outcomes, nil
}

func (ocm OutcomeModel) GetLatestOutcomesByIDs(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) (outcomes []*entity.Outcome, err error) {
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		cond1 := da.OutcomeCondition{
			IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		}
		total, otcs1, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, &cond1)
		if err != nil {
			log.Error(ctx, "GetLatestOutcomesByIDs: SearchOutcome failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			return err
		}
		if total == 0 {
			log.Debug(ctx, "GetLatestOutcomesByIDs: SearchOutcome return empty",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			outcomes = []*entity.Outcome{}
			return nil
		}
		cond2 := da.OutcomeCondition{}
		for _, o := range otcs1 {
			cond2.IDs.Strings = append(cond2.IDs.Strings, o.LatestID)
		}
		cond2.IDs.Valid = true
		total, otcs2, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, &cond2)
		if err != nil {
			log.Error(ctx, "GetLatestOutcomesByIDs: SearchOutcome failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", cond2.IDs.Strings))
			return err
		}
		if total == 0 {
			log.Debug(ctx, "GetLatestOutcomesByIDs: SearchOutcome return empty",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", cond2.IDs.Strings))
			outcomes = []*entity.Outcome{}
		} else {
			outcomes = otcs2
		}
		return nil
	})
	return
}

func (ocm OutcomeModel) getAuthorNameByID(ctx context.Context, id string) (name string, err error) {
	//TODO:
	if err != nil {
		log.Error(ctx, "getAuthorNameByID failed",
			log.Err(err),
			log.String("author_id", id))
	}
	return
}

func (ocm OutcomeModel) getOrganizationNameByID(ctx context.Context, id string) (orgName string, err error) {
	//TODO:
	if err != nil {
		log.Error(ctx, "getOrganizationNameByID failed",
			log.Err(err),
			log.String("org_id", id))
	}
	return
}

func (ocm OutcomeModel) getRootOrganizationByOrgID(ctx context.Context, id string) (orgID, orgName string, err error) {
	// TODO:
	if err != nil {
		log.Error(ctx, "getRootOrganizationByOrgID failed",
			log.Err(err),
			log.String("org_id", id))
	}
	return
}

func (ocm OutcomeModel) getRootOrganizationByAuthorID(ctx context.Context, id string) (orgID, orgName string, err error) {
	// TODO:
	if err != nil {
		log.Error(ctx, "getRootOrganizationByOrgID failed",
			log.Err(err),
			log.String("author_id", id))
	}
	return
}

func (ocm OutcomeModel) getShortCode(ctx context.Context) (shortcode string, err error) {
	// TODO:
	if err != nil {
		log.Error(ctx, "getRootOrganizationByOrgID failed",
			log.Err(err))
	}
	return
}

func (ocm OutcomeModel) lockOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) (err error) {
	// must in a transaction
	if outcome.LockedBy != "" && outcome.LockedBy != "-" {
		err = ErrContentAlreadyLocked
		log.Warn(ctx, "lockOutcome: invalid lock status",
			log.Err(err),
			log.String("op", operator.UserID))
		return
	}
	outcome.LockedBy = operator.UserID
	err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "lockOutcome: UpdateOutcome failed",
			log.Err(err),
			log.String("op", operator.UserID))
	}
	return
}

func (ocm OutcomeModel) unlockOutcome(ctx context.Context, tx *dbo.DBContext, otcid string) (err error) {
	// must in a transaction
	outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, otcid)
	if err != nil {
		log.Error(ctx, "unlockOutcome: GetOutcomeByID failed",
			log.String("outcome_id", otcid))
		return
	}
	if outcome.LockedBy != "" && outcome.LockedBy != "-" {
		outcome.LockedBy = "-"
		err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, outcome)
		if err != nil {
			log.Error(ctx, "unlockOutcome: UpdateOutcome failed",
				log.String("outcome_id", otcid))
			return
		}
	}
	return
}

func (ocm OutcomeModel) deleteOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	// must in a transaction
	if outcome.LockedBy != "" && outcome.LockedBy != "-" {
		err = ErrContentAlreadyLocked
		log.Error(ctx, "deleteOutcome: invalid lock status",
			log.Err(err),
			log.Any("outcome", outcome))
		return
	}
	err = da.GetOutcomeDA().DeleteOutcome(ctx, tx, outcome.ID)
	if err != nil {
		log.Error(ctx, "deleteOutcome: DeleteOutcome failed",
			log.Err(err),
			log.Any("outcome", outcome))
		return
	}
	if outcome.SourceID != "" && outcome.SourceID != outcome.ID {
		err = ocm.unlockOutcome(ctx, tx, outcome.SourceID)
		//if gorm.IsRecordNotFoundError(err) {
		//	log.Error(ctx, "deleteOutcome: unlockOutcome maybe inconsistency",
		//		log.Any("outcome", outcome))
		//	err = nil
		//}
		if err != nil {
			log.Error(ctx, "deleteOutcome: unlockOutcome failed",
				log.Any("outcome", outcome))
			return
		}
	}
	return
}
