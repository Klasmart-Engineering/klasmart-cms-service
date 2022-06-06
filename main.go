package main

import (
	"context"
	"os"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/decorator"
	"github.com/KL-Engineering/kidsloop-cache/cache"
	"github.com/KL-Engineering/kidsloop-cms-service/api"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/KL-Engineering/kidsloop-cms-service/model/storage"
	"github.com/KL-Engineering/kidsloop-cms-service/utils/kl2cache"
	"github.com/KL-Engineering/tracecontext"
)

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
func initCache(ctx context.Context) {
	da.InitRedis(ctx)
	initDataSource()

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
		badaCtx, ok := tracecontext.GetTraceContext(ctx)
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
	}), log.WithLogLevel(config.Get().Log.Level))
	log.ReplaceGlobals(logger)
}

// @title KidsLoop 2.0 REST API
// @version 1.0.0
// @description "KidsLoop 2.0 backend rest api
// @termsOfService http://swagger.io/terms/
// @host https://kl2-test.kidsloop.net/v1
func main() {
	ctx := context.Background()

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
	// previous log used the default logger, here we replace it with env configured params
	initLogger()

	log.Debug(ctx, "load config successfully", log.Any("config", config.Get()))

	da.InitMySQL(ctx)
	log.Debug(ctx, "init db successfully")

	initCache(ctx)
	log.Debug(ctx, "init cache successfully")

	storage.DefaultStorage()
	log.Debug(ctx, "init storage successfully")

	if os.Getenv("env") == "HTTP" {
		decorator.Setenv(decorator.EnvHTTP)
	} else {
		decorator.Setenv(decorator.EnvLAMBDA)
	}

	go decorator.RunWithHTTPHandler(api.NewServer(), ":8088")

	log.Debug(ctx, "init api server successfully")

	select {}
}
