package da

import (
	"context"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/ro"
)

func InitMySQL(ctx context.Context) {
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		dbConf := config.Get().DBConfig
		c.DBType = dbo.NewRelicMySQL
		c.MaxIdleConns = dbConf.MaxIdleConns
		c.MaxOpenConns = dbConf.MaxOpenConns
		c.ConnMaxIdleTime = dbConf.ConnMaxIdleTime
		c.ConnMaxLifetime = dbConf.ConnMaxLifetime
		c.ConnectionString = dbConf.ConnectionString
		c.LogLevel = dbo.Info
		c.SlowThreshold = dbConf.SlowThreshold
	})
	if err != nil {
		log.Panic(ctx, "create dbo failed", log.Err(err), log.Any("config", config.Get().DBConfig))
	}
	dbo.ReplaceGlobal(dboHandler)
}

func InitRedis(ctx context.Context) {
	ro.SetConfig(config.Get().RedisConfig.Option)
	ro.MustGetRedis(ctx)
}
