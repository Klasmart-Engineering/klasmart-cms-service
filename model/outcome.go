package model

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"gitlab.badanamu.com.cn/calmisland/ro"
)

type IOutcomeModel interface {
	CreateLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) error
	GetLearningOutcomeByID(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) (*entity.Outcome, error)
	UpdateLearningOutcome(ctx context.Context, outcome *entity.Outcome, operator *entity.Operator) error
	DeleteLearningOutcome(ctx context.Context, outcomeID string, operator *entity.Operator) error
	SearchLearningOutcome(ctx context.Context, tx *dbo.DBContext, condition *entity.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error)

	LockLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) (string, error)

	PublishLearningOutcome(ctx context.Context, outcomeID string, scope string, operator *entity.Operator) error
	BulkPubLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, scope string, operator *entity.Operator) error
	BulkDelLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) error

	SearchPrivateOutcomes(ctx context.Context, tx *dbo.DBContext, condition *entity.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error)
	SearchPendingOutcomes(ctx context.Context, tx *dbo.DBContext, condition *entity.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error)

	GetLearningOutcomesByIDs(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) ([]*entity.Outcome, error)
	GetLatestOutcomesByIDs(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) ([]*entity.Outcome, error)

	ApproveLearningOutcome(ctx context.Context, outcomeID string, operator *entity.Operator) error
	RejectLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, reason string, operator *entity.Operator) error
}

type OutcomeModel struct {
}

func (ocm OutcomeModel) CreateLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) (err error) {
	// outcome get value from api lay, this lay add some information
	outcome.ID = utils.NewID()
	outcome.AncestorID = outcome.ID
	//outcome.SourceID = outcome.ID
	outcome.AuthorID = operator.UserID
	outcome.AuthorName, err = ocm.getAuthorNameByID(ctx, outcome.AuthorID)
	if err != nil {
		log.Error(ctx, "CreateLearningOutcome: getAuthorNameByID failed",
			log.String("op", outcome.AuthorID),
			log.Any("outcome", outcome))
		return
	}
	outcome.OrganizationID = operator.OrgID
	//outcome.OrganizationID, _, err = ocm.getRootOrganizationByAuthorID(ctx, operator.UserID)
	//if err != nil {
	//	log.Error(ctx, "CreateLearningOutcome: getRootOrganizationByAuthorID failed",
	//		log.String("op", operator.UserID),
	//		log.Any("outcome", outcome))
	//	return
	//}
	outcome.Shortcode, err = ocm.getShortCode(ctx, outcome.OrganizationID)
	if err != nil {
		log.Error(ctx, "CreateLearningOutcome: getShortCode failed",
			log.String("op", operator.UserID),
			log.Any("outcome", outcome))
		return
	}
	outcome.PublishStatus = entity.OutcomeStatusDraft
	err = da.GetOutcomeDA().CreateOutcome(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "CreateLearningOutcome: CreateOutcome failed",
			log.String("op", operator.UserID),
			log.Any("outcome", outcome))
	}
	return err
}

func (ocm OutcomeModel) GetLearningOutcomeByID(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) (*entity.Outcome, error) {
	outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
	if err == dbo.ErrRecordNotFound {
		return nil, ErrResourceNotFound
	}
	if err != nil {
		log.Error(ctx, "GetLearningOutcomeByID: GetOutcomeByID failed",
			log.String("op", operator.UserID),
			log.String("outcome_id", outcomeID))
		return nil, err
	}
	return outcome, nil
}

func (ocm OutcomeModel) UpdateLearningOutcome(ctx context.Context, outcome *entity.Outcome, operator *entity.Operator) error {
	err := dbo.GetTrans(ctx, func(cxt context.Context, tx *dbo.DBContext) error {
		data, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcome.ID)
		if err == dbo.ErrRecordNotFound {
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "UpdateLearningOutcome: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.Any("data", outcome))
			return err
		}
		if data.PublishStatus != entity.OutcomeStatusDraft && data.PublishStatus != entity.OutcomeStatusRejected {
			log.Error(ctx, "UpdateLearningOutcome: publish status not allowed edit",
				log.String("op", operator.UserID),
				log.Any("data", data))
			return ErrInvalidPublishStatus
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
	})
	return err
}

func (ocm OutcomeModel) DeleteLearningOutcome(ctx context.Context, outcomeID string, operator *entity.Operator) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
		if err != nil && err != dbo.ErrRecordNotFound {
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

func (ocm OutcomeModel) SearchLearningOutcome(ctx context.Context, tx *dbo.DBContext, condition *entity.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error) {
	if condition.OrganizationID == "" {
		//orgID, _, err := ocm.getRootOrganizationByAuthorID(ctx, user.UserID)
		//if err != nil {
		//	log.Error(ctx, "SearchLearningOutcome: getRootOrganizationByAuthorID failed",
		//		log.String("op", user.UserID),
		//		log.Any("condition", condition))
		//	return 0, nil, err
		//}
		condition.OrganizationID = user.OrgID
	}
	if condition.PublishStatus == "" { // Must search published outcomes
		condition.PublishStatus = entity.OutcomeStatusPublished
	}
	total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, da.NewOutcomeCondition(condition))
	if err != nil {
		log.Error(ctx, "SearchLearningOutcome: DeleteOutcome failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	return total, outcomes, nil
}

func (ocm OutcomeModel) LockLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, operator *entity.Operator) (string, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeLock)
	if err != nil {
		log.Error(ctx, "LockLearningOutcome: NewLock failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.String("outcome_id", outcomeID))
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()
	var newVersion entity.Outcome
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
		if err == dbo.ErrRecordNotFound {
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "LockLearningOutcome: GetOutcomeByID failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}

		if outcome.LockedBy == operator.UserID {
			copyValue, err := da.GetOutcomeDA().GetOutcomeBySourceID(ctx, tx, outcomeID)
			if err != nil {
				log.Error(ctx, "LockLearningOutcome: GetOutcomeBySourceID failed",
					log.String("op", operator.UserID),
					log.String("outcome_id", outcomeID))
				return err
			}
			if copyValue.PublishStatus == entity.OutcomeStatusDraft {
				newVersion = *copyValue
				return nil
			} else {
				log.Error(ctx, "LockLearningOutcome: copyValue status not draft",
					log.String("op", operator.UserID),
					log.Any("copy", copyValue))
				return ErrContentAlreadyLocked
			}
		}

		err = ocm.lockOutcome(ctx, tx, outcome, operator)
		if err != nil {
			return err
		}
		newVersion = outcome.Clone()
		err = da.GetOutcomeDA().CreateOutcome(ctx, tx, &newVersion)
		if err != nil {
			log.Error(ctx, "LockLearningOutcome: CreateOutcome failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID),
				log.Any("outcome", newVersion))
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return newVersion.ID, nil
}

func (ocm OutcomeModel) PublishLearningOutcome(ctx context.Context, outcomeID string, scope string, operator *entity.Operator) error {
	if scope == "" {
		//scopeID, _, err := ocm.getRootOrganizationByAuthorID(ctx, operator.UserID)
		//if err != nil {
		//	log.Error(ctx, "PublishLearningOutcome: getRootOrganizationByAuthorID failed",
		//		log.String("op", operator.UserID),
		//		log.String("outcome_id", outcomeID))
		//}
		scope = operator.OrgID
	}
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
		if err == dbo.ErrRecordNotFound {
			err = ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "PublishLearningOutcome: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = outcome.SetStatus(ctx, entity.OutcomeStatusPending)
		if err != nil {
			log.Error(ctx, "PublishLearningOutcome: SetStatus failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return ErrInvalidContentStatusToPublish
		}
		if outcome.PublishScope != "" && outcome.PublishScope != scope {
			log.Error(ctx, "PublishLearningOutcome: scope mismatch",
				log.String("op", operator.UserID),
				log.String("scope", scope),
				log.Any("outcome", outcome))
			return ErrInvalidContentStatusToPublish
		}
		outcome.PublishScope = scope
		outcome.UpdateAt = time.Now().Unix()
		err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, outcome)
		if err != nil {
			log.Error(ctx, "PublishLearningOutcome: UpdateOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		return nil
	})
	return err
}

func (ocm OutcomeModel) BulkPubLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, scope string, operator *entity.Operator) error {
	if scope == "" {
		//scopeID, _, err := ocm.getRootOrganizationByAuthorID(ctx, operator.UserID)
		//if err != nil {
		//	log.Error(ctx, "PublishLearningOutcome: getRootOrganizationByAuthorID failed",
		//		log.String("op", operator.UserID),
		//		log.Strings("outcome_ids", outcomeIDs))
		//}
		scope = operator.OrgID
	}
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
			err = o.SetStatus(ctx, entity.OutcomeStatusPublished)
			if err != nil {
				log.Error(ctx, "BulkPubLearningOutcome: SetStatus failed",
					log.String("op", operator.UserID),
					log.Any("outcome", o))
				return ErrInvalidContentStatusToPublish
			}
			if o.PublishScope != "" && o.PublishScope != scope {
				log.Error(ctx, "PublishLearningOutcome: scope mismatch",
					log.String("op", operator.UserID),
					log.String("scope", scope),
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

func (ocm OutcomeModel) SearchPrivateOutcomes(ctx context.Context, tx *dbo.DBContext, condition *entity.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error) {
	if condition.PublishStatus != entity.OutcomeStatusDraft &&
		condition.PublishStatus != entity.OutcomeStatusPending &&
		condition.PublishStatus != entity.OutcomeStatusRejected {
		log.Error(ctx, "SearchPrivateOutcomes: SearchPendingOutcomes failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, ErrBadRequest
	}
	condition.AuthorID = user.UserID
	total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, da.NewOutcomeCondition(condition))
	if err != nil {
		log.Error(ctx, "BulkDelLearningOutcome: DeleteOutcome failed",
			log.Err(err),
			log.String("op", user.UserID),
			log.Any("outcome", ocm))
		return 0, nil, err
	}
	return total, outcomes, nil
}

func (ocm OutcomeModel) SearchPendingOutcomes(ctx context.Context, tx *dbo.DBContext, condition *entity.OutcomeCondition, user *entity.Operator) (int, []*entity.Outcome, error) {
	if condition.PublishStatus != entity.OutcomeStatusPending {
		log.Error(ctx, "SearchPendingOutcomes: SearchPendingOutcomes failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, ErrBadRequest
	}
	orgID, _, err := ocm.getRootOrganizationByOrgID(ctx, user.UserID)
	if err != nil {
		log.Error(ctx, "SearchPendingOutcomes: SearchPendingOutcomes failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	condition.PublishScope = orgID
	total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, da.NewOutcomeCondition(condition))
	if err != nil {
		log.Error(ctx, "SearchPendingOutcomes: SearchOutcome failed",
			log.Err(err),
			log.String("op", user.UserID),
			log.Any("outcome", ocm))
		return 0, nil, err
	}
	return total, outcomes, nil
}

func (ocm OutcomeModel) ApproveLearningOutcome(ctx context.Context, outcomeID string, operator *entity.Operator) error {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeReview)
	if err != nil {
		log.Error(ctx, "ApproveLearningOutcome: NewLock failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.String("outcome_id", outcomeID))
		return err
	}
	locker.Lock()
	defer locker.Unlock()
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
		if err == dbo.ErrRecordNotFound || gorm.IsRecordNotFoundError(err) {
			log.Warn(ctx, "ApproveLearningOutcome: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "ApproveLearningOutcome: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = outcome.SetStatus(ctx, entity.OutcomeStatusPublished)
		if err != nil {
			log.Error(ctx, "ApproveLearningOutcome: SetStatus failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return ErrInvalidPublishStatus
		}
		if outcome.LatestID == "" {
			outcome.LatestID = outcome.ID
		}
		err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, outcome)
		if err != nil {
			log.Error(ctx, "ApproveLearningOutcome: UpdateOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = ocm.hideParent(ctx, tx, outcome)
		if err != nil {
			log.Error(ctx, "ApproveLearningOutcome: hideParent failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = ocm.updateLatestToHead(ctx, tx, outcome)
		if err != nil {
			log.Error(ctx, "ApproveLearningOutcome: updateLatestToHead failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		return nil
	})
	return err
}

func (ocm OutcomeModel) RejectLearningOutcome(ctx context.Context, tx *dbo.DBContext, outcomeID string, reason string, operator *entity.Operator) error {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeReview)
	if err != nil {
		log.Error(ctx, "RejectLearningOutcome: NewLock failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.String("outcome_id", outcomeID))
		return err
	}
	locker.Lock()
	defer locker.Unlock()
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
		if err == dbo.ErrRecordNotFound || gorm.IsRecordNotFoundError(err) {
			log.Warn(ctx, "RejectLearningOutcome: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "RejectLearningOutcome: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = outcome.SetStatus(ctx, entity.OutcomeStatusRejected)
		outcome.RejectReason = reason
		if err != nil {
			log.Error(ctx, "RejectLearningOutcome: SetStatus failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return ErrInvalidPublishStatus
		}
		err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, outcome)
		if err != nil {
			log.Error(ctx, "RejectLearningOutcome: UpdateOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		return nil
	})
	return err
}

func (ocm OutcomeModel) GetLearningOutcomesByIDs(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) ([]*entity.Outcome, error) {
	condition := da.OutcomeCondition{
		IDs:            dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		IncludeDeleted: true,
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

func (ocm OutcomeModel) getShortCode(ctx context.Context, orgID string) (shortcode string, err error) {
	redisKey := fmt.Sprintf("%s:%s", da.RedisKeyPrefixOutcomeShortcode, orgID)
	num, err := ro.MustGetRedis(ctx).Incr(redisKey).Result()
	if err != nil {
		log.Error(ctx, "getShortCode failed",
			log.Err(err))
	}
	shortcode = PaddingStr(NumToBHex(int(num), constant.ShortcodeBaseCustom), constant.ShortcodeShowLength)
	return
}

func (ocm OutcomeModel) lockOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome, operator *entity.Operator) (err error) {
	// must in a transaction
	if outcome.PublishStatus != entity.OutcomeStatusPublished {
		err = ErrInvalidPublishStatus
		log.Warn(ctx, "lockOutcome: invalid lock status",
			log.Err(err),
			log.String("op", operator.UserID))
		return
	}
	if outcome.LockedBy != "" && outcome.LockedBy != constant.LockedByNoBody {
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
	if outcome.LockedBy != "" && outcome.LockedBy != constant.LockedByNoBody {
		outcome.LockedBy = constant.LockedByNoBody
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
	if outcome.LockedBy != "" && outcome.LockedBy != constant.LockedByNoBody {
		err = ErrContentAlreadyLocked
		log.Error(ctx, "deleteOutcome: invalid lock status",
			log.Err(err),
			log.Any("outcome", outcome))
		return
	}
	err = da.GetOutcomeDA().DeleteOutcome(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "deleteOutcome: DeleteOutcome failed",
			log.Err(err),
			log.Any("outcome", outcome))
		return
	}
	if outcome.SourceID != "" && outcome.SourceID != outcome.ID {
		err = ocm.unlockOutcome(ctx, tx, outcome.SourceID)
		// TODO: data maybe inconsistency, but seems can be ignore
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

func (ocm OutcomeModel) hideParent(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	// must in a transaction
	if outcome.SourceID == "" || outcome.SourceID == outcome.ID {
		return
	}
	parent, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcome.SourceID)
	// TODO: data maybe inconsistency, but seems can be ignore
	//if gorm.IsRecordNotFoundError(err) {
	//	log.Error(ctx, "hideParent: data maybe inconsistency",
	//		log.Any("outcome", outcome))
	//	err = nil
	//}
	if err != nil {
		log.Error(ctx, "hideParent: GetOutcomeByID failed",
			log.Any("outcome", outcome))
		return
	}
	parent.LockedBy = constant.LockedByNoBody
	err = parent.SetStatus(ctx, entity.OutcomeStatusHidden)
	if err != nil {
		log.Error(ctx, "hideParent: SetStatus failed",
			log.Any("outcome", parent))
		return ErrInvalidPublishStatus
	}
	err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, parent)
	if err != nil {
		log.Error(ctx, "hideParent: UpdateOutcome failed",
			log.Any("outcome", parent))
		return err
	}
	return nil
}

func (ocm OutcomeModel) updateLatestToHead(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	// must in a transaction
	if outcome.LatestID == outcome.ID {
		return nil
	}
	err = da.GetOutcomeDA().UpdateLatestHead(ctx, tx, outcome.LatestID, outcome.ID)
	return
}

func (ocm OutcomeModel) getAuthorNameByID(ctx context.Context, id string) (name string, err error) {
	provider := external.GetUserServiceProvider()
	user, err := provider.GetUserInfoByID(ctx, id)
	if err != nil {
		log.Error(ctx, "getAuthorNameByID: GetUserInfoByID failed",
			log.Err(err),
			log.String("user_id", id))
		return "", err
	}
	return user.UserName, nil
}

func (ocm OutcomeModel) getOrganizationNameByID(ctx context.Context, id string) (orgName string, err error) {
	provider := external.GetOrganizationServiceProvider()
	orgs, err := provider.BatchGet(ctx, []string{id})
	if err != nil {
		log.Error(ctx, "getOrganizationNameByID: BatchGet failed",
			log.Err(err),
			log.String("org_id", id))
		return "", err
	}
	if len(orgs) == 0 {
		log.Error(ctx, "getOrganizationNameByID: org list is empty",
			log.String("org_id", id))
		return "", nil
	}
	return orgs[0].Name, nil
}

func (ocm OutcomeModel) getRootOrganizationByOrgID(ctx context.Context, id string) (orgID, orgName string, err error) {
	provider := external.GetOrganizationServiceProvider()
	orgs, err := provider.GetParents(ctx, orgID)
	if err != nil {
		log.Error(ctx, "getRootOrganizationByOrgID: GetMyTopOrg failed",
			log.Err(err),
			log.String("org_id", id))
		return "", "", err
	}
	if len(orgs) == 0 {
		log.Error(ctx, "getRootOrganizationByOrgID: parents is empty",
			log.Err(err),
			log.String("org_id", id))
		return "", "", nil
	}
	root := orgs[0]
	for i := 0; i < len(orgs); i++ {
		if root.ParentID != "" && root.ParentID != root.ID {
			for _, o := range orgs {
				if o.ID == root.ParentID {
					root = o
					break
				}
			}
		}
	}
	return root.ID, root.Name, nil
}

func (ocm OutcomeModel) getRootOrganizationByAuthorID(ctx context.Context, id string) (orgID, orgName string, err error) {
	provider := external.GetUserServiceProvider()
	user, err := provider.GetUserInfoByID(ctx, id)
	if err != nil {
		log.Error(ctx, "getRootOrganizationByAuthorID failed",
			log.Err(err),
			log.String("user_id", id))
	}
	return ocm.getRootOrganizationByOrgID(ctx, user.OrgID)
}

var (
	_outcomeModel     IOutcomeModel
	_outcomeModelOnce sync.Once
)

func GetOutcomeModel() IOutcomeModel {
	_outcomeModelOnce.Do(func() {
		_outcomeModel = new(OutcomeModel)
	})
	return _outcomeModel
}

var num2char = "0123456789abcdefghijklmnopqrstuvwxyz"

func NumToBHex(num int, n int) string {
	numStr := ""
	for num != 0 {
		yu := num % n
		numStr = string(num2char[yu]) + numStr
		num = num / n
	}
	return strings.ToUpper(numStr)
}

func PaddingStr(s string, l int) string {
	if l <= len(s) {
		return s
	}
	return strings.Repeat("0", l-len(s)) + s
}
