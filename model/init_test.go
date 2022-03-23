package model

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/test/utils"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/kl2cache"
	"gitlab.badanamu.com.cn/calmisland/ro"
)

func initCache() {
	ro.SetConfig(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
		Password: config.Get().RedisConfig.Password,
	})
	//initDataSource()

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

func TestMain(m *testing.M) {
	ctx := context.Background()
	utils.InitConfig(ctx)
	da.InitMySQL(ctx)
	da.InitRedis(ctx)
	initDataSource()
	exitVal := m.Run()
	os.Exit(exitVal)
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
