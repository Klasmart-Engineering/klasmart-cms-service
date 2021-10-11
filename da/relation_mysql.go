package da

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type RelationSqlDA struct {
	dbo.BaseDA
}

func (a RelationSqlDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, table string, masterType entity.RelationType, masterIDs []string) error {
	if len(masterIDs) > 0 {
		sql := fmt.Sprintf("delete from %s where master_id in (?) and master_type = ?", table)
		err := tx.Exec(sql, masterIDs, masterType).Error
		if err != nil {
			log.Error(ctx, "Replace: exec del sql failed",
				log.Err(err),
				log.Strings("master", masterIDs),
				log.String("type", string(masterType)),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}

func (a RelationSqlDA) InsertTx(ctx context.Context, tx *dbo.DBContext, table string, relations []*entity.Relation) error {
	if len(relations) > 0 {
		now := time.Now().Unix()
		var models []entity.MilestoneRelation
		for i := range relations {
			models = append(models, entity.MilestoneRelation{
				Relation: entity.Relation{
					MasterID:     relations[i].MasterID,
					RelationID:   relations[i].RelationID,
					RelationType: relations[i].RelationType,
					MasterType:   relations[i].MasterType,
					CreateAt:     now + int64(i),
					UpdateAt:     now + int64(i),
				},
			})
		}
		_, err := a.InsertInBatchesTx(ctx, tx, &models, len(models))
		if err != nil {
			log.Error(ctx, "db exec batchInsert milestones_relations sql error",
				log.Err(err),
				log.Any("relation", relations))
			return err
		}
		return nil
	}
	return nil
}

var relationDA *RelationSqlDA
var _relationDAOnce sync.Once

func GetRelationDA() *RelationSqlDA {
	_relationDAOnce.Do(func() {
		relationDA = new(RelationSqlDA)
	})
	return relationDA
}

type RelationCondition struct {
	MasterIDs  dbo.NullStrings
	MasterType sql.NullString

	IncludeDeleted bool
	OrderBy        RelationOrderBy `json:"order_by"`
	Pager          dbo.Pager
}

func (c *RelationCondition) GetConditions() ([]string, []interface{}) {
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

type RelationOrderBy int

const (
	_ RelationOrderBy = iota
	OrderByRelationUpdateAt
	OrderByRelationUpdateAtDesc
)

func (c *RelationCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *RelationCondition) GetOrderBy() string {
	switch c.OrderBy {
	case OrderByRelationUpdateAt:
		return "update_at"
	case OrderByRelationUpdateAtDesc:
		return "update_at desc"
	default:
		return "update_at desc"
	}
}

type IRelationDA interface {
	TableName() string
	MasterType() string
}
