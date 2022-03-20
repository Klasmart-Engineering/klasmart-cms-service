package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"gitlab.badanamu.com.cn/calmisland/common-cn/common"
	"gitlab.badanamu.com.cn/calmisland/common-cn/helper"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/api"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/kl2cache"
	"gitlab.badanamu.com.cn/calmisland/ro"
)

func initDB() {
	ctx := context.Background()
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		dbConf := config.Get().DBConfig
		c.DBType = dbo.NewRelicMySQL
		c.ShowLog = dbConf.ShowLog
		c.ShowSQL = dbConf.ShowSQL
		c.MaxIdleConns = dbConf.MaxIdleConns
		c.MaxOpenConns = dbConf.MaxOpenConns
		c.ConnMaxIdleTime = dbConf.ConnMaxIdleTime
		c.ConnMaxLifetime = dbConf.ConnMaxLifetime
		c.ConnectionString = dbConf.ConnectionString
		c.LogLevel = dbo.Info
		c.SlowThreshold = dbConf.SlowThreshold
	})
	if err != nil {
		log.Error(ctx, "create dbo failed", log.Err(err))
		panic(err)

	}
	dbo.ReplaceGlobal(dboHandler)
}

func initDataSource() {
	//init querier
	ctx := context.Background()
	engine := cache.GetCacheEngine()
	engine.SetExpire(ctx, constant.MaxCacheExpire)
	engine.OpenCache(ctx, config.Get().RedisConfig.OpenCache)
	cache.GetPassiveCacheRefresher().SetUpdateFrequency(constant.MaxCacheExpire, constant.MinCacheExpire)

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

	ctx := context.Background()
	conf := config.Get()
	err := kl2cache.Init(ctx,
		kl2cache.OptEnable(conf.RedisConfig.OpenCache),
		kl2cache.OptRedis(conf.RedisConfig.Host, conf.RedisConfig.Port, conf.RedisConfig.Password),
		kl2cache.OptStrategyFixed(constant.MaxCacheExpire),
	)
	if err != nil {
		log.Panic(ctx, "kl2cache.Init failed", log.Err(err))
	}
}

func initLogger() {
	logger := log.New(log.WithDynamicFields(func(ctx context.Context) (fields []log.Field) {
		badaCtx, ok := helper.GetBadaCtx(ctx)
		if !ok {
			return
		}

		if badaCtx.CurrTid != "" {
			fields = append(fields, log.String("currTid", badaCtx.CurrTid))
		}

		if badaCtx.PrevTid != "" {
			fields = append(fields, log.String("prevTid", badaCtx.PrevTid))
		}

		if badaCtx.EntryTid != "" {
			fields = append(fields, log.String("entryTid", badaCtx.EntryTid))
		}

		return
	}))
	log.ReplaceGlobals(logger)
}

// @title KidsLoop 2.0 REST API
// @version 1.0.0
// @description "KidsLoop 2.0 backend rest api
// @termsOfService http://swagger.io/terms/
// @host https://kl2-test.kidsloop.net/v1
func main() {
	ctx := context.Background()

	initLogger()

	log.Info(ctx, "start kidsloop2 api service")
	defer func() {
		if err := recover(); err != nil {
			log.Info(ctx, "kidsloop2 api service stopped", log.Any("err", err))
		} else {
			log.Info(ctx, "kidsloop2 api service stopped")
		}
	}()

	log.Debug(ctx, "build information",
		log.String("gitHash", constant.GitHash),
		log.String("buildTimestamp", constant.BuildTimestamp),
		log.String("latestMigrate", constant.LatestMigrate))

	// read config
	config.LoadEnvConfig()

	log.Debug(ctx, "load config successfully", log.Any("config", config.Get()))

	// init database connection
	initDB()

	log.Debug(ctx, "init db successfully")
	initCache()

	log.Debug(ctx, "init cache successfully")
	// init dynamodb connection
	storage.DefaultStorage()

	log.Debug(ctx, "init storage successfully")

	if os.Getenv("env") == "HTTP" {
		common.Setenv(common.EnvHTTP)
	} else {
		common.Setenv(common.EnvLAMBDA)
	}

	go common.RunWithHTTPHandler(api.NewServer(), ":8088")

	log.Debug(ctx, "init api server successfully")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
