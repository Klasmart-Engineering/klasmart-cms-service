package da

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/ro"
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

func TestOutcomeRedis_GetConditions(t *testing.T) {
	c := &OutcomeCondition{
		PublishStatus:  dbo.NullStrings{Strings: []string{"published"}, Valid: true},
		OrganizationID: sql.NullString{String: "50dbe15c-1db5-4b22-8819-34562a2f79e4", Valid: true},
		FuzzyKey:       sql.NullString{String: "ss", Valid: true},
		AuthorIDs: dbo.NullStrings{Strings: []string{
			"685ec351-f369-5c80-85e1-1321dbc1c275",
			"bcbd3c50-f9d0-520a-99f4-f2bb89db0da3",
			"9fb01c8e-6d5f-5e48-a205-007f4d9d3bdf",
			"28b6a2a2-50e4-55ce-aad3-18eea8febeb5",
		}, Valid: true},
	}
	c.GetConditions()
}

func TestOutcomeRedis_CleanOutcomeConditionCache(t *testing.T) {
	config.LoadEnvConfig()
	ro.SetConfig(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
		Password: config.Get().RedisConfig.Password,
	})
	condition := OutcomeCondition{
		IDs: dbo.NullStrings{Strings: []string{"123"}, Valid: true},
	}
	GetOutcomeRedis().CleanOutcomeConditionCache(context.Background(), &condition)
}
