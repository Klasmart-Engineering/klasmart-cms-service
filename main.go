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
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"go.uber.org/zap"
)

func main() {
	//获取数据库连接
	config.LoadEnvConfig()
	// init logger
	logger, _ := zap.NewDevelopment(zap.AddCaller())
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
