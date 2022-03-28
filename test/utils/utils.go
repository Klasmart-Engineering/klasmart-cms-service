package utils

import (
	"context"
	"os"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"

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
	loadAMSConfig()
	loadAssessmentServiceConfig()
	loadDataServiceConfig()

	log.Debug(ctx, "load config success",
		log.Any("config", config.Get()))
}

func loadDBEnvConfig(ctx context.Context) {
	cfg := config.Get()
	cfg.DBConfig.ConnectionString = assertGetEnv("connection_string")
	maxOpenConnsStr := os.Getenv("max_open_conns")
	maxIdleConnsStr := os.Getenv("max_idle_conns")
	showLogStr := os.Getenv("show_log")
	showSQLStr := os.Getenv("show_sql")

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

func loadAMSConfig() {
	cfg := config.Get()
	cfg.AMS.EndPoint = "https://api.alpha.kidsloop.net/user/" //os.Getenv("ams_endpoint")
}

func loadAssessmentServiceConfig() {
	cfg := config.Get()
	cfg.H5P.EndPoint = "https://api.alpha.kidsloop.net/assessment/graphql/" //os.Getenv("h5p_endpoint")
}

func loadDataServiceConfig() {
	cfg := config.Get()
	cfg.DataService.EndPoint = os.Getenv("data_service_endpoint")
	cfg.DataService.AuthorizedKey = os.Getenv("data_service_api_key")
	cfg.DataService.PublicAuthorizedKey = os.Getenv("data_service_public_key")
}
