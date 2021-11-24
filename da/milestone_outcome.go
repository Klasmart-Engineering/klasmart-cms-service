package da

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IMilestoneOutcomeDA interface {
	dbo.DataAccesser
	// TODO: remove duplicate interfaces
	SearchTx(ctx context.Context, tx *dbo.DBContext, condition *MilestoneOutcomeCondition) ([]*entity.MilestoneOutcome, error)
	DeleteTx(ctx context.Context, tx *dbo.DBContext, milestoneIDs []string) error
	BatchCountTx(ctx context.Context, tx *dbo.DBContext, generalIDs, normalIDs []string) (map[string]int, error)
}

type milestoneOutcomeDA struct {
	dbo.BaseDA
}

var _milestoneOutcomeDA IMilestoneOutcomeDA
var _milestoneOutcomeOnce sync.Once

func GetMilestoneOutcomeDA() IMilestoneOutcomeDA {
	_milestoneOutcomeOnce.Do(func() {
		_milestoneOutcomeDA = new(milestoneOutcomeDA)
	})
	return _milestoneOutcomeDA
}

type MilestoneOutcomeCondition struct {
	MilestoneID      sql.NullString
	MilestoneIDs     dbo.NullStrings
	NotMilestoneID   sql.NullString
	OutcomeAncestor  sql.NullString
	OutcomeAncestors dbo.NullStrings
	IncludeDeleted   bool
	OrderBy          MilestoneOutcomeOrderBy `json:"order_by"`
	Pager            dbo.Pager
}
type MilestoneOutcomeOrderBy int

const (
	_ MilestoneOutcomeOrderBy = iota
	OrderByMilestoneOutcomeUpdatedAt
	OrderByMilestoneOutcomeUpdatedAtDesc
)

func (c *MilestoneOutcomeCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	if c.MilestoneID.Valid {
		wheres = append(wheres, "milestone_id = ?")
		params = append(params, c.MilestoneID.String)
	}

	if c.MilestoneIDs.Valid {
		wheres = append(wheres, "milestone_id in (?)")
		params = append(params, c.MilestoneIDs.Strings)
	}

	if c.NotMilestoneID.Valid {
		wheres = append(wheres, "milestone_id <> ?")
		params = append(params, c.NotMilestoneID.String)
	}

	if c.OutcomeAncestor.Valid {
		wheres = append(wheres, "outcome_ancestor = ?")
		params = append(params, c.OutcomeAncestor.String)
	}

	if c.OutcomeAncestors.Valid {
		wheres = append(wheres, "outcome_ancestor in (?)")
		params = append(params, c.OutcomeAncestors.Strings)
	}

	if !c.IncludeDeleted {
		wheres = append(wheres, "COALESCE(delete_at, 0) = 0")
	}

	return wheres, params
}

func (c *MilestoneOutcomeCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *MilestoneOutcomeCondition) GetOrderBy() string {
	switch c.OrderBy {
	case OrderByMilestoneOutcomeUpdatedAt:
		return "update_at"
	case OrderByMilestoneOutcomeUpdatedAtDesc:
		return "update_at desc"
	default:
		return "update_at desc"

	}
}

func (mso milestoneOutcomeDA) SearchTx(ctx context.Context, tx *dbo.DBContext, condition *MilestoneOutcomeCondition) ([]*entity.MilestoneOutcome, error) {
	var result []*entity.MilestoneOutcome
	_, err := mso.BaseDA.PageTx(ctx, tx, condition, &result)
	if err != nil {
		log.Error(ctx, "SearchTx: PageTx failed",
			log.Err(err),
			log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

func (mso milestoneOutcomeDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, milestoneIDs []string) error {
	table := entity.MilestoneOutcome{}.TableName()
	if len(milestoneIDs) > 0 {
		sql := fmt.Sprintf("delete from %s where milestone_id in (?)", table)
		err := tx.Exec(sql, milestoneIDs).Error
		if err != nil {
			log.Error(ctx, "DeleteTx: exec del sql failed",
				log.Err(err),
				log.Strings("milestone", milestoneIDs),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}

func (mso milestoneOutcomeDA) BatchCountTx(ctx context.Context, tx *dbo.DBContext, generalIDs, normalIDs []string) (map[string]int, error) {
	var generalResults, normalResults []*struct {
		MilestoneID string `gorm:"column:milestone_id" json:"milestone_id"`
		Count       int    `gorm:"column:count" json:"count"`
	}

	generalSql := fmt.Sprintf("select gmo.milestone_id, count(gmo.outcome_ancestor) as count from "+
		"(select any_value(milestone_id) as milestone_id, outcome_ancestor from %s where outcome_ancestor in "+
		"(select outcome_ancestor from %s where milestone_id in (?)) "+
		"group by outcome_ancestor having count(milestone_id) = 1) as gmo group by gmo.milestone_id", entity.MilestoneOutcome{}.TableName(), entity.MilestoneOutcome{}.TableName())

	err := tx.Raw(generalSql, generalIDs).Find(&generalResults).Error
	if err != nil {
		log.Error(ctx, "Count: exec sql failed",
			log.Strings("general", normalIDs),
			log.String("sql", generalSql))
		return nil, err
	}

	normalSql := fmt.Sprintf("select milestone_id, count(outcome_ancestor) as count from %s where milestone_id in (?) group by milestone_id", entity.MilestoneOutcome{}.TableName())
	err = tx.Raw(normalSql, normalIDs).Find(&normalResults).Error
	if err != nil {
		log.Error(ctx, "Count: exec sql failed",
			log.Strings("normal", normalIDs),
			log.String("sql", normalSql))
		return nil, err
	}

	counts := make(map[string]int)
	for i := range generalResults {
		counts[generalResults[i].MilestoneID] = generalResults[i].Count
	}
	for i := range normalResults {
		counts[normalResults[i].MilestoneID] = normalResults[i].Count
	}
	return counts, nil
}
