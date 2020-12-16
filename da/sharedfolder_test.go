package da

import (
	"context"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func TestCreateSharedFolderRecordTable(t *testing.T) {
	dsn := "root:Badanamu123456@tcp(192.168.1.234:3310)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open("mysql", dsn)
	if !assert.NoError(t, err) {
		return
	}
	db.LogMode(true)
	db.AutoMigrate(entity.SharedFolderRecord{})
}

func TestAddAndBatchAddSharedFolderReocrds(t *testing.T) {
	ctx := context.Background()
	err := GetSharedFolderDA().AddSharedFolderRecord(ctx, dbo.MustGetDB(ctx), entity.SharedFolderRecord{
		OrgID:    "1",
		FolderID: "abbc",
		Creator:  "2",
	})
	if !assert.NoError(t, err) {
		return
	}
	t.Log("add:", err)
	err = GetSharedFolderDA().BatchAddSharedFolderRecord(ctx, dbo.MustGetDB(ctx), []*entity.SharedFolderRecord{
		{
			OrgID:    "1",
			FolderID: "abbc2",
			Creator:  "2",
		}, {
			OrgID:    "1",
			FolderID: "abbc3",
			Creator:  "2",
		}, {
			OrgID:    "1",
			FolderID: "abbc4",
			Creator:  "2",
		},
	})
	t.Log("batch add", err)

	if !assert.NoError(t, err) {
		return
	}
	res, err := GetSharedFolderDA().SearchSharedFolderRecords(ctx, dbo.MustGetDB(ctx), SharedFolderCondition{
		OrgIDs: []string{"1"},
	})
	if !assert.NoError(t, err) {
		return
	}
	t.Log("res:", res)
}
