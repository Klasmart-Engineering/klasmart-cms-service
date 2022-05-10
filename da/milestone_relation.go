package da

import (
	"context"
	"database/sql"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type IMilestoneRelationDA interface {
	dbo.DataAccesser
	// Deprecated
	DeleteTx(ctx context.Context, tx *dbo.DBContext, masterIDs []string) error
}

type milestoneRelationDA struct {
	dbo.BaseDA
}

var _milestoneRelationDA IMilestoneRelationDA
var _milestoneRelationDAOnce sync.Once

func GetMilestoneRelationDA() IMilestoneRelationDA {
	_milestoneRelationDAOnce.Do(func() {
		_milestoneRelationDA = new(milestoneRelationDA)
	})
	return _milestoneRelationDA
}

func (milestoneRelationDA) TableName() string {
	return entity.MilestoneRelationTable
}

func (mas milestoneRelationDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, masterIDs []string) error {
	if len(masterIDs) > 0 {
		tx.ResetCondition()
		err := tx.Where("master_id in (?)", masterIDs).Debug().
			Table(entity.MilestoneRelation{}.TableName()).
			Delete(entity.MilestoneRelation{}).
			Error
		if err != nil {
			log.Error(ctx, "delete record error",
				log.Err(err),
				log.Strings("masterIDs", masterIDs))
			return err
		}
	}

	return nil
}

type MilestoneRelationCondition struct {
	MasterIDs  dbo.NullStrings
	MasterType sql.NullString

	IncludeDeleted bool
	OrderBy        MilestoneRelationOrderBy
	Pager          dbo.Pager
}

func (c *MilestoneRelationCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	if c.MasterIDs.Valid {
		wheres = append(wheres, "master_id in (?)")
		params = append(params, c.MasterIDs.Strings)
	}

	if c.MasterType.Valid {
		wheres = append(wheres, "master_type = ?")
		params = append(params, c.MasterType.String)
	}

	if !c.IncludeDeleted {
		wheres = append(wheres, "delete_at is null")
	}
	return wheres, params
}

type MilestoneRelationOrderBy int

const (
	_ MilestoneRelationOrderBy = iota
	OrderByMilestoneRelationUpdateAt
	OrderByMilestoneRelationUpdateAtDesc
)

func (c *MilestoneRelationCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *MilestoneRelationCondition) GetOrderBy() string {
	switch c.OrderBy {
	case OrderByMilestoneRelationUpdateAt:
		return "update_at"
	case OrderByMilestoneRelationUpdateAtDesc:
		return "update_at desc"
	default:
		return "update_at desc"
	}
}
