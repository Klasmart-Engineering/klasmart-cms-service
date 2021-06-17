package da

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"os"
	"testing"
)

func setup() {
	config.Set(&config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: os.Getenv("connection_string"),
			MaxOpenConns:     8,
			MaxIdleConns:     8,
			ShowLog:          true,
			ShowSQL:          true,
		},
		RedisConfig: config.RedisConfig{
			OpenCache: true,
			Host:      os.Getenv("redis_host"),
			Port:      16379,
			Password:  "",
		},
	})
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		dbConf := config.Get().DBConfig
		c.ShowLog = dbConf.ShowLog
		c.ShowSQL = dbConf.ShowSQL
		c.MaxIdleConns = dbConf.MaxIdleConns
		c.MaxOpenConns = dbConf.MaxOpenConns
		c.ConnectionString = dbConf.ConnectionString
	})
	if err != nil {
		log.Error(context.TODO(), "create dbo failed", log.Err(err))
		panic(err)
	}
	dbo.ReplaceGlobal(dboHandler)
	ro.SetConfig(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
		Password: config.Get().RedisConfig.Password,
	})
}

func initOperator() *entity.Operator {
	return &entity.Operator{
		OrgID:  "8a31ebab-b879-4790-af99-ee4941a778b3",
		UserID: "2013e53e-52dd-5e1c-af0b-b503e31c8a59",
	}
}
func TestGetMilestoneDA_BatchUnLock(t *testing.T) {
	setup()
	ctx := context.TODO()
	err := GetMilestoneDA().BatchUnLock(ctx, dbo.MustGetDB(ctx), []string{"607e7089eb753a919be411cf", "6094f1e9de3afd06e06fc5d8"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("ok")
}

func TestCountTx(t *testing.T) {
	setup()
	ctx := context.TODO()
	result, err := GetMilestoneOutcomeDA().CountTx(ctx, dbo.MustGetDB(ctx), []string{"607e7089eb753a919be411cf"}, []string{"6094f1e9de3afd06e06fc5d8"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(result)
}
