package model

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/pkg/errgroup"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrMilestoneInvalidData = errors.New("invalid milestone data")
)

var (
	_milestoneModel     IMilestoneModel
	_milestoneModelOnce sync.Once
)

type IMilestoneModel interface {
	Create(ctx context.Context, op *entity.Operator, milestone *entity.Milestone, outcomeAncestors []string) error
	Obtain(ctx context.Context, op *entity.Operator, milestoneID string) (*MilestoneDetailView, error)
	Update(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, milestone *entity.Milestone, outcomeIDs []string) error
	Delete(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, IDs []string) error
	Search(ctx context.Context, op *entity.Operator, condition *entity.MilestoneCondition) (*SearchMilestoneResponse, error)
	Occupy(ctx context.Context, op *entity.Operator, milestoneID string) (*entity.Milestone, error)
	// Deprecated
	// Publish(ctx context.Context, op *entity.Operator, IDs []string) error
	BulkPublish(ctx context.Context, op *entity.Operator, milestoneIDs []string) error
	BulkApprove(ctx context.Context, op *entity.Operator, milestoneIDs []string) error
	BulkReject(ctx context.Context, op *entity.Operator, milestoneIDs []string, reason string) error
	GenerateShortcode(ctx context.Context, op *entity.Operator) (string, error)

	CreateGeneral(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, orgID string) (*entity.Milestone, error)
	ObtainGeneral(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, orgID string) (*entity.Milestone, error)
	BindToGeneral(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) error

	ShortcodeProvider
}

type MilestoneModel struct {
	milestoneDA         da.IMilestoneDA
	milestoneRelationDA da.IMilestoneRelationDA
	milestoneOutcomeDA  da.IMilestoneOutcomeDA
	outcomeSetDA        da.IOutcomeSetDA
	outcomeRelationDA   da.IOutcomeRelationDA

	organizationService external.OrganizationServiceProvider
	userService         external.UserServiceProvider
	programService      external.ProgramServiceProvider
	subjectService      external.SubjectServiceProvider
	categoryService     external.CategoryServiceProvider
	subCategoryService  external.SubCategoryServiceProvider
	gradeService        external.GradeServiceProvider
	ageService          external.AgeServiceProvider
}

func GetMilestoneModel() IMilestoneModel {
	_milestoneModelOnce.Do(func() {
		_milestoneModel = &MilestoneModel{
			milestoneDA:         da.GetMilestoneDA(),
			milestoneRelationDA: da.GetMilestoneRelationDA(),
			milestoneOutcomeDA:  da.GetMilestoneOutcomeDA(),
			outcomeSetDA:        da.GetOutcomeSetDA(),
			outcomeRelationDA:   da.GetOutcomeRelationDA(),

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
	return _milestoneModel
}

func (m MilestoneModel) Create(ctx context.Context, op *entity.Operator, milestone *entity.Milestone, outcomeAncestors []string) error {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixShortcodeMute, entity.KindMileStone, op.OrgID)
	if err != nil {
		log.Error(ctx, "CreateMilestone: NewLock failed",
			log.Err(err),
			log.Any("op", op),
			log.Any("milestone", milestone))
		return err
	}
	locker.Lock()
	defer locker.Unlock()
	milestone.ID = utils.NewID()
	milestone.SourceID = milestone.ID
	milestone.AncestorID = milestone.ID
	milestone.LatestID = milestone.ID
	milestone.CreateAt = time.Now().Unix()
	milestone.UpdateAt = milestone.CreateAt
	milestone.Status = entity.OutcomeStatusDraft
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		exists, err := m.IsShortcodeExists(ctx, op, tx, milestone.AncestorID, milestone.Shortcode)
		if err != nil {
			log.Error(ctx, "CreateMilestone: IsShortcodeExists failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}
		if exists {
			return constant.ErrConflict
		}
		err = da.GetMilestoneDA().Create(ctx, tx, milestone)
		if err != nil {
			log.Error(ctx, "CreateMilestone: Create failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}

		length := len(outcomeAncestors)
		milestoneOutcomes := make([]*entity.MilestoneOutcome, length)
		currentTime := time.Now().Unix() + int64(length)
		for i := range outcomeAncestors {
			milestoneOutcome := entity.MilestoneOutcome{
				MilestoneID:     milestone.ID,
				OutcomeAncestor: outcomeAncestors[i],
				CreateAt:        currentTime,
				UpdateAt:        currentTime,
			}
			// milestoneOutcomes sort: first in last out
			milestoneOutcomes[length-1-i] = &milestoneOutcome

			currentTime--
		}

		_, err = da.GetMilestoneOutcomeDA().InsertInBatchesTx(ctx, tx, milestoneOutcomes, len(milestoneOutcomes))
		if err != nil {
			log.Error(ctx, "CreateMilestone: da.GetMilestoneOutcomeDA().InsertTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestoneOutcomes", milestoneOutcomes))
			return err
		}

		_, err = da.GetMilestoneRelationDA().InsertTx(ctx, tx, m.collectRelation(milestone))
		if err != nil {
			log.Error(ctx, "CreateMilestone: InsertTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}
		return nil
	})
	m.RemoveShortcode(ctx, op, milestone.Shortcode)
	return err
}

func (m MilestoneModel) Obtain(ctx context.Context, op *entity.Operator, milestoneID string) (*MilestoneDetailView, error) {
	var milestone *entity.Milestone
	err := m.milestoneDA.Get(ctx, milestoneID, &milestone)
	if err != nil {
		log.Error(ctx, "m.milestoneDA.Get error",
			log.Err(err),
			log.String("milestoneID", milestoneID))
		return nil, err
	}

	return m.transformToMilestoneDetailView(ctx, op, milestone)
}

func (m MilestoneModel) Update(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, milestone *entity.Milestone, outcomeAncestors []string) error {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixShortcodeMute, entity.KindMileStone, op.OrgID)
	if err != nil {
		log.Error(ctx, "Update: NewLock failed",
			log.Err(err),
			log.Any("op", op),
			log.Any("milestone", milestone))
		return err
	}
	locker.Lock()
	defer locker.Unlock()
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		oldMilestone, err := da.GetMilestoneDA().GetByID(ctx, tx, milestone.ID)
		if err != nil {
			log.Error(ctx, "Update: GetByID failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}
		if oldMilestone.HasLocked() {
			log.Debug(ctx, "Update: lockedBy user",
				log.Any("op", op),
				log.Any("milestone", oldMilestone))
			return &ErrContentAlreadyLocked{LockedBy: &external.User{
				ID: oldMilestone.LockedBy,
			}}
		}

		if oldMilestone.Status != entity.MilestoneStatusDraft && oldMilestone.Status != entity.MilestoneStatusRejected {
			log.Error(ctx, "Update: status not allowed edit",
				log.Any("op", op),
				log.Any("milestone", oldMilestone))
			return ErrInvalidPublishStatus
		}

		if !m.allowUpdateMilestone(ctx, op, perms, oldMilestone) {
			log.Warn(ctx, "Update: no permission",
				log.Any("op", op),
				log.Any("perms", perms),
				log.Any("milestone", oldMilestone))
			return constant.ErrOperateNotAllowed
		}

		if oldMilestone.Shortcode != milestone.Shortcode {
			exists, err := m.IsShortcodeExists(ctx, op, tx, oldMilestone.AncestorID, milestone.Shortcode)
			if err != nil {
				log.Error(ctx, "Update: isOccupied failed",
					log.Err(err),
					log.Any("op", op),
					log.Any("milestone", oldMilestone),
					log.Any("shortCode", milestone.Shortcode))
				return err
			}
			if exists {
				return constant.ErrConflict
			}
		}
		newMilsestone := m.updateMilestone(oldMilestone, milestone)
		_, err = da.GetMilestoneDA().UpdateTx(ctx, tx, newMilsestone)
		if err != nil {
			log.Error(ctx, "Update: Update failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", newMilsestone))
			return err
		}

		var needDeleteOutcomeMilestoneID []string
		length := len(outcomeAncestors)
		milestoneOutcomes := make([]*entity.MilestoneOutcome, length)
		currentTime := time.Now().Unix() + int64(length)
		for i := range outcomeAncestors {
			milestoneOutcome := entity.MilestoneOutcome{
				MilestoneID:     newMilsestone.ID,
				OutcomeAncestor: outcomeAncestors[i],
				CreateAt:        currentTime,
				UpdateAt:        currentTime,
			}
			// milestoneOutcomes sort: first in last out
			milestoneOutcomes[length-1-i] = &milestoneOutcome
			currentTime--
		}

		needDeleteOutcomeMilestoneID = append(needDeleteOutcomeMilestoneID, oldMilestone.ID)
		err = da.GetMilestoneOutcomeDA().DeleteTx(ctx, tx, needDeleteOutcomeMilestoneID)
		if err != nil {
			log.Error(ctx, "Update: DeleteTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", oldMilestone))
			return err
		}
		_, err = da.GetMilestoneOutcomeDA().InsertInBatchesTx(ctx, tx, milestoneOutcomes, len(milestoneOutcomes))
		if err != nil {
			log.Error(ctx, "Update: da.GetMilestoneOutcomeDA().InsertInBatchesTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestoneOutcomes", milestoneOutcomes))
			return err
		}

		err = da.GetMilestoneRelationDA().DeleteTx(ctx, tx, []string{oldMilestone.ID})
		if err != nil {
			log.Error(ctx, "Update: DeleteTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", oldMilestone))
			return err
		}
		_, err = da.GetMilestoneRelationDA().InsertTx(ctx, tx, m.collectRelation(newMilsestone))
		if err != nil {
			log.Error(ctx, "Update: InsertTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", newMilsestone))
		}
		return nil
	})
	return err
}

func (m MilestoneModel) Delete(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, IDs []string) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, milestones, err := da.GetMilestoneDA().Search(ctx, tx, &da.MilestoneCondition{
			IDs: dbo.NullStrings{Strings: IDs, Valid: true},
		})
		if err != nil {
			log.Error(ctx, "Delete: Search failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", IDs))
			return err
		}

		for _, ms := range milestones {
			if ms.HasLocked() {
				log.Error(ctx, "Delete: invalid lock status",
					log.Err(err),
					log.Any("milestone", ms))
				return NewErrContentAlreadyLocked(ctx, ms.LockedBy, op)
			}

			if !m.allowDeleteMilestone(ctx, op, perms, ms) {
				log.Warn(ctx, "Delete: no permission",
					log.Any("op", op),
					log.Any("perms", perms),
					log.Any("milestone", ms))
				return constant.ErrOperateNotAllowed
			}
		}

		err = da.GetMilestoneDA().BatchDelete(ctx, tx, IDs)
		if err != nil {
			log.Error(ctx, "Delete: BatchDelete failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestoneIDs", IDs))
			return err
		}

		err = da.GetMilestoneDA().BatchUnLock(ctx, tx, m.getParentIDs(ctx, milestones))
		if err != nil {
			log.Debug(ctx, "Delete: BatchUnLock failed",
				log.Any("op", op),
				log.Strings("ids", IDs),
				log.Any("milestone", milestones))
			return err
		}

		err = da.GetMilestoneOutcomeDA().DeleteTx(ctx, tx, IDs)
		if err != nil {
			log.Error(ctx, "Delete: DeleteTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", IDs))
			return err
		}

		err = da.GetMilestoneRelationDA().DeleteTx(ctx, tx, IDs)
		if err != nil {
			log.Error(ctx, "Delete: DeleteTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", IDs))
			return err
		}
		return nil
	})
	return err
}

func (m MilestoneModel) Search(ctx context.Context, op *entity.Operator, condition *entity.MilestoneCondition) (*SearchMilestoneResponse, error) {
	var milestones []*entity.Milestone
	daCondition := &da.MilestoneCondition{
		ID:          sql.NullString{String: condition.ID, Valid: condition.ID != ""},
		IDs:         dbo.NullStrings{Strings: condition.IDs, Valid: len(condition.IDs) > 0},
		Description: sql.NullString{String: condition.Description, Valid: condition.Description != ""},
		Name:        sql.NullString{String: condition.Name, Valid: condition.Name != ""},
		Shortcode:   sql.NullString{String: condition.Shortcode, Valid: condition.Shortcode != ""},
		SearchKey:   sql.NullString{String: condition.SearchKey, Valid: condition.SearchKey != ""},

		AuthorID:  sql.NullString{String: condition.AuthorID, Valid: condition.AuthorID != ""},
		AuthorIDs: dbo.NullStrings{Strings: condition.AuthorIDs, Valid: len(condition.AuthorIDs) > 0},

		OrganizationID: sql.NullString{String: condition.OrganizationID, Valid: condition.OrganizationID != ""},
		Status:         sql.NullString{String: condition.Status, Valid: condition.Status != ""},
		OrderBy:        da.NewMilestoneOrderBy(condition.OrderBy),
		Pager:          utils.GetDboPager(condition.Page, condition.PageSize),
	}
	total, err := m.milestoneDA.Page(ctx, daCondition, &milestones)
	if err != nil {
		log.Error(ctx, "m.milestoneDA.Page error",
			log.Err(err),
			log.Any("daCondition", daCondition))
		return nil, err
	}

	if len(milestones) == 0 {
		return &SearchMilestoneResponse{
			Total:      total,
			Milestones: []*SearchMilestoneView{},
		}, nil
	}

	searchMilestoneView, err := m.transformToSearchMilestoneView(ctx, op, milestones)
	if err != nil {
		log.Error(ctx, "m.transformToSearchMilestoneView error",
			log.Err(err),
			log.Any("milestones", milestones))
		return nil, err
	}

	return &SearchMilestoneResponse{
		Total:      total,
		Milestones: searchMilestoneView,
	}, nil
}

func (m MilestoneModel) Occupy(ctx context.Context, op *entity.Operator, milestoneID string) (*entity.Milestone, error) {
	var newVersion *entity.Milestone
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		milestone, err := da.GetMilestoneDA().GetByID(ctx, tx, milestoneID)
		if err != nil {
			log.Error(ctx, "Occupy: GetByID failed",
				log.Err(err),
				log.Any("op", op),
				log.String("milestone", milestoneID))
			return err
		}

		// NKL-1021
		// if ms.Type == entity.GeneralMilestoneType {
		// 	log.Warn(ctx, "Occupy: can not operate general milestone", log.Any("milestone", ms))
		// 	return constant.ErrOperateNotAllowed
		// }

		if milestone.LockedBy == op.UserID {
			copyVersion, err := m.getBySourceID(ctx, tx, milestone.ID)
			if err != nil {
				log.Debug(ctx, "Occupy: getBySourceID failed",
					log.Any("op", op),
					log.Any("milestone", milestone))
				return err
			}
			if copyVersion.Status == entity.MilestoneStatusDraft {
				newVersion = copyVersion
				return nil
			}
			log.Error(ctx, "Lock: copyVersion status not draft",
				log.Any("op", op),
				log.Any("copyVersion", copyVersion))
			return NewErrContentAlreadyLocked(ctx, milestone.LockedBy, op)
		}

		if milestone.HasLocked() {
			log.Warn(ctx, "Occupy: already locked",
				log.Any("milestone", milestone))
			return NewErrContentAlreadyLocked(ctx, milestone.LockedBy, op)
		}

		newVersion, err = m.copy(op, milestone)
		if err != nil {
			log.Error(ctx, "Occupy: Copy failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}

		milestone.LockedBy = op.UserID
		_, err = da.GetMilestoneDA().UpdateTx(ctx, tx, milestone)
		if err != nil {
			log.Error(ctx, "Occupy: Update failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}
		err = da.GetMilestoneDA().Create(ctx, tx, newVersion)
		if err != nil {
			log.Error(ctx, "Occupy: Create failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("newVersion", newVersion))
		}

		milestoneOutcomes, err := da.GetMilestoneOutcomeDA().SearchTx(ctx, tx, &da.MilestoneOutcomeCondition{
			MilestoneID: sql.NullString{String: milestoneID, Valid: true},
			OrderBy:     da.OrderByMilestoneOutcomeUpdatedAt,
		})
		if err != nil {
			log.Error(ctx, "Occupy: Search failed",
				log.Err(err),
				log.Any("op", op),
				log.String("milestone", milestoneID))
			return err
		}

		for i := range milestoneOutcomes {
			milestoneOutcomes[i].MilestoneID = newVersion.ID
			milestoneOutcomes[i].ID = 0
		}

		_, err = da.GetMilestoneOutcomeDA().InsertInBatchesTx(ctx, tx, milestoneOutcomes, len(milestoneOutcomes))
		if err != nil {
			log.Error(ctx, "Occupy: da.GetMilestoneOutcomeDA().InsertInBatchesTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestoneOutcomes", milestoneOutcomes))
			return err
		}

		var milestoneRelations []*entity.MilestoneRelation
		err = da.GetMilestoneRelationDA().QueryTx(ctx, tx, &da.MilestoneRelationCondition{
			MasterIDs:  dbo.NullStrings{Strings: []string{milestoneID}, Valid: true},
			MasterType: sql.NullString{String: string(entity.MilestoneType), Valid: true},
		}, &milestoneRelations)
		if err != nil {
			log.Error(ctx, "Occupy: Search failed",
				log.Err(err),
				log.Any("op", op))
			return err
		}

		for i := range milestoneRelations {
			milestoneRelations[i].MasterID = newVersion.ID
			milestoneRelations[i].ID = 0
		}

		_, err = da.GetMilestoneRelationDA().InsertTx(ctx, tx, milestoneRelations)
		if err != nil {
			log.Error(ctx, "Occupy: InsertTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestoneRelations", milestoneRelations))
			return err
		}
		return nil
	})
	return newVersion, err
}

// Deprecated: not implement approve process
// func (m MilestoneModel) Publish(ctx context.Context, op *entity.Operator, IDs []string) error {
// 	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
// 		_, milestones, err := da.GetMilestoneDA().Search(ctx, tx, &da.MilestoneCondition{
// 			IDs: dbo.NullStrings{Strings: IDs, Valid: true},
// 		})
// 		if err != nil {
// 			log.Error(ctx, "Publish: Search failed",
// 				log.Err(err),
// 				log.Any("op", op),
// 				log.Strings("milestone", IDs))
// 			return err
// 		}
// 		publishIDs, hideIDs, ancestorLatest, err := m.canPublish(ctx, milestones)
// 		if err != nil {
// 			log.Error(ctx, "Publish: canPublish failed",
// 				log.Err(err),
// 				log.Any("op", op),
// 				log.Strings("milestone", IDs))
// 			return err
// 		}
// 		err = da.GetMilestoneDA().BatchPublish(ctx, tx, publishIDs)
// 		if err != nil {
// 			log.Error(ctx, "Publish: BatchPublish failed",
// 				log.Err(err),
// 				log.Any("op", op),
// 				log.Strings("publish", publishIDs))
// 			return err
// 		}
// 		err = da.GetMilestoneDA().BatchHide(ctx, tx, hideIDs)
// 		if err != nil {
// 			log.Error(ctx, "Publish: BatchHide failed",
// 				log.Err(err),
// 				log.Any("op", op),
// 				log.Strings("hide", hideIDs))
// 			return err
// 		}
// 		err = da.GetMilestoneOutcomeDA().DeleteTx(ctx, tx, hideIDs)
// 		if err != nil {
// 			log.Error(ctx, "Publish: DeleteTx failed",
// 				log.Any("op", op),
// 				log.Strings("hide", hideIDs))
// 			return err
// 		}
// 		err = da.GetMilestoneDA().BatchUpdateLatest(ctx, tx, ancestorLatest)
// 		if err != nil {
// 			log.Error(ctx, "Publish: BatchUpdateLatest failed",
// 				log.Err(err),
// 				log.Any("op", op),
// 				log.Any("ancestor", ancestorLatest))
// 			return err
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (m MilestoneModel) BulkPublish(ctx context.Context, op *entity.Operator, milestoneIDs []string) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		condition := &da.MilestoneCondition{
			IDs: dbo.NullStrings{Strings: milestoneIDs, Valid: true},
		}
		_, milestones, err := da.GetMilestoneDA().Search(ctx, tx, condition)
		if err != nil {
			log.Error(ctx, "BulkPublish: Search failed",
				log.Any("op", op),
				log.Any("condition", condition))
			return err
		}

		if len(milestones) == 0 {
			log.Warn(ctx, "BulkPublish: SearchOutcome failed",
				log.Any("op", op),
				log.Any("condition", condition))
			return ErrResourceNotFound
		}

		for _, ms := range milestones {
			if ms.AuthorID != op.UserID {
				log.Warn(ctx, "BulkPublish: must by self",
					log.Any("op", op),
					log.Any("milestone", ms))
				return ErrNoAuth
			}

			ok := ms.SetStatus(entity.MilestoneStatusPending)
			if !ok {
				log.Error(ctx, "BulkPublish: SetStatus failed",
					log.Any("milestone", ms))
				return ErrInvalidContentStatusToPublish
			}

			_, err = da.GetMilestoneDA().UpdateTx(ctx, tx, ms)
			if err != nil {
				log.Error(ctx, "BulkPublish: UpdateOutcome failed",
					log.Any("milestone", ms))
				return err
			}
		}
		return nil
	})
	return err
}

func (m MilestoneModel) BulkApprove(ctx context.Context, op *entity.Operator, milestoneIDs []string) error {
	for _, msID := range milestoneIDs {
		locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeReview, msID)
		if err != nil {
			log.Error(ctx, "BulkApprove: NewLock failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone_id", msID))
			return err
		}
		locker.Lock()
		defer locker.Unlock()
	}
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, milestones, err := da.GetMilestoneDA().Search(ctx, tx, &da.MilestoneCondition{
			IDs: dbo.NullStrings{Strings: milestoneIDs, Valid: true},
		})
		if len(milestones) == 0 {
			log.Warn(ctx, "BulkApprove: Search failed",
				log.Any("op", op),
				log.Strings("milestoneIDs", milestoneIDs))
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "BulkApprove: Search failed",
				log.Any("op", op),
				log.Strings("milestoneIDs", milestoneIDs))
			return err
		}
		for _, milestone := range milestones {
			ok := milestone.SetStatus(entity.MilestoneStatusPublished)
			if !ok {
				log.Error(ctx, "BulkApprove: SetStatus failed",
					log.Any("op", op),
					log.Any("milestone", milestone))
				return ErrInvalidPublishStatus
			}

			_, err = da.GetMilestoneDA().UpdateTx(ctx, tx, milestone)
			if err != nil {
				log.Error(ctx, "BulkApprove: Update failed",
					log.Any("op", op),
					log.Any("milestone", milestone))
				return err
			}

			if !milestone.IsAncestor() {
				parent, err := da.GetMilestoneDA().GetByID(ctx, tx, milestone.SourceID)
				if err != nil {
					log.Error(ctx, "BulkApprove: GetByID failed",
						log.Any("milestone", milestone))
					return err
				}
				parent.LockedBy = constant.LockedByNoBody
				ok = parent.SetStatus(entity.OutcomeStatusHidden)
				if !ok {
					log.Error(ctx, "BulkApprove: SetStatus failed",
						log.Any("milestone", milestone))
					return ErrInvalidPublishStatus
				}
				_, err = da.GetMilestoneDA().UpdateTx(ctx, tx, parent)
				if err != nil {
					log.Error(ctx, "BulkApprove: Update failed",
						log.Any("milestone", milestone))
					return err
				}
			}

			if !milestone.IsLatest() {
				err = da.GetMilestoneDA().UpdateLatest(ctx, tx, milestone.AncestorID, milestone.ID)
				if err != nil {
					log.Error(ctx, "BulkApprove: UpdateLatest failed",
						log.Any("milestone", milestone))
					return err
				}
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

func (m MilestoneModel) BulkReject(ctx context.Context, op *entity.Operator, milestoneIDs []string, reason string) error {
	for _, msID := range milestoneIDs {
		locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeReview, msID)
		if err != nil {
			log.Error(ctx, "BulkReject: NewLock failed",
				log.Err(err),
				log.Any("op", op),
				log.String("msID", msID))
			return err
		}
		locker.Lock()
		defer locker.Unlock()
	}
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, milestones, err := da.GetMilestoneDA().Search(ctx, tx, &da.MilestoneCondition{
			IDs: dbo.NullStrings{Strings: milestoneIDs, Valid: true},
		})
		if len(milestones) == 0 {
			log.Warn(ctx, "BulkReject: Search failed",
				log.Any("op", op),
				log.Strings("milestoneIDs", milestoneIDs))
			return ErrResourceNotFound
		}
		if err != nil {
			log.Error(ctx, "BulkReject: Search failed",
				log.Any("op", op),
				log.Strings("milestoneIDs", milestoneIDs))
			return err
		}

		for _, milestone := range milestones {
			milestone.RejectReason = reason
			ok := milestone.SetStatus(entity.OutcomeStatusRejected)
			if !ok {
				log.Error(ctx, "BulkReject: SetStatus failed",
					log.Any("op", op),
					log.Any("milestone", milestone))
				return ErrInvalidPublishStatus
			}
			_, err = da.GetMilestoneDA().UpdateTx(ctx, tx, milestone)
			if err != nil {
				log.Error(ctx, "BulkReject: Update failed",
					log.Any("op", op),
					log.Any("milestone", milestone))
				return err
			}
		}
		return nil
	})

	return err
}

func (m MilestoneModel) GenerateShortcode(ctx context.Context, op *entity.Operator) (string, error) {
	var shortcode string
	var index int
	cursor, err := m.Current(ctx, op)
	if err != nil {
		log.Debug(ctx, "GenerateShortcode: Current failed",
			log.Any("op", op),
			log.Int("cursor", cursor))
		return "", err
	}
	shortcodeModel := GetShortcodeModel()
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		index, shortcode, err = shortcodeModel.generate(ctx, op, tx, cursor+1, m)
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
	err = m.Cache(ctx, op, index, shortcode)
	return shortcode, err
}

func (m MilestoneModel) CreateGeneral(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, orgID string) (*entity.Milestone, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixGeneralMilestoneMute, entity.KindMileStone, op.OrgID)
	if err != nil {
		log.Error(ctx, "CreateGeneral: NewLock failed",
			log.Err(err),
			log.Any("op", op))
		return nil, err
	}
	locker.Lock()
	defer locker.Unlock()
	if orgID == "" {
		orgID = op.OrgID
	}
	general, err := m.ObtainGeneral(ctx, op, tx, orgID)
	if err != nil && err != constant.ErrRecordNotFound {
		log.Debug(ctx, "CreateGeneral: ObtainGeneral failed", log.Any("op", op))
		return nil, err
	}
	if general != nil {
		return general, nil
	}
	general = m.buildGeneral(ctx, op)
	err = da.GetMilestoneDA().Create(ctx, tx, general)
	if err != nil {
		log.Debug(ctx, "CreateGeneral: Create failed", log.Any("op", op),
			log.Any("milestone", general))
		return nil, err
	}
	return general, nil
}

func (m MilestoneModel) ObtainGeneral(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, orgID string) (*entity.Milestone, error) {
	_, milestones, err := da.GetMilestoneDA().Search(ctx, tx, &da.MilestoneCondition{
		OrganizationID: sql.NullString{String: orgID, Valid: true},
		Type:           sql.NullString{String: string(entity.GeneralMilestoneType), Valid: true},
	})
	if err != nil {
		log.Error(ctx, "ObtainGeneral:Search failed",
			log.Any("op", op),
			log.String("org", orgID))
		return nil, err
	}

	if len(milestones) == 0 {
		log.Error(ctx, "ObtainGeneral: not found",
			log.Any("op", op),
			log.String("org", orgID))
		return nil, constant.ErrRecordNotFound
	}

	if len(milestones) > 1 {
		log.Error(ctx, "ObtainGeneral: should only one",
			log.Any("op", op),
			log.String("org", orgID))
		return nil, constant.ErrInternalServer
	}
	return milestones[0], nil
}

func (m MilestoneModel) BindToGeneral(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) error {
	general, err := m.ObtainGeneral(ctx, op, tx, outcome.OrganizationID)
	if err != nil && err != constant.ErrRecordNotFound {
		log.Error(ctx, "BindToGeneral: ObtainGeneral failed",
			log.Any("op", op),
			log.Any("outcome", outcome))
		return err
	}
	if err == constant.ErrRecordNotFound {
		general, err = m.CreateGeneral(ctx, op, tx, outcome.OrganizationID)
		if err != nil {
			log.Error(ctx, "BindToGeneral: CreateGeneral failed",
				log.Any("op", op),
				log.String("org", outcome.OrganizationID))
			return err
		}
	} else {
		milestoneOutcomes, err := da.GetMilestoneOutcomeDA().SearchTx(ctx, tx, &da.MilestoneOutcomeCondition{
			MilestoneID:     sql.NullString{String: general.ID, Valid: true},
			OutcomeAncestor: sql.NullString{String: outcome.AncestorID, Valid: true},
		})
		if err != nil {
			log.Error(ctx, "BindToGeneral: SearchTx failed",
				log.Any("op", op),
				log.Any("outcome", outcome))
			return err
		}
		if len(milestoneOutcomes) != 0 {
			return nil
		}
	}
	now := time.Now().Unix()
	milestoneOutcome := &entity.MilestoneOutcome{
		MilestoneID:     general.ID,
		OutcomeAncestor: outcome.AncestorID,
		CreateAt:        now,
		UpdateAt:        now,
	}

	_, err = da.GetMilestoneOutcomeDA().InsertTx(ctx, tx, milestoneOutcome)
	if err != nil {
		log.Error(ctx, "BindToGeneral: da.GetMilestoneOutcomeDA().InsertTx failed",
			log.Any("op", op),
			log.Any("general", general),
			log.Any("milestoneOutcome", milestoneOutcome))
		return err
	}
	return nil
}

// implement ShortcodeProvider interface
func (m MilestoneModel) Current(ctx context.Context, op *entity.Operator) (int, error) {
	return da.GetShortcodeRedis(ctx).Get(ctx, op, string(entity.KindMileStone))
}

func (m MilestoneModel) Intersect(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, shortcodes []string) (map[string]bool, error) {
	_, milestones, err := da.GetMilestoneDA().Search(ctx, tx, &da.MilestoneCondition{
		Shortcodes:     dbo.NullStrings{Strings: shortcodes, Valid: true},
		OrganizationID: sql.NullString{String: op.OrgID, Valid: true},
		OrderBy:        da.OrderByMilestoneShortcode,
	})
	if err != nil {
		log.Debug(ctx, "Intersect: Search failed",
			log.Any("op", op),
			log.Strings("shortcode", shortcodes))
		return nil, err
	}
	mapShortcode := make(map[string]bool)
	for i := range milestones {
		mapShortcode[milestones[i].Shortcode] = true
	}
	return mapShortcode, nil
}

func (m MilestoneModel) IsShortcodeExists(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, ancestor string, shortcode string) (bool, error) {
	_, milestones, err := da.GetMilestoneDA().Search(ctx, tx, &da.MilestoneCondition{
		OrganizationID: sql.NullString{String: op.OrgID, Valid: true},
		Shortcode:      sql.NullString{String: shortcode, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "IsShortcodeExists: Search failed",
			log.String("org", op.OrgID),
			log.String("shortcode", shortcode))
		return false, err
	}
	for i := range milestones {
		if ancestor != milestones[i].AncestorID {
			return true, nil
		}
	}
	return false, nil
}

func (m MilestoneModel) IsShortcodeCached(ctx context.Context, op *entity.Operator, shortcode string) (bool, error) {
	exists, err := da.GetShortcodeRedis(ctx).IsCached(ctx, op, string(entity.KindMileStone), shortcode)
	if err != nil {
		log.Error(ctx, "IsCached: redis access failed",
			log.Err(err),
			log.Any("op", op),
			log.String("shortcode", shortcode))
		return false, err
	}
	return exists, nil
}

func (m MilestoneModel) RemoveShortcode(ctx context.Context, op *entity.Operator, shortcode string) error {
	err := da.GetShortcodeRedis(ctx).Remove(ctx, op, string(entity.KindMileStone), shortcode)
	if err != nil {
		log.Error(ctx, "RemoveShortcode: redis access failed",
			log.Err(err),
			log.Any("op", op),
			log.String("shortcode", shortcode))
		return err
	}
	return nil
}

func (m MilestoneModel) Cache(ctx context.Context, op *entity.Operator, cursor int, shortcode string) error {
	err := da.GetShortcodeRedis(ctx).Cache(ctx, op, string(entity.KindMileStone), cursor, shortcode)
	if err != nil {
		log.Error(ctx, "Cache: redis access failed",
			log.Err(err),
			log.Any("op", op),
			log.Int("cursor", cursor),
			log.String("shortcode", shortcode))
		return err
	}
	return nil
}

func (m MilestoneModel) ShortcodeLength() int {
	return constant.ShortcodeShowLength
}

func (m MilestoneModel) canBeDeleted(ctx context.Context, milestones []*entity.Milestone, perms map[external.PermissionName]bool) ([]string, error) {
	var milestoneIDs []string
	for i := range milestones {
		if milestones[i].HasLocked() {
			log.Warn(ctx, "canBeDeleted: locked",
				log.Any("milestone", milestones[i]))
			return nil, &ErrContentAlreadyLocked{LockedBy: &external.User{
				ID: milestones[i].LockedBy,
			}}
		}

		switch milestones[i].Status {
		case entity.MilestoneStatusDraft, entity.MilestoneStatusHidden:

			if !perms[external.DeleteUnpublishedMilestone] {
				log.Warn(ctx, "canBeDeleted: has no perm",
					log.Any("perms", perms),
					log.Any("milestone", milestones[i]))
				return nil, constant.ErrOperateNotAllowed
			}
		case entity.OutcomeStatusPublished:
			if !perms[external.DeletePublishedMilestone] {
				log.Warn(ctx, "canBeDeleted: has no perm",
					log.Any("perms", perms),
					log.Any("milestone", milestones[i]))
				return nil, constant.ErrOperateNotAllowed
			}
			// NKL-1021
			// if milestones[i].Type == entity.GeneralMilestoneType {
			// 	log.Warn(ctx, "canBeDeleted: can not operate general milestone",
			// 		log.Any("milestone", milestones[i]))
			// 	return nil, constant.ErrOperateNotAllowed
			// }
		default:
			log.Warn(ctx, "canBeDeleted: status not allowed",
				log.Any("milestone", milestones[i]))
			return nil, constant.ErrOperateNotAllowed
		}

		if milestones[i].DeleteAt == 0 {
			milestoneIDs = append(milestoneIDs, milestones[i].ID)
		}
	}
	return milestoneIDs, nil
}

func (m MilestoneModel) getParentIDs(ctx context.Context, milestones []*entity.Milestone) (parentIDs []string) {
	for i := range milestones {
		if !milestones[i].IsAncestor() && milestones[i].SourceID != "" {
			parentIDs = append(parentIDs, milestones[i].SourceID)
		}
	}

	return
}

func (m MilestoneModel) getBySourceID(ctx context.Context, tx *dbo.DBContext, sourceID string) (*entity.Milestone, error) {
	_, res, err := da.GetMilestoneDA().Search(ctx, tx, &da.MilestoneCondition{
		SourceID: sql.NullString{String: sourceID, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "getBySourceID: failed",
			log.String("source_id", sourceID))
		return nil, err
	}

	// TODO: optimisation logic
	for _, v := range res {
		if v.SourceID != v.ID {
			return v, nil
		}
	}

	log.Debug(ctx, "getBySourceID: error",
		log.String("source_id", sourceID),
		log.Any("milestones", res))
	return nil, constant.ErrInternalServer
}

func (m MilestoneModel) canPublish(ctx context.Context, milestones []*entity.Milestone) (publishIDs, hideIDs []string, ancestorLatest map[string]string, err error) {
	ancestorLatest = make(map[string]string)
	for i := range milestones {
		switch milestones[i].Status {
		case entity.OutcomeStatusDraft:
			publishIDs = append(publishIDs, milestones[i].ID)
			if milestones[i].SourceID != milestones[i].ID {
				hideIDs = append(hideIDs, milestones[i].SourceID)
			}
			ancestorLatest[milestones[i].AncestorID] = milestones[i].ID
		case entity.OutcomeStatusPublished:
			log.Info(ctx, "canPublish: do nothing",
				log.Any("milestone", milestones[i]))
		default:
			log.Warn(ctx, "canPublish: status not allowed",
				log.Any("milestone", milestones[i]))
			err = constant.ErrOperateNotAllowed
			return
		}
	}
	return
}

func (m MilestoneModel) buildGeneral(ctx context.Context, op *entity.Operator) *entity.Milestone {
	ID := utils.NewID()
	now := time.Now().Unix()
	general := &entity.Milestone{
		ID:   ID,
		Name: entity.GeneralMilestoneName,
		//Shortcode:      entity.GeneralMilestoneShortcode,
		OrganizationID: op.OrgID,
		Type:           entity.GeneralMilestoneType,
		Status:         entity.OutcomeStatusPublished,
		AncestorID:     ID,
		SourceID:       ID,
		LatestID:       ID,
		CreateAt:       now,
		UpdateAt:       now,
	}
	return general
}

func (m MilestoneModel) collectRelation(ms *entity.Milestone) []*entity.MilestoneRelation {
	relations := make([]*entity.MilestoneRelation, 0, len(ms.Programs)+len(ms.Subjects)+len(ms.Categories)+len(ms.Subcategories)+len(ms.Grades)+len(ms.Ages))
	for i := range ms.Programs {
		relation := entity.MilestoneRelation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Programs[i],
			RelationType: entity.ProgramType,
		}
		relations = append(relations, &relation)
	}

	for i := range ms.Subjects {
		relation := entity.MilestoneRelation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Subjects[i],
			RelationType: entity.SubjectType,
		}
		relations = append(relations, &relation)
	}

	for i := range ms.Categories {
		relation := entity.MilestoneRelation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Categories[i],
			RelationType: entity.CategoryType,
		}
		relations = append(relations, &relation)
	}

	for i := range ms.Subcategories {
		relation := entity.MilestoneRelation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Subcategories[i],
			RelationType: entity.SubcategoryType,
		}
		relations = append(relations, &relation)
	}

	for i := range ms.Grades {
		relation := entity.MilestoneRelation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Grades[i],
			RelationType: entity.GradeType,
		}
		relations = append(relations, &relation)
	}

	for i := range ms.Ages {
		relation := entity.MilestoneRelation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Ages[i],
			RelationType: entity.AgeType,
		}
		relations = append(relations, &relation)
	}
	return relations
}

func (m MilestoneModel) fillRelation(ms *entity.Milestone, relations []*entity.MilestoneRelation) {

	for i := range relations {
		switch relations[i].RelationType {
		case entity.ProgramType:
			ms.Programs = append(ms.Programs, relations[i].RelationID)
		case entity.SubjectType:
			ms.Subjects = append(ms.Subjects, relations[i].RelationID)
		case entity.CategoryType:
			ms.Categories = append(ms.Categories, relations[i].RelationID)
		case entity.SubcategoryType:
			ms.Subcategories = append(ms.Subcategories, relations[i].RelationID)
		case entity.GradeType:
			ms.Grades = append(ms.Grades, relations[i].RelationID)
		case entity.AgeType:
			ms.Ages = append(ms.Ages, relations[i].RelationID)
		}
	}
}

func (m *MilestoneModel) copy(op *entity.Operator, ms *entity.Milestone) (*entity.Milestone, error) {
	if ms.Status == entity.MilestoneStatusHidden {
		return nil, constant.ErrOutOfDate
	}
	if ms.Status != entity.MilestoneStatusPublished {
		return nil, constant.ErrOperateNotAllowed
	}
	now := time.Now().Unix()
	milestone := &entity.Milestone{
		ID:             utils.NewID(),
		Name:           ms.Name,
		Shortcode:      ms.Shortcode,
		OrganizationID: op.OrgID,
		AuthorID:       op.UserID,
		Type:           ms.Type,
		Description:    ms.Description,
		LoCounts:       ms.LoCounts,

		Status: entity.OutcomeStatusDraft,

		AncestorID: ms.AncestorID,
		SourceID:   ms.ID,
		LatestID:   ms.LatestID,
		CreateAt:   now,
		UpdateAt:   now,
	}

	return milestone, nil
}

func (m *MilestoneModel) updateMilestone(old *entity.Milestone, new *entity.Milestone) *entity.Milestone {
	milestone := *old
	milestone.Name = new.Name
	milestone.Shortcode = new.Shortcode
	milestone.Description = new.Description
	milestone.Programs = new.Programs
	milestone.Subjects = new.Subjects
	milestone.Categories = new.Categories
	milestone.Subcategories = new.Subcategories
	milestone.Grades = new.Grades
	milestone.Ages = new.Ages
	milestone.Status = entity.MilestoneStatusDraft
	return &milestone
}

func (m *MilestoneModel) allowUpdateMilestone(ctx context.Context, operator *entity.Operator, perms map[external.PermissionName]bool, milestone *entity.Milestone) bool {
	if perms[external.EditUnpublishedMilestone] && milestone.Status != entity.MilestoneStatusPublished {
		return true
	}

	if (perms[external.EditMyUnpublishedMilestone] && milestone.AuthorID == operator.UserID) &&
		(milestone.Status != entity.MilestoneStatusPublished && milestone.Status != entity.MilestoneStatusPending) {
		return true
	}

	if perms[external.EditPublishedMilestone] && milestone.Status == entity.MilestoneStatusPublished {
		return true
	}

	if perms[external.EditMyUnpublishedMilestone] &&
		milestone.Status != entity.MilestoneStatusPublished &&
		milestone.AuthorID == operator.UserID {
		return true
	}

	return false
}

func (m *MilestoneModel) allowDeleteMilestone(ctx context.Context, operator *entity.Operator, perms map[external.PermissionName]bool, milestone *entity.Milestone) bool {
	if (milestone.Status == entity.MilestoneStatusDraft || milestone.Status == entity.MilestoneStatusRejected) &&
		perms[external.DeleteUnpublishedMilestone] {
		return true
	}

	if (milestone.Status == entity.MilestoneStatusDraft || milestone.Status == entity.MilestoneStatusRejected) &&
		perms[external.DeleteMyUnpublishedMilestone] && milestone.AuthorID == operator.UserID {
		return true
	}

	if milestone.Status == entity.MilestoneStatusPending && perms[external.DeleteOrgPendingMilestone] && milestone.OrganizationID == operator.OrgID {
		return true
	}

	if milestone.Status == entity.MilestoneStatusPending && perms[external.DeleteMyPendingMilestone] && milestone.AuthorID == operator.UserID {
		return true
	}

	if milestone.Status == entity.MilestoneStatusPublished && perms[external.DeletePublishedMilestone] {
		return true
	}

	return false
}

func (m *MilestoneModel) transformToSearchMilestoneView(ctx context.Context, operator *entity.Operator, milestones []*entity.Milestone) ([]*SearchMilestoneView, error) {
	result := make([]*SearchMilestoneView, len(milestones))
	resultMap := make(map[string]*SearchMilestoneView, len(milestones))
	milestoneIDs := make([]string, len(milestones))
	var lockedMilestoneIDs []string
	var userIDs []string
	for i, milestone := range milestones {
		result[i] = &SearchMilestoneView{
			MilestoneID: milestone.ID,
			Name:        milestone.Name,
			Shortcode:   milestone.Shortcode,
			Type:        milestone.Type,
			Status:      string(milestone.Status),
			Program:     []*Program{},
			Category:    []*Category{},
			LockedBy:    milestone.LockedBy,
			CreateAt:    milestone.CreateAt,
		}

		if milestone.HasLocked() {
			userIDs = append(userIDs, milestone.LockedBy)
			lockedMilestoneIDs = append(lockedMilestoneIDs, milestone.ID)
		}

		userIDs = append(userIDs, milestone.AuthorID)
		milestoneIDs[i] = milestone.ID
		resultMap[milestone.ID] = result[i]
	}

	var milestoneRelations []*entity.MilestoneRelation
	var programIDs []string
	var categoryIDs []string

	err := m.milestoneRelationDA.Query(ctx, &da.MilestoneRelationCondition{
		MasterIDs: dbo.NullStrings{
			Strings: milestoneIDs,
			Valid:   true,
		},
	}, &milestoneRelations)
	if err != nil {
		log.Error(ctx, "m.milestoneRelationDA.Query error",
			log.Err(err),
			log.Strings("milestoneIDs", milestoneIDs))
		return nil, err
	}

	for _, milestoneRelation := range milestoneRelations {
		switch milestoneRelation.RelationType {
		case entity.ProgramType:
			programIDs = append(programIDs, milestoneRelation.RelationID)
		case entity.CategoryType:
			categoryIDs = append(categoryIDs, milestoneRelation.RelationID)
		}
	}

	programIDs = utils.StableSliceDeduplication(programIDs)
	categoryIDs = utils.StableSliceDeduplication(categoryIDs)

	g := new(errgroup.Group)
	var userNameMap map[string]string
	var programNameMap map[string]string
	var categoryNameMap map[string]string
	var outcomeCountMap map[string]int
	lockedMilestoneChildrenMap := make(map[string]*entity.Milestone)

	// get user name map
	if len(userIDs) > 0 {
		g.Go(func() error {
			userNames, err := m.userService.BatchGetNameMap(ctx, operator, userIDs)
			if err != nil {
				log.Error(ctx, "m.userService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("userIDs", userIDs))
				return err
			}

			userNameMap = userNames

			return nil
		})
	}

	// get program info
	if len(programIDs) > 0 {
		g.Go(func() error {
			programNames, err := m.programService.BatchGetNameMap(ctx, operator, programIDs)
			if err != nil {
				log.Error(ctx, "m.programService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("programIDs", programIDs))
				return err
			}

			programNameMap = programNames

			return nil
		})
	}

	// get category info
	if len(categoryIDs) > 0 {
		g.Go(func() error {
			categoryNames, err := m.categoryService.BatchGetNameMap(ctx, operator, categoryIDs)
			if err != nil {
				log.Error(ctx, "m.categoryService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("categoryIDs", categoryIDs))
				return err
			}

			categoryNameMap = categoryNames

			return nil
		})
	}

	// TODO: bad database query
	// get outcomes count
	g.Go(func() error {
		countMap, err := m.milestoneOutcomeDA.BatchCountTx(ctx, dbo.MustGetDB(ctx),
			[]string{}, milestoneIDs)
		if err != nil {
			log.Error(ctx, "m.milestoneOutcomeDA.BatchCountTx", log.Err(err),
				log.Strings("milestoneIDs", milestoneIDs))
			return err
		}

		outcomeCountMap = countMap

		return nil
	})

	// get locked milestone children
	if len(lockedMilestoneIDs) > 0 {
		g.Go(func() error {
			var lockedMilestoneChildren []*entity.Milestone
			lockedChildrenCondition := &da.MilestoneCondition{
				IncludeDeleted: false,
				SourceIDs:      dbo.NullStrings{Strings: lockedMilestoneIDs, Valid: true},
			}

			err := m.milestoneDA.Query(ctx, lockedChildrenCondition, &lockedMilestoneChildren)
			if err != nil {
				log.Error(ctx, "m.milestoneDA.Query error",
					log.Err(err),
					log.Any("condition", lockedChildrenCondition))
				return err
			}

			for _, lockedMilestone := range lockedMilestoneChildren {
				lockedMilestoneChildrenMap[lockedMilestone.SourceID] = lockedMilestone
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		log.Error(ctx, "transformToSearchMilestoneView error",
			log.Err(err))
		return nil, err
	}

	for i, milestone := range milestones {
		milestoneView := result[i]

		// fill author name
		if userName, ok := userNameMap[milestone.AuthorID]; ok {
			milestoneView.Author = AuthorView{
				AuthorID:   milestone.AuthorID,
				AuthorName: userName,
			}
		} else {
			log.Error(ctx, "author not found", log.String("userID", milestone.AuthorID))
		}

		// fill outcome count
		if outcomeCount, ok := outcomeCountMap[milestone.ID]; ok {
			milestoneView.OutcomeCount = outcomeCount
		} else {
			log.Debug(ctx, "milestone outcome not found", log.Any("milestone", milestone))
		}

		if milestone.HasLocked() {
			if lockedMilestoneChildren, ok := lockedMilestoneChildrenMap[milestone.ID]; ok {
				milestoneView.LockedLocation = []string{string(lockedMilestoneChildren.Status)}
				milestoneView.LastEditedAt = lockedMilestoneChildren.CreateAt
				if userName, ok := userNameMap[milestone.LockedBy]; ok {
					milestoneView.LastEditedBy = userName
				} else {
					log.Error(ctx, "locked by user not found", log.String("userID", milestone.LockedBy))
				}
			} else {
				log.Error(ctx, "fill milestone locked info error", log.Any("milestone", milestone))
				return nil, ErrMilestoneInvalidData
			}
		}
	}

	// fill program and category
	for _, milestoneRelation := range milestoneRelations {
		if milestoneView, ok := resultMap[milestoneRelation.MasterID]; ok {
			switch milestoneRelation.RelationType {
			case entity.ProgramType:
				if programName, ok := programNameMap[milestoneRelation.RelationID]; ok {
					milestoneView.Program = append(milestoneView.Program,
						&Program{
							ProgramID:   milestoneRelation.RelationID,
							ProgramName: programName,
						})
				} else {
					log.Error(ctx, "program record not found", log.String("programID", milestoneRelation.RelationID))
				}
			case entity.CategoryType:
				if categoryName, ok := categoryNameMap[milestoneRelation.RelationID]; ok {
					milestoneView.Category = append(milestoneView.Category,
						&Category{
							CategoryID:   milestoneRelation.RelationID,
							CategoryName: categoryName,
						})
				} else {
					log.Error(ctx, "category record not found", log.String("categoryID", milestoneRelation.RelationID))
				}
			}
		}
	}

	return result, nil
}

func (m *MilestoneModel) transformToMilestoneDetailView(ctx context.Context, operator *entity.Operator, milestone *entity.Milestone) (*MilestoneDetailView, error) {
	result := &MilestoneDetailView{
		MilestoneID:  milestone.ID,
		Name:         milestone.Name,
		Shortcode:    milestone.Shortcode,
		Description:  milestone.Description,
		Type:         milestone.Type,
		Status:       string(milestone.Status),
		RejectReason: milestone.RejectReason,
		// init zero value
		Program:     []*Program{},
		Subject:     []*Subject{},
		Category:    []*Category{},
		SubCategory: []*SubCategory{},
		Age:         []*Age{},
		Grade:       []*Grade{},
		Outcomes:    []*MilestoneOutcomeView{},
		CreateAt:    milestone.CreateAt,
	}

	userIDs := []string{milestone.AuthorID}
	var milestoneRelations []*entity.MilestoneRelation
	var programIDs []string
	var subjectIDs []string
	var categoryIDs []string
	var subCategoryIDs []string
	var ageIDs []string
	var gradeIDs []string

	err := m.milestoneRelationDA.Query(ctx, &da.MilestoneRelationCondition{
		MasterIDs: dbo.NullStrings{
			Strings: []string{milestone.ID},
			Valid:   true,
		},
	}, &milestoneRelations)
	if err != nil {
		log.Error(ctx, "m.milestoneRelationDA.Query error",
			log.Err(err),
			log.String("milsestoneID", milestone.ID))
		return nil, err
	}

	for _, milestoneRelation := range milestoneRelations {
		switch milestoneRelation.RelationType {
		case entity.ProgramType:
			programIDs = append(programIDs, milestoneRelation.RelationID)
		case entity.SubjectType:
			subjectIDs = append(subjectIDs, milestoneRelation.RelationID)
		case entity.CategoryType:
			categoryIDs = append(categoryIDs, milestoneRelation.RelationID)
		case entity.SubcategoryType:
			subCategoryIDs = append(subCategoryIDs, milestoneRelation.RelationID)
		case entity.GradeType:
			gradeIDs = append(gradeIDs, milestoneRelation.RelationID)
		case entity.AgeType:
			ageIDs = append(ageIDs, milestoneRelation.RelationID)
		}
	}

	programIDs = utils.StableSliceDeduplication(programIDs)
	subjectIDs = utils.StableSliceDeduplication(subjectIDs)
	categoryIDs = utils.StableSliceDeduplication(categoryIDs)
	subCategoryIDs = utils.StableSliceDeduplication(subCategoryIDs)
	gradeIDs = utils.StableSliceDeduplication(gradeIDs)
	ageIDs = utils.StableSliceDeduplication(ageIDs)

	g := new(errgroup.Group)
	var orgNameMap map[string]string
	var userNameMap map[string]string
	var programNameMap map[string]string
	var subjectNameMap map[string]string
	var categoryNameMap map[string]string
	var subCategoryNameMap map[string]string
	var gradeNameMap map[string]string
	var ageNameMap map[string]string
	milestoneOutcomeViews := []*MilestoneOutcomeView{}

	g.Go(func() error {
		orgNames, err := m.organizationService.BatchGetNameMap(ctx, operator, []string{milestone.OrganizationID})
		if err != nil {
			log.Error(ctx, "m.organizationService.BatchGetNameMap error",
				log.Err(err),
				log.Strings("userIDs", userIDs))
			return err
		}

		orgNameMap = orgNames

		return nil
	})

	// get user name map
	if len(userIDs) > 0 {
		g.Go(func() error {
			userNames, err := m.userService.BatchGetNameMap(ctx, operator, userIDs)
			if err != nil {
				log.Error(ctx, "m.userService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("userIDs", userIDs))
				return err
			}

			userNameMap = userNames

			return nil
		})
	}

	// get program info
	if len(programIDs) > 0 {
		g.Go(func() error {
			programNames, err := m.programService.BatchGetNameMap(ctx, operator, programIDs)
			if err != nil {
				log.Error(ctx, "m.programService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("programIDs", programIDs))
				return err
			}

			programNameMap = programNames

			return nil
		})
	}

	// get subject info
	if len(subjectIDs) > 0 {
		g.Go(func() error {
			subjectNames, err := m.subjectService.BatchGetNameMap(ctx, operator, subjectIDs)
			if err != nil {
				log.Error(ctx, "m.subjectService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("subjectIDs", subjectIDs))
				return err
			}

			subjectNameMap = subjectNames

			return nil
		})
	}

	// get category info
	if len(categoryIDs) > 0 {
		g.Go(func() error {
			categoryNames, err := m.categoryService.BatchGetNameMap(ctx, operator, categoryIDs)
			if err != nil {
				log.Error(ctx, "m.categoryService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("categoryIDs", categoryIDs))
				return err
			}

			categoryNameMap = categoryNames

			return nil
		})
	}

	// get sub category info
	if len(subCategoryIDs) > 0 {
		g.Go(func() error {
			subCategoryNames, err := m.subCategoryService.BatchGetNameMap(ctx, operator, subCategoryIDs)
			if err != nil {
				log.Error(ctx, "m.subCategoryService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("subCategoryIDs", subCategoryIDs))
				return err
			}

			subCategoryNameMap = subCategoryNames

			return nil
		})
	}

	// get grade info
	if len(gradeIDs) > 0 {
		g.Go(func() error {
			gradeNames, err := m.gradeService.BatchGetNameMap(ctx, operator, gradeIDs)
			if err != nil {
				log.Error(ctx, "m.gradeService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("gradeIDs", gradeIDs))
				return err
			}

			gradeNameMap = gradeNames

			return nil
		})
	}

	// get age info
	if len(ageIDs) > 0 {
		g.Go(func() error {
			ageNames, err := m.ageService.BatchGetNameMap(ctx, operator, ageIDs)
			if err != nil {
				log.Error(ctx, "m.ageService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("ageIDs", ageIDs))
				return err
			}

			ageNameMap = ageNames

			return nil
		})
	}

	// TODO: bad database query
	// get outcomes
	g.Go(func() error {
		milestoneOutcomeRelations, err := m.milestoneOutcomeDA.SearchTx(ctx, dbo.MustGetDB(ctx),
			&da.MilestoneOutcomeCondition{
				MilestoneID: sql.NullString{String: milestone.ID, Valid: true},
			})
		if err != nil {
			log.Error(ctx, "m.milestoneOutcomeDA.SearchTx error", log.Err(err),
				log.Any("milestone", milestone))
			return err
		}

		if len(milestoneOutcomeRelations) == 0 {
			log.Debug(ctx, "milestone outcome relation not found", log.Any("milestone", milestone))
			return nil
		}

		outcomeAncestorIDs := make([]string, len(milestoneOutcomeRelations))
		for i, milestoneOutcomeRelation := range milestoneOutcomeRelations {
			outcomeAncestorIDs[i] = milestoneOutcomeRelation.OutcomeAncestor
		}

		outcomeAncestorIDs = utils.StableSliceDeduplication(outcomeAncestorIDs)

		outcomes, err := GetOutcomeModel().GetLatestByAncestors(ctx, operator, dbo.MustGetDB(ctx), outcomeAncestorIDs)
		if err != nil {
			log.Error(ctx, "GetOutcomeModel().GetLatestByAncestors error",
				log.Err(err),
				log.Strings("outcomeAncestorIDs", outcomeAncestorIDs))
			return err
		}

		milestoneOutcomeViewList, err := m.transformToMilestoneOutcomeView(ctx, operator, outcomes)
		if err != nil {
			log.Error(ctx, " m.transformToMilestoneOutcomeView error",
				log.Err(err),
				log.Any("outcomes", outcomes))
			return err
		}

		outcomesMap := make(map[string]*MilestoneOutcomeView, len(outcomes))
		for _, milestoneOutcomeView := range milestoneOutcomeViewList {
			outcomesMap[milestoneOutcomeView.AncestorID] = milestoneOutcomeView
		}

		for _, outcomeAncestorID := range outcomeAncestorIDs {
			milestoneOutcomeViews = append(milestoneOutcomeViews, outcomesMap[outcomeAncestorID])
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Error(ctx, "transformToMilestoneDetailView error",
			log.Err(err))
		return nil, err
	}

	// fill author name
	if userName, ok := userNameMap[milestone.AuthorID]; ok {
		result.Author = &AuthorView{
			AuthorID:   milestone.AuthorID,
			AuthorName: userName,
		}
	} else {
		log.Error(ctx, "author not found", log.String("userID", milestone.AuthorID))
	}

	// fill org name
	if orgName, ok := orgNameMap[milestone.OrganizationID]; ok {
		result.Organization = &OrganizationView{
			OrganizationID:   milestone.OrganizationID,
			OrganizationName: orgName,
		}
	} else {
		log.Error(ctx, "organization not found", log.String("orgID", milestone.OrganizationID))
	}

	// fill outcome milestones
	result.Outcomes = milestoneOutcomeViews

	// fill program, subject, category, subCategory, grade, age
	for _, milestoneRelation := range milestoneRelations {
		switch milestoneRelation.RelationType {
		case entity.ProgramType:
			if programName, ok := programNameMap[milestoneRelation.RelationID]; ok {
				result.Program = append(result.Program,
					&Program{
						ProgramID:   milestoneRelation.RelationID,
						ProgramName: programName,
					})
			} else {
				log.Error(ctx, "program record not found", log.String("programID", milestoneRelation.RelationID))
			}
		case entity.SubjectType:
			if subjectName, ok := subjectNameMap[milestoneRelation.RelationID]; ok {
				result.Subject = append(result.Subject,
					&Subject{
						SubjectID:   milestoneRelation.RelationID,
						SubjectName: subjectName,
					})
			} else {
				log.Error(ctx, "subject record not found", log.String("subjectID", milestoneRelation.RelationID))
			}
		case entity.CategoryType:
			if categoryName, ok := categoryNameMap[milestoneRelation.RelationID]; ok {
				result.Category = append(result.Category,
					&Category{
						CategoryID:   milestoneRelation.RelationID,
						CategoryName: categoryName,
					})
			} else {
				log.Error(ctx, "category record not found", log.String("categoryID", milestoneRelation.RelationID))
			}
		case entity.SubcategoryType:
			if subCategoryName, ok := subCategoryNameMap[milestoneRelation.RelationID]; ok {
				result.SubCategory = append(result.SubCategory,
					&SubCategory{
						SubCategoryID:   milestoneRelation.RelationID,
						SubCategoryName: subCategoryName,
					})
			} else {
				log.Error(ctx, "subCategory record not found", log.String("subCategoryID", milestoneRelation.RelationID))
			}
		case entity.GradeType:
			if gradeName, ok := gradeNameMap[milestoneRelation.RelationID]; ok {
				result.Grade = append(result.Grade,
					&Grade{
						GradeID:   milestoneRelation.RelationID,
						GradeName: gradeName,
					})
			} else {
				log.Error(ctx, "grade record not found", log.String("gradeID", milestoneRelation.RelationID))
			}
		case entity.AgeType:
			if ageName, ok := ageNameMap[milestoneRelation.RelationID]; ok {
				result.Age = append(result.Age,
					&Age{
						AgeID:   milestoneRelation.RelationID,
						AgeName: ageName,
					})
			} else {
				log.Error(ctx, "age record not found", log.String("ageID", milestoneRelation.RelationID))
			}
		}
	}

	return result, nil
}

func (m *MilestoneModel) transformToMilestoneOutcomeView(ctx context.Context, operator *entity.Operator, outcomes []*entity.Outcome) ([]*MilestoneOutcomeView, error) {
	result := make([]*MilestoneOutcomeView, len(outcomes))
	resultMap := make(map[string]*MilestoneOutcomeView, len(outcomes))
	outcomeIDs := make([]string, len(outcomes))
	var outcomeRelations []*entity.OutcomeRelation
	var programIDs []string
	var categoryIDs []string

	for i, outcome := range outcomes {
		result[i] = &MilestoneOutcomeView{
			OutcomeID:   outcome.ID,
			OutcomeName: outcome.Name,
			Shortcode:   outcome.Shortcode,
			AncestorID:  outcome.AncestorID,
			Assumed:     outcome.Assumed,
			// init zero value
			Program:       []Program{},
			Developmental: []Developmental{},
			Sets:          []*OutcomeSetCreateView{},
		}

		outcomeIDs[i] = outcome.ID
		resultMap[outcome.ID] = result[i]
	}

	err := m.outcomeRelationDA.Query(ctx, &da.OutcomeRelationCondition{
		MasterIDs: dbo.NullStrings{
			Strings: outcomeIDs,
			Valid:   true,
		},
	}, &outcomeRelations)
	if err != nil {
		log.Error(ctx, "m.outcomeRelationDA.Query error",
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
	var programNameMap map[string]string
	var categoryNameMap map[string]string

	// get outcome set
	g.Go(func() error {
		outcomeSets, err := m.outcomeSetDA.SearchSetsByOutcome(ctx, dbo.MustGetDB(ctx), outcomeIDs)
		if err != nil {
			log.Error(ctx, "m.outcomeSetDA.SearchSetsByOutcome error",
				log.Err(err),
				log.Strings("outcomeIDs", outcomeIDs))
			return err
		}

		outcomeSetMap = outcomeSets
		return nil
	})

	// get program info
	if len(programIDs) > 0 {
		g.Go(func() error {
			programNames, err := m.programService.BatchGetNameMap(ctx, operator, programIDs)
			if err != nil {
				log.Error(ctx, "m.programService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("programIDs", programIDs))
				return err
			}

			programNameMap = programNames

			return nil
		})
	}

	// get category info
	if len(categoryIDs) > 0 {
		g.Go(func() error {
			categoryNames, err := m.categoryService.BatchGetNameMap(ctx, operator, categoryIDs)
			if err != nil {
				log.Error(ctx, "m.categoryService.BatchGetNameMap error",
					log.Err(err),
					log.Strings("categoryIDs", categoryIDs))
				return err
			}

			categoryNameMap = categoryNames

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		log.Error(ctx, "transformToOutcomeView error",
			log.Err(err))
		return nil, err
	}

	for i, outcome := range outcomes {
		milestoneOutcomeView := result[i]
		// fill outcome sets
		if outcomeSets, ok := outcomeSetMap[outcome.ID]; ok {
			for _, set := range outcomeSets {
				milestoneOutcomeView.Sets = append(milestoneOutcomeView.Sets, &OutcomeSetCreateView{
					SetID:   set.ID,
					SetName: set.Name,
				})
			}
		}
	}

	// fill program and category
	for _, outcomeRelation := range outcomeRelations {
		if outcomeView, ok := resultMap[outcomeRelation.MasterID]; ok {
			switch outcomeRelation.RelationType {
			case entity.ProgramType:
				if programName, ok := programNameMap[outcomeRelation.RelationID]; ok {
					outcomeView.Program = append(outcomeView.Program,
						Program{
							ProgramID:   outcomeRelation.RelationID,
							ProgramName: programName,
						})
				} else {
					log.Error(ctx, "program record not found", log.String("programID", outcomeRelation.RelationID))
				}
			case entity.CategoryType:
				if categoryName, ok := categoryNameMap[outcomeRelation.RelationID]; ok {
					outcomeView.Developmental = append(outcomeView.Developmental,
						Developmental{
							DevelopmentalID:   outcomeRelation.RelationID,
							DevelopmentalName: categoryName,
						})
				} else {
					log.Error(ctx, "category record not found", log.String("categoryID", outcomeRelation.RelationID))
				}
			}
		}
	}

	return result, nil
}
