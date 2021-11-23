package utils

import (
	"context"
	"os"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func InitOperator(ctx context.Context, tokenString, orgID string) *entity.Operator {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		log.Panic(ctx, "invalid token", log.Err(err))
	}

	var userID string
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if id, ok := claims["id"]; ok {
			userID = id.(string)
		}
	}

	if userID == "" {
		log.Panic(ctx, "invalid token, id is empty", log.String("token", tokenString))
	}

	return &entity.Operator{
		OrgID:  orgID,
		UserID: userID,
		Token:  tokenString,
	}
}

func InitConfig(ctx context.Context) {
	config.Set(&config.Config{})
	loadDBEnvConfig(ctx)
	log.Debug(ctx, "load config success",
		log.Any("config", config.Get()))
}

func InitDB(ctx context.Context) {
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		dbConf := config.Get().DBConfig
		c.ShowLog = dbConf.ShowLog
		c.ShowSQL = dbConf.ShowSQL
		c.MaxIdleConns = dbConf.MaxIdleConns
		c.MaxOpenConns = dbConf.MaxOpenConns
		c.ConnectionString = dbConf.ConnectionString
	})
	if err != nil {
		log.Panic(ctx, "initDB failed",
			log.Err(err))
	}
	dbo.ReplaceGlobal(dboHandler)
}

func loadDBEnvConfig(ctx context.Context) {
	cfg := config.Get()
	cfg.DBConfig.ConnectionString = assertGetEnv("connection_string")
	maxOpenConnsStr := assertGetEnv("max_open_conns")
	maxIdleConnsStr := assertGetEnv("max_idle_conns")
	showLogStr := assertGetEnv("show_log")
	showSQLStr := assertGetEnv("show_sql")

	maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
	if err != nil {
		log.Error(ctx, "Can't parse max_open_conns", log.Err(err))
		maxOpenConns = 16
	}
	cfg.DBConfig.MaxOpenConns = maxOpenConns

	maxIdleConns, err := strconv.Atoi(maxIdleConnsStr)
	if err != nil {
		log.Error(ctx, "Can't parse max_idle_conns", log.Err(err))
		maxIdleConns = 16
	}
	cfg.DBConfig.MaxIdleConns = maxIdleConns

	showLog, err := strconv.ParseBool(showLogStr)
	if err != nil {
		log.Error(ctx, "Can't parse show_log", log.Err(err))
		showLog = true
	}
	cfg.DBConfig.ShowLog = showLog

	showSQL, err := strconv.ParseBool(showSQLStr)
	if err != nil {
		log.Error(ctx, "Can't parse show_sql", log.Err(err))
		showSQL = true
	}
	cfg.DBConfig.ShowSQL = showSQL
	config.Set(cfg)
}

func assertGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Panic(context.TODO(), "Environment is nil", log.String("key", key))
	}
	return value
}
