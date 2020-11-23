package da

import (
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"testing"
)

func initDB() {
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		c.ShowLog = true
		c.ShowSQL = true
		c.MaxIdleConns = 2
		c.MaxOpenConns = 4
		c.ConnectionString = "root:Badanamu123456@tcp(192.168.1.234:3310)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local"
	})
	if err != nil {
		log.Error(context.TODO(), "create dbo failed", log.Err(err))
		panic(err)
	}
	config.Set(&config.Config{
		RedisConfig:     config.RedisConfig{
			OpenCache: false,
			Host:      "",
			Port:      0,
			Password:  "",
		},
	})
	dbo.ReplaceGlobal(dboHandler)
}

func TestMain(m *testing.M) {
	fmt.Println("begin test")
	initDB()
	m.Run()
	fmt.Println("end test")
}

func TestCreateTable(t *testing.T) {
	dsn := "root:Badanamu123456@tcp(192.168.1.234:3310)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open("mysql", dsn)
	if !assert.NoError(t, err) {
		return
	}
	db.LogMode(true)
	db.AutoMigrate(entity.FolderItem{})
}

func TestSearchFolderContent(t *testing.T) {
	total, folderContent, err := GetContentDA().SearchFolderContentUnsafe(context.Background(), dbo.MustGetDB(context.Background()), CombineConditions{
		SourceCondition: &ContentCondition{Name: "test",
			OrderBy: ContentOrderByCreatedAtDesc,
			Pager: utils.Pager{
				PageIndex: 0,
				PageSize:  10,
			},
		},
		TargetCondition: &ContentCondition{Name: "test2",
			OrderBy: ContentOrderByCreatedAtDesc,
			Pager: utils.Pager{
				PageIndex: 0,
				PageSize:  10,
			},
		},
	}, FolderCondition{
		Name:              "root",
	})
	if err != nil{
		t.Error(err)
		return
	}
	t.Log(total)
	for i := range folderContent{
		t.Logf("%#v\n", folderContent[i])
	}
}