package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

type IMilestoneModel interface {
	Create(ctx context.Context, op *entity.Operator, milestone *entity.Milestone, outcomeAncestors []string) error
	Obtain(ctx context.Context, op *entity.Operator, milestoneID string) (*entity.Milestone, error)
	Update(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, milestone *entity.Milestone, outcomeIDs []string) error
	Delete(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, IDs []string) error
	Search(ctx context.Context, op *entity.Operator, condition *entity.MilestoneCondition) (int, []*entity.Milestone, error)
	Occupy(ctx context.Context, op *entity.Operator, milestoneID string) (*entity.Milestone, error)
	Publish(ctx context.Context, op *entity.Operator, IDs []string) error
	SaveAndPublish(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, milestone *entity.Milestone, outcomeIDs []string) error
}

type MilestoneModel struct {
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
	milestone.Status = "draft"
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		occupied, err := GetShortcodeModel().isOccupied(ctx, tx, entity.Milestone{}.TableName(), op.OrgID, milestone.AncestorID, milestone.Shortcode)
		if err != nil {
			log.Error(ctx, "CreateMilestone: isOccupied failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}
		if occupied {
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
		milestoneOutcomes := make([]*entity.MilestoneOutcome, len(outcomeAncestors))
		for i := range outcomeAncestors {
			milestoneOutcome := entity.MilestoneOutcome{
				MilestoneID:     milestone.ID,
				OutcomeAncestor: outcomeAncestors[i],
			}
			milestoneOutcomes[i] = &milestoneOutcome
		}
		err = da.GetMilestoneOutcomeDA().InsertTx(ctx, tx, milestoneOutcomes)
		if err != nil {
			log.Error(ctx, "CreateMilestone: Replace failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}
		err = da.GetMilestoneRelationDA().InsertTx(ctx, tx, m.CollectRelation(milestone))
		if err != nil {
			log.Error(ctx, "CreateMilestone: InsertTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}
		return nil
	})
	da.GetShortcodeCacheDA().Remove(ctx, entity.KindMileStone, op.OrgID, milestone.Shortcode)
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
			log.Error(ctx, "Obtain: Search failed",
				log.Err(err),
				log.Any("op", op),
				log.String("milestone", milestoneID))
			return err
		}
		m.FillRelation(milestone, relations)
		milestoneOutcomes, err := da.GetMilestoneOutcomeDA().SearchTx(ctx, tx, &da.MilestoneOutcomeCondition{
			MilestoneID: sql.NullString{String: milestoneID, Valid: true},
		})
		bindLength := len(milestoneOutcomes)
		if bindLength == 0 {
			log.Info(ctx, "Obtain: no outcome bind",
				log.String("milestone", milestoneID))
			return nil
		}

		outcomeAncestors := make([]string, bindLength)
		for i := range milestoneOutcomes {
			outcomeAncestors[i] = milestoneOutcomes[i].OutcomeAncestor
		}
		outcomes, err := GetOutcomeModel().GetLatestOutcomesByAncestors(ctx, op, tx, outcomeAncestors)
		if err != nil {
			log.Error(ctx, "Obtain: GetLatestOutcomesByAncestors failed",
				log.Err(err),
				log.Strings("ancestors", outcomeAncestors))
			return err
		}
		milestone.Outcomes = outcomes
		milestone.LoCounts = len(outcomes)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return milestone, nil
}

func (m MilestoneModel) Update(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, milestone *entity.Milestone, outcomeAncestors []string) error {
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
	exists, err := GetShortcodeModel().isCached(ctx, entity.KindMileStone, op.OrgID, milestone.Shortcode)
	if err != nil {
		log.Error(ctx, "Update: isCached failed",
			log.Err(err),
			log.Any("op", op),
			log.Any("milestone", milestone))
		return err
	}
	if exists {
		return constant.ErrConflict
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		ms, err := da.GetMilestoneDA().GetByID(ctx, tx, milestone.ID)
		if err != nil {
			log.Error(ctx, "Update: GetByID failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}
		if ms.LockedBy != "" {
			return &ErrContentAlreadyLocked{LockedBy: &external.User{
				ID: ms.LockedBy,
			}}
		}
		switch ms.Status {
		case entity.OutcomeStatusDraft:
			if !perms[external.EditUnpublishedMilestone] {
				log.Error(ctx, "Update: perm failed",
					log.Err(err),
					log.Any("perms", perms),
					log.Any("op", op),
					log.Any("milestone", ms))
				return constant.ErrOperateNotAllowed
			}
		case entity.OutcomeStatusPublished:
			if !perms[external.EditPublishedMilestone] {
				log.Error(ctx, "Update: perm failed",
					log.Err(err),
					log.Any("perms", perms),
					log.Any("op", op),
					log.Any("milestone", ms))
				return constant.ErrOperateNotAllowed
			}
		default:
			return constant.ErrOperateNotAllowed
		}

		if ms.Shortcode != milestone.Shortcode {
			ms.Shortcode = milestone.Shortcode
			exists, err := GetShortcodeModel().isOccupied(ctx, tx, entity.Milestone{}.TableName(), op.OrgID, ms.AncestorID, ms.Shortcode)
			if err != nil {
				log.Error(ctx, "Update: isOccupied failed",
					log.Err(err),
					log.Any("op", op),
					log.Any("milestone", ms))
				return err
			}
			if exists {
				return constant.ErrConflict
			}
		}
		m.UpdateMilestone(milestone, ms)
		err = da.GetMilestoneDA().Update(ctx, tx, ms)
		if err != nil {
			log.Error(ctx, "Update: Update failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
			return err
		}

		milestoneOutcomes := make([]*entity.MilestoneOutcome, len(outcomeAncestors))
		for i := range outcomeAncestors {
			milestoneOutcome := entity.MilestoneOutcome{
				MilestoneID:     ms.ID,
				OutcomeAncestor: outcomeAncestors[i],
			}
			milestoneOutcomes[i] = &milestoneOutcome
		}
		err = da.GetMilestoneOutcomeDA().DeleteTx(ctx, tx, []string{ms.ID})
		if err != nil {
			log.Error(ctx, "Update: DeleteTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
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
		err = da.GetMilestoneRelationDA().DeleteTx(ctx, tx, []string{ms.ID})
		if err != nil {
			log.Error(ctx, "Update: DeleteTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
			return err
		}
		err = da.GetMilestoneRelationDA().InsertTx(ctx, tx, m.CollectRelation(ms))
		if err != nil {
			log.Error(ctx, "Update: InsertTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (m MilestoneModel) canBeDeleted(ctx context.Context, milestones []*entity.Milestone, perms map[external.PermissionName]bool) ([]string, error) {
	var milestoneIDs []string
	for i := range milestones {
		if milestones[i].LockedBy != "" {
			log.Warn(ctx, "canBeDeleted: locked",
				log.Any("milestone", milestones[i]))
			return nil, &ErrContentAlreadyLocked{LockedBy: &external.User{
				ID: milestones[i].LockedBy,
			}}
		}

		switch milestones[i].Status {
		case entity.OutcomeStatusDraft, entity.OutcomeStatusHidden:
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
		deleteIDs, err := m.canBeDeleted(ctx, milestones, perms)
		if err != nil {
			return err
		}
		err = da.GetMilestoneDA().BatchDelete(ctx, tx, deleteIDs)
		if err != nil {
			log.Error(ctx, "Delete: BatchDelete failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", deleteIDs))
			return err
		}
		err = da.GetMilestoneOutcomeDA().DeleteTx(ctx, tx, deleteIDs)
		if err != nil {
			log.Error(ctx, "Delete: DeleteTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", deleteIDs))
			return err
		}
		//err = da.GetRelationDA().Replace(ctx, tx, entity.OutcomeRelationTable, deleteIDs, entity.MilestoneType, nil)
		err = da.GetMilestoneRelationDA().DeleteTx(ctx, tx, deleteIDs)
		if err != nil {
			log.Error(ctx, "Delete: DeleteTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", deleteIDs))
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
			OrderBy:        da.OrderByMilestoneCreatedAtDesc,
			Pager:          utils.GetDboPager(condition.Page, condition.PageSize),
		})
		if err != nil {
			log.Error(ctx, "Search: Search failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("cond", condition))
			return err
		}

		if len(milestones) == 0 {
			log.Info(ctx, "Search: not found",
				log.Any("op", op),
				log.Any("cond", condition))
			return nil
		}
		milestoneIDs := make([]string, len(milestones))
		for i := range milestones {
			milestoneIDs[i] = milestones[i].ID
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
					m.FillRelation(milestones[j], []*entity.Relation{relations[i]})
					break
				}
			}
		}
		counts, err := da.GetMilestoneOutcomeDA().CountTx(ctx, tx, milestoneIDs)
		if err != nil {
			log.Error(ctx, "Search: Count failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", milestoneIDs))
			return err
		}
		for i := range milestones {
			milestones[i].LoCounts = counts[milestones[i].ID]
		}
		return nil
	})
	if err != nil {
		return 0, nil, err
	}
	return count, milestones, nil
}

func (m MilestoneModel) Occupy(ctx context.Context, op *entity.Operator, milestoneID string) (*entity.Milestone, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixMilestoneMute)
	if err != nil {
		log.Error(ctx, "Occupy: NewLock failed",
			log.Err(err),
			log.String("op", op.UserID),
			log.String("milestone", milestoneID))
		return nil, err
	}
	locker.Lock()
	defer locker.Unlock()
	var milestone *entity.Milestone
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		ms, err := da.GetMilestoneDA().GetByID(ctx, tx, milestoneID)
		if err != nil {
			log.Error(ctx, "Occupy: GetByID failed",
				log.Err(err),
				log.Any("op", op),
				log.String("milestone", milestoneID))
			return err
		}
		if ms.LockedBy != "" {
			log.Warn(ctx, "Occupy: already locked", log.Any("milestone", ms))
			return &ErrContentAlreadyLocked{LockedBy: &external.User{
				ID: ms.LockedBy,
			}}
		}
		milestone, err = m.Copy(op, ms)
		if err != nil {
			log.Error(ctx, "Occupy: Copy failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
			return err
		}
		ms.LockedBy = op.UserID
		err = da.GetMilestoneDA().Update(ctx, tx, ms)
		if err != nil {
			log.Error(ctx, "Occupy: Update failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
			return err
		}
		err = da.GetMilestoneDA().Create(ctx, tx, milestone)
		if err != nil {
			log.Error(ctx, "Occupy: Create failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
		}
		milestoneOutcomes, err := da.GetMilestoneOutcomeDA().SearchTx(ctx, tx, &da.MilestoneOutcomeCondition{
			MilestoneID: sql.NullString{String: milestoneID, Valid: true},
		})
		if err != nil {
			log.Error(ctx, "Occupy: Search failed",
				log.Err(err),
				log.Any("op", op),
				log.String("milestone", milestoneID))
			return err
		}
		for i := range milestoneOutcomes {
			milestoneOutcomes[i].MilestoneID = milestone.ID
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
			relations[i].MasterID = milestone.ID
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
	if err != nil {
		return nil, err
	}
	return milestone, nil
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
func (m MilestoneModel) Publish(ctx context.Context, op *entity.Operator, IDs []string) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, milestones, err := da.GetMilestoneDA().Search(ctx, tx, &da.MilestoneCondition{
			IDs: dbo.NullStrings{Strings: IDs, Valid: true},
		})
		if err != nil {
			log.Error(ctx, "Publish: Search failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", IDs))
			return err
		}
		publishIDs, hideIDs, ancestorLatest, err := m.canPublish(ctx, milestones)
		if err != nil {
			log.Error(ctx, "Publish: canPublish failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", IDs))
			return err
		}
		err = da.GetMilestoneDA().BatchPublish(ctx, tx, publishIDs)
		if err != nil {
			log.Error(ctx, "Publish: BatchPublish failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("publish", publishIDs))
			return err
		}
		err = da.GetMilestoneDA().BatchHide(ctx, tx, hideIDs)
		if err != nil {
			log.Error(ctx, "Publish: BatchHide failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("publish", publishIDs))
			return err
		}
		err = da.GetMilestoneDA().BatchUpdateLatest(ctx, tx, ancestorLatest)
		if err != nil {
			log.Error(ctx, "Publish: BatchUpdateLatest failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("publish", publishIDs))
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (m MilestoneModel) SaveAndPublish(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, milestone *entity.Milestone, outcomeAncestors []string) error {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixShortcodeMute, entity.KindMileStone, op.OrgID)
	if err != nil {
		log.Error(ctx, "SaveAndPublish: NewLock failed",
			log.Err(err),
			log.Any("op", op),
			log.Any("milestone", milestone))
		return err
	}
	locker.Lock()
	defer locker.Unlock()
	exists, err := GetShortcodeModel().isCached(ctx, entity.KindMileStone, op.OrgID, milestone.Shortcode)
	if err != nil {
		log.Error(ctx, "SaveAndPublish: isCached failed",
			log.Err(err),
			log.Any("op", op),
			log.Any("milestone", milestone))
		return err
	}
	if exists {
		return constant.ErrConflict
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		ms, err := da.GetMilestoneDA().GetByID(ctx, tx, milestone.ID)
		if err != nil {
			log.Error(ctx, "SaveAndPublish: GetByID failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}
		if ms.LockedBy != "" {
			return &ErrContentAlreadyLocked{LockedBy: &external.User{
				ID: ms.LockedBy,
			}}
		}
		switch ms.Status {
		case entity.OutcomeStatusDraft:
			if !perms[external.CreateMilestone] {
				log.Error(ctx, "SaveAndPublish: perm failed",
					log.Err(err),
					log.Any("perms", perms),
					log.Any("op", op),
					log.Any("milestone", ms))
				return constant.ErrOperateNotAllowed
			}
		default:
			log.Warn(ctx, "SaveAndPublish: status not allowed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
			return constant.ErrOperateNotAllowed
		}

		if ms.Shortcode != milestone.Shortcode {
			ms.Shortcode = milestone.Shortcode
			exists, err := GetShortcodeModel().isOccupied(ctx, tx, entity.Milestone{}.TableName(), op.OrgID, ms.AncestorID, ms.Shortcode)
			if err != nil {
				log.Error(ctx, "SaveAndPublish: isOccupied failed",
					log.Err(err),
					log.Any("op", op),
					log.Any("milestone", ms))
				return err
			}
			if exists {
				log.Warn(ctx, "SaveAndPublish: isOccupied",
					log.Err(err),
					log.Any("op", op),
					log.Any("milestone", ms))
				return constant.ErrConflict
			}
		}
		m.UpdateMilestone(milestone, ms)
		ms.Status = entity.OutcomeStatusPublished
		err = da.GetMilestoneDA().Update(ctx, tx, ms)
		if err != nil {
			log.Error(ctx, "Update: Update failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
			return err
		}

		err = da.GetMilestoneDA().BatchHide(ctx, tx, []string{milestone.SourceID})
		if err != nil {
			log.Error(ctx, "Publish: BatchHide failed",
				log.Any("op", op),
				log.Any("milestone", ms))
			return err
		}
		ancestorLatest := make(map[string]string)
		ancestorLatest[ms.AncestorID] = ms.ID
		err = da.GetMilestoneDA().BatchUpdateLatest(ctx, tx, ancestorLatest)
		if err != nil {
			log.Error(ctx, "Publish: BatchUpdateLatest failed",
				log.Any("op", op),
				log.Any("milestone", ancestorLatest))
			return err
		}
		milestoneOutcomes := make([]*entity.MilestoneOutcome, len(outcomeAncestors))
		for i := range outcomeAncestors {
			milestoneOutcome := entity.MilestoneOutcome{
				MilestoneID:     ms.ID,
				OutcomeAncestor: outcomeAncestors[i],
			}
			milestoneOutcomes[i] = &milestoneOutcome
		}
		err = da.GetMilestoneOutcomeDA().DeleteTx(ctx, tx, []string{ms.ID})
		if err != nil {
			log.Error(ctx, "Update: DeleteTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
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
		err = da.GetMilestoneRelationDA().DeleteTx(ctx, tx, []string{ms.ID})
		if err != nil {
			log.Error(ctx, "Update: DeleteTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
			return err
		}
		err = da.GetMilestoneRelationDA().InsertTx(ctx, tx, m.CollectRelation(ms))
		if err != nil {
			log.Error(ctx, "Update: InsertTx failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

var (
	_milestoneModel     IMilestoneModel
	_milestoneModelOnce sync.Once
)

func GetMilestoneModel() IMilestoneModel {
	_milestoneModelOnce.Do(func() {
		_milestoneModel = new(MilestoneModel)
	})
	return _milestoneModel
}

func (m MilestoneModel) CollectRelation(ms *entity.Milestone) []*entity.Relation {
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

func (m MilestoneModel) FillRelation(ms *entity.Milestone, relations []*entity.Relation) {

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

func (m *MilestoneModel) Copy(op *entity.Operator, ms *entity.Milestone) (*entity.Milestone, error) {
	if ms.Status != entity.OutcomeStatusPublished {
		return nil, constant.ErrOperateNotAllowed
	}
	milestone := &entity.Milestone{
		ID:             utils.NewID(),
		Name:           ms.Name,
		Shortcode:      ms.Shortcode,
		OrganizationID: op.OrgID,
		AuthorID:       op.UserID,
		Description:    ms.Description,
		LoCounts:       ms.LoCounts,

		Status: entity.OutcomeStatusDraft,

		AncestorID: ms.AncestorID,
		SourceID:   ms.ID,
	}
	milestone.SourceID = ms.ID
	milestone.LatestID = milestone.ID
	return milestone, nil
}

func (m *MilestoneModel) UpdateMilestone(milestone *entity.Milestone, ms *entity.Milestone) {
	ms.Name = milestone.Name
	ms.Description = milestone.Description
	ms.Programs = milestone.Programs
	ms.Subjects = milestone.Subjects
	ms.Categories = milestone.Categories
	ms.Subcategories = milestone.Subcategories
	ms.Grades = milestone.Grades
	ms.Ages = milestone.Ages
}
