package model

import (
	"context"
	"database/sql"
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
)

type IOutcomeModel interface {
	Create(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) error
	Get(ctx context.Context, operator *entity.Operator, outcomeID string) (*entity.Outcome, error)
	Update(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) error
	Delete(ctx context.Context, operator *entity.Operator, outcomeID string) error
	Search(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error)

	Lock(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeID string) (string, error)

	Publish(ctx context.Context, operator *entity.Operator, outcomeID string, scope string) error
	BulkPublish(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string, scope string) error
	BulkDelete(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) error

	SearchPrivate(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error)
	SearchPending(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error)

	GetByIDs(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) ([]*entity.Outcome, error)
	GetLatestByIDs(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) ([]*entity.Outcome, error)

	Approve(ctx context.Context, operator *entity.Operator, outcomeID string) error
	Reject(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeID string, reason string) error

	BulkApprove(ctx context.Context, operator *entity.Operator, outcomeIDs []string) error
	BulkReject(ctx context.Context, operator *entity.Operator, outcomeIDs []string, reason string) error

	GetLatestByIDsMapResult(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (map[string]*entity.Outcome, error)

	HasLocked(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (bool, error)
	GetLatestByAncestors(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, ancestoryIDs []string) ([]*entity.Outcome, error)
	GenerateShortcode(ctx context.Context, op *entity.Operator) (string, error)

	ShortcodeProvider
}

type OutcomeModel struct {
}

func (ocm OutcomeModel) GenerateShortcode(ctx context.Context, op *entity.Operator) (string, error) {
	return GetShortcodeModel().Generate(ctx, op, ocm)
}
func (ocm OutcomeModel) Current(ctx context.Context, op *entity.Operator) (int, error) {
	return da.GetShortcodeRedis(ctx).Get(ctx, op, string(entity.KindOutcome))
}

func (ocm OutcomeModel) Intersect(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, shortcodes []string) (map[string]bool, error) {
	_, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, op, tx, &da.OutcomeCondition{
		Shortcodes:     dbo.NullStrings{Strings: shortcodes, Valid: true},
		OrganizationID: sql.NullString{String: op.OrgID, Valid: true},
		OrderBy:        da.OrderByShortcode,
	})
	if err != nil {
		log.Debug(ctx, "Intersect: Search failed",
			log.Any("op", op),
			log.Strings("shortcode", shortcodes))
		return nil, err
	}
	mapShortcode := make(map[string]bool)
	for i := range outcomes {
		mapShortcode[outcomes[i].Shortcode] = true
	}
	return mapShortcode, nil
}

func (ocm OutcomeModel) ShortcodeLength() int {
	return constant.ShortcodeShowLength
}

func (ocm OutcomeModel) IsShortcodeCached(ctx context.Context, op *entity.Operator, shortcode string) (bool, error) {
	exists, err := da.GetShortcodeRedis(ctx).IsCached(ctx, op, string(entity.KindOutcome), shortcode)
	if err != nil {
		log.Debug(ctx, "IsCached: redis access failed",
			log.Any("op", op),
			log.String("shortcode", shortcode))
		return false, err
	}
	return exists, nil
}

func (ocm OutcomeModel) IsShortcodeExists(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, ancestor string, shortcode string) (bool, error) {
	_, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, op, tx, &da.OutcomeCondition{
		OrganizationID: sql.NullString{String: op.OrgID, Valid: true},
		Shortcodes:     dbo.NullStrings{Strings: []string{shortcode}, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "IsShortcodeExists: Search failed",
			log.String("org", op.OrgID),
			log.String("shortcode", shortcode))
		return false, err
	}
	for i := range outcomes {
		if ancestor != outcomes[i].AncestorID {
			return true, nil
		}
	}
	return false, nil
}

func (ocm OutcomeModel) RemoveShortcode(ctx context.Context, op *entity.Operator, shortcode string) error {
	err := da.GetShortcodeRedis(ctx).Remove(ctx, op, string(entity.KindOutcome), shortcode)
	if err != nil {
		log.Error(ctx, "RemoveShortcode: redis access failed",
			log.Err(err),
			log.Any("op", op),
			log.String("shortcode", shortcode))
		return err
	}
	return nil
}

func (ocm OutcomeModel) Cache(ctx context.Context, op *entity.Operator, cursor int, shortcode string) error {
	err := da.GetShortcodeRedis(ctx).Cache(ctx, op, string(entity.KindOutcome), cursor, shortcode)
	if err != nil {
		log.Debug(ctx, "Cache: redis access failed",
			log.Any("op", op),
			log.Int("cursor", cursor),
			log.String("shortcode", shortcode))
		return err
	}
	return nil
}

func (ocm OutcomeModel) Create(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) (err error) {
	// outcome get value from api lay, this lay add some information
	outcome.AuthorName, err = ocm.getAuthorNameByID(ctx, operator, operator.UserID)
	if err != nil {
		log.Error(ctx, "Create: getAuthorNameByID failed",
			log.String("op", outcome.AuthorID),
			log.Any("outcome", outcome))
		return
	}

	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixShortcodeMute, entity.KindOutcome, operator.OrgID)
	if err != nil {
		log.Error(ctx, "Create: NewLock failed",
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
		exists, err := ocm.IsShortcodeExists(ctx, operator, tx, outcome.AncestorID, outcome.Shortcode)
		if err != nil {
			log.Error(ctx, "Create: IsShortcodeExistInDBWithOtherAncestor failed",
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
			log.Error(ctx, "Create: BindByOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = da.GetOutcomeDA().CreateOutcome(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "Create: CreateOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = da.GetOutcomeRelationDA().InsertTx(ctx, tx, ocm.CollectRelation(outcome))
		if err != nil {
			log.Error(ctx, "Create: InsertTx failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		return nil
	})
	ocm.RemoveShortcode(ctx, operator, outcome.Shortcode)
	if err != nil {
		return err
	}
	da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, operator, nil)
	return
}

func (ocm OutcomeModel) Get(ctx context.Context, operator *entity.Operator, outcomeID string) (outcome *entity.Outcome, err error) {

	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		var err error
		outcome, err = da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
		if err == dbo.ErrRecordNotFound {
			log.Warn(ctx, "Get: not found",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "Get: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		relations, err := da.GetOutcomeRelationDA().SearchTx(ctx, tx, &da.RelationCondition{
			MasterIDs:  dbo.NullStrings{Strings: []string{outcomeID}, Valid: true},
			MasterType: sql.NullString{String: string(entity.OutcomeType), Valid: true},
		})
		if err != nil {
			log.Error(ctx, "Get: SearchTx failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		ocm.FillRelation(outcome, relations)

		milestoneOutcomes, err := da.GetMilestoneOutcomeDA().SearchTx(ctx, tx, &da.MilestoneOutcomeCondition{
			OutcomeAncestor: sql.NullString{String: outcome.AncestorID, Valid: true},
		})
		if err != nil {
			log.Error(ctx, "Get: SearchTx failed",
				log.String("op", operator.UserID),
				log.String("outcome", outcomeID))
			return err
		}
		milestoneIDs := make([]string, len(milestoneOutcomes))
		for i := range milestoneOutcomes {
			milestoneIDs[i] = milestoneOutcomes[i].MilestoneID
		}
		if len(milestoneIDs) == 0 {
			log.Warn(ctx, "Get: haven't bind milestone",
				log.String("op", operator.UserID),
				log.String("outcome", outcomeID))
			return nil
		}
		_, milestones, err := da.GetMilestoneDA().Search(ctx, tx, &da.MilestoneCondition{
			IDs: dbo.NullStrings{Strings: milestoneIDs, Valid: true},
		})
		if err != nil {
			log.Error(ctx, "Get: Search failed",
				log.String("op", operator.UserID),
				log.String("outcome", outcomeID))
			return err
		}
		outcome.Milestones = milestones
		return nil
	})
	return
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

func (ocm OutcomeModel) Update(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) error {
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, []external.PermissionName{
		external.EditMyUnpublishedLearningOutcome,
		external.EditOrgUnpublishedLearningOutcome,
	})
	if err != nil {
		log.Error(ctx, "Update:HasOrganizationPermissions failed", log.Any("op", operator), log.Err(err))
		return err
	}
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixShortcodeMute, entity.KindOutcome, operator.OrgID)
	if err != nil {
		log.Error(ctx, "Update: NewLock failed",
			log.Err(err),
			log.Any("op", operator),
			log.Any("outcome", outcome))
		return err
	}
	locker.Lock()
	defer locker.Unlock()
	exists, err := ocm.IsShortcodeCached(ctx, operator, outcome.Shortcode)
	if err != nil {
		log.Error(ctx, "Update: IsCached failed",
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
			log.Error(ctx, "Update: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.Any("data", data))
			return err
		}
		if !allowEditOutcome(ctx, operator, perms, data) {
			log.Warn(ctx, "Update: no permission",
				log.Any("op", operator),
				log.Any("perms", perms),
				log.Any("data", data))
			return constant.ErrOperateNotAllowed
		}
		if data.PublishStatus != entity.OutcomeStatusDraft && data.PublishStatus != entity.OutcomeStatusRejected {
			log.Error(ctx, "Update: publish status not allowed edit",
				log.String("op", operator.UserID),
				log.Any("data", data))
			return ErrInvalidPublishStatus
		}

		if data.Shortcode != outcome.Shortcode {
			exists, err := ocm.IsShortcodeExists(ctx, operator, tx, data.AncestorID, outcome.Shortcode)
			if err != nil {
				log.Error(ctx, "Update: IsShortcodeExists failed",
					log.Err(err),
					log.Any("op", operator),
					log.Any("outcome", outcome))
				return err
			}
			if exists {
				return constant.ErrConflict
			}
		}

		ocm.UpdateOutcome(outcome, data)
		err = ocm.updateOutcomeSet(ctx, operator, tx, outcome.ID, outcome.Sets)
		if err != nil {
			log.Error(ctx, "Update: updateOutcomeSet failed",
				log.String("op", operator.UserID),
				log.Any("data", outcome))
			return err
		}
		// because of cache, follow statements need be at last
		err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, data)
		if err != nil {
			log.Error(ctx, "Update: UpdateOutcome failed",
				log.String("op", operator.UserID),
				log.Any("data", outcome))
			return err
		}
		err = da.GetOutcomeRelationDA().DeleteTx(ctx, tx, []string{outcome.ID})
		if err != nil {
			log.Error(ctx, "Update: DeleteTx failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = da.GetOutcomeRelationDA().InsertTx(ctx, tx, ocm.CollectRelation(outcome))
		if err != nil {
			log.Error(ctx, "Update: InsertTx failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		return nil
	})
	return err
}

func (ocm OutcomeModel) Delete(ctx context.Context, operator *entity.Operator, outcomeID string) error {
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, []external.PermissionName{
		external.DeleteMyUnpublishedLearningOutcome,
		external.DeleteOrgUnpublishedLearningOutcome,
		external.DeleteMyPendingLearningOutcome,
		external.DeleteOrgPendingLearningOutcome,
		external.DeletePublishedLearningOutcome,
	})
	if err != nil {
		log.Error(ctx, "Delete:HasOrganizationPermissions failed", log.Any("op", operator), log.Err(err))
		return err
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
		if err != nil && err != dbo.ErrRecordNotFound {
			log.Error(ctx, "Delete: no permission",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		if !allowDeleteOutcome(ctx, operator, perms, outcome) {
			log.Warn(ctx, "Delete: no permission", log.Any("op", operator),
				log.Any("perms", perms), log.Any("outcome", outcome))
			return constant.ErrOperateNotAllowed
		}
		err = ocm.deleteOutcome(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "Delete: deleteOutcome failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = da.GetOutcomeSetDA().DeleteBoundOutcomeSet(ctx, tx, outcome.ID)
		if err != nil {
			log.Error(ctx, "Delete: deleteOutcome failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = da.GetOutcomeRelationDA().DeleteTx(ctx, tx, []string{outcome.ID})
		if err != nil {
			log.Error(ctx, "Delete: DeleteTx failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = da.GetMilestoneDA().UnbindOutcomes(ctx, tx, []string{outcome.AncestorID})
		if err != nil {
			log.Error(ctx, "Delete: UnbindOutcomes failed",
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

func (ocm OutcomeModel) search(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error) {
	var err error
	total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, op, tx, da.NewOutcomeCondition(condition))
	if err != nil {
		log.Error(ctx, "Search: SearchOutcome failed",
			log.String("op", op.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	if len(outcomes) == 0 {
		log.Info(ctx, "Search: not found",
			log.Any("op", op),
			log.Any("cond", condition))
		return total, outcomes, nil
	}
	outcomeIDs := make([]string, len(outcomes))
	for i := range outcomes {
		outcomeIDs[i] = outcomes[i].ID
	}

	relations, err := da.GetOutcomeRelationDA().SearchTx(ctx, tx, &da.RelationCondition{
		MasterIDs:  dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		MasterType: sql.NullString{String: string(entity.OutcomeType), Valid: true},
	})
	if err != nil {
		log.Error(ctx, "Search: SearchTx failed",
			log.Err(err),
			log.Any("op", op),
			log.Strings("outcome", outcomeIDs))
		return 0, nil, err
	}
	outcomeRelations := make(map[string][]*entity.Relation)
	for i := range relations {
		outcomeRelations[relations[i].MasterID] = append(outcomeRelations[relations[i].MasterID], relations[i])
	}
	for _, outcome := range outcomes {
		ocm.FillRelation(outcome, outcomeRelations[outcome.ID])
	}
	return total, outcomes, nil
}

func (ocm OutcomeModel) Search(ctx context.Context, user *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error) {
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
		log.Error(ctx, "Search: fillAuthorIDs failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	err = ocm.fillIDsBySetName(ctx, user, condition)
	if err != nil {
		log.Error(ctx, "Search: fillIDsBySetName failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}

	var total int
	var outcomes []*entity.Outcome
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		var err error
		total, outcomes, err = ocm.search(ctx, user, tx, condition)
		if err != nil {
			log.Error(ctx, "Search: search failed",
				log.String("op", user.UserID),
				log.Any("condition", condition))
			return err
		}
		return nil
	})
	return total, outcomes, err
}

func (ocm OutcomeModel) Lock(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeID string) (string, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeLock)
	if err != nil {
		log.Error(ctx, "Lock: NewLock failed",
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
			log.Error(ctx, "Lock: GetOutcomeByID failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}

		if outcome.LockedBy == operator.UserID {
			copyValue, err := da.GetOutcomeDA().GetOutcomeBySourceID(ctx, operator, tx, outcomeID)
			if err != nil {
				log.Error(ctx, "Lock: GetOutcomeBySourceID failed",
					log.String("op", operator.UserID),
					log.String("outcome_id", outcomeID))
				return err
			}
			if copyValue.PublishStatus == entity.OutcomeStatusDraft {
				newVersion = *copyValue
				return nil
			}
			log.Error(ctx, "Lock: copyValue status not draft",
				log.String("op", operator.UserID),
				log.Any("copy", copyValue))
			return NewErrContentAlreadyLocked(ctx, outcome.LockedBy, operator)
		}

		err = ocm.lockOutcome(ctx, operator, tx, outcome)
		if err != nil {
			return err
		}
		newVersion = ocm.Clone(operator, outcome)
		err = GetOutcomeSetModel().BindByOutcome(ctx, operator, tx, &newVersion)
		if err != nil {
			log.Error(ctx, "Lock: BindByOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = da.GetOutcomeDA().CreateOutcome(ctx, operator, tx, &newVersion)
		if err != nil {
			log.Error(ctx, "Lock: CreateOutcome failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID),
				log.Any("outcome", newVersion))
			return err
		}
		relations, err := da.GetOutcomeRelationDA().SearchTx(ctx, tx, &da.RelationCondition{
			MasterIDs:  dbo.NullStrings{Strings: []string{outcome.ID}, Valid: true},
			MasterType: sql.NullString{String: string(entity.OutcomeType), Valid: true},
		})
		if err != nil {
			log.Error(ctx, "Lock: SearchTx failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID),
				log.Any("outcome", newVersion))
			return err
		}
		for i := range relations {
			relations[i].MasterID = newVersion.ID
		}
		err = da.GetOutcomeRelationDA().InsertTx(ctx, tx, relations)
		if err != nil {
			log.Error(ctx, "Lock: InsertTx failed",
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

func (ocm OutcomeModel) Publish(ctx context.Context, operator *entity.Operator, outcomeID string, scope string) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, tx, outcomeID)
		if err == dbo.ErrRecordNotFound {
			err = ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "Publish: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = ocm.SetStatus(ctx, outcome, entity.OutcomeStatusPending)
		if err != nil {
			log.Error(ctx, "Publish: SetStatus failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return ErrInvalidContentStatusToPublish
		}
		if outcome.PublishScope != "" && outcome.PublishScope != scope {
			log.Error(ctx, "Publish: scope mismatch",
				log.String("op", operator.UserID),
				log.String("scope", scope),
				log.Any("outcome", outcome))
			return ErrInvalidContentStatusToPublish
		}
		outcome.PublishScope = scope
		outcome.UpdateAt = time.Now().Unix()
		err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "Publish: UpdateOutcome failed",
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

func (ocm OutcomeModel) BulkPublish(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string, scope string) error {
	if scope == "" {
		//scopeID, _, err := ocm.getRootOrganizationByAuthorID(ctx, operator.UserID)
		//if err != nil {
		//	log.Error(ctx, "Publish: getRootOrganizationByAuthorID failed",
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
			log.Error(ctx, "BulkPublish: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Int("total", total),
				log.Strings("outcome_id", outcomeIDs))
			return err
		}
		if total == 0 {
			log.Warn(ctx, "BulkPublish: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Int("total", total),
				log.Strings("outcome_id", outcomeIDs))
			return ErrResourceNotFound
		}
		for _, o := range outcomes {
			err = ocm.SetStatus(ctx, o, entity.OutcomeStatusPublished)
			if err != nil {
				log.Error(ctx, "BulkPublish: SetStatus failed",
					log.String("op", operator.UserID),
					log.Any("outcome", o))
				return ErrInvalidContentStatusToPublish
			}
			if o.PublishScope != "" && o.PublishScope != scope {
				log.Error(ctx, "Publish: scope mismatch",
					log.String("op", operator.UserID),
					log.String("scope", scope),
					log.Any("outcome", o))
				return ErrInvalidContentStatusToPublish
			}
			err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, o)
			if err != nil {
				log.Error(ctx, "BulkPublish: UpdateOutcome failed",
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

func (ocm OutcomeModel) BulkDelete(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) error {
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, []external.PermissionName{
		external.DeleteMyUnpublishedLearningOutcome,
		external.DeleteOrgUnpublishedLearningOutcome,
		external.DeleteMyPendingLearningOutcome,
		external.DeleteOrgPendingLearningOutcome,
		external.DeletePublishedLearningOutcome,
	})
	if err != nil {
		log.Error(ctx, "BulkDelete:HasOrganizationPermissions failed", log.Any("op", operator), log.Err(err))
		return err
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		condition := da.OutcomeCondition{
			IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		}
		total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &condition)
		if err != nil {
			log.Error(ctx, "BulkDelete: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Int("total", total),
				log.Strings("outcome_ids", outcomeIDs))
			return err
		}

		if len(outcomes) > 0 && !allowDeleteOutcome(ctx, operator, perms, outcomes[0]) {
			log.Warn(ctx, "BulkDelete: no permission", log.Any("op", operator),
				log.Any("perms", perms), log.Any("outcome", outcomes[0]))
			return constant.ErrOperateNotAllowed
		}
		ancestorIDs := make([]string, len(outcomes))
		for i, o := range outcomes {
			err = ocm.deleteOutcome(ctx, operator, tx, o)
			if err != nil {
				log.Error(ctx, "BulkDelete: DeleteOutcome failed",
					log.String("op", operator.UserID),
					log.String("outcome_id", o.ID))
				return err
			}
			ancestorIDs[i] = o.AncestorID
		}

		err = da.GetOutcomeRelationDA().DeleteTx(ctx, tx, outcomeIDs)
		if err != nil {
			log.Error(ctx, "BulkDelete: DeleteTx failed",
				log.String("op", operator.UserID),
				log.Strings("outcome_id", outcomeIDs))
			return err
		}
		err = da.GetMilestoneDA().UnbindOutcomes(ctx, tx, ancestorIDs)
		if err != nil {
			log.Error(ctx, "BulkDelete: UnbindOutcomes failed",
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

func (ocm OutcomeModel) SearchPrivate(ctx context.Context, user *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error) {
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, user, []external.PermissionName{
		external.ViewMyUnpublishedLearningOutcome,  // my draft & my rejected
		external.ViewOrgUnpublishedLearningOutcome, // org draft & org waiting for approved & org rejected
		external.ViewMyPendingLearningOutcome,      // my waiting for approved
	})
	if err != nil {
		log.Error(ctx, "SearchPrivate: HasOrganizationPermissions failed",
			log.Any("op", user), log.Err(err))
		return 0, nil, constant.ErrInternalServer
	}
	condition.OrganizationID = user.OrgID
	if !allowSearchPrivate(ctx, user, perms, condition) {
		log.Warn(ctx, "SearchPrivate: no permission",
			log.Any("op", user),
			log.Any("perms", perms),
			log.Any("cond", condition))
		return 0, nil, constant.ErrOperateNotAllowed
	}

	err = ocm.fillAuthorIDs(ctx, user, condition)
	if err != nil {
		log.Error(ctx, "SearchPrivate: fillAuthorIDs failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	err = ocm.fillIDsBySetName(ctx, user, condition)
	if err != nil {
		log.Error(ctx, "SearchPrivate: fillIDsBySetName failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}

	var total int
	var outcomes []*entity.Outcome
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		var err error
		total, outcomes, err = ocm.search(ctx, user, tx, condition)
		if err != nil {
			log.Error(ctx, "Search: search failed",
				log.String("op", user.UserID),
				log.Any("condition", condition))
			return err
		}
		return nil
	})
	return total, outcomes, err
}

func (ocm OutcomeModel) SearchPending(ctx context.Context, user *entity.Operator, tx *dbo.DBContext, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error) {
	if condition.PublishStatus != entity.OutcomeStatusPending {
		log.Warn(ctx, "SearchPending: SearchPending failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, ErrBadRequest
	}
	// as there is no level,orgID is the same as [user.OrgID]
	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, user, external.ViewOrgPendingLearningOutcome)
	if !hasPerm {
		log.Warn(ctx, "SearchPending: no permission",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, constant.ErrOperateNotAllowed
	}
	condition.PublishScope = user.OrgID
	err = ocm.fillAuthorIDs(ctx, user, condition)
	if err != nil {
		log.Error(ctx, "SearchPending: fillAuthorIDs failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	err = ocm.fillIDsBySetName(ctx, user, condition)
	if err != nil {
		log.Error(ctx, "SearchPending: fillIDsBySetName failed",
			log.String("op", user.UserID),
			log.Any("condition", condition))
		return 0, nil, err
	}
	var total int
	var outcomes []*entity.Outcome
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		var err error
		total, outcomes, err = ocm.search(ctx, user, tx, condition)
		if err != nil {
			log.Error(ctx, "Search: search failed",
				log.String("op", user.UserID),
				log.Any("condition", condition))
			return err
		}
		return nil
	})
	return total, outcomes, err
}

func (ocm OutcomeModel) Approve(ctx context.Context, operator *entity.Operator, outcomeID string) error {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeReview)
	if err != nil {
		log.Error(ctx, "Approve: NewLock failed",
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
			log.Warn(ctx, "Approve: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "Approve: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = ocm.SetStatus(ctx, outcome, entity.OutcomeStatusPublished)
		if err != nil {
			log.Error(ctx, "Approve: SetStatus failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return ErrInvalidPublishStatus
		}
		if outcome.LatestID == "" {
			outcome.LatestID = outcome.ID
		}
		err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "Approve: UpdateOutcome failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = ocm.hideParent(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "Approve: hideParent failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return err
		}
		err = ocm.updateLatestToHead(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "Approve: updateLatestToHead failed",
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

func (ocm OutcomeModel) Reject(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeID string, reason string) error {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeReview)
	if err != nil {
		log.Error(ctx, "Reject: NewLock failed",
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
			log.Warn(ctx, "Reject: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "Reject: GetOutcomeByID failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID))
			return err
		}
		err = ocm.SetStatus(ctx, outcome, entity.OutcomeStatusRejected)
		outcome.RejectReason = reason
		if err != nil {
			log.Error(ctx, "Reject: SetStatus failed",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return ErrInvalidPublishStatus
		}
		err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, outcome)
		if err != nil {
			log.Error(ctx, "Reject: UpdateOutcome failed",
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

func (ocm OutcomeModel) BulkApprove(ctx context.Context, operator *entity.Operator, outcomeIDs []string) error {
	for _, o := range outcomeIDs {
		locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeReview, o)
		if err != nil {
			log.Error(ctx, "Reject: NewLock failed",
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
			log.Warn(ctx, "BulkApprove: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "BulkApprove: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			return err
		}
		for _, outcome := range outcomes {
			err = ocm.SetStatus(ctx, outcome, entity.OutcomeStatusPublished)
			if err != nil {
				log.Error(ctx, "BulkApprove: SetStatus failed",
					log.String("op", operator.UserID),
					log.Any("outcome", outcome))
				return ErrInvalidPublishStatus
			}
			if outcome.LatestID == "" {
				outcome.LatestID = outcome.ID
			}
			err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, outcome)
			if err != nil {
				log.Error(ctx, "BulkApprove: UpdateOutcome failed",
					log.String("op", operator.UserID),
					log.Any("outcome", outcome))
				return err
			}
			err = ocm.hideParent(ctx, operator, tx, outcome)
			if err != nil {
				log.Error(ctx, "BulkApprove: hideParent failed",
					log.String("op", operator.UserID),
					log.Any("outcome", outcome))
				return err
			}
			err = ocm.updateLatestToHead(ctx, operator, tx, outcome)
			if err != nil {
				log.Error(ctx, "BulkApprove: updateLatestToHead failed",
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
func (ocm OutcomeModel) BulkReject(ctx context.Context, operator *entity.Operator, outcomeIDs []string, reason string) error {
	for _, o := range outcomeIDs {
		locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeReview, o)
		if err != nil {
			log.Error(ctx, "Reject: NewLock failed",
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
			log.Warn(ctx, "BulkReject: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "BulkReject: SearchOutcome failed",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			return err
		}
		for _, outcome := range outcomes {
			err = ocm.SetStatus(ctx, outcome, entity.OutcomeStatusRejected)
			outcome.RejectReason = reason
			if err != nil {
				log.Error(ctx, "BulkReject: SetStatus failed",
					log.String("op", operator.UserID),
					log.Any("outcome", outcome))
				return ErrInvalidPublishStatus
			}
			err = da.GetOutcomeDA().UpdateOutcome(ctx, operator, tx, outcome)
			if err != nil {
				log.Error(ctx, "BulkReject: UpdateOutcome failed",
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

func (ocm OutcomeModel) GetByIDs(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) ([]*entity.Outcome, error) {
	condition := da.OutcomeCondition{
		IDs:            dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		IncludeDeleted: true,
	}
	_, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &condition)
	if err != nil {
		log.Error(ctx, "GetByIDs: SearchOutcome failed",
			log.Err(err),
			log.String("op", operator.UserID),
			log.Any("outcome", ocm))
		return nil, err
	}
	err = ocm.fillRelation(ctx, operator, tx, outcomes)
	if err != nil {
		log.Error(ctx, "GetByIDs: fillRelation failed",
			log.Err(err),
			log.String("op", operator.UserID))
		return nil, err
	}
	return outcomes, nil
}

func (ocm OutcomeModel) GetLatestByIDs(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (outcomes []*entity.Outcome, err error) {
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		cond1 := da.OutcomeCondition{
			IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		}
		total, otcs1, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &cond1)
		if err != nil {
			log.Error(ctx, "GetLatestByIDs: SearchOutcome failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			return err
		}
		if total == 0 {
			log.Debug(ctx, "GetLatestByIDs: SearchOutcome return empty",
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
			log.Error(ctx, "GetLatestByIDs: SearchOutcome failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", cond2.IDs.Strings))
			return err
		}
		if total == 0 {
			log.Debug(ctx, "GetLatestByIDs: SearchOutcome return empty",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", cond2.IDs.Strings))
			outcomes = []*entity.Outcome{}
		} else {
			outcomes = otcs2
			err = ocm.fillRelation(ctx, operator, tx, outcomes)
			if err != nil {
				log.Error(ctx, "GetLatestByIDs: fillRelation failed",
					log.Err(err),
					log.String("op", operator.UserID))
				return err
			}
		}
		return nil
	})
	return
}

func (ocm OutcomeModel) GetLatestByIDsMapResult(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (latests map[string]*entity.Outcome, err error) {
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		cond1 := da.OutcomeCondition{
			IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
		}
		total, otcs1, err := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &cond1)
		if err != nil {
			log.Error(ctx, "GetLatestByIDs: SearchOutcome failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", outcomeIDs))
			return err
		}
		if total == 0 {
			log.Debug(ctx, "GetLatestByIDs: SearchOutcome return empty",
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
			log.Error(ctx, "GetLatestByIDs: SearchOutcome failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", cond2.IDs.Strings))
			return err
		}
		if total == 0 {
			log.Debug(ctx, "GetLatestByIDs: SearchOutcome return empty",
				log.String("op", operator.UserID),
				log.Strings("outcome_ids", cond2.IDs.Strings))
			return constant.ErrRecordNotFound
		}
		err = ocm.fillRelation(ctx, operator, tx, otcs2)
		if err != nil {
			log.Error(ctx, "GetLatestByIDs: fillRelation failed",
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
		return nil
	})
	return
}

func (ocm OutcomeModel) HasLocked(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (bool, error) {
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

func (ocm OutcomeModel) GetLatestByAncestors(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, ancestorIDs []string) (outcomes []*entity.Outcome, err error) {
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, outcomes, err = da.GetOutcomeDA().SearchOutcome(ctx, op, tx, &da.OutcomeCondition{
			AncestorIDs:   dbo.NullStrings{Strings: ancestorIDs, Valid: true},
			PublishStatus: dbo.NullStrings{Strings: []string{entity.OutcomeStatusPublished}, Valid: true},
		})
		if err != nil {
			log.Error(ctx, "GetLatestByAncestors: SearchOutcome failed",
				log.Err(err),
				log.String("op", op.UserID),
				log.Strings("ancestor", ancestorIDs))
			return err
		}
		err = ocm.fillRelation(ctx, op, tx, outcomes)
		if err != nil {
			log.Error(ctx, "GetLatestByAncestors: fillRelation failed",
				log.Err(err),
				log.String("op", op.UserID),
				log.Strings("ancestor", ancestorIDs))
			return err
		}
		return nil
	})
	return
}

func (ocm OutcomeModel) fillRelation(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomes []*entity.Outcome) error {
	if len(outcomes) > 0 {
		masterIDs := make([]string, len(outcomes))
		for i := range outcomes {
			masterIDs[i] = outcomes[i].ID
		}
		relations, err := da.GetOutcomeRelationDA().SearchTx(ctx, tx, &da.RelationCondition{
			MasterIDs:  dbo.NullStrings{Strings: masterIDs, Valid: true},
			MasterType: sql.NullString{String: string(entity.OutcomeType), Valid: true},
		})
		if err != nil {
			log.Error(ctx, "fillRelation: Search failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.Any("outcomes", outcomes))
			return err
		}
		for i := range relations {
			for j := range outcomes {
				ocm.FillRelation(outcomes[j], []*entity.Relation{relations[i]})
				break
			}
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
	err = ocm.SetStatus(ctx, parent, entity.OutcomeStatusHidden)
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

func (ocm OutcomeModel) CollectRelation(oc *entity.Outcome) []*entity.Relation {
	relations := make([]*entity.Relation, 0, len(oc.Programs)+len(oc.Subjects)+len(oc.Categories)+len(oc.Subcategories)+len(oc.Grades)+len(oc.Ages))
	for i := range oc.Programs {
		relation := entity.Relation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Programs[i],
			RelationType: entity.ProgramType,
		}
		relations = append(relations, &relation)
	}

	for i := range oc.Subjects {
		relation := entity.Relation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Subjects[i],
			RelationType: entity.SubjectType,
		}
		relations = append(relations, &relation)
	}

	for i := range oc.Categories {
		relation := entity.Relation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Categories[i],
			RelationType: entity.CategoryType,
		}
		relations = append(relations, &relation)
	}

	for i := range oc.Subcategories {
		relation := entity.Relation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Subcategories[i],
			RelationType: entity.SubcategoryType,
		}
		relations = append(relations, &relation)
	}

	for i := range oc.Grades {
		relation := entity.Relation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Grades[i],
			RelationType: entity.GradeType,
		}
		relations = append(relations, &relation)
	}

	for i := range oc.Ages {
		relation := entity.Relation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Ages[i],
			RelationType: entity.AgeType,
		}
		relations = append(relations, &relation)
	}
	return relations
}

func (ocm OutcomeModel) FillRelation(oc *entity.Outcome, relations []*entity.Relation) {
	if len(relations) == 0 {
		program := strings.TrimSpace(oc.Program)
		if program != "" {
			oc.Programs = strings.Split(program, entity.JoinComma)
		}
		subject := strings.TrimSpace(oc.Subject)
		if subject != "" {
			oc.Subjects = strings.Split(subject, entity.JoinComma)
		}

		category := strings.TrimSpace(oc.Developmental)
		if category != "" {
			oc.Categories = strings.Split(category, entity.JoinComma)
		}

		subcategories := strings.TrimSpace(oc.Skills)
		if subcategories != "" {
			oc.Subcategories = strings.Split(subcategories, entity.JoinComma)
		}

		grade := strings.TrimSpace(oc.Grade)
		if grade != "" {
			oc.Grades = strings.Split(grade, entity.JoinComma)
		}

		age := strings.TrimSpace(oc.Age)
		if age != "" {
			oc.Ages = strings.Split(age, entity.JoinComma)
		}
		return
	}
	for i := range relations {
		switch relations[i].RelationType {
		case entity.ProgramType:
			oc.Programs = append(oc.Programs, relations[i].RelationID)
		case entity.SubjectType:
			oc.Subjects = append(oc.Subjects, relations[i].RelationID)
		case entity.CategoryType:
			oc.Categories = append(oc.Categories, relations[i].RelationID)
		case entity.SubcategoryType:
			oc.Subcategories = append(oc.Subcategories, relations[i].RelationID)
		case entity.GradeType:
			oc.Grades = append(oc.Grades, relations[i].RelationID)
		case entity.AgeType:
			oc.Ages = append(oc.Ages, relations[i].RelationID)
		}
	}
	if len(oc.Programs) > 0 {
		oc.Program = strings.Join(oc.Programs, entity.JoinComma)
	}
	if len(oc.Subjects) > 0 {
		oc.Subject = strings.Join(oc.Subjects, entity.JoinComma)
	}
	if len(oc.Developmental) > 0 {
		oc.Developmental = strings.Join(oc.Categories, entity.JoinComma)
	}
	if len(oc.Skills) > 0 {
		oc.Skills = strings.Join(oc.Subcategories, entity.JoinComma)
	}
	if len(oc.Grades) > 0 {
		oc.Grade = strings.Join(oc.Grades, entity.JoinComma)
	}
	if len(oc.Ages) > 0 {
		oc.Age = strings.Join(oc.Ages, entity.JoinComma)
	}
}
func (ocm OutcomeModel) UpdateOutcome(data *entity.Outcome, oc *entity.Outcome) {
	if data.Name != "" {
		oc.Name = data.Name
	}

	oc.Assumed = data.Assumed
	oc.Program = data.Program
	oc.Subject = data.Subject
	oc.Developmental = data.Developmental
	oc.Skills = data.Skills
	oc.Age = data.Age
	oc.Grade = data.Grade
	oc.EstimatedTime = data.EstimatedTime
	oc.Keywords = data.Keywords
	oc.Description = data.Description
	oc.PublishStatus = entity.OutcomeStatusDraft
	oc.Shortcode = data.Shortcode
	oc.Sets = data.Sets
	oc.UpdateAt = time.Now().Unix()
}

func (ocm OutcomeModel) Clone(op *entity.Operator, oc *entity.Outcome) entity.Outcome {
	now := time.Now().Unix()
	return entity.Outcome{
		ID:            utils.NewID(),
		AncestorID:    oc.AncestorID,
		Shortcode:     oc.Shortcode,
		Name:          oc.Name,
		Program:       oc.Program,
		Subject:       oc.Subject,
		Developmental: oc.Developmental,
		Skills:        oc.Skills,
		Age:           oc.Age,
		Grade:         oc.Grade,
		Keywords:      oc.Keywords,
		Description:   oc.Description,

		EstimatedTime:  oc.EstimatedTime,
		AuthorID:       op.UserID,
		AuthorName:     oc.AuthorName,
		OrganizationID: oc.OrganizationID,

		PublishStatus: entity.OutcomeStatusDraft,
		PublishScope:  oc.PublishScope,
		LatestID:      oc.LatestID,
		Sets:          oc.Sets,

		Version:  1,
		SourceID: oc.ID,
		Assumed:  oc.Assumed,

		CreateAt: now,
		UpdateAt: now,
	}
}

func (ocm OutcomeModel) SetStatus(ctx context.Context, oc *entity.Outcome, status entity.OutcomeStatus) error {
	switch status {
	case entity.OutcomeStatusHidden:
		if ocm.allowedToHidden(oc) {
			oc.PublishStatus = entity.OutcomeStatusHidden
			return nil
		}
	case entity.OutcomeStatusPending:
		if ocm.allowedToPending(oc) {
			oc.PublishStatus = entity.OutcomeStatusPending
			return nil
		}
	case entity.OutcomeStatusPublished:
		if ocm.allowedToBeReviewed(oc) {
			oc.PublishStatus = entity.OutcomeStatusPublished
			return nil
		}
	case entity.OutcomeStatusRejected:
		if ocm.allowedToBeReviewed(oc) {
			oc.PublishStatus = entity.OutcomeStatusRejected
			return nil
		}
	}
	log.Error(ctx, "SetStatus failed",
		log.Err(constant.ErrForbidden),
		log.String("status", string(status)))
	return constant.ErrForbidden
}

func (ocm OutcomeModel) allowedToArchive(oc *entity.Outcome) bool {
	switch oc.PublishStatus {
	case entity.OutcomeStatusPublished:
		return true
	}
	return false
}

func (ocm OutcomeModel) allowedToAttachment(oc *entity.Outcome) bool {
	// TODO
	return false
}

func (ocm OutcomeModel) allowedToPending(oc *entity.Outcome) bool {
	switch oc.PublishStatus {
	case entity.OutcomeStatusDraft, entity.OutcomeStatusRejected:
		return true
	}
	return false
}

func (ocm OutcomeModel) allowedToBeReviewed(oc *entity.Outcome) bool {
	switch oc.PublishStatus {
	case entity.OutcomeStatusPending:
		return true
	}
	return false
}

func (ocm OutcomeModel) allowedToHidden(oc *entity.Outcome) bool {
	switch oc.PublishStatus {
	case entity.OutcomeStatusPublished:
		return true
	}
	return false
}
