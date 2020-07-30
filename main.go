package main

import (
	"os"
	"os/signal"
	"syscall"

	"gitlab.badanamu.com.cn/calmisland/common-cn/common"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/api"
	_ "gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/storage"
	"go.uber.org/zap"
)

func main() {
	// init logger
	lc := zap.NewDevelopmentConfig()
	lc.Encoding = "json"

	logger, _ := lc.Build()
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	// init dynamodb connection
	dynamodb.GetClient()
	storage.DefaultStorage()

	zap.L().Info("start kidsloop2 api service")
	defer zap.L().Info("kidsloop2 api service stopped")

	common.Setenv(common.EnvLAMBDA)
	go common.RunWithHTTPHandler(api.NewServer(), "")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
