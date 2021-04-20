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

type AttachSqlDA struct {
	dbo.BaseDA
}

func (a AttachSqlDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, table string, masterType string, masterIDs []string) error {
	if len(masterIDs) > 0 {
		sql := fmt.Sprintf("delete from %s where master_id in (?) and master_type = ?", table)
		err := tx.Exec(sql, masterIDs, masterType).Error
		if err != nil {
			log.Error(ctx, "Replace: exec del sql failed",
				log.Err(err),
				log.Strings("master", masterIDs),
				log.String("type", masterType),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}

func (a AttachSqlDA) InsertTx(ctx context.Context, tx *dbo.DBContext, table string, attaches []*entity.Attach) error {
	if len(attaches) > 0 {
		values := make([][]interface{}, len(attaches))
		now := time.Now().Unix()
		for i := range attaches {
			values[i] = []interface{}{
				attaches[i].MasterID,
				attaches[i].AttachID,
				attaches[i].AttachType,
				attaches[i].MasterType,
				now + int64(i),
				now + int64(i),
			}
		}
		format, result := SQLBatchInsert(table, []string{"master_id", "attach_id", "attach_type", "master_type", "create_at", "update_at"}, values)
		err := tx.Exec(format, result...).Error
		if err != nil {
			log.Error(ctx, "Replace: exec insert sql failed",
				log.Err(err),
				log.Any("attach", attaches),
				log.String("sql", format))
			return err
		}
	}
	return nil
}

//func (a AttachSqlDA) SearchTx(ctx context.Context, tx *dbo.DBContext, table string, masterIDs []string, masterType string) ([]*entity.Attach, error) {
//	a.BaseDA.PageTx(ctx, tx)
//	//sql := fmt.Sprintf("select * from %s where master_id in (?) and master_type = ? and delete_at is null order by update_at", table)
//	//var attaches []*entity.Attach
//	//err := tx.Raw(sql, masterIDs, masterType).Find(&attaches).Error
//	//if err != nil {
//	//	log.Error(ctx, "Search: exec sql failed",
//	//		log.Err(err),
//	//		log.Strings("master", masterIDs),
//	//		log.String("type", masterType),
//	//		log.String("sql", sql))
//	//	return nil, err
//	//}
//	return attaches, nil
//}

var attachDA *AttachSqlDA
var _attachDAOnce sync.Once

func GetAttachDA() *AttachSqlDA {
	_attachDAOnce.Do(func() {
		attachDA = new(AttachSqlDA)
	})
	return attachDA
}

type AttachCondition struct {
	MasterIDs  dbo.NullStrings
	MasterType sql.NullString

	IncludeDeleted bool
	OrderBy        AttachOrderBy `json:"order_by"`
	Pager          dbo.Pager
}

func (c *AttachCondition) GetConditions() ([]string, []interface{}) {
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

type AttachOrderBy int

const (
	_ AttachOrderBy = iota
	OrderByAttachUpdateAt
	OrderByAttachUpdateAtDesc
)

func (c *AttachCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *AttachCondition) GetOrderBy() string {
	switch c.OrderBy {
	case OrderByAttachUpdateAt:
		return "update_at"
	case OrderByAttachUpdateAtDesc:
		return "update_at desc"
	default:
		return "update_at desc"
	}
}

type IAttachDA interface {
	TableName() string
	MasterType() string
}
