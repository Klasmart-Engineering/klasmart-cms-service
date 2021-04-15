package model

import (
	"context"
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
)

type IOutcomeModel interface {
	CreateLearningOutcome(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) error
	GetLearningOutcomeByID(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeID string) (*entity.Outcome, error)
	UpdateLearningOutcome(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) error
	DeleteLearningOutcome(ctx context.Context, operator *entity.Operator, outcomeID string) error
	SearchLearningOutcome(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error)

	LockLearningOutcome(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeID string) (string, error)

	PublishLearningOutcome(ctx context.Context, operator *entity.Operator, outcomeID string, scope string) error
	BulkPubLearningOutcome(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string, scope string) error
	BulkDelLearningOutcome(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) error

	SearchPrivateOutcomes(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error)
	SearchPendingOutcomes(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error)

	GetLearningOutcomesByIDs(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) ([]*entity.Outcome, error)
	GetLatestOutcomesByIDs(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) ([]*entity.Outcome, error)

	ApproveLearningOutcome(ctx context.Context, operator *entity.Operator, outcomeID string) error
	RejectLearningOutcome(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeID string, reason string) error

	BulkApproveLearningOutcome(ctx context.Context, operator *entity.Operator, outcomeIDs []string) error
	BulkRejectLearningOutcome(ctx context.Context, operator *entity.Operator, outcomeIDs []string, reason string) error

	GetLatestOutcomesByIDsMapResult(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (map[string]*entity.Outcome, error)

	HasLockedOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (bool, error)
}

type OutcomeModel struct {
}

func (ocm OutcomeModel) CreateLearningOutcome(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) (err error) {
	// outcome get value from api lay, this lay add some information
	outcome.AuthorName, err = ocm.getAuthorNameByID(ctx, operator, operator.UserID)
	if err != nil {
		log.Error(ctx, "CreateLearningOutcome: getAuthorNameByID failed",
			log.String("op", outcome.AuthorID),
			log.Any("outcome", outcome))
		return
	}

	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixShortcodeMute, constant.ShortcodeOutcomeKind, operator.OrgID)
	if err != nil {
		log.Error(ctx, "CreateLearningOutcome: NewLock failed",
			log.Err(err),
			log.Any("op", operator),
			log.Any("outcome", outcome))
		return err
	}
	locker.Lock()
	defer locker.Unlock()
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		outcome.ID = utils.NewID()
		outcome.AncestorID = outcome.ID
		outcome.AuthorID = operator.UserID
		outcome.OrganizationID = operator.OrgID
		outcome.PublishStatus = entity.OutcomeStatusDraft
		exists, err := GetShortcodeModel().isOccupied(ctx, tx, entity.KindOutcome, operator.OrgID, outcome.AncestorID, outcome.Shortcode)
		if err != nil {
			log.Error(ctx, "CreateLearningOutcome: IsShortcodeExistInDBWithOtherAncestor failed",
				log.Err(err),
				log.Any("op", operator),
				log.Any("outcome", outcome))
			return err
		}
		if exists {
			return constant.ErrConflict
		}
		err = GetOutcomeSetModel().BindByOutcome(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "CreateLearningOutcome: BindByOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = da.GetOutcomeDA().CreateOutcome(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "CreateLearningOutcome: CreateOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = da.GetAttachDA().Replace(ctx, tx, entity.AttachOutcomeTable, nil, entity.OutcomeType, outcome.CollectAttach())
		if err != nil {
			log.Error(ctx, "CreateLearningOutcome: Replace failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		return nil
	})
	da.GetShortcodeCacheDA().Remove(ctx, constant.ShortcodeOutcomeKind, operator.OrgID, outcome.Shortcode)
	if err != nil {
		return err
	}
	da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, operator, nil)
	return
}

func (ocm OutcomeModel) GetLearningOutcomeByID(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeID string) (*entity.Outcome, error) {
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
	attaches, err := da.GetAttachDA().SearchAttaches(ctx, tx, entity.AttachOutcomeTable, []string{outcomeID}, entity.OutcomeType)
	if err != nil {
		log.Error(ctx, "GetLearningOutcomeByID: SearchAttaches failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.String("outcome_id", outcomeID))
	}
	outcome.FillAttach(attaches)
	return outcome, nil
}

func (ocm OutcomeModel) updateOutcomeSet(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcomeID string, sets []*entity.Set) error {
	err := da.GetOutcomeSetDA().DeleteBoundOutcomeSet(ctx, tx, outcomeID)
	if err != nil {
		log.Error(ctx, "updateOutcomeSet: DeleteBoundOutcomeSet failed",
			log.Err(err),
			log.String("outcome", outcomeID),
			log.Any("sets", sets))
		return err
	}
	outcomeSets := make([]*entity.OutcomeSet, len(sets))
	for i := range sets {
		outcomeSet := entity.OutcomeSet{
			OutcomeID: outcomeID,
			SetID:     sets[i].ID,
		}
		outcomeSets[i] = &outcomeSet
	}
	if len(outcomeSets) == 0 {
		log.Info(ctx, "updateOutcomeSet: delete all bound sets",
			log.String("outcome", outcomeID))
		return nil
	}
	err = da.GetOutcomeSetDA().BindOutcomeSet(ctx, op, tx, outcomeSets)
	if err != nil {
		log.Error(ctx, "updateOutcomeSet: BindOutcomeSet failed",
			log.Err(err),
			log.String("outcome", outcomeID),
			log.Any("sets", sets))
		return err
	}
	return nil
}

func (ocm OutcomeModel) UpdateLearningOutcome(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) error {
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, []external.PermissionName{
		external.EditMyUnpublishedLearningOutcome,
		external.EditOrgUnpublishedLearningOutcome,
	})
	if err != nil {
		log.Error(ctx, "UpdateLearningOutcome:HasOrganizationPermissions failed", log.Any("op", operator), log.Err(err))
		return err
	}
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixShortcodeMute, constant.ShortcodeOutcomeKind, operator.OrgID)
	if err != nil {
		log.Error(ctx, "UpdateLearningOutcome: NewLock failed",
			log.Err(err),
			log.Any("op", operator),
			log.Any("outcome", outcome))
		return err
	}
	locker.Lock()
	defer locker.Unlock()
	exists, err := GetShortcodeModel().isCached(ctx, entity.KindOutcome, operator.OrgID, outcome.Shortcode)
	if err != nil {
		log.Error(ctx, "UpdateLearningOutcome: IsShortcodeExistInRedis failed",
			log.Err(err),
			log.Any("op", operator),
			log.Any("outcome", outcome))
		return err
	}
	if exists {
		return constant.ErrConflict
	}
	err = dbo.GetTrans(ctx, func(cxt context.Context, tx *dbo.DBContext) error {
		data, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcome.ID)
		if err == dbo.ErrRecordNotFound {
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "UpdateLearningOutcome: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.Any("data", data))
			return err
		}
		if !allowEditOutcome(ctx, operator, perms, data) {
			log.Warn(ctx, "UpdateLearningOutcome: no permission",
				log.Any("op", operator),
				log.Any("perms", perms),
				log.Any("data", data))
			return constant.ErrOperateNotAllowed
		}
		if data.PublishStatus != entity.OutcomeStatusDraft && data.PublishStatus != entity.OutcomeStatusRejected {
			log.Error(ctx, "UpdateLearningOutcome: publish status not allowed edit",
				log.String("op", operator.UserID),
				log.Any("data", data))
			return ErrInvalidPublishStatus
		}

		if data.Shortcode != outcome.Shortcode {
			exists, err := GetShortcodeModel().isOccupied(ctx, tx, entity.KindOutcome, operator.OrgID, data.AncestorID, outcome.Shortcode)
			if err != nil {
				log.Error(ctx, "UpdateLearningOutcome: IsShortcodeExistInDBWithOtherAncestor failed",
					log.Err(err),
					log.Any("op", operator),
					log.Any("outcome", outcome))
				return err
			}
			if exists {
				return constant.ErrConflict
			}
		}

		data.Update(outcome)
		err = ocm.updateOutcomeSet(ctx, operator, tx, outcome.ID, outcome.Sets)
		if err != nil {
			log.Error(ctx, "UpdateLearningOutcome: updateOutcomeSet failed",
				log.String("op", operator.UserID),
				log.Any("data", outcome))
			return err
		}
		// because of cache, follow statements need be at last
		err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, data)
		if err != nil {
			log.Error(ctx, "UpdateLearningOutcome: UpdateOutcome failed",
				log.String("op", operator.UserID),
				log.Any("data", outcome))
			return err
		}
		err = da.GetAttachDA().Replace(ctx, tx, entity.AttachOutcomeTable, []string{outcome.ID}, entity.OutcomeType, outcome.CollectAttach())
		if err != nil {
			log.Error(ctx, "UpdateLearningOutcome: Replace failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		return nil
	})
	return err
}

func (ocm OutcomeModel) DeleteLearningOutcome(ctx context.Context, operator *entity.Operator, outcomeID string) error {
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, []external.PermissionName{
		external.DeleteMyUnpublishedLearningOutcome,
		external.DeleteOrgUnpublishedLearningOutcome,
		external.DeleteMyPendingLearningOutcome,
		external.DeleteOrgPendingLearningOutcome,
		external.DeletePublishedLearningOutcome,
	})
	if err != nil {
		log.Error(ctx, "DeleteLearningOutcome:HasOrganizationPermissions failed", log.Any("op", operator), log.Err(err))
		return err
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
		if err != nil && err != dbo.ErrRecordNotFound {
			log.Error(ctx, "DeleteLearningOutcome: no permission",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		if !allowDeleteOutcome(ctx, operator, perms, outcome) {
			log.Warn(ctx, "DeleteLearningOutcome: no permission", log.Any("op", operator),
				log.Any("perms", perms), log.Any("outcome", outcome))
			return constant.ErrOperateNotAllowed
		}
		err = ocm.deleteOutcome(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "DeleteLearningOutcome: deleteOutcome failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = da.GetOutcomeSetDA().DeleteBoundOutcomeSet(ctx, tx, outcome.ID)
		if err != nil {
			log.Error(ctx, "DeleteLearningOutcome: deleteOutcome failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = da.GetAttachDA().Replace(ctx, tx, entity.AttachOutcomeTable, []string{outcome.ID}, entity.OutcomeType, nil)
		if err != nil {
			log.Error(ctx, "DeleteLearningOutcome: Replace failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = da.GetMilestoneDA().UnbindOutcomes(ctx, tx, []string{outcome.ID})
		if err != nil {
			log.Error(ctx, "DeleteLearningOutcome: UnbindOutcomes failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		return nil
	})

	if err == nil {
		da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, operator, nil)
	}
	return err
}

func (ocm OutcomeModel) fillAuthorIDs(ctx context.Context, op *entity.Operator, condition *entity.OutcomeCondition) error {
	var authorName string
	if condition.FuzzyKey != "" {
		authorName = condition.FuzzyKey
	}
	if condition.FuzzyKey == "" && condition.AuthorName != "" {
		authorName = condition.AuthorName
		// if no authName matched, it shouldn't match outcome. but not ignore the query condition
		condition.AuthorIDs = append(condition.AuthorIDs, "")
	}

	if authorName != "" {
		users, err := external.GetUserServiceProvider().Query(ctx, op, op.OrgID, authorName)
		if err != nil {
			log.Error(ctx, "fillAuthorIDs: GetUserServiceProvider failed",
				log.Any("op", op),
				log.Any("condition", condition))
			return err
		}
		for _, u := range users {
			condition.AuthorIDs = append(condition.AuthorIDs, u.ID)
		}
	}
	return nil
}

func (ocm OutcomeModel) fillIDsBySetName(ctx context.Context, op *entity.Operator, condition *entity.OutcomeCondition) error {
	var setName string
	if condition.FuzzyKey != "" {
		setName = condition.FuzzyKey
	}
	if condition.FuzzyKey == "" && condition.SetName != "" {
		setName = condition.SetName
		// if no setName matched, it shouldn't match outcome. but not ignore the query condition
		condition.IDs = append(condition.IDs, "")
	}

	if setName != "" {
		outcomeSets, err := da.GetOutcomeSetDA().SearchOutcomeBySetName(ctx, op, setName)
		if err != nil {
			log.Error(ctx, "fillIDsBySetName: SearchOutcomeBySetName failed",
				log.Err(err),
				log.Any("op", op),
				log.String("set_name", condition.SetName))
			return err
		}
		for i := range outcomeSets {
			condition.IDs = append(condition.IDs, outcomeSets[i].OutcomeID)
		}
	}
	return nil
}

func (ocm OutcomeModel) SearchLearningOutcome(ctx context.Context, user *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error) {
	if condition.OrganizationID == "" {
		condition.OrganizationID = user.OrgID
	}
	if condition.AuthorName == constant.Self {
		condition.AuthorID = user.UserID
		condition.AuthorName = ""
	}
	if condition.PublishStatus == "" { // Must search published outcomes
		condition.PublishStatus = entity.OutcomeStatusPublished
	}

	err := ocm.fillAuthorIDs(ctx, user, condition)
	if err != nil {
		log.Error(ctx, "SearchLearningOutcome: fillAuthorIDs failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	err = ocm.fillIDsBySetName(ctx, user, condition)
	if err != nil {
		log.Error(ctx, "SearchLearningOutcome: fillIDsBySetName failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}

	total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, user, tx, da.NewOutcomeCondition(condition))
	if err != nil {
		log.Error(ctx, "SearchLearningOutcome: SearchOutcome failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	err = ocm.fillAttach(ctx, user, tx, outcomes)
	if err != nil {
		log.Error(ctx, "SearchLearningOutcome: fillAttach failed",
			log.Err(err),
			log.String("op", user.UserID))
		return 0, nil, err
	}
	return total, outcomes, nil
}

func (ocm OutcomeModel) LockLearningOutcome(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeID string) (string, error) {
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
			copyValue, err := da.GetOutcomeDA().GetOutcomeBySourceID(ctx, operator, tx, outcomeID)
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
				return NewErrContentAlreadyLocked(ctx, outcome.LockedBy, operator)
			}
		}

		err = ocm.lockOutcome(ctx, operator, tx, outcome)
		if err != nil {
			return err
		}
		newVersion = outcome.Clone(operator)
		err = GetOutcomeSetModel().BindByOutcome(ctx, operator, tx, &newVersion)
		if err != nil {
			log.Error(ctx, "LockLearningOutcome: BindByOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = da.GetOutcomeDA().CreateOutcome(ctx, operator, tx, &newVersion)
		if err != nil {
			log.Error(ctx, "LockLearningOutcome: CreateOutcome failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID),
				log.Any("outcome", newVersion))
			return err
		}
		attaches, err := da.GetAttachDA().SearchAttaches(ctx, tx, entity.AttachOutcomeTable, []string{outcome.ID}, entity.OutcomeType)
		if err != nil {
			log.Error(ctx, "LockLearningOutcome: CreateOutcome failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID),
				log.Any("outcome", newVersion))
			return err
		}
		for i := range attaches {
			attaches[i].MasterID = newVersion.ID
		}
		err = da.GetAttachDA().Replace(ctx, tx, entity.AttachOutcomeTable, nil, entity.OutcomeType, attaches)
		if err != nil {
			log.Error(ctx, "LockLearningOutcome: Replace failed",
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
	da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, operator, nil)
	return newVersion.ID, nil
}

func (ocm OutcomeModel) PublishLearningOutcome(ctx context.Context, operator *entity.Operator, outcomeID string, scope string) error {
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
		err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "PublishLearningOutcome: UpdateOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		return nil
	})
	if err == nil {
		da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, operator, nil)
	}
	return err
}

func (ocm OutcomeModel) BulkPubLearningOutcome(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string, scope string) error {
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
		total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &condition)
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
			err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, o)
			if err != nil {
				log.Error(ctx, "BulkPubLearningOutcome: UpdateOutcome failed",
					log.String("op", operator.UserID),
					log.Any("outcome", o))
				return err
			}
		}
		return nil
	})
	if err == nil {
		da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, operator, nil)
	}
	return err
}

func (ocm OutcomeModel) BulkDelLearningOutcome(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) error {
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, []external.PermissionName{
		external.DeleteMyUnpublishedLearningOutcome,
		external.DeleteOrgUnpublishedLearningOutcome,
		external.DeleteMyPendingLearningOutcome,
		external.DeleteOrgPendingLearningOutcome,
		external.DeletePublishedLearningOutcome,
	})
	if err != nil {
		log.Error(ctx, "BulkDelLearningOutcome:HasOrganizationPermissions failed", log.Any("op", operator), log.Err(err))
		return err
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		condition := da.OutcomeCondition{
			IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		}
		total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &condition)
		if err != nil {
			log.Error(ctx, "BulkDelLearningOutcome: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Int("total", total),
				log.Strings("outcome_ids", outcomeIDs))
			return err
		}

		if len(outcomes) > 0 && !allowDeleteOutcome(ctx, operator, perms, outcomes[0]) {
			log.Warn(ctx, "BulkDelLearningOutcome: no permission", log.Any("op", operator),
				log.Any("perms", perms), log.Any("outcome", outcomes[0]))
			return constant.ErrOperateNotAllowed
		}
		for _, o := range outcomes {
			err = ocm.deleteOutcome(ctx, operator, tx, o)
			if err != nil {
				log.Error(ctx, "BulkDelLearningOutcome: DeleteOutcome failed",
					log.String("op", operator.UserID),
					log.String("outcome_id", o.ID))
				return err
			}
		}

		err = da.GetAttachDA().Replace(ctx, tx, entity.AttachOutcomeTable, outcomeIDs, entity.OutcomeType, nil)
		if err != nil {
			log.Error(ctx, "BulkDelLearningOutcome: Replace failed",
				log.String("op", operator.UserID),
				log.Strings("outcome_id", outcomeIDs))
			return err
		}
		err = da.GetMilestoneDA().UnbindOutcomes(ctx, tx, outcomeIDs)
		if err != nil {
			log.Error(ctx, "BulkDelLearningOutcome: UnbindOutcomes failed",
				log.String("op", operator.UserID),
				log.Strings("outcome_id", outcomeIDs))
			return err
		}
		return nil
	})
	if err == nil {
		da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, operator, nil)
	}
	return err
}

func (ocm OutcomeModel) SearchPrivateOutcomes(ctx context.Context, user *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error) {
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, user, []external.PermissionName{
		external.ViewMyUnpublishedLearningOutcome,  // my draft & my rejected
		external.ViewOrgUnpublishedLearningOutcome, // org draft & org waiting for approved & org rejected
		external.ViewMyPendingLearningOutcome,      // my waiting for approved
	})
	if err != nil {
		log.Error(ctx, "SearchPrivateOutcomes: HasOrganizationPermissions failed",
			log.Any("op", user), log.Err(err))
		return 0, nil, constant.ErrInternalServer
	}
	condition.OrganizationID = user.OrgID
	if !allowSearchPrivate(ctx, user, perms, condition) {
		log.Warn(ctx, "SearchPrivateOutcomes: no permission",
			log.Any("op", user),
			log.Any("perms", perms),
			log.Any("cond", condition))
		return 0, nil, constant.ErrOperateNotAllowed
	}

	err = ocm.fillAuthorIDs(ctx, user, condition)
	if err != nil {
		log.Error(ctx, "SearchPrivateOutcomes: fillAuthorIDs failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	err = ocm.fillIDsBySetName(ctx, user, condition)
	if err != nil {
		log.Error(ctx, "SearchPrivateOutcomes: fillIDsBySetName failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}

	total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, user, tx, da.NewOutcomeCondition(condition))
	if err != nil {
		log.Error(ctx, "SearchPrivateOutcomes: SearchOutcome failed",
			log.Err(err),
			log.String("op", user.UserID),
			log.Any("outcome", ocm))
		return 0, nil, err
	}
	err = ocm.fillAttach(ctx, user, tx, outcomes)
	if err != nil {
		log.Error(ctx, "SearchPrivateOutcomes: fillAttach failed",
			log.Err(err),
			log.String("op", user.UserID))
		return 0, nil, err
	}
	return total, outcomes, nil
}

func (ocm OutcomeModel) SearchPendingOutcomes(ctx context.Context, user *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error) {
	if condition.PublishStatus != entity.OutcomeStatusPending {
		log.Warn(ctx, "SearchPendingOutcomes: SearchPendingOutcomes failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, ErrBadRequest
	}
	// as there is no level,orgID is the same as [user.OrgID]
	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, user, external.ViewOrgPendingLearningOutcome)
	if !hasPerm {
		log.Warn(ctx, "SearchPendingOutcomes: no permission",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, constant.ErrOperateNotAllowed
	}
	condition.PublishScope = user.OrgID
	err = ocm.fillAuthorIDs(ctx, user, condition)
	if err != nil {
		log.Error(ctx, "SearchPendingOutcomes: fillAuthorIDs failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	err = ocm.fillIDsBySetName(ctx, user, condition)
	if err != nil {
		log.Error(ctx, "SearchPendingOutcomes: fillIDsBySetName failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, user, tx, da.NewOutcomeCondition(condition))
	if err != nil {
		log.Error(ctx, "SearchPendingOutcomes: SearchOutcome failed",
			log.Err(err),
			log.String("op", user.UserID),
			log.Any("outcome", ocm))
		return 0, nil, err
	}
	err = ocm.fillAttach(ctx, user, tx, outcomes)
	if err != nil {
		log.Error(ctx, "SearchPendingOutcomes: fillAttach failed",
			log.Err(err),
			log.String("op", user.UserID))
		return 0, nil, err
	}
	return total, outcomes, nil
}

func (ocm OutcomeModel) ApproveLearningOutcome(ctx context.Context, operator *entity.Operator, outcomeID string) error {
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
		err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "ApproveLearningOutcome: UpdateOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = ocm.hideParent(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "ApproveLearningOutcome: hideParent failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = ocm.updateLatestToHead(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "ApproveLearningOutcome: updateLatestToHead failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		return nil
	})
	if err == nil {
		da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, operator, nil)
	}
	return err
}

func (ocm OutcomeModel) RejectLearningOutcome(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeID string, reason string) error {
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
		err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "RejectLearningOutcome: UpdateOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		return nil
	})
	if err == nil {
		da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, operator, nil)
	}
	return err
}

func (ocm OutcomeModel) BulkApproveLearningOutcome(ctx context.Context, operator *entity.Operator, outcomeIDs []string) error {
	for _, o := range outcomeIDs {
		locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeReview, o)
		if err != nil {
			log.Error(ctx, "RejectLearningOutcome: NewLock failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.String("outcome_id", o))
			return err
		}
		locker.Lock()
		defer locker.Unlock()
	}
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &da.OutcomeCondition{IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true}})
		if len(outcomes) == 0 {
			log.Warn(ctx, "BulkApproveLearningOutcome: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "BulkApproveLearningOutcome: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			return err
		}
		for _, outcome := range outcomes {
			err = outcome.SetStatus(ctx, entity.OutcomeStatusPublished)
			if err != nil {
				log.Error(ctx, "BulkApproveLearningOutcome: SetStatus failed",
					log.String("op", operator.UserID),
					log.Any("outcome", outcome))
				return ErrInvalidPublishStatus
			}
			if outcome.LatestID == "" {
				outcome.LatestID = outcome.ID
			}
			err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, outcome)
			if err != nil {
				log.Error(ctx, "BulkApproveLearningOutcome: UpdateOutcome failed",
					log.String("op", operator.UserID),
					log.Any("outcome", outcome))
				return err
			}
			err = ocm.hideParent(ctx, operator, tx, outcome)
			if err != nil {
				log.Error(ctx, "BulkApproveLearningOutcome: hideParent failed",
					log.String("op", operator.UserID),
					log.Any("outcome", outcome))
				return err
			}
			err = ocm.updateLatestToHead(ctx, operator, tx, outcome)
			if err != nil {
				log.Error(ctx, "BulkApproveLearningOutcome: updateLatestToHead failed",
					log.String("op", operator.UserID),
					log.Any("outcome", outcome))
				return err
			}
		}
		return nil
	})
	if err == nil {
		da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, operator, nil)
	}
	return err
}
func (ocm OutcomeModel) BulkRejectLearningOutcome(ctx context.Context, operator *entity.Operator, outcomeIDs []string, reason string) error {
	for _, o := range outcomeIDs {
		locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeReview, o)
		if err != nil {
			log.Error(ctx, "RejectLearningOutcome: NewLock failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.String("outcome_id", o))
			return err
		}
		locker.Lock()
		defer locker.Unlock()
	}
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &da.OutcomeCondition{IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true}})
		if len(outcomes) == 0 {
			log.Warn(ctx, "BulkRejectLearningOutcome: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "BulkRejectLearningOutcome: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			return err
		}
		for _, outcome := range outcomes {
			err = outcome.SetStatus(ctx, entity.OutcomeStatusRejected)
			outcome.RejectReason = reason
			if err != nil {
				log.Error(ctx, "BulkRejectLearningOutcome: SetStatus failed",
					log.String("op", operator.UserID),
					log.Any("outcome", outcome))
				return ErrInvalidPublishStatus
			}
			err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, outcome)
			if err != nil {
				log.Error(ctx, "BulkRejectLearningOutcome: UpdateOutcome failed",
					log.String("op", operator.UserID),
					log.Any("outcome", outcome))
				return err
			}
		}
		return nil
	})
	if err == nil {
		da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, operator, nil)
	}
	return err
}

func (ocm OutcomeModel) GetLearningOutcomesByIDs(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) ([]*entity.Outcome, error) {
	condition := da.OutcomeCondition{
		IDs:            dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		IncludeDeleted: true,
	}
	_, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &condition)
	if err != nil {
		log.Error(ctx, "GetLearningOutcomesByIDs: SearchOutcome failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.Any("outcome", ocm))
		return nil, err
	}
	err = ocm.fillAttach(ctx, operator, tx, outcomes)
	if err != nil {
		log.Error(ctx, "GetLearningOutcomesByIDs: fillAttach failed",
			log.Err(err),
			log.String("op", operator.UserID))
		return nil, err
	}
	return outcomes, nil
}

func (ocm OutcomeModel) GetLatestOutcomesByIDs(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (outcomes []*entity.Outcome, err error) {
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		cond1 := da.OutcomeCondition{
			IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		}
		total, otcs1, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &cond1)
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
		total, otcs2, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &cond2)
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
			err = ocm.fillAttach(ctx, operator, tx, outcomes)
			if err != nil {
				log.Error(ctx, "GetLatestOutcomesByIDs: fillAttach failed",
					log.Err(err),
					log.String("op", operator.UserID))
				return err
			}
		}
		return nil
	})
	return
}

func (ocm OutcomeModel) GetLatestOutcomesByIDsMapResult(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (latests map[string]*entity.Outcome, err error) {
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		cond1 := da.OutcomeCondition{
			IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		}
		total, otcs1, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &cond1)
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
			return constant.ErrRecordNotFound
		}
		cond2 := da.OutcomeCondition{}
		for _, o := range otcs1 {
			cond2.IDs.Strings = append(cond2.IDs.Strings, o.LatestID)
		}
		cond2.IDs.Valid = true
		total, otcs2, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &cond2)
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
			return constant.ErrRecordNotFound
		} else {
			err = ocm.fillAttach(ctx, operator, tx, otcs2)
			if err != nil {
				log.Error(ctx, "GetLatestOutcomesByIDs: fillAttach failed",
					log.Err(err),
					log.String("op", operator.UserID))
				return err
			}
			latests = make(map[string]*entity.Outcome, len(otcs1))
			for _, o := range otcs1 {
				for _, l := range otcs2 {
					if o.LatestID == l.ID {
						latests[l.ID] = l
						break
					}
				}
			}
		}
		return nil
	})
	return
}

func (ocm OutcomeModel) HasLockedOutcome(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (bool, error) {
	if len(outcomeIDs) == 0 {
		return false, nil
	}
	_, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &da.OutcomeCondition{
		IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
	})
	if err != nil {
		return false, err
	}
	for i := range outcomes {
		if outcomes[i].LockedBy != "" {
			return true, nil
		}
	}
	return false, nil
}

func (ocm OutcomeModel) fillAttach(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomes []*entity.Outcome) error {
	masterIDs := make([]string, len(outcomes))
	for i := range outcomes {
		masterIDs[i] = outcomes[i].ID
	}
	attaches, err := da.GetAttachDA().SearchAttaches(ctx, tx, entity.AttachOutcomeTable, masterIDs, entity.OutcomeType)
	if err != nil {
		log.Error(ctx, "GetLatestOutcomesByIDs: fillAttach failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.Any("outcomes", outcomes))
		return err
	}
	for i := range attaches {
		for j := range outcomes {
			outcomes[j].FillAttach([]*entity.Attach{attaches[i]})
			break
		}
	}
	return nil
}

func (ocm OutcomeModel) lockOutcome(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	// must in a transaction
	if outcome.PublishStatus != entity.OutcomeStatusPublished {
		err = ErrInvalidPublishStatus
		log.Warn(ctx, "lockOutcome: invalid lock status",
			log.Err(err),
			log.String("op", operator.UserID))
		return
	}
	if outcome.LockedBy != "" && outcome.LockedBy != constant.LockedByNoBody {
		err = NewErrContentAlreadyLocked(ctx, outcome.LockedBy, operator)
		log.Warn(ctx, "lockOutcome: invalid lock status",
			log.Err(err),
			log.String("op", operator.UserID))
		return
	}
	outcome.LockedBy = operator.UserID
	err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, outcome)
	if err != nil {
		log.Error(ctx, "lockOutcome: UpdateOutcome failed",
			log.Err(err),
			log.String("op", operator.UserID))
	}
	return
}

func (ocm OutcomeModel) unlockOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, otcid string) (err error) {
	// must in a transaction
	outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, otcid)
	if err != nil {
		log.Error(ctx, "unlockOutcome: GetOutcomeByID failed",
			log.String("outcome_id", otcid))
		return
	}
	if outcome.LockedBy != "" && outcome.LockedBy != constant.LockedByNoBody {
		outcome.LockedBy = constant.LockedByNoBody
		err = da.GetOutcomeDA().UpdateOutcome(ctx, op, tx, outcome)
		if err != nil {
			log.Error(ctx, "unlockOutcome: UpdateOutcome failed",
				log.String("outcome_id", otcid))
			return
		}
	}
	return
}

func (ocm OutcomeModel) deleteOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	// must in a transaction
	if outcome.LockedBy != "" && outcome.LockedBy != constant.LockedByNoBody {
		err = NewErrContentAlreadyLocked(ctx, outcome.LockedBy, op)
		log.Error(ctx, "deleteOutcome: invalid lock status",
			log.Err(err),
			log.Any("outcome", outcome))
		return
	}
	err = da.GetOutcomeDA().DeleteOutcome(ctx, op, tx, outcome)
	if err != nil {
		log.Error(ctx, "deleteOutcome: DeleteOutcome failed",
			log.Err(err),
			log.Any("outcome", outcome))
		return
	}
	if outcome.SourceID != "" && outcome.SourceID != outcome.ID {
		err = ocm.unlockOutcome(ctx, op, tx, outcome.SourceID)
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

func (ocm OutcomeModel) hideParent(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
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
	err = da.GetOutcomeDA().UpdateOutcome(ctx, op, tx, parent)
	if err != nil {
		log.Error(ctx, "hideParent: UpdateOutcome failed",
			log.Any("outcome", parent))
		return err
	}
	return nil
}

func (ocm OutcomeModel) updateLatestToHead(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	// must in a transaction
	if outcome.LatestID == outcome.ID {
		return nil
	}
	err = da.GetOutcomeDA().UpdateLatestHead(ctx, op, tx, outcome.LatestID, outcome.ID)
	return
}

func (ocm OutcomeModel) getAuthorNameByID(ctx context.Context, operator *entity.Operator, id string) (name string, err error) {
	provider := external.GetUserServiceProvider()
	user, err := provider.Get(ctx, operator, id)
	if err != nil {
		log.Error(ctx, "getAuthorNameByID: GetUserInfoByID failed",
			log.Err(err),
			log.String("user_id", id))
		return "", err
	}
	return user.Name, nil
}

func allowDeleteOutcome(ctx context.Context, operator *entity.Operator, perms map[external.PermissionName]bool, outcome *entity.Outcome) bool {
	if (outcome.PublishStatus == entity.OutcomeStatusDraft || outcome.PublishStatus == entity.OutcomeStatusRejected) &&
		perms[external.DeleteOrgUnpublishedLearningOutcome] {
		return true
	}

	if (outcome.PublishStatus == entity.OutcomeStatusDraft || outcome.PublishStatus == entity.OutcomeStatusRejected) &&
		(perms[external.DeleteMyUnpublishedLearningOutcome] && outcome.AuthorID == operator.UserID) {
		return true
	}

	if outcome.PublishStatus == entity.OutcomeStatusPending && perms[external.DeleteMyPendingLearningOutcome] {
		return true
	}

	if outcome.PublishStatus == entity.OutcomeStatusPending &&
		(perms[external.DeleteMyPendingLearningOutcome] && outcome.AuthorID == operator.UserID) {
		return true
	}

	if outcome.PublishStatus == entity.OutcomeStatusPublished && perms[external.DeletePublishedLearningOutcome] {
		return true
	}

	return false
}

func allowEditOutcome(ctx context.Context, operator *entity.Operator, perms map[external.PermissionName]bool, outcome *entity.Outcome) bool {
	if perms[external.EditOrgUnpublishedLearningOutcome] && outcome.PublishStatus != entity.OutcomeStatusPublished {
		return true
	}

	if (perms[external.EditMyUnpublishedLearningOutcome] && outcome.AuthorID == operator.UserID) &&
		(outcome.PublishStatus != entity.OutcomeStatusPublished && outcome.PublishStatus != entity.OutcomeStatusPending) {
		return true
	}

	return false
}

func allowSearchPrivate(ctx context.Context, operator *entity.Operator, perms map[external.PermissionName]bool, cond *entity.OutcomeCondition) bool {
	if (cond.PublishStatus == entity.OutcomeStatusDraft ||
		cond.PublishStatus == entity.OutcomeStatusRejected ||
		cond.PublishStatus == entity.OutcomeStatusPending) &&
		perms[external.ViewOrgUnpublishedLearningOutcome] {
		if cond.AuthorName == constant.Self {
			cond.AuthorID = operator.UserID
			cond.AuthorName = ""
		}
		return true
	}

	if (cond.PublishStatus == entity.OutcomeStatusDraft ||
		cond.PublishStatus == entity.OutcomeStatusRejected) &&
		perms[external.ViewMyUnpublishedLearningOutcome] {
		cond.AuthorID = operator.UserID
		return true
	}

	if cond.PublishStatus == entity.OutcomeStatusPending && perms[external.ViewMyPendingLearningOutcome] {
		cond.AuthorID = operator.UserID
		return true
	}

	return false
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
