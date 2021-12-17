package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IOutcomeRelationDA interface {
	dbo.DataAccesser

	// Deprecated
	DeleteTx(ctx context.Context, tx *dbo.DBContext, masterIDs []string) error
}

type outcomeRelationDA struct {
	dbo.BaseDA
}

// Deprecated
func (*outcomeRelationDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, masterIDs []string) error {
	if len(masterIDs) > 0 {
		tx.ResetCondition()
		err := tx.Where("master_id in (?)", masterIDs).
			Delete(entity.OutcomeRelation{}).
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

var _outcomeRelationDA IOutcomeRelationDA
var _outcomeRelationDAOnce sync.Once

func GetOutcomeRelationDA() IOutcomeRelationDA {
	_outcomeRelationDAOnce.Do(func() {
		_outcomeRelationDA = new(outcomeRelationDA)
	})
	return _outcomeRelationDA
}

type OutcomeRelationCondition struct {
	MasterIDs      dbo.NullStrings
	IncludeDeleted bool
	OrderBy        OutcomeRelationOrderBy
	Pager          dbo.Pager
}

func (c *OutcomeRelationCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	if c.MasterIDs.Valid {
		wheres = append(wheres, "master_id in (?)")
		params = append(params, c.MasterIDs.Strings)
	}

	if !c.IncludeDeleted {
		wheres = append(wheres, "delete_at is null")
	}
	return wheres, params
}

type OutcomeRelationOrderBy int

const (
	_ OutcomeRelationOrderBy = iota
	OrderByOutcomeRelationUpdateAt
	OrderByOutcomeRelationUpdateAtDesc
)

func (c *OutcomeRelationCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *OutcomeRelationCondition) GetOrderBy() string {
	switch c.OrderBy {
	case OrderByOutcomeRelationUpdateAt:
		return "update_at"
	case OrderByOutcomeRelationUpdateAtDesc:
		return "update_at desc"
	default:
		return "update_at desc"
	}
}
