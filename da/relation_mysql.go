package da

import (
	"context"
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
	"time"
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
		values := make([][]interface{}, len(relations))
		now := time.Now().Unix()
		for i := range relations {
			values[i] = []interface{}{
				relations[i].MasterID,
				relations[i].RelationID,
				relations[i].RelationType,
				relations[i].MasterType,
				now + int64(i),
				now + int64(i),
			}
		}
		format, result := SQLBatchInsert(table, []string{"master_id", "relation_id", "relation_type", "master_type", "create_at", "update_at"}, values)
		err := tx.Exec(format, result...).Error
		if err != nil {
			log.Error(ctx, "Replace: exec insert sql failed",
				log.Err(err),
				log.Any("relation", relations),
				log.String("sql", format))
			return err
		}
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
