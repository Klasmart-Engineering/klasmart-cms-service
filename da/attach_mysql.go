package da

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
	"sync"
	"time"
)

type AttachSqlDA struct {
	dbo.BaseDA
}

func (a AttachSqlDA) Replace(ctx context.Context, tx *dbo.DBContext, table string, masterIDs []string, masterType string, attaches []*entity.Attach) error {
	if len(masterIDs) > 0 {
		sql := fmt.Sprintf("delete from %s where master_id in (?) and master_type = ?", table)
		err := tx.Exec(sql, masterIDs, masterType).Error
		if err != nil {
			log.Error(ctx, "Replace: exec del sql failed",
				log.Err(err),
				log.Strings("master", masterIDs),
				log.String("type", masterType),
				log.Any("attaches", attaches),
				log.String("sql", sql))
			return err
		}
	}
	if len(attaches) > 0 {
		values := make([]string, len(attaches))
		now := time.Now().Unix()
		for i := range attaches {
			tm := now + int64(i)
			masterID := attaches[i].MasterID
			masterType := attaches[i].MasterType
			attachID := attaches[i].AttachID
			attachType := attaches[i].AttachType
			value := fmt.Sprintf("select '%s', '%s', '%s', '%s', %d, %d where not exists(select * from %s where master_id='%s' and master_type='%s' and attach_id='%s' and attach_type ='%s' and delete_at is null)",
				masterID, attachID, attachType, masterType, tm, tm, table, masterID, masterType, attachID, attachType)
			values[i] = value
		}
		sql := fmt.Sprintf("insert into %s(master_id, attach_id, attach_type, master_type, create_at, update_at) (%s)", table, strings.Join(values, " union "))
		err := tx.Exec(sql).Error
		if err != nil {
			log.Error(ctx, "Replace: exec insert sql failed",
				log.Err(err),
				log.Strings("master", masterIDs),
				log.String("type", masterType),
				log.Any("attach", attaches),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}

func (a AttachSqlDA) Search(ctx context.Context, tx *dbo.DBContext, table string, masterIDs []string, masterType string) ([]*entity.Attach, error) {
	sql := fmt.Sprintf("select * from %s where master_id in (?) and master_type = ? and delete_at is null order by update_at", table)
	var attaches []*entity.Attach
	err := tx.Raw(sql, masterIDs, masterType).Find(&attaches).Error
	if err != nil {
		log.Error(ctx, "Search: exec sql failed",
			log.Err(err),
			log.Strings("master", masterIDs),
			log.String("type", masterType),
			log.String("sql", sql))
		return nil, err
	}
	return attaches, nil
}

var attachDA *AttachSqlDA
var _attachDAOnce sync.Once

func GetAttachDA() *AttachSqlDA {
	_attachDAOnce.Do(func() {
		attachDA = new(AttachSqlDA)
	})
	return attachDA
}
