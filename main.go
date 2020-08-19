package main

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"os"
	"os/signal"
	"syscall"

	"gitlab.badanamu.com.cn/calmisland/common-cn/common"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/api"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/storage"
)

func initDBO(){
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
}

func main() {
	//获取数据库连接
	config.LoadEnvConfig()

	//设置数据库配置
	//initDBO()
	dynamodb.GetClient()
	//ro.SetConfig(&redis.ClusterOptions{})

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
