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
)

type IMilestoneModel interface {
	Create(ctx context.Context, op *entity.Operator, milestone *entity.Milestone, outcomeIDs []string) error
	Obtain(ctx context.Context, op *entity.Operator, milestoneID string) (*entity.Milestone, error)
	Update(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, milestone *entity.Milestone, outcomeIDs []string) error
	Delete(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, IDs []string) error
	Search(ctx context.Context, op *entity.Operator, condition *entity.MilestoneCondition) (int, []*entity.Milestone, error)
	Occupy(ctx context.Context, op *entity.Operator, milestoneID string) (*entity.Milestone, error)
	Publish(ctx context.Context, op *entity.Operator, IDs []string) error
}

type MilestoneModel struct {
}

func (m MilestoneModel) Create(ctx context.Context, op *entity.Operator, milestone *entity.Milestone, outcomeIDs []string) error {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixMilestoneMute, op.OrgID)
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
	milestone.Status = "draft"
	milestone.LoCounts = len(outcomeIDs)
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
		milestoneOutcomes := make([]*entity.MilestoneOutcome, len(outcomeIDs))
		for i := range outcomeIDs {
			milestoneOutcome := entity.MilestoneOutcome{
				MilestoneID: milestone.ID,
				OutcomeID:   outcomeIDs[i],
			}
			milestoneOutcomes[i] = &milestoneOutcome
		}
		err = da.GetMilestoneOutcomeDA().Replace(ctx, tx, nil, milestoneOutcomes)
		if err != nil {
			log.Error(ctx, "CreateMilestone: Replace failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", milestone))
			return err
		}
		err = da.GetMilestoneDA().ReplaceAttaches(ctx, tx, nil, entity.MilestoneType, milestone.CollectAttach())
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
			// TODO
			return err
		}
		attaches, err := da.GetMilestoneDA().SearchAttaches(ctx, tx, []string{milestone.ID}, entity.MilestoneType)
		if err != nil {
			// TODO
			return err
		}
		milestone.FillAttach(attaches)
		milestoneOutcomes, err := da.GetMilestoneOutcomeDA().Search(ctx, tx, milestoneID)
		bindLength := len(milestoneOutcomes)
		if bindLength == 0 {
			return nil
		}

		outcomeIDs := make([]string, bindLength)
		for i := range milestoneOutcomes {
			outcomeIDs[i] = milestoneOutcomes[i].OutcomeID
		}
		outcomes, err := GetOutcomeModel().GetLatestOutcomesByIDs(ctx, op, tx, outcomeIDs)
		if err != nil {
			// TODO
			return err
		}
		milestone.Outcomes = outcomes
		return nil
	})
	if err != nil {
		return nil, err
	}
	return milestone, nil
}

func (m MilestoneModel) Update(ctx context.Context, op *entity.Operator, perms map[external.PermissionName]bool, milestone *entity.Milestone, outcomeIDs []string) error {
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
			exists, err := GetShortcodeModel().isOccupied(ctx, tx, entity.KindMileStone, op.OrgID, ms.AncestorID, ms.Shortcode)
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
		ms.Update(milestone)
		err = da.GetMilestoneDA().Update(ctx, tx, ms)
		if err != nil {
			log.Error(ctx, "Update: Update failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
			return err
		}

		milestoneOutcomes := make([]*entity.MilestoneOutcome, len(outcomeIDs))
		for i := range outcomeIDs {
			milestoneOutcome := entity.MilestoneOutcome{
				MilestoneID: ms.ID,
				OutcomeID:   outcomeIDs[i],
			}
			milestoneOutcomes[i] = &milestoneOutcome
		}
		err = da.GetMilestoneOutcomeDA().Replace(ctx, tx, []string{ms.ID}, milestoneOutcomes)
		if err != nil {
			log.Error(ctx, "Update: Replace failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone", ms))
			return err
		}
		err = da.GetMilestoneDA().ReplaceAttaches(ctx, tx, []string{ms.ID}, entity.MilestoneType, ms.CollectAttach())
		if err != nil {
			log.Error(ctx, "Update: Replace failed",
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
		err = da.GetMilestoneOutcomeDA().Replace(ctx, tx, deleteIDs, nil)
		if err != nil {
			log.Error(ctx, "Delete: Replace failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", deleteIDs))
			return err
		}
		err = da.GetMilestoneDA().ReplaceAttaches(ctx, tx, deleteIDs, entity.MilestoneType, nil)
		if err != nil {
			log.Error(ctx, "Delete: Replace failed",
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

		attaches, err := da.GetMilestoneDA().SearchAttaches(ctx, tx, milestoneIDs, entity.MilestoneType)
		if err != nil {
			log.Error(ctx, "Search: SearchAttaches failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("milestone", milestoneIDs))
			return err
		}
		for i := range attaches {
			for j := range milestones {
				if attaches[i].MasterID == milestones[j].ID {
					milestones[j].FillAttach([]*entity.Attach{attaches[i]})
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return 0, nil, err
	}
	return count, milestones, nil
}

func (m MilestoneModel) Occupy(ctx context.Context, op *entity.Operator, milestoneID string) (*entity.Milestone, error) {
	var milestone *entity.Milestone
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
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
		milestone, err = ms.Copy(op)
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
		milestoneOutcomes, err := da.GetMilestoneOutcomeDA().Search(ctx, tx, milestoneID)
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
		err = da.GetMilestoneOutcomeDA().Replace(ctx, tx, nil, milestoneOutcomes)
		if err != nil {
			log.Error(ctx, "Occupy: Replace failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("milestone_outcome", milestoneOutcomes))
			return err
		}
		attaches, err := da.GetMilestoneDA().SearchAttaches(ctx, tx, []string{milestoneID}, entity.MilestoneType)
		if err != nil {
			log.Error(ctx, "Occupy: SearchAttaches failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("attach", attaches))
			return err
		}
		for i := range attaches {
			attaches[i].MasterID = milestone.ID
		}
		err = da.GetMilestoneDA().ReplaceAttaches(ctx, tx, nil, entity.MilestoneType, attaches)
		if err != nil {
			log.Error(ctx, "Occupy: ReplaceAttache failed",
				log.Err(err),
				log.Any("op", op),
				log.Any("attach", attaches))
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
		err = da.GetMilestoneDA().BatchPublish(ctx, tx, publishIDs, hideIDs, ancestorLatest)
		if err != nil {
			log.Error(ctx, "Publish: BatchPublish failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("publish", publishIDs),
				log.Strings("hidden", hideIDs),
				log.Any("ancestor_latest", ancestorLatest))
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
