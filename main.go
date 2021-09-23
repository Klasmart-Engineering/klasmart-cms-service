package main

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"gitlab.badanamu.com.cn/calmisland/common-cn/common"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/api"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"gitlab.badanamu.com.cn/calmisland/ro"

	logger "gitlab.badanamu.com.cn/calmisland/common-cn/logger"
)

func initDB() {
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
}

func initDataSource() {
	//init querier
	ctx := context.Background()
	engine := cache.GetCacheEngine()
	engine.AddDataSource(ctx, external.GetUserServiceProvider())
	engine.AddDataSource(ctx, external.GetTeacherServiceProvider())
	engine.AddDataSource(ctx, external.GetSubjectServiceProvider())
	engine.AddDataSource(ctx, external.GetSubCategoryServiceProvider())
	engine.AddDataSource(ctx, external.GetStudentServiceProvider())
	engine.AddDataSource(ctx, external.GetSchoolServiceProvider())
	engine.AddDataSource(ctx, external.GetProgramServiceProvider())
	engine.AddDataSource(ctx, external.GetOrganizationServiceProvider())
	engine.AddDataSource(ctx, external.GetGradeServiceProvider())
	engine.AddDataSource(ctx, external.GetClassServiceProvider())
	engine.AddDataSource(ctx, external.GetCategoryServiceProvider())
	engine.AddDataSource(ctx, external.GetAgeServiceProvider())
}
func initCache() {
	ro.SetConfig(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
		Password: config.Get().RedisConfig.Password,
	})
	initDataSource()
}

// @title KidsLoop 2.0 REST API
// @version 1.0.0
// @description "KidsLoop 2.0 backend rest api
// @termsOfService http://swagger.io/terms/
// @host https://kl2-test.kidsloop.net/v1
func main() {
	log.Info(context.TODO(), "start kidsloop2 api service")
	defer func() {
		if err := recover(); err != nil {
			log.Info(context.TODO(), "kidsloop2 api service stopped", log.Any("err", err))
		} else {
			log.Info(context.TODO(), "kidsloop2 api service stopped")
		}
	}()

	// temp solution, will remove in next version
	logger.SetLevel(logrus.DebugLevel)

	log.Debug(context.TODO(), "build information",
		log.String("gitHash", constant.GitHash),
		log.String("buildTimestamp", constant.BuildTimestamp),
		log.String("latestMigrate", constant.LatestMigrate))

	// read config
	config.LoadEnvConfig()

	log.Debug(context.TODO(), "load config success", log.Any("config", config.Get()))

	// init database connection
	initDB()

	log.Debug(context.TODO(), "init db success")
	initCache()

	log.Debug(context.TODO(), "init cache success")
	// init dynamodb connection
	storage.DefaultStorage()

	log.Debug(context.TODO(), "init storage success")

	if os.Getenv("env") == "HTTP" {
		common.Setenv(common.EnvHTTP)
	} else {
		common.Setenv(common.EnvLAMBDA)
	}

	go common.RunWithHTTPHandler(api.NewServer(), ":8088")

	log.Debug(context.TODO(), "init api server success")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
