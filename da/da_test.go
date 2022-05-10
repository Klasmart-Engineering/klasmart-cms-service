package da

import (
	"context"
	"os"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/config"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	config.LoadDBEnvConfig(ctx)
	InitMySQL(ctx)

	config.LoadRedisEnvConfig(ctx)
	InitRedis(ctx)

	os.Exit(m.Run())
}
