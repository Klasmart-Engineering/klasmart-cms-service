package da

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/ro"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func initDB() {
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		c.ShowLog = true
		c.ShowSQL = true
		c.MaxIdleConns = 2
		c.MaxOpenConns = 4
		//c.ConnectionString = "root:Passw0rd@tcp(127.0.0.1:3306)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local"
		c.ConnectionString = os.Getenv("connection_string")
	})
	if err != nil {
		log.Error(context.TODO(), "create dbo failed", log.Err(err))
		panic(err)
	}
	config.Set(&config.Config{
		RedisConfig: config.RedisConfig{
			OpenCache: false,
			Host:      "",
			Port:      0,
			Password:  "",
		},
	})
	dbo.ReplaceGlobal(dboHandler)
}

func initRedis() {
	config.Set(&config.Config{
		RedisConfig: config.RedisConfig{
			OpenCache: true,
			Host:      "localhost",
			Port:      6379,
		},
	})
	ro.SetConfig(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
		Password: config.Get().RedisConfig.Password,
	})
}

func TestMain(m *testing.M) {
	fmt.Println("begin test")
	initDB()
	initRedis()
	m.Run()
	fmt.Println("end test")
}

func TestCreateTable(t *testing.T) {
	dsn := "root:Badanamu123456@tcp(127.0.0.1:3306)/kidsloop2_temp?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open("mysql", dsn)
	if !assert.NoError(t, err) {
		return
	}
	db.LogMode(true)
	db.AutoMigrate(entity.FolderItem{})
	db.AutoMigrate(entity.Content{})
}

func TestGroupQuery(t *testing.T) {
	fids := []string{"6071665ed41edf4799723253"}
	res, err := GetFolderDA().BatchGetFolderItemsCount(context.Background(), dbo.MustGetDB(context.Background()),
		fids)
	if err != nil {
		t.Error(err)
		return
	}
	for i := range res {
		t.Logf("%#v", res[i])
	}
}

func TestUpdateItemCountQuery(t *testing.T) {
	fids := []*entity.UpdateFolderItemsCountRequest{
		{
			ID:    "6071665ed41edf4799723253",
			Count: 6,
		},
	}
	err := GetFolderDA().BatchUpdateFolderItemsCount(context.Background(), dbo.MustGetDB(context.Background()),
		fids)
	if err != nil {
		t.Error(err)
		return
	}

}

//func TestSearchFolderContent(t *testing.T) {
//	total, folderContent, err := GetContentDA().SearchFolderContentUnsafe(context.Background(), dbo.MustGetDB(context.Background()), CombineConditions{
//		SourceCondition: &ContentCondition{Name: "test",
//			OrderBy: ContentOrderByCreatedAtDesc,
//			Pager: utils.Pager{
//				PageIndex: 0,
//				PageSize:  10,
//			},
//		},
//		TargetCondition: &ContentCondition{Name: "test2",
//			OrderBy: ContentOrderByCreatedAtDesc,
//			Pager: utils.Pager{
//				PageIndex: 0,
//				PageSize:  10,
//			},
//		},
//	}, FolderCondition{
//		Name: "plans and materials",
//	})
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	t.Log(total)
//	for i := range folderContent {
//		t.Logf("%#v\n", folderContent[i])
//	}
//}

func TestBatchUpdatePath(t *testing.T) {
	fids := []string{
		"5fc9edc9bfbf99d0a0eb2435",
		"5fc9edbebfbf99d0a0eb242d",
		"5fc9edb4bfbf99d0a0eb2422",
		"5fc9edaba06736d33312750e",
		"5fc9eda4bfbf99d0a0eb241a",
	}
	err := GetFolderDA().BatchReplaceFolderPath(context.Background(), dbo.MustGetDB(context.Background()),
		fids, "/", "/5fca13af8c9ae169ed002557/5fc88381285b97fd15b7a58f/")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("done")
}

func TestBatchUpdatePathPrefix(t *testing.T) {
	fids := []string{
		"5fb791ccb6574c2e81d2be7b",
	}
	err := GetFolderDA().BatchUpdateFolderPathPrefix(context.Background(), dbo.MustGetDB(context.Background()),
		fids, "/5fca13af8c9ae169ed002557/5fc88381285b97fd15b7a58f")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("done")
}

func TestGorm2TB(t *testing.T) {
	dbo.WithShowSQL(true)
	ctx := context.Background()
	tx := dbo.MustGetDB(ctx)
	//err := tx.Order("update_at desc").Limit(1).Exec("select * from cms_contents where delete_at=0", &entity.Content{}).Error
	err := tx.Select("* from cms_contents where delete_at=0").Order("update_at desc").Limit(1).Find(&entity.Content{}).Error

	//err := tx.Select("update cms_contents set delete_at = delete_at").Order("update_at desc").Limit(1).Find(&entity.Content{}).Error
	//err := tx.Raw("select * from cms_contents where delete_at=0").Order("update_at desc").Limit(1).Find(&entity.Content{}).Error

	//"where  order by "
	//db = db.Order("update_at desc")
	//db = db.Limit(1)
	//err := db.Find(&entity.Content{}).Error
	if err != nil {
		t.Fatal(err)
	}
}
