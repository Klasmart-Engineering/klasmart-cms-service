package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"os"
	"os/signal"
	"syscall"

	"gitlab.badanamu.com.cn/calmisland/common-cn/common"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/api"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/storage"
)

func initDB(){
	if config.Get().DBConfig.DBMode == "mysql"{
		dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
			dbConf := config.Get().DBConfig
			c.ShowLog = dbConf.ShowLog
			c.ShowSQL = dbConf.ShowSQL
			c.MaxIdleConns = dbConf.MaxIdleConns
			c.MaxOpenConns = dbConf.MaxOpenConns
			c.ConnectionString = dbConf.ConnectionString
		})
		if err != nil{
			log.Error(context.TODO(), "create dbo failed", log.Err(err))
			panic(err)
		}
		dbo.ReplaceGlobal(dboHandler)
	}else{
		dynamodb.GetClient()
	}

}
func initCache(){
	if config.Get().RedisConfig.OpenCache {
		ro.SetConfig(&redis.Options{
			Addr:               fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
			Password:           config.Get().RedisConfig.Password,
		})
	}
}

func main() {
	//获取数据库连接
	config.LoadEnvConfig()

	//设置数据库配置
	initDB()
	initCache()

	// init dynamodb connection
	storage.DefaultStorage()

	log.Info(context.TODO(), "start kidsloop2 api service")
	defer log.Info(context.TODO(), "kidsloop2 api service stopped")

	common.Setenv(common.EnvHTTP)
	go common.RunWithHTTPHandler(api.NewServer(), ":8088")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
