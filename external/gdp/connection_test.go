package gdp

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	config.LoadAMSConfig(ctx)
	os.Exit(m.Run())
}
