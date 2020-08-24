package config

import (
	"context"
	"os"
	"strconv"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

type Config struct {
	StorageConfig StorageConfig
	CDNConfig     CDNConfig
	Schedule      ScheduleConfig `json:"daschedule" yaml:"daschedule"`
	DBConfig      DBConfig
	RedisConfig   RedisConfig
}

var config *Config

type RedisConfig struct {
	OpenCache bool
	Host      string
	Port      int
	Password  string
}
type DBConfig struct {
	DBMode string `json:"db_mode"`

	ConnectionString string `json:"connection_string"`
	MaxOpenConns     int    `json:"max_open_conns"`
	MaxIdleConns     int    `json:"max_idle_conns"`
	ShowLog          bool   `json:"show_log"`
	ShowSQL          bool   `json:"show_sql"`

	DynamoEndPoint string `json:"dynamo_end_point"`
	DynamoRegion   string `json:"dynamo_region"`
}

type StorageConfig struct {
	Accelerate    bool   `yaml:"accelerate"`
	CloudEnv      string `yaml:"cloud_env"`
	StorageBucket string `yaml:"storage_bucket"`
	StorageRegion string `yaml:"storage_region"`
}

type CDNConfig struct {
	CDNOpen bool
	CDNMode string

	CDNPath       string
	CDNKeyId      string
	CDNPrivateKey string

	CDNServicePath  string
	CDNServiceToken string
}

type ScheduleConfig struct {
	MaxRepeatYear int `json:"max_repeat_year" yaml:"max_repeat_year"`
}

func assertGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Panic(context.TODO(), "Environment is nil")
	}
	return value
}

func LoadEnvConfig() {
	ctx := context.TODO()
	config = new(Config)
	loadStorageEnvConfig(ctx)
	loadDBEnvConfig(ctx)
	loadRedisEnvConfig(ctx)
}
func loadStorageEnvConfig(ctx context.Context) {
	config.StorageConfig.CloudEnv = assertGetEnv("cloud_env")
	config.StorageConfig.StorageBucket = assertGetEnv("storage_bucket")
	config.StorageConfig.StorageRegion = assertGetEnv("storage_region")

	accelerateStr := assertGetEnv("storage_accelerate")
	accelerate, err := strconv.ParseBool(accelerateStr)
	if err != nil {
		log.Panic(ctx, "Can't parse storage_accelerate",
			log.Err(err),
			log.String("accelerateStr", accelerateStr))
	}
	config.StorageConfig.Accelerate = accelerate

	cdnOpenStr := assertGetEnv("cdn_open")
	cdnOpen, err := strconv.ParseBool(cdnOpenStr)
	if err != nil {
		log.Panic(ctx, "Can't parse cdn_open",
			log.Err(err),
			log.String("cdnOpenStr", cdnOpenStr))
	}
	config.CDNConfig.CDNOpen = cdnOpen

	if cdnOpen {
		config.CDNConfig.CDNMode = assertGetEnv("cdn_mode")
		if config.CDNConfig.CDNMode == "service" {
			config.CDNConfig.CDNPath = assertGetEnv("cdn_path")
			config.CDNConfig.CDNServicePath = assertGetEnv("cdn_service_path")
			config.CDNConfig.CDNServiceToken = assertGetEnv("cdn_service_token")
		} else if config.CDNConfig.CDNMode == "key" {
			config.CDNConfig.CDNPath = assertGetEnv("cdn_path")
			config.CDNConfig.CDNKeyId = assertGetEnv("cdn_key_id")
			config.CDNConfig.CDNPrivateKey = assertGetEnv("cdn_private_key")
		} else {
			log.Panic(ctx, "Unsupported cdn_mode", log.String("CDNMode", config.CDNConfig.CDNMode))
		}
	}

	maxRepeatYearStr := strings.TrimSpace(os.Getenv("max_repeat_year"))
	if maxRepeatYearStr == "" {
		config.Schedule.MaxRepeatYear = 2
	} else {
		i, err := strconv.Atoi(maxRepeatYearStr)
		if err != nil {
			log.Panic(ctx, "parse env max_repeat_year failed", log.String("max_repeat_year", maxRepeatYearStr))
		}
		config.Schedule.MaxRepeatYear = i
	}

}

func loadRedisEnvConfig(ctx context.Context) {
	openCacheStr := os.Getenv("open_cache")
	openCache, _ := strconv.ParseBool(openCacheStr)
	config.RedisConfig.OpenCache = openCache
	if openCache {
		host := assertGetEnv("redis_host")
		portStr := assertGetEnv("redis_port")
		password := os.Getenv("redis_password")
		config.RedisConfig.Host = host
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Error(ctx, "Can't parse redis_port")
			port = 3306
		}
		config.RedisConfig.Port = port
		config.RedisConfig.Password = password
	}
}

func loadDBEnvConfig(ctx context.Context) {
	config.DBConfig.DBMode = os.Getenv("db_env")

	if config.DBConfig.DBMode == "mysql" {
		config.DBConfig.ConnectionString = assertGetEnv("connection_string")
		maxOpenConnsStr := assertGetEnv("max_open_conns")
		maxIdleConnsStr := assertGetEnv("max_idle_conns")
		showLogStr := assertGetEnv("show_log")
		showSQLStr := assertGetEnv("show_sql")

		maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
		if err != nil {
			log.Error(ctx, "Can't parse max_open_conns")
			maxOpenConns = 16
		}
		config.DBConfig.MaxOpenConns = maxOpenConns

		maxIdleConns, err := strconv.Atoi(maxIdleConnsStr)
		if err != nil {
			log.Error(ctx, "Can't parse max_idle_conns")
			maxOpenConns = 16
		}
		config.DBConfig.MaxIdleConns = maxIdleConns

		showLog, err := strconv.ParseBool(showLogStr)
		if err != nil {
			log.Error(ctx, "Can't parse show_log")
			showLog = true
		}
		config.DBConfig.ShowLog = showLog

		showSQL, err := strconv.ParseBool(showSQLStr)
		if err != nil {
			log.Error(ctx, "Can't parse show_sql")
			showLog = true
		}
		config.DBConfig.ShowSQL = showSQL
	} else {
		config.DBConfig.DynamoEndPoint = assertGetEnv("dynamo_end_point")
		config.DBConfig.DynamoRegion = assertGetEnv("dynamo_region")
	}

}

func Get() *Config {
	return config
}
