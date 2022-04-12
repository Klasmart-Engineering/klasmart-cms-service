package da

import (
	"context"
	"os"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	config.LoadDBEnvConfig(ctx)
	InitMySQL(ctx)

	config.LoadRedisEnvConfig(ctx)
	InitRedis(ctx)

	os.Exit(m.Run())
}
