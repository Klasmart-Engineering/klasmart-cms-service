package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gitlab.badanamu.com.cn/calmisland/common-cn/common"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/api"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/storage"
)

func main() {
	//获取数据库连接
	config.LoadEnvConfig()

	// init dynamodb connection
	dynamodb.GetClient()
	storage.DefaultStorage()

	log.Info(context.TODO(), "start kidsloop2 api service")
	defer log.Info(context.TODO(), "kidsloop2 api service stopped")

	common.Setenv(common.EnvHTTP)
	go common.RunWithHTTPHandler(api.NewServer(), ":8088")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
