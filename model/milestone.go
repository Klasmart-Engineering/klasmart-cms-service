package model

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	_milestoneModel     IMilestoneModel
	_milestoneModelOnce sync.Once
)

type IMilestoneModel interface {
	Create(ctx context.Context, op *entity.Operator, milestone *entity.Milestone, outcomeAncestors []string) error
	Obtain(ctx context.Context, op *entity.Operator, milestoneID string) (*entity.Milestone, error)
	Update(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, milestone *entity.Milestone, outcomeIDs []string) error
	Delete(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, IDs []string) error
	Search(ctx context.Context, op *entity.Operator, condition *entity.MilestoneCondition) (int, []*entity.Milestone, error)
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
}

func GetMilestoneModel() IMilestoneModel {
	_milestoneModelOnce.Do(func() {
		_milestoneModel = new(MilestoneModel)
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
		for i := range outcomeAncestors {
			milestoneOutcome := entity.MilestoneOutcome{
				MilestoneID:     milestone.ID,
				OutcomeAncestor: outcomeAncestors[i],
			}
			// milestoneOutcomes sort: first in last out
			milestoneOutcomes[length-1-i] = &milestoneOutcome
		}
		err = da.GetMilestoneOutcomeDA().InsertTx(ctx, tx, milestoneOutcomes)
		if err != nil {
			log.Error(ctx, "CreateMilestone: Replace failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}
		err = da.GetMilestoneRelationDA().InsertTx(ctx, tx, m.collectRelation(milestone))
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

func (m MilestoneModel) Obtain(ctx context.Context, op *entity.Operator, milestoneID string) (*entity.Milestone, error) {
	var milestone *entity.Milestone
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		var err error
		milestone, err = da.GetMilestoneDA().GetByID(ctx, tx, milestoneID)
		if err != nil {
			log.Error(ctx, "Obtain: GetByID failed",
				log.Err(err),
				log.Any("op", op),
				log.String("milestone", milestoneID))
			return err
		}
		relations, err := da.GetMilestoneRelationDA().SearchTx(ctx, tx, &da.RelationCondition{
			MasterIDs:  dbo.NullStrings{Strings: []string{milestoneID}, Valid: true},
			MasterType: sql.NullString{String: string(entity.MilestoneType), Valid: true},
		})
		if err != nil {
			log.Error(ctx, "Obtain: Search failed", log.Any("op", op), log.String("milestone", milestoneID))
			return err
		}
		m.fillRelation(milestone, relations)
		milestoneOutcomes, err := da.GetMilestoneOutcomeDA().SearchTx(ctx, tx, &da.MilestoneOutcomeCondition{
			MilestoneID: sql.NullString{String: milestoneID, Valid: true},
		})
		if err != nil {
			log.Error(ctx, "Obtain: SearchTx failed", log.Any("op", op), log.String("milestone", milestoneID))
			return err
		}
		bindLength := len(milestoneOutcomes)
		if bindLength == 0 {
			log.Info(ctx, "Obtain: no outcome bind", log.String("milestone", milestoneID))
			return nil
		}

		outcomeAncestors := make([]string, 0, bindLength)
		for i := range milestoneOutcomes {
			outcomeAncestors = append(outcomeAncestors, milestoneOutcomes[i].OutcomeAncestor)
		}

		// NKL-1021
		// if milestone.Type == entity.GeneralMilestoneType {
		// 	intersect, err := da.GetMilestoneOutcomeDA().SearchTx(ctx, tx, &da.MilestoneOutcomeCondition{
		// 		OutcomeAncestors: dbo.NullStrings{Strings: outcomeAncestors, Valid: true},
		// 		NotMilestoneID:   sql.NullString{String: milestone.ID, Valid: true},
		// 	})
		// 	if err != nil {
		// 		log.Debug(ctx, "Obtain: exclude normal bind from general",
		// 			log.Any("milestone", milestone),
		// 			log.Strings("ancestors", outcomeAncestors))
		// 		return err
		// 	}
		// 	if len(intersect) > 0 {
		// 		intersectMap := make(map[string]bool)
		// 		for i := range intersect {
		// 			intersectMap[intersect[i].OutcomeAncestor] = true
		// 		}
		// 		outcomeAncestors = make([]string, 0, bindLength-len(intersectMap))
		// 		for i := range milestoneOutcomes {
		// 			if !intersectMap[milestoneOutcomes[i].OutcomeAncestor] {
		// 				outcomeAncestors = append(outcomeAncestors, milestoneOutcomes[i].OutcomeAncestor)
		// 			}
		// 		}
		// 	}
		// }

		outcomeAncestors = utils.StableSliceDeduplication(outcomeAncestors)

		outcomes, err := GetOutcomeModel().GetLatestByAncestors(ctx, op, tx, outcomeAncestors)
		if err != nil {
			log.Error(ctx, "Obtain: GetLatestByAncestors failed",
				log.Err(err),
				log.Strings("ancestors", outcomeAncestors))
			return err
		}

		outcomesMap := make(map[string]*entity.Outcome, len(outcomes))
		for i := range outcomes {
			outcomesMap[outcomes[i].AncestorID] = outcomes[i]
		}
		milestone.Outcomes = make([]*entity.Outcome, len(outcomes))
		for i := range outcomeAncestors {
			milestone.Outcomes[i] = outcomesMap[outcomeAncestors[i]]
		}
		milestone.LoCounts = len(milestone.Outcomes)
		return nil
	})
	return milestone, err
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
		err = da.GetMilestoneDA().Update(ctx, tx, newMilsestone)
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
		for i := range outcomeAncestors {
			milestoneOutcome := entity.MilestoneOutcome{
				MilestoneID:     newMilsestone.ID,
				OutcomeAncestor: outcomeAncestors[i],
			}
			// milestoneOutcomes sort: first in last out
			milestoneOutcomes[length-1-i] = &milestoneOutcome
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
		err = da.GetMilestoneOutcomeDA().InsertTx(ctx, tx, milestoneOutcomes)
		if err != nil {
			log.Error(ctx, "Update: InsertTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestoneOutcomes))
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
		err = da.GetMilestoneRelationDA().InsertTx(ctx, tx, m.collectRelation(newMilsestone))
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

func (m MilestoneModel) Search(ctx context.Context, op *entity.Operator, condition *entity.MilestoneCondition) (int, []*entity.Milestone, error) {
	var count int
	var milestones []*entity.Milestone
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		var err error
		count, milestones, err = da.GetMilestoneDA().Search(ctx, tx, &da.MilestoneCondition{
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
		})
		if err != nil {
			log.Error(ctx, "Search: Search failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("condition", condition))
			return err
		}

		if len(milestones) == 0 {
			log.Debug(ctx, "Search: not found",
				log.Any("op", op),
				log.Any("condition", condition))
			// NKL-1021
			// if condition.Status != entity.OutcomeStatusPublished {
			// 	return nil
			// }

			// general, err := m.CreateGeneral(ctx, op, tx, "")
			// if err != nil {
			// 	log.Error(ctx, "Search: CreateGeneral failed",
			// 		log.Any("op", op))
			// 	return err
			// }
			// count = 1
			// milestones = append(milestones, general)
			return nil
		}

		var generalIDs, normalIDs []string
		milestoneIDs := make([]string, len(milestones))
		lockedMilestoneIDs := make([]string, 0)
		for i := range milestones {
			if milestones[i].Type == entity.GeneralMilestoneType {
				generalIDs = append(generalIDs, milestones[i].ID)
			}
			if milestones[i].Type == entity.CustomMilestoneType {
				normalIDs = append(normalIDs, milestones[i].ID)
			}
			milestoneIDs[i] = milestones[i].ID
			if milestones[i].HasLocked() {
				lockedMilestoneIDs = append(lockedMilestoneIDs, milestones[i].ID)
			}
		}

		lockedChildrenMap := make(map[string]*entity.Milestone, len(lockedMilestoneIDs))
		if len(lockedMilestoneIDs) > 0 {
			lockedChildrenCondition := &da.MilestoneCondition{
				IncludeDeleted: false,
				SourceIDs:      dbo.NullStrings{Strings: lockedMilestoneIDs, Valid: true},
			}

			_, lockedChildren, err := da.GetMilestoneDA().Search(ctx, tx, lockedChildrenCondition)
			if err != nil {
				log.Error(ctx, "Search: Search failed",
					log.String("op", op.UserID),
					log.Any("condition", lockedChildrenCondition))
				return err
			}

			for _, v := range lockedChildren {
				// ancestor's source_id is itself
				if !v.IsAncestor() {
					lockedChildrenMap[v.SourceID] = v
				}
			}
		}

		relations, err := da.GetMilestoneRelationDA().SearchTx(ctx, tx, &da.RelationCondition{
			MasterIDs:  dbo.NullStrings{Strings: milestoneIDs, Valid: true},
			MasterType: sql.NullString{String: string(entity.MilestoneType), Valid: true},
		})
		if err != nil {
			log.Error(ctx, "Search: Search failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", milestoneIDs))
			return err
		}
		for i := range relations {
			for j := range milestones {
				if relations[i].MasterID == milestones[j].ID {
					m.fillRelation(milestones[j], []*entity.Relation{relations[i]})
					break
				}
			}
		}
		counts, err := da.GetMilestoneOutcomeDA().CountTx(ctx, tx, generalIDs, normalIDs)
		if err != nil {
			log.Error(ctx, "Search: Count failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("general", generalIDs),
				log.Strings("normal", normalIDs))
			return err
		}
		for i := range milestones {
			milestones[i].LoCounts = counts[milestones[i].ID]
			milestones[i].EditingMilestone = lockedChildrenMap[milestones[i].ID]
		}
		return nil
	})
	if err != nil {
		return 0, nil, err
	}
	return count, milestones, nil
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
		err = da.GetMilestoneDA().Update(ctx, tx, milestone)
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
		}

		err = da.GetMilestoneOutcomeDA().InsertTx(ctx, tx, milestoneOutcomes)
		if err != nil {
			log.Error(ctx, "Occupy: InsertTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone_outcome", milestoneOutcomes))
			return err
		}

		relations, err := da.GetMilestoneRelationDA().SearchTx(ctx, tx, &da.RelationCondition{
			MasterIDs:  dbo.NullStrings{Strings: []string{milestoneID}, Valid: true},
			MasterType: sql.NullString{String: string(entity.MilestoneType), Valid: true},
		})
		if err != nil {
			log.Error(ctx, "Occupy: Search failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("relation", relations))
			return err
		}

		for i := range relations {
			relations[i].MasterID = newVersion.ID
		}

		err = da.GetMilestoneRelationDA().InsertTx(ctx, tx, relations)
		if err != nil {
			log.Error(ctx, "Occupy: InsertTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("relation", relations))
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

			err = da.GetMilestoneDA().Update(ctx, tx, ms)
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

			err = da.GetMilestoneDA().Update(ctx, tx, milestone)
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
				err = da.GetMilestoneDA().Update(ctx, tx, parent)
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
			err = da.GetMilestoneDA().Update(ctx, tx, milestone)
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
	err = da.GetMilestoneOutcomeDA().InsertTx(ctx, tx, []*entity.MilestoneOutcome{milestoneOutcome})
	if err != nil {
		log.Error(ctx, "BindToGeneral: InsertTx failed",
			log.Any("op", op),
			log.Any("general", general),
			log.Any("outcome", outcome))
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
		log.Debug(ctx, "IsCached: redis access failed",
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
		log.Debug(ctx, "Cache: redis access failed",
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

func (m MilestoneModel) collectRelation(ms *entity.Milestone) []*entity.Relation {
	relations := make([]*entity.Relation, 0, len(ms.Programs)+len(ms.Subjects)+len(ms.Categories)+len(ms.Subcategories)+len(ms.Grades)+len(ms.Ages))
	for i := range ms.Programs {
		relation := entity.Relation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Programs[i],
			RelationType: entity.ProgramType,
		}
		relations = append(relations, &relation)
	}

	for i := range ms.Subjects {
		relation := entity.Relation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Subjects[i],
			RelationType: entity.SubjectType,
		}
		relations = append(relations, &relation)
	}

	for i := range ms.Categories {
		relation := entity.Relation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Categories[i],
			RelationType: entity.CategoryType,
		}
		relations = append(relations, &relation)
	}

	for i := range ms.Subcategories {
		relation := entity.Relation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Subcategories[i],
			RelationType: entity.SubcategoryType,
		}
		relations = append(relations, &relation)
	}

	for i := range ms.Grades {
		relation := entity.Relation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Grades[i],
			RelationType: entity.GradeType,
		}
		relations = append(relations, &relation)
	}

	for i := range ms.Ages {
		relation := entity.Relation{
			MasterID:     ms.ID,
			MasterType:   entity.MilestoneType,
			RelationID:   ms.Ages[i],
			RelationType: entity.AgeType,
		}
		relations = append(relations, &relation)
	}
	return relations
}

func (m MilestoneModel) fillRelation(ms *entity.Milestone, relations []*entity.Relation) {

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
