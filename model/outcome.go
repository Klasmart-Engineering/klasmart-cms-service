package model

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"golang.org/x/sync/errgroup"

	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrOutcomeInvalidData = errors.New("invalid outcome data")
)

type IOutcomeModel interface {
	Create(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) error
	Update(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) error
	Delete(ctx context.Context, operator *entity.Operator, outcomeID string) error

	Get(ctx context.Context, operator *entity.Operator, outcomeID string) (*OutcomeDetailView, error)
	GetByIDs(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) ([]*entity.Outcome, error)
	GetLatestOutcomes(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (map[string]*entity.Outcome, []string, error)
	GetLatestByIDsMapResult(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (map[string]*entity.Outcome, error)
	GetLatestByAncestors(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, ancestoryIDs []string) ([]*entity.Outcome, error)

	Search(ctx context.Context, operator *entity.Operator, condition *entity.OutcomeCondition) (*SearchOutcomeResponse, error)
	SearchWithoutRelation(ctx context.Context, operator *entity.Operator, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error)
	SearchPrivate(ctx context.Context, operator *entity.Operator, condition *entity.OutcomeCondition) (*SearchOutcomeResponse, error)
	SearchPending(ctx context.Context, operator *entity.Operator, condition *entity.OutcomeCondition) (*SearchOutcomeResponse, error)
	SearchPublished(ctx context.Context, operator *entity.Operator, condition *entity.OutcomeCondition) (*SearchPublishedOutcomeResponse, error)

	Lock(ctx context.Context, operator *entity.Operator, outcomeID string) (string, error)
	HasLocked(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (bool, error)
	Publish(ctx context.Context, operator *entity.Operator, outcomeID string, scope string) error
	BulkPublish(ctx context.Context, operator *entity.Operator, outcomeIDs []string, scope string) error
	BulkDelete(ctx context.Context, operator *entity.Operator, outcomeIDs []string) error

	Approve(ctx context.Context, operator *entity.Operator, outcomeID string) error
	Reject(ctx context.Context, operator *entity.Operator, outcomeID string, reason string) error

	BulkApprove(ctx context.Context, operator *entity.Operator, outcomeIDs []string) error
	BulkReject(ctx context.Context, operator *entity.Operator, outcomeIDs []string, reason string) error

	GenerateShortcode(ctx context.Context, op *entity.Operator) (string, error)

	ShortcodeProvider
}

type OutcomeModel struct {
	outcomeDA          da.IOutcomeDA
	outcomeRelationDA  da.IOutcomeRelationDA
	outcomeSetDA       da.IOutcomeSetDA
	milestoneDA        da.IMilestoneDA
	milestoneOutcomeDA da.IMilestoneOutcomeDA

	organizationService external.OrganizationServiceProvider
	userService         external.UserServiceProvider
	programService      external.ProgramServiceProvider
	subjectService      external.SubjectServiceProvider
	categoryService     external.CategoryServiceProvider
	subCategoryService  external.SubCategoryServiceProvider
	gradeService        external.GradeServiceProvider
	ageService          external.AgeServiceProvider
}

var (
	_outcomeModel     IOutcomeModel
	_outcomeModelOnce sync.Once
)

func GetOutcomeModel() IOutcomeModel {
	_outcomeModelOnce.Do(func() {
		_outcomeModel = &OutcomeModel{
			outcomeDA:          da.GetOutcomeDA(),
			outcomeRelationDA:  da.GetOutcomeRelationDA(),
			outcomeSetDA:       da.GetOutcomeSetDA(),
			milestoneDA:        da.GetMilestoneDA(),
			milestoneOutcomeDA: da.GetMilestoneOutcomeDA(),

			organizationService: external.GetOrganizationServiceProvider(),
			userService:         external.GetUserServiceProvider(),
			programService:      external.GetProgramServiceProvider(),
			subjectService:      external.GetSubjectServiceProvider(),
			categoryService:     external.GetCategoryServiceProvider(),
			subCategoryService:  external.GetSubCategoryServiceProvider(),
			gradeService:        external.GetGradeServiceProvider(),
			ageService:          external.GetAgeServiceProvider(),
		}
	})
	return _outcomeModel
}

func (ocm OutcomeModel) GenerateShortcode(ctx context.Context, op *entity.Operator) (string, error) {
	var shortcode string
	var index int
	cursor, err := ocm.Current(ctx, op)
	if err != nil {
		log.Debug(ctx, "GenerateShortcode: Current failed",
			log.Any("op", op),
			log.Int("cursor", cursor))
		return "", err
	}
	shortcodeModel := GetShortcodeModel()
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		index, shortcode, err = shortcodeModel.generate(ctx, op, tx, cursor+1, ocm)
		if err != nil {
			log.Debug(ctx, "GenerateShortcode",
				log.Any("op", op),
				log.Int("cursor", cursor))
			return err
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	err = ocm.Cache(ctx, op, index, shortcode)
	return shortcode, err
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
		log.Error(ctx, "IsCached: redis access failed",
			log.Err(err),
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

		outcomeRelations := ocm.CollectRelation(outcome)
		_, err = da.GetOutcomeRelationDA().InsertInBatchesTx(ctx, tx, outcomeRelations, len(outcomeRelations))
		if err != nil {
			log.Error(ctx, "Create: InsertInBatchesTx failed",
				log.Any("op", operator),
				log.Any("outcome", outcome),
				log.Any("outcomeRelations", outcomeRelations))
			return err
		}

		return nil
	})
	ocm.RemoveShortcode(ctx, operator, outcome.Shortcode)
	if err != nil {
		return err
	}
	return
}

func (ocm OutcomeModel) Get(ctx context.Context, operator *entity.Operator, outcomeID string) (*OutcomeDetailView, error) {
	var outcome entity.Outcome
	err := ocm.outcomeDA.Get(ctx, outcomeID, &outcome)
	if err != nil {
		log.Error(ctx, "ocm.outcomeDA.Get error",
			log.Err(err),
			log.String("outcome_id", outcomeID))
		return nil, err
	}

	return ocm.transformToOutcomeDetailView(ctx, operator, &outcome)
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

		outcomeRelations := ocm.CollectRelation(outcome)
		_, err = da.GetOutcomeRelationDA().InsertInBatchesTx(ctx, tx, outcomeRelations, len(outcomeRelations))
		if err != nil {
			log.Error(ctx, "Update: InsertInBatchesTx failed",
				log.Any("op", operator),
				log.Any("outcome", outcome),
				log.Any("outcomeRelations", outcomeRelations))
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

func (ocm OutcomeModel) Search(ctx context.Context, op *entity.Operator, condition *entity.OutcomeCondition) (*SearchOutcomeResponse, error) {
	err := ocm.fillAuthorIDs(ctx, op, condition)
	if err != nil {
		log.Error(ctx, "Search: fillAuthorIDs failed",
			log.Err(err),
			log.String("op", op.UserID),
			log.Any("condition", condition))
		return nil, err
	}

	err = ocm.fillIDsBySetName(ctx, op, condition)
	if err != nil {
		log.Error(ctx, "Search: fillIDsBySetName failed",
			log.Err(err),
			log.String("op", op.UserID),
			log.Any("condition", condition))
		return nil, err
	}

	daCondition := da.NewOutcomeCondition(condition)
	var outcomes []*entity.Outcome
	total, err := ocm.outcomeDA.Page(ctx, daCondition, &outcomes)
	if err != nil {
		log.Error(ctx, "ocm.outcomeDA.Page error",
			log.Err(err),
			log.Any("daCondition", daCondition))
		return nil, err
	}

	if len(outcomes) == 0 {
		log.Debug(ctx, "outcome record not found",
			log.Any("op", op),
			log.Any("daCondition", daCondition))
		return &SearchOutcomeResponse{
			Total: total,
			List:  []*OutcomeView{},
		}, nil
	}

	outcomeViewList, err := ocm.transformToOutcomeView(ctx, op, outcomes)
	if err != nil {
		log.Error(ctx, "ocm.transformToOutcomeView error",
			log.Err(err),
			log.Any("outcomes", outcomes))
		return nil, err
	}

	return &SearchOutcomeResponse{
		Total: total,
		List:  outcomeViewList,
	}, nil
}

func (ocm OutcomeModel) SearchWithoutRelation(ctx context.Context, user *entity.Operator, condition *entity.OutcomeCondition) (int, []*entity.Outcome, error) {
	var outcomes []*entity.Outcome
	total, err := da.GetOutcomeDA().Page(ctx, da.NewOutcomeCondition(condition), &outcomes)
	if err != nil {
		log.Error(ctx, "SearchWithoutRelation: da.GetOutcomeDA().Query failed",
			log.Any("op", user),
			log.Any("condition", condition))
		return 0, nil, err
	}

	return total, outcomes, err
}

func (ocm OutcomeModel) Lock(ctx context.Context, operator *entity.Operator, outcomeID string) (string, error) {
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
		var outcomeRelations []*entity.OutcomeRelation
		err = da.GetOutcomeRelationDA().QueryTx(ctx, tx, &da.OutcomeRelationCondition{
			MasterIDs:  dbo.NullStrings{Strings: []string{outcome.ID}, Valid: true},
			MasterType: sql.NullString{String: string(entity.OutcomeType), Valid: true},
		}, &outcomeRelations)
		if err != nil {
			log.Error(ctx, "Lock: QueryTx failed",
				log.String("op", operator.UserID),
				log.String("outcome_id", outcomeID),
				log.Any("outcome", newVersion))
			return err
		}
		for i := range outcomeRelations {
			outcomeRelations[i].MasterID = newVersion.ID
			outcomeRelations[i].ID = 0
		}

		_, err = da.GetOutcomeRelationDA().InsertInBatchesTx(ctx, tx, outcomeRelations, len(outcomeRelations))
		if err != nil {
			log.Error(ctx, "Lock: InsertInBatchesTx failed",
				log.Err(err),
				log.Any("op", operator),
				log.Any("outcomeRelations", outcomeRelations))
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
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
		if outcome.AuthorID != operator.UserID {
			log.Warn(ctx, "Publish: must by self",
				log.String("op", operator.UserID),
				log.Any("outcome", outcome))
			return ErrNoAuth
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
	return err
}

func (ocm OutcomeModel) BulkPublish(ctx context.Context, operator *entity.Operator, outcomeIDs []string, scope string) error {
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
			if o.AuthorID != operator.UserID {
				log.Warn(ctx, "BulkPublish: must by self",
					log.String("op", operator.UserID),
					log.Any("outcome", o))
				return ErrNoAuth
			}
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
	return err
}

func (ocm OutcomeModel) BulkDelete(ctx context.Context, operator *entity.Operator, outcomeIDs []string) error {
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

		if len(outcomes) == 0 {
			log.Warn(ctx, "BulkDelete: not found", log.Any("op", operator),
				log.Any("perms", perms), log.Strings("outcome_ids", outcomeIDs))
			return nil
		}

		for i := range outcomes {
			if !allowDeleteOutcome(ctx, operator, perms, outcomes[i]) {
				log.Warn(ctx, "BulkDelete: no permission", log.Any("op", operator),
					log.Any("perms", perms), log.Any("outcome", outcomes[i]))
				return constant.ErrOperateNotAllowed
			}
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
	return err
}

func (ocm OutcomeModel) SearchPrivate(ctx context.Context, op *entity.Operator, condition *entity.OutcomeCondition) (*SearchOutcomeResponse, error) {
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.ViewMyUnpublishedLearningOutcome,  // my draft & my rejected
		external.ViewOrgUnpublishedLearningOutcome, // org draft & org waiting for approved & org rejected
		external.ViewMyPendingLearningOutcome,      // my waiting for approved
	})
	if err != nil {
		log.Error(ctx, "SearchPrivate: HasOrganizationPermissions failed",
			log.Any("op", op), log.Err(err))
		return nil, constant.ErrInternalServer
	}
	condition.OrganizationID = op.OrgID
	if !allowSearchPrivate(ctx, op, perms, condition) {
		log.Warn(ctx, "SearchPrivate: no permission",
			log.Any("op", op),
			log.Any("perms", perms),
			log.Any("cond", condition))
		return nil, constant.ErrOperateNotAllowed
	}

	err = ocm.fillAuthorIDs(ctx, op, condition)
	if err != nil {
		log.Error(ctx, "SearchPrivate: fillAuthorIDs failed",
			log.String("op", op.UserID),
			log.Any("condition", condition))
		return nil, err
	}
	err = ocm.fillIDsBySetName(ctx, op, condition)
	if err != nil {
		log.Error(ctx, "SearchPrivate: fillIDsBySetName failed",
			log.String("op", op.UserID),
			log.Any("condition", condition))
		return nil, err
	}

	daCondition := da.NewOutcomeCondition(condition)
	var outcomes []*entity.Outcome
	total, err := ocm.outcomeDA.Page(ctx, daCondition, &outcomes)
	if err != nil {
		log.Error(ctx, "ocm.outcomeDA.Page error",
			log.Err(err),
			log.Any("daCondition", daCondition))
		return nil, err
	}

	if len(outcomes) == 0 {
		log.Debug(ctx, "outcome record not found",
			log.Any("op", op),
			log.Any("daCondition", daCondition))
		return &SearchOutcomeResponse{
			Total: total,
			List:  []*OutcomeView{},
		}, nil
	}

	outcomeViewList, err := ocm.transformToOutcomeView(ctx, op, outcomes)
	if err != nil {
		log.Error(ctx, "ocm.transformToOutcomeView error",
			log.Err(err),
			log.Any("outcomes", outcomes))
		return nil, err
	}

	return &SearchOutcomeResponse{
		Total: total,
		List:  outcomeViewList,
	}, nil
}

func (ocm OutcomeModel) SearchPending(ctx context.Context, op *entity.Operator, condition *entity.OutcomeCondition) (*SearchOutcomeResponse, error) {
	if condition.PublishStatus != entity.OutcomeStatusPending {
		log.Warn(ctx, "SearchPending: SearchPending failed",
			log.String("op", op.UserID),
			log.Any("condition", condition))
		return nil, ErrBadRequest
	}
	// as there is no level,orgID is the same as [user.OrgID]
	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ViewOrgPendingLearningOutcome)
	if !hasPerm {
		log.Warn(ctx, "SearchPending: no permission",
			log.String("op", op.UserID),
			log.Any("condition", condition))
		return nil, constant.ErrOperateNotAllowed
	}
	condition.PublishScope = op.OrgID
	err = ocm.fillAuthorIDs(ctx, op, condition)
	if err != nil {
		log.Error(ctx, "SearchPending: fillAuthorIDs failed",
			log.String("op", op.UserID),
			log.Any("condition", condition))
		return nil, err
	}
	err = ocm.fillIDsBySetName(ctx, op, condition)
	if err != nil {
		log.Error(ctx, "SearchPending: fillIDsBySetName failed",
			log.String("op", op.UserID),
			log.Any("condition", condition))
		return nil, err
	}

	daCondition := da.NewOutcomeCondition(condition)
	var outcomes []*entity.Outcome
	total, err := ocm.outcomeDA.Page(ctx, daCondition, &outcomes)
	if err != nil {
		log.Error(ctx, "ocm.outcomeDA.Page error",
			log.Err(err),
			log.Any("daCondition", daCondition))
		return nil, err
	}

	if len(outcomes) == 0 {
		log.Debug(ctx, "outcome record not found",
			log.Any("op", op),
			log.Any("daCondition", daCondition))
		return &SearchOutcomeResponse{
			Total: total,
			List:  []*OutcomeView{},
		}, nil
	}

	outcomeViewList, err := ocm.transformToOutcomeView(ctx, op, outcomes)
	if err != nil {
		log.Error(ctx, "ocm.transformToOutcomeView error",
			log.Err(err),
			log.Any("outcomes", outcomes))
		return nil, err
	}

	return &SearchOutcomeResponse{
		Total: total,
		List:  outcomeViewList,
	}, nil
}

func (o OutcomeModel) SearchPublished(ctx context.Context, op *entity.Operator, condition *entity.OutcomeCondition) (*SearchPublishedOutcomeResponse, error) {
	var outcomes []*entity.Outcome
	daCondition := da.NewOutcomeCondition(condition)
	total, err := o.outcomeDA.Page(ctx, daCondition, &outcomes)
	if err != nil {
		log.Error(ctx, "o.outcomeDA.Page error",
			log.Any("op", op),
			log.Any("daCondition", daCondition),
			log.Err(err))
		return nil, err
	}

	if len(outcomes) == 0 {
		log.Debug(ctx, "outcome record not found",
			log.Any("op", op),
			log.Any("daCondition", daCondition))
		return &SearchPublishedOutcomeResponse{
			Total: total,
			List:  []*PublishedOutcomeView{},
		}, nil
	}

	publishedOutcomeViewList, err := o.transformToPublishedOutcomeView(ctx, op, outcomes)
	if err != nil {
		log.Error(ctx, "o.transformToPublishedOutcomeView error",
			log.Err(err),
			log.Any("op", op),
			log.Any("outcomes", outcomes))
		return nil, err
	}

	return &SearchPublishedOutcomeResponse{
		Total: total,
		List:  publishedOutcomeViewList,
	}, err
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
		// NKL-1021
		// err = GetMilestoneModel().BindToGeneral(ctx, operator, tx, outcome)
		// if err != nil {
		// 	log.Error(ctx, "Approve: BindToGeneral failed",
		// 		log.String("op", operator.UserID),
		// 		log.Any("outcome", outcome))
		// 	return err
		// }
		return nil
	})
	return err
}

func (ocm OutcomeModel) Reject(ctx context.Context, operator *entity.Operator, outcomeID string, reason string) error {
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
			// NKL-1021
			// err = GetMilestoneModel().BindToGeneral(ctx, operator, tx, outcome)
			// if err != nil {
			// 	log.Error(ctx, "BulkApprove: BindToGeneral failed",
			// 		log.String("op", operator.UserID),
			// 		log.Any("outcome", outcome))
			// 	return err
			// }
		}
		return nil
	})
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

// map key is outcome id from outcomeIDs, []string is original outcome id order by name asc
func (ocm OutcomeModel) GetLatestOutcomes(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (map[string]*entity.Outcome, []string, error) {
	result := make(map[string]*entity.Outcome, len(outcomeIDs))
	var sortedOutcomeID []string
	var outcomes []*entity.Outcome
	outcomeCondition := &da.OutcomeCondition{
		IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
	}

	err := da.GetOutcomeDA().QueryTx(ctx, tx, outcomeCondition, &outcomes)
	if err != nil {
		log.Error(ctx, "da.GetOutcomeDA().QueryTx error",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("outcomeCondition", outcomeCondition))
		return nil, nil, err
	}

	if len(outcomes) == 0 {
		log.Debug(ctx, "outcome not found", log.Any("outcomeCondition", outcomeCondition))
		return result, sortedOutcomeID, nil
	}

	latestOutcomeIDs := make([]string, len(outcomes))
	for i, v := range outcomes {
		latestOutcomeIDs[i] = v.LatestID
	}
	latestOutcomeCondition := &da.OutcomeCondition{
		IDs:     dbo.NullStrings{Strings: latestOutcomeIDs, Valid: true},
		OrderBy: da.OrderByName,
	}
	var latestOutcomes []*entity.Outcome
	err = da.GetOutcomeDA().QueryTx(ctx, tx, latestOutcomeCondition, &latestOutcomes)
	if err != nil {
		log.Error(ctx, "da.GetOutcomeDA().QueryTx error",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("latestOutcomeCondition", latestOutcomeCondition))
		return nil, nil, err
	}

	if len(latestOutcomes) == 0 {
		log.Debug(ctx, "latest outcome not found", log.Any("latestOutcomeCondition", latestOutcomeCondition))
		return result, sortedOutcomeID, nil
	}

	err = ocm.fillRelation(ctx, operator, tx, latestOutcomes)
	if err != nil {
		log.Error(ctx, "ocm.fillRelation failed",
			log.Err(err),
			log.Any("latestOutcomes", latestOutcomes))
		return nil, nil, err
	}

	for _, latestOutcome := range latestOutcomes {
		for _, outcome := range outcomes {
			if outcome.LatestID == latestOutcome.ID {
				sortedOutcomeID = append(sortedOutcomeID, outcome.ID)
				result[outcome.ID] = latestOutcome
				break
			}
		}
	}

	return result, sortedOutcomeID, nil
}

func (ocm OutcomeModel) GetLatestByIDsMapResult(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (latests map[string]*entity.Outcome, err error) {
	cond1 := da.OutcomeCondition{
		IDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
	}
	total, otcs1, err1 := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &cond1)
	if err1 != nil {
		log.Error(ctx, "GetLatestByIDs: SearchOutcome failed",
			log.Err(err1),
			log.String("op", operator.UserID),
			log.Strings("outcome_ids", outcomeIDs))
		return nil, err1
	}
	if total == 0 {
		log.Debug(ctx, "GetLatestByIDs: SearchOutcome return empty",
			log.String("op", operator.UserID),
			log.Strings("outcome_ids", outcomeIDs))
		return nil, constant.ErrRecordNotFound
	}
	cond2 := da.OutcomeCondition{}
	for _, o := range otcs1 {
		cond2.IDs.Strings = append(cond2.IDs.Strings, o.LatestID)
	}
	cond2.IDs.Valid = true
	total, otcs2, err1 := da.GetOutcomeDA().SearchOutcome(ctx, operator, tx, &cond2)
	if err1 != nil {
		log.Error(ctx, "GetLatestByIDs: SearchOutcome failed",
			log.Err(err1),
			log.String("op", operator.UserID),
			log.Strings("outcome_ids", cond2.IDs.Strings))
		return nil, err1
	}
	if total == 0 {
		log.Debug(ctx, "GetLatestByIDs: SearchOutcome return empty",
			log.String("op", operator.UserID),
			log.Strings("outcome_ids", cond2.IDs.Strings))
		return nil, constant.ErrRecordNotFound
	}
	err1 = ocm.fillRelation(ctx, operator, tx, otcs2)
	if err1 != nil {
		log.Error(ctx, "GetLatestByIDs: fillRelation failed",
			log.Err(err1),
			log.String("op", operator.UserID))
		return nil, err1
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
	_, outcomes, err = da.GetOutcomeDA().SearchOutcome(ctx, op, tx, &da.OutcomeCondition{
		AncestorIDs:   dbo.NullStrings{Strings: ancestorIDs, Valid: true},
		PublishStatus: dbo.NullStrings{Strings: []string{entity.OutcomeStatusPublished}, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "GetLatestByAncestors: SearchOutcome failed",
			log.Err(err),
			log.String("op", op.UserID),
			log.Strings("ancestor", ancestorIDs))
		return
	}
	err = ocm.fillRelation(ctx, op, tx, outcomes)
	if err != nil {
		log.Error(ctx, "GetLatestByAncestors: fillRelation failed",
			log.Err(err),
			log.String("op", op.UserID),
			log.Strings("ancestor", ancestorIDs))
		return
	}
	return
}

func (ocm OutcomeModel) fillRelation(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomes []*entity.Outcome) error {
	if len(outcomes) > 0 {
		masterIDs := make([]string, len(outcomes))
		for i := range outcomes {
			masterIDs[i] = outcomes[i].ID
		}

		var outcomeRelations []*entity.OutcomeRelation
		err := da.GetOutcomeRelationDA().QueryTx(ctx, tx, &da.OutcomeRelationCondition{
			MasterIDs:  dbo.NullStrings{Strings: masterIDs, Valid: true},
			MasterType: sql.NullString{String: string(entity.OutcomeType), Valid: true},
		}, &outcomeRelations)
		if err != nil {
			log.Error(ctx, "fillRelation: Query failed",
				log.Err(err),
				log.String("op", operator.UserID),
				log.Any("outcomes", outcomes))
			return err
		}
		for _, outcome := range outcomes {
			var temp []*entity.OutcomeRelation
			for _, outcomeRelation := range outcomeRelations {
				if outcomeRelation.MasterID == outcome.ID {
					temp = append(temp, outcomeRelation)
				}
			}
			ocm.FillRelation(ctx, outcome, temp)
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

	if outcome.PublishStatus == entity.OutcomeStatusPending && perms[external.DeleteMyPendingLearningOutcome] && outcome.AuthorID == operator.UserID {
		return true
	}

	if outcome.PublishStatus == entity.OutcomeStatusPending && perms[external.DeleteOrgPendingLearningOutcome] && outcome.OrganizationID == operator.OrgID {
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

func (ocm OutcomeModel) CollectRelation(oc *entity.Outcome) []*entity.OutcomeRelation {
	outcomeRelations := make([]*entity.OutcomeRelation, 0, len(oc.Programs)+len(oc.Subjects)+len(oc.Categories)+len(oc.Subcategories)+len(oc.Grades)+len(oc.Ages))
	for i := range oc.Programs {
		outcomeRelation := entity.OutcomeRelation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Programs[i],
			RelationType: entity.ProgramType,
		}
		outcomeRelations = append(outcomeRelations, &outcomeRelation)
	}

	for i := range oc.Subjects {
		outcomeRelation := entity.OutcomeRelation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Subjects[i],
			RelationType: entity.SubjectType,
		}
		outcomeRelations = append(outcomeRelations, &outcomeRelation)
	}

	for i := range oc.Categories {
		outcomeRelation := entity.OutcomeRelation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Categories[i],
			RelationType: entity.CategoryType,
		}
		outcomeRelations = append(outcomeRelations, &outcomeRelation)
	}

	for i := range oc.Subcategories {
		outcomeRelation := entity.OutcomeRelation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Subcategories[i],
			RelationType: entity.SubcategoryType,
		}
		outcomeRelations = append(outcomeRelations, &outcomeRelation)
	}

	for i := range oc.Grades {
		outcomeRelation := entity.OutcomeRelation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Grades[i],
			RelationType: entity.GradeType,
		}
		outcomeRelations = append(outcomeRelations, &outcomeRelation)
	}

	for i := range oc.Ages {
		outcomeRelation := entity.OutcomeRelation{
			MasterID:     oc.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   oc.Ages[i],
			RelationType: entity.AgeType,
		}
		outcomeRelations = append(outcomeRelations, &outcomeRelation)
	}
	return outcomeRelations
}

// TODO: Kyle: outcome relation data sync check
func (ocm OutcomeModel) FillRelation(ctx context.Context, oc *entity.Outcome, relations []*entity.OutcomeRelation) {
	log.Debug(ctx, "fill relation",
		log.Any("outcome", oc),
		log.Any("relations", relations),
	)
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
	} else {
		oc.Programs = strings.Split(oc.Program, entity.JoinComma)
	}
	if len(oc.Subjects) > 0 {
		oc.Subject = strings.Join(oc.Subjects, entity.JoinComma)
	} else {
		oc.Subjects = strings.Split(oc.Subject, entity.JoinComma)
	}
	if len(oc.Categories) > 0 {
		oc.Developmental = strings.Join(oc.Categories, entity.JoinComma)
	} else {
		oc.Categories = strings.Split(oc.Developmental, entity.JoinComma)
	}
	if len(oc.Subcategories) > 0 {
		oc.Skills = strings.Join(oc.Subcategories, entity.JoinComma)
	} else {
		oc.Subcategories = strings.Split(oc.Skills, entity.JoinComma)
	}
	if len(oc.Grades) > 0 {
		oc.Grade = strings.Join(oc.Grades, entity.JoinComma)
	} else {
		oc.Grades = strings.Split(oc.Grade, entity.JoinComma)
	}
	if len(oc.Ages) > 0 {
		oc.Age = strings.Join(oc.Ages, entity.JoinComma)
	} else {
		oc.Ages = strings.Split(oc.Age, entity.JoinComma)
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

func (o OutcomeModel) transformToPublishedOutcomeView(ctx context.Context, operator *entity.Operator, outcomes []*entity.Outcome) ([]*PublishedOutcomeView, error) {
	result := make([]*PublishedOutcomeView, len(outcomes))
	outcomeMap := make(map[string]*PublishedOutcomeView, len(outcomes))
	outcomeIDs := make([]string, len(outcomes))
	for i, outcome := range outcomes {
		result[i] = &PublishedOutcomeView{
			OutcomeID:   outcome.ID,
			OutcomeName: outcome.Name,
			Shortcode:   outcome.Shortcode,
			Assumed:     outcome.Assumed,
			// init zero value
			Sets:           []*OutcomeSetCreateView{},
			ProgramIDs:     []string{},
			SubjectIDs:     []string{},
			CategoryIDs:    []string{},
			SubcategoryIDs: []string{},
			GradeIDs:       []string{},
			AgeIDs:         []string{},
		}

		outcomeIDs[i] = outcome.ID
		outcomeMap[outcome.ID] = result[i]
	}

	g := new(errgroup.Group)
	var outcomeSetMap map[string][]*entity.Set
	var outcomeRelations []*entity.OutcomeRelation

	// get outcome set
	g.Go(func() error {
		outcomeSets, err := o.outcomeSetDA.SearchSetsByOutcome(ctx, dbo.MustGetDB(ctx), outcomeIDs)
		if err != nil {
			log.Error(ctx, "o.outcomeSetDA.SearchSetsByOutcome error",
				log.Err(err),
				log.Strings("outcomeIDs", outcomeIDs))
			return err
		}

		outcomeSetMap = outcomeSets
		return nil
	})

	// get outcome relations
	g.Go(func() error {
		err := o.outcomeRelationDA.Query(ctx, &da.OutcomeRelationCondition{
			MasterIDs: dbo.NullStrings{
				Strings: outcomeIDs,
				Valid:   true,
			},
		}, &outcomeRelations)
		if err != nil {
			log.Error(ctx, "o.outcomeRelationDA.Query error",
				log.Err(err),
				log.Strings("outcomeIDs", outcomeIDs))
			return err
		}

		return nil
	})

	for _, outcomeRelation := range outcomeRelations {
		if outcome, ok := outcomeMap[outcomeRelation.MasterID]; ok {
			switch outcomeRelation.RelationType {
			case entity.ProgramType:
				outcome.ProgramIDs = append(outcome.ProgramIDs, outcomeRelation.RelationID)
			case entity.SubjectType:
				outcome.SubjectIDs = append(outcome.SubjectIDs, outcomeRelation.RelationID)
			case entity.CategoryType:
				outcome.CategoryIDs = append(outcome.CategoryIDs, outcomeRelation.RelationID)
			case entity.SubcategoryType:
				outcome.SubcategoryIDs = append(outcome.SubcategoryIDs, outcomeRelation.RelationID)
			case entity.GradeType:
				outcome.GradeIDs = append(outcome.GradeIDs, outcomeRelation.RelationID)
			case entity.AgeType:
				outcome.AgeIDs = append(outcome.AgeIDs, outcomeRelation.RelationID)
			}
		}
	}

	for i, outcome := range outcomes {
		outcomeView := result[i]
		// fill outcome sets
		if outcomeSets, ok := outcomeSetMap[outcome.ID]; ok {
			for _, set := range outcomeSets {
				outcomeView.Sets = append(outcomeView.Sets, &OutcomeSetCreateView{
					SetID:   set.ID,
					SetName: set.Name,
				})
			}
		}
	}

	return result, nil
}

func (o OutcomeModel) transformToOutcomeView(ctx context.Context, operator *entity.Operator, outcomes []*entity.Outcome) ([]*OutcomeView, error) {
	result := make([]*OutcomeView, len(outcomes))
	resultMap := make(map[string]*OutcomeView, len(outcomes))
	outcomeIDs := make([]string, len(outcomes))
	var lockedOutcomeIDs []string
	var userIDs []string
	var outcomeRelations []*entity.OutcomeRelation
	var programIDs []string
	var categoryIDs []string

	for i, outcome := range outcomes {
		result[i] = &OutcomeView{
			OutcomeID:     outcome.ID,
			AncestorID:    outcome.AncestorID,
			OutcomeName:   outcome.Name,
			Shortcode:     outcome.Shortcode,
			Assumed:       outcome.Assumed,
			LockedBy:      outcome.LockedBy,
			AuthorID:      outcome.AuthorID,
			AuthorName:    outcome.AuthorName,
			PublishStatus: string(outcome.PublishStatus),
			CreatedAt:     outcome.CreateAt,
			UpdatedAt:     outcome.UpdateAt,

			// init zero value
			LockedLocation: []string{},
			Program:        []Program{},
			Developmental:  []Developmental{},
			Sets:           []*OutcomeSetCreateView{},
		}

		if outcome.HasLocked() {
			userIDs = append(userIDs, outcome.LockedBy)
			lockedOutcomeIDs = append(lockedOutcomeIDs, outcome.ID)
		}

		userIDs = append(userIDs, outcome.AuthorID)
		outcomeIDs[i] = outcome.ID
		resultMap[outcome.ID] = result[i]
	}

	err := o.outcomeRelationDA.Query(ctx, &da.OutcomeRelationCondition{
		MasterIDs: dbo.NullStrings{
			Strings: outcomeIDs,
			Valid:   true,
		},
	}, &outcomeRelations)
	if err != nil {
		log.Error(ctx, "o.outcomeRelationDA.Query error",
			log.Err(err),
			log.Strings("outcomeIDs", outcomeIDs))
		return nil, err
	}

	for _, outcomeRelation := range outcomeRelations {
		switch outcomeRelation.RelationType {
		case entity.ProgramType:
			programIDs = append(programIDs, outcomeRelation.RelationID)
		case entity.CategoryType:
			categoryIDs = append(categoryIDs, outcomeRelation.RelationID)
		}
	}

	programIDs = utils.StableSliceDeduplication(programIDs)
	categoryIDs = utils.StableSliceDeduplication(categoryIDs)

	g := new(errgroup.Group)
	var outcomeSetMap map[string][]*entity.Set
	var userNameMap map[string]string
	lockedOutcomeChildrenMap := make(map[string]*entity.Outcome)
	var programMap map[string]*external.Program
	var categoryMap map[string]*external.Category

	// get outcome set
	g.Go(func() error {
		outcomeSets, err := o.outcomeSetDA.SearchSetsByOutcome(ctx, dbo.MustGetDB(ctx), outcomeIDs)
		if err != nil {
			log.Error(ctx, "o.outcomeSetDA.SearchSetsByOutcome error",
				log.Err(err),
				log.Strings("outcomeIDs", outcomeIDs))
			return err
		}

		outcomeSetMap = outcomeSets
		return nil
	})

	// get user name map
	if len(userIDs) > 0 {
		g.Go(func() error {
			users, err := o.userService.BatchGetNameMap(ctx, operator, userIDs)
			if err != nil {
				log.Error(ctx, "s.userService.BatchGetMap error",
					log.Err(err),
					log.Strings("userIDs", userIDs))
				return err
			}

			userNameMap = users

			return nil
		})
	}

	// get locked outcome children
	if len(lockedOutcomeIDs) > 0 {
		g.Go(func() error {
			var lockedOutcomeChildren []*entity.Outcome
			lockedChildrenCondition := &da.OutcomeCondition{
				IncludeDeleted: false,
				SourceIDs:      dbo.NullStrings{Strings: lockedOutcomeIDs, Valid: true},
			}

			err := o.outcomeDA.Query(ctx, lockedChildrenCondition, &lockedOutcomeChildren)
			if err != nil {
				log.Error(ctx, "o.outcomeDA.Query",
					log.Err(err),
					log.Any("condition", lockedChildrenCondition))
				return err
			}

			for _, lockedOutcome := range lockedOutcomeChildren {
				lockedOutcomeChildrenMap[lockedOutcome.SourceID] = lockedOutcome
			}

			return nil
		})
	}

	// get program info
	if len(programIDs) > 0 {
		g.Go(func() error {
			programs, err := o.programService.BatchGetMap(ctx, operator, programIDs)
			if err != nil {
				log.Error(ctx, "o.programService.BatchGetMap error",
					log.Err(err),
					log.Strings("programIDs", programIDs))
				return err
			}

			programMap = programs

			return nil
		})
	}

	// get category info
	if len(categoryIDs) > 0 {
		g.Go(func() error {
			categories, err := o.categoryService.BatchGetMap(ctx, operator, categoryIDs)
			if err != nil {
				log.Error(ctx, "o.categoryService.BatchGetMap error",
					log.Err(err),
					log.Strings("categoryIDs", categoryIDs))
				return err
			}

			categoryMap = categories

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		log.Error(ctx, "transformToOutcomeView error",
			log.Err(err))
		return nil, err
	}

	for i, outcome := range outcomes {
		outcomeView := result[i]
		// fill outcome sets
		if outcomeSets, ok := outcomeSetMap[outcome.ID]; ok {
			for _, set := range outcomeSets {
				outcomeView.Sets = append(outcomeView.Sets, &OutcomeSetCreateView{
					SetID:   set.ID,
					SetName: set.Name,
				})
			}
		}

		// fill author name
		if userName, ok := userNameMap[outcome.AuthorID]; ok {
			outcomeView.AuthorName = userName
		} else {
			log.Error(ctx, "author not found", log.String("userID", outcome.AuthorID))
		}

		if outcome.HasLocked() {
			if lockedOutcomeChildren, ok := lockedOutcomeChildrenMap[outcome.ID]; ok {
				outcomeView.LockedLocation = []string{string(lockedOutcomeChildren.PublishStatus)}
				outcomeView.LastEditedAt = lockedOutcomeChildren.CreateAt
				if userName, ok := userNameMap[outcome.LockedBy]; ok {
					outcomeView.LastEditedBy = userName
				} else {
					log.Error(ctx, "locked by user not found", log.String("userID", outcome.LockedBy))
				}
			} else {
				log.Error(ctx, "fill outcome locked info error", log.Any("outcome", outcome))
				return nil, ErrOutcomeInvalidData
			}
		}
	}

	// fill program and category
	for _, outcomeRelation := range outcomeRelations {
		if outcomeView, ok := resultMap[outcomeRelation.MasterID]; ok {
			switch outcomeRelation.RelationType {
			case entity.ProgramType:
				if program, ok := programMap[outcomeRelation.RelationID]; ok {
					outcomeView.Program = append(outcomeView.Program,
						Program{
							ProgramID:   program.ID,
							ProgramName: program.Name,
						})
				} else {
					log.Error(ctx, "program record not found", log.String("programID", outcomeRelation.RelationID))
				}
			case entity.CategoryType:
				if category, ok := categoryMap[outcomeRelation.RelationID]; ok {
					outcomeView.Developmental = append(outcomeView.Developmental,
						Developmental{
							DevelopmentalID:   category.ID,
							DevelopmentalName: category.Name,
						})
				} else {
					log.Error(ctx, "category record not found", log.String("categoryID", outcomeRelation.RelationID))
				}
			}
		}
	}

	return result, nil
}

func (o OutcomeModel) transformToOutcomeDetailView(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) (*OutcomeDetailView, error) {
	var userIDs []string

	result := &OutcomeDetailView{
		OutcomeID:      outcome.ID,
		AncestorID:     outcome.AncestorID,
		OrganizationID: outcome.OrganizationID,
		OutcomeName:    outcome.Name,
		Shortcode:      outcome.Shortcode,
		Description:    outcome.Description,
		AuthorID:       outcome.AuthorID,
		AuthorName:     outcome.AuthorName,
		Assumed:        outcome.Assumed,
		PublishStatus:  string(outcome.PublishStatus),
		RejectReason:   outcome.RejectReason,
		LockedBy:       outcome.LockedBy,
		Keywords:       strings.Split(outcome.Keywords, ","),
		CreatedAt:      outcome.CreateAt,
		UpdatedAt:      outcome.UpdateAt,
		// init zero value
		LockedLocation: []string{},
		Program:        []Program{},
		Subject:        []Subject{},
		Developmental:  []Developmental{},
		Skills:         []Skill{},
		Age:            []Age{},
		Grade:          []Grade{},
		Sets:           []*OutcomeSetCreateView{},
		Milestones:     []*Milestone{},
	}

	if outcome.HasLocked() {
		userIDs = append(userIDs, outcome.LockedBy)
	}
	userIDs = append(userIDs, outcome.AuthorID)

	var outcomeRelations []*entity.OutcomeRelation
	var programIDs []string
	var subjectIDs []string
	var categoryIDs []string
	var subCategoryIDs []string
	var ageIDs []string
	var gradeIDs []string

	err := o.outcomeRelationDA.Query(ctx, &da.OutcomeRelationCondition{
		MasterIDs: dbo.NullStrings{
			Strings: []string{outcome.ID},
			Valid:   true,
		},
	}, &outcomeRelations)
	if err != nil {
		log.Error(ctx, "o.outcomeRelationDA.Query error",
			log.Err(err),
			log.String("outcomeID", outcome.ID))
		return nil, err
	}

	for _, outcomeRelation := range outcomeRelations {
		switch outcomeRelation.RelationType {
		case entity.ProgramType:
			programIDs = append(programIDs, outcomeRelation.RelationID)
		case entity.SubjectType:
			subjectIDs = append(subjectIDs, outcomeRelation.RelationID)
		case entity.CategoryType:
			categoryIDs = append(categoryIDs, outcomeRelation.RelationID)
		case entity.SubcategoryType:
			subCategoryIDs = append(subCategoryIDs, outcomeRelation.RelationID)
		case entity.GradeType:
			gradeIDs = append(gradeIDs, outcomeRelation.RelationID)
		case entity.AgeType:
			ageIDs = append(ageIDs, outcomeRelation.RelationID)
		}
	}

	programIDs = utils.StableSliceDeduplication(programIDs)
	subjectIDs = utils.StableSliceDeduplication(subjectIDs)
	categoryIDs = utils.StableSliceDeduplication(categoryIDs)
	subCategoryIDs = utils.StableSliceDeduplication(subCategoryIDs)
	gradeIDs = utils.StableSliceDeduplication(gradeIDs)
	ageIDs = utils.StableSliceDeduplication(ageIDs)

	g := new(errgroup.Group)
	var outcomeSetMap map[string][]*entity.Set
	var userNameMap map[string]string
	lockedOutcomeChildrenMap := make(map[string]*entity.Outcome)
	var programMap map[string]*external.Program
	var subjectMap map[string]*external.Subject
	var categoryMap map[string]*external.Category
	var subCategoryMap map[string]*external.SubCategory
	var gradeMap map[string]*external.Grade
	var ageMap map[string]*external.Age
	outcomeMilestones := []*Milestone{}
	var orgNameMap map[string]string

	// get org name
	g.Go(func() error {
		orgNames, err := o.organizationService.BatchGetNameMap(ctx, operator, []string{outcome.OrganizationID})
		if err != nil {
			log.Error(ctx, "o.organizationService.BatchGetNameMap error",
				log.Err(err),
				log.Strings("userIDs", userIDs))
			return err
		}

		orgNameMap = orgNames

		return nil
	})

	// get outcome set
	g.Go(func() error {
		outcomeSets, err := o.outcomeSetDA.SearchSetsByOutcome(ctx, dbo.MustGetDB(ctx), []string{outcome.ID})
		if err != nil {
			log.Error(ctx, "o.outcomeSetDA.SearchSetsByOutcome error",
				log.Err(err),
				log.String("outcomeID", outcome.ID))
			return err
		}

		outcomeSetMap = outcomeSets
		return nil
	})

	// get user name map
	if len(userIDs) > 0 {
		g.Go(func() error {
			users, err := o.userService.BatchGetNameMap(ctx, operator, userIDs)
			if err != nil {
				log.Error(ctx, "s.userService.BatchGetMap error",
					log.Err(err),
					log.Strings("userIDs", userIDs))
				return err
			}

			userNameMap = users

			return nil
		})
	}

	// get locked outcome children
	if outcome.HasLocked() {
		g.Go(func() error {
			var lockedOutcomeChildren []*entity.Outcome
			lockedChildrenCondition := &da.OutcomeCondition{
				IncludeDeleted: false,
				SourceID:       sql.NullString{String: outcome.ID, Valid: true},
			}

			err := o.outcomeDA.Query(ctx, lockedChildrenCondition, &lockedOutcomeChildren)
			if err != nil {
				log.Error(ctx, "o.outcomeDA.Query",
					log.Err(err),
					log.Any("condition", lockedChildrenCondition))
				return err
			}

			for _, lockedOutcome := range lockedOutcomeChildren {
				lockedOutcomeChildrenMap[lockedOutcome.SourceID] = lockedOutcome
			}

			return nil
		})
	}

	// get program info
	if len(programIDs) > 0 {
		g.Go(func() error {
			programs, err := o.programService.BatchGetMap(ctx, operator, programIDs)
			if err != nil {
				log.Error(ctx, "o.programService.BatchGetMap error",
					log.Err(err),
					log.Strings("programIDs", programIDs))
				return err
			}

			programMap = programs

			return nil
		})
	}

	// get subject info
	if len(subjectIDs) > 0 {
		g.Go(func() error {
			subjects, err := o.subjectService.BatchGetMap(ctx, operator, subjectIDs)
			if err != nil {
				log.Error(ctx, "o.subjectService.BatchGetMap error",
					log.Err(err),
					log.Strings("subjectIDs", subjectIDs))
				return err
			}

			subjectMap = subjects

			return nil
		})
	}

	// get category info
	if len(categoryIDs) > 0 {
		g.Go(func() error {
			categories, err := o.categoryService.BatchGetMap(ctx, operator, categoryIDs)
			if err != nil {
				log.Error(ctx, "o.categoryService.BatchGetMap error",
					log.Err(err),
					log.Strings("categoryIDs", categoryIDs))
				return err
			}

			categoryMap = categories

			return nil
		})
	}

	// get sub category info
	if len(subCategoryIDs) > 0 {
		g.Go(func() error {
			subCategories, err := o.subCategoryService.BatchGetMap(ctx, operator, subCategoryIDs)
			if err != nil {
				log.Error(ctx, "o.subCategoryService.BatchGetMap error",
					log.Err(err),
					log.Strings("subCategoryIDs", subCategoryIDs))
				return err
			}

			subCategoryMap = subCategories

			return nil
		})
	}

	// get grade info
	if len(gradeIDs) > 0 {
		g.Go(func() error {
			grades, err := o.gradeService.BatchGetMap(ctx, operator, gradeIDs)
			if err != nil {
				log.Error(ctx, "o.gradeService.BatchGetMap error",
					log.Err(err),
					log.Strings("gradeIDs", gradeIDs))
				return err
			}

			gradeMap = grades

			return nil
		})
	}

	// get age info
	if len(ageIDs) > 0 {
		g.Go(func() error {
			ages, err := o.ageService.BatchGetMap(ctx, operator, ageIDs)
			if err != nil {
				log.Error(ctx, "o.ageService.BatchGetMap error",
					log.Err(err),
					log.Strings("ageIDs", ageIDs))
				return err
			}

			ageMap = ages

			return nil
		})
	}

	// TODO: bad database query
	// get milestones
	g.Go(func() error {
		milestoneOutcomes, err := o.milestoneOutcomeDA.SearchTx(ctx, dbo.MustGetDB(ctx),
			&da.MilestoneOutcomeCondition{
				OutcomeAncestor: sql.NullString{String: outcome.AncestorID, Valid: true},
			})
		if err != nil {
			log.Error(ctx, "o.milestoneOutcomeDA.SearchTx error",
				log.Err(err),
				log.Any("outcome", outcome))
			return err
		}

		milestoneIDs := make([]string, len(milestoneOutcomes))
		for i := range milestoneOutcomes {
			milestoneIDs[i] = milestoneOutcomes[i].MilestoneID
		}

		if len(milestoneIDs) == 0 {
			log.Debug(ctx, "milestone outcome relation record not found", log.Any("outcome", outcome))
			return nil
		}

		_, milestones, err := o.milestoneDA.Search(ctx, dbo.MustGetDB(ctx), &da.MilestoneCondition{
			IDs: dbo.NullStrings{Strings: milestoneIDs, Valid: true},
			// TODO: why use OutcomeStatus in milestong?
			Statuses: dbo.NullStrings{Strings: []string{entity.OutcomeStatusPublished, entity.OutcomeStatusDraft}, Valid: true},
			OrderBy:  da.OrderByMilestoneUpdatedAtDesc,
		})
		if err != nil {
			log.Error(ctx, "o.MilestoneDA.Search error",
				log.Strings("milestoneIDs", milestoneIDs))
			return err
		}

		for _, milestone := range milestones {
			if milestone.Status == entity.OutcomeStatusDraft && milestone.SourceID != milestone.ID {
				continue
			}
			// NML-1021
			// if milestones[i].Type == entity.GeneralMilestoneType && len(milestones) != 1 {
			// 	continue
			// }
			outcomeMilestones = append(outcomeMilestones, &Milestone{
				MilestoneID:   milestone.ID,
				MilestoneName: milestone.Name,
			})
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Error(ctx, "transformToOutcomeView error",
			log.Err(err))
		return nil, err
	}

	// fill org name
	if orgName, ok := orgNameMap[outcome.OrganizationID]; ok {
		result.OrganizationName = orgName
	} else {
		log.Error(ctx, "organization not found", log.String("orgID", outcome.OrganizationID))
	}

	// fill author name
	if userName, ok := userNameMap[outcome.AuthorID]; ok {
		result.AuthorName = userName
	} else {
		log.Error(ctx, "author not found", log.String("userID", outcome.AuthorID))
	}

	// fill outcome sets
	if outcomeSets, ok := outcomeSetMap[outcome.ID]; ok {
		for _, set := range outcomeSets {
			result.Sets = append(result.Sets, &OutcomeSetCreateView{
				SetID:   set.ID,
				SetName: set.Name,
			})
		}
	}

	// fill outcome milestones
	result.Milestones = outcomeMilestones

	if outcome.HasLocked() {
		if lockedOutcomeChildren, ok := lockedOutcomeChildrenMap[outcome.ID]; ok {
			result.LockedLocation = []string{string(lockedOutcomeChildren.PublishStatus)}
			result.LastEditedAt = lockedOutcomeChildren.CreateAt
			if userName, ok := userNameMap[outcome.LockedBy]; ok {
				result.LastEditedBy = userName
			} else {
				log.Error(ctx, "locked by user not found", log.String("userID", outcome.LockedBy))
			}
		} else {
			log.Error(ctx, "fill outcome locked info error", log.Any("outcome", outcome))
			return nil, ErrOutcomeInvalidData
		}
	}

	// fill program, subject, category, subCategory, grade, age
	for _, outcomeRelation := range outcomeRelations {
		switch outcomeRelation.RelationType {
		case entity.ProgramType:
			if program, ok := programMap[outcomeRelation.RelationID]; ok {
				result.Program = append(result.Program,
					Program{
						ProgramID:   program.ID,
						ProgramName: program.Name,
					})
			} else {
				log.Error(ctx, "program record not found", log.String("programID", outcomeRelation.RelationID))
			}
		case entity.SubjectType:
			if subject, ok := subjectMap[outcomeRelation.RelationID]; ok {
				result.Subject = append(result.Subject,
					Subject{
						SubjectID:   subject.ID,
						SubjectName: subject.Name,
					})
			} else {
				log.Error(ctx, "subject record not found", log.String("subjectID", outcomeRelation.RelationID))
			}
		case entity.CategoryType:
			if category, ok := categoryMap[outcomeRelation.RelationID]; ok {
				result.Developmental = append(result.Developmental,
					Developmental{
						DevelopmentalID:   category.ID,
						DevelopmentalName: category.Name,
					})
			} else {
				log.Error(ctx, "category record not found", log.String("categoryID", outcomeRelation.RelationID))
			}
		case entity.SubcategoryType:
			if subCategory, ok := subCategoryMap[outcomeRelation.RelationID]; ok {
				result.Skills = append(result.Skills,
					Skill{
						SkillID:   subCategory.ID,
						SkillName: subCategory.Name,
					})
			} else {
				log.Error(ctx, "subCategory record not found", log.String("subCategoryID", outcomeRelation.RelationID))
			}
		case entity.GradeType:
			if grade, ok := gradeMap[outcomeRelation.RelationID]; ok {
				result.Grade = append(result.Grade,
					Grade{
						GradeID:   grade.ID,
						GradeName: grade.Name,
					})
			} else {
				log.Error(ctx, "grade record not found", log.String("gradeID", outcomeRelation.RelationID))
			}
		case entity.AgeType:
			if age, ok := ageMap[outcomeRelation.RelationID]; ok {
				result.Age = append(result.Age,
					Age{
						AgeID:   age.ID,
						AgeName: age.Name,
					})
			} else {
				log.Error(ctx, "age record not found", log.String("ageID", outcomeRelation.RelationID))
			}
		}
	}

	return result, nil
}
