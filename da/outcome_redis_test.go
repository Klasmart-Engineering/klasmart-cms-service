package da

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"testing"
)

func TestOutcomeRedis_GetOutcomeCacheBySearchCondition(t *testing.T) {
	config.LoadEnvConfig()
	ro.SetConfig(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
		Password: config.Get().RedisConfig.Password,
	})

	result := GetOutcomeRedis().GetOutcomeCacheBySearchCondition(context.Background(), &OutcomeCondition{
		IDs: dbo.NullStrings{Strings: []string{"1234"}, Valid: true},
	})
	if result != nil {
		fmt.Println("----------")
	}
}

func TestOutcomeRedis_SaveOutcomeCacheListBySearchCondition(t *testing.T) {
	config.LoadEnvConfig()
	ro.SetConfig(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
		Password: config.Get().RedisConfig.Password,
	})

	GetOutcomeRedis().SaveOutcomeCacheListBySearchCondition(context.Background(),
		&OutcomeCondition{
		IDs: dbo.NullStrings{Strings: []string{"123"}, Valid: true},
	}, &OutcomeListWithKey{0, nil})
}