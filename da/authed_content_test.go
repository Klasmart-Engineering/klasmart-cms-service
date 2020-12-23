package da

import (
	"context"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func TestCreateAuthedContentTable(t *testing.T) {
	dsn := "root:Badanamu123456@tcp(192.168.1.234:3310)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open("mysql", dsn)
	if !assert.NoError(t, err) {
		return
	}
	db.LogMode(true)
	db.AutoMigrate(entity.AuthedContentRecord{})
	db.AutoMigrate(entity.SharedFolderRecord{})
}

func TestAddAndBatchAddRecords(t *testing.T) {
	ctx := context.Background()
	err := GetAuthedContentRecordsDA().AddAuthedContent(ctx, dbo.MustGetDB(ctx), entity.AuthedContentRecord{
		OrgID:     "1",
		ContentID: "abbc",
		Creator:   "2",
	})
	if !assert.NoError(t, err) {
		return
	}
	t.Log("add:", err)
	err = GetAuthedContentRecordsDA().BatchAddAuthedContent(ctx, dbo.MustGetDB(ctx), []*entity.AuthedContentRecord{
		{
			OrgID:     "1",
			ContentID: "abbc2",
			Creator:   "2",
		}, {
			OrgID:     "1",
			ContentID: "abbc3",
			Creator:   "2",
		}, {
			OrgID:     "1",
			ContentID: "abbc4",
			Creator:   "2",
		},
	})
	t.Log("batch add", err)

	if !assert.NoError(t, err) {
		return
	}
	total, res, err := GetAuthedContentRecordsDA().SearchAuthedContentRecords(ctx, dbo.MustGetDB(ctx), AuthedContentCondition{
		OrgIDs: []string{"1"},
	})
	if !assert.NoError(t, err) {
		return
	}
	t.Log("total:", total)
	t.Log("res:", res)
}

func TestSearchRecords(t *testing.T) {
	ctx := context.Background()

	total, res, err := GetAuthedContentRecordsDA().SearchAuthedContentRecords(ctx, dbo.MustGetDB(ctx), AuthedContentCondition{
		OrgIDs: []string{"1"},
	})
	if !assert.NoError(t, err) {
		return
	}
	t.Log("total:", total)
	for i := range res {
		t.Logf("res: %#v", res[i])
	}
}
