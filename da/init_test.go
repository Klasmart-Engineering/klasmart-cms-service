package da

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/test/utils"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	logger.SetLevel(logrus.DebugLevel)
	utils.InitConfig(ctx)
	utils.InitDB(ctx)
	exitVal := m.Run()
	os.Exit(exitVal)
}
