package config

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"os"
	"strconv"
	"strings"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type Config struct {
	StorageConfig StorageConfig  `yaml:"storage_config"`
	CDNConfig     CDNConfig      `yaml:"cdn_config"`
	Schedule      ScheduleConfig `json:"schedule" yaml:"schedule"`
	DBConfig      DBConfig       `yaml:"db_config"`
	RedisConfig   RedisConfig    `yaml:"redis_config"`

	CryptoConfig    CryptoConfig    `yaml:"crypto_config"`
	LiveTokenConfig LiveTokenConfig `yaml:"live_token_config"`
}

var config *Config

type CryptoConfig struct {
	PrivateKey string `yaml:"crypto_private_key"`
}

type RedisConfig struct {
	OpenCache bool   `yaml:"open_cache"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Password  string `yaml:"password"`
}

type DBConfig struct {
	DBMode string `yaml:"db_mode"`

	ConnectionString string `yaml:"connection_string"`
	MaxOpenConns     int    `yaml:"max_open_conns"`
	MaxIdleConns     int    `yaml:"max_idle_conns"`
	ShowLog          bool   `yaml:"show_log"`
	ShowSQL          bool   `yaml:"show_sql"`

	DynamoEndPoint string `yaml:"dynamo_end_point"`
	DynamoRegion   string `yaml:"dynamo_region"`
}

type StorageConfig struct {
	Accelerate    bool   `yaml:"accelerate"`
	CloudEnv      string `yaml:"cloud_env"`
	StorageBucket string `yaml:"storage_bucket"`
	StorageRegion string `yaml:"storage_region"`
}

type CDNConfig struct {
	CDNOpen bool   `yaml:"cdn_open"`
	CDNMode string `yaml:"cdn_mode"`

	CDNPath       string `yaml:"cdn_path"`
	CDNKeyId      string `yaml:"cdn_key_id"`
	CDNPrivateKey string `yaml:"cdn_private_key"`

	CDNServicePath  string `yaml:"cdn_service_path"`
	CDNServiceToken string `yaml:"cdn_service_token"`
}

type ScheduleConfig struct {
	MaxRepeatYear   int           `json:"max_repeat_year" yaml:"max_repeat_year"`
	CacheExpiration time.Duration `yaml:"cache_expiration"`
}

type LiveTokenConfig struct {
	PrivateKey interface{} `yaml:"private_key"`
	//PublicKey  string      `yaml:"public_key"`
}

func assertGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Panic(context.TODO(), "Environment is nil", log.String("key", key))
	}
	return value
}

func LoadEnvConfig() {
	ctx := context.TODO()
	config = new(Config)
	loadStorageEnvConfig(ctx)
	loadDBEnvConfig(ctx)
	loadRedisEnvConfig(ctx)
	loadScheduleEnvConfig(ctx)
	loadCryptoEnvConfig(ctx)
	loadLiveTokenEnvConfig(ctx)
}

func loadCryptoEnvConfig(ctx context.Context) {
	config.CryptoConfig.PrivateKey = os.Getenv("crypto_private_key")
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
		config.CDNConfig.CDNPath = assertGetEnv("cdn_path")
		if config.CDNConfig.CDNMode == "service" {
			config.CDNConfig.CDNServicePath = assertGetEnv("cdn_service_path")
			config.CDNConfig.CDNServiceToken = assertGetEnv("cdn_service_token")
		} else if config.CDNConfig.CDNMode == "key" {
			config.CDNConfig.CDNKeyId = assertGetEnv("cdn_key_id")
			config.CDNConfig.CDNPrivateKey = assertGetEnv("cdn_private_key")
		} else {
			log.Panic(ctx, "Unsupported cdn_mode", log.String("CDNMode", config.CDNConfig.CDNMode))
		}
	}
}

func loadRedisEnvConfig(ctx context.Context) {
	openCacheStr := os.Getenv("open_cache")
	openCache, _ := strconv.ParseBool(openCacheStr)
	config.RedisConfig.OpenCache = openCache
	// if openCache {
	host := assertGetEnv("redis_host")
	portStr := assertGetEnv("redis_port")
	password := os.Getenv("redis_password")
	config.RedisConfig.Host = host
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Error(ctx, "Can't parse redis_port", log.Err(err))
		port = 3306
	}
	config.RedisConfig.Port = port
	config.RedisConfig.Password = password
	// }
}

func loadScheduleEnvConfig(ctx context.Context) {
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
	cacheExpiration, err := time.ParseDuration(os.Getenv("cache_expiration"))
	if err != nil {
		config.Schedule.CacheExpiration = constant.ScheduleDefaultCacheExpiration
	} else {
		config.Schedule.CacheExpiration = cacheExpiration
	}
}

func loadDBEnvConfig(ctx context.Context) {
	config.DBConfig.ConnectionString = assertGetEnv("connection_string")
	maxOpenConnsStr := assertGetEnv("max_open_conns")
	maxIdleConnsStr := assertGetEnv("max_idle_conns")
	showLogStr := assertGetEnv("show_log")
	showSQLStr := assertGetEnv("show_sql")

	maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
	if err != nil {
		log.Error(ctx, "Can't parse max_open_conns", log.Err(err))
		maxOpenConns = 16
	}
	config.DBConfig.MaxOpenConns = maxOpenConns

	maxIdleConns, err := strconv.Atoi(maxIdleConnsStr)
	if err != nil {
		log.Error(ctx, "Can't parse max_idle_conns", log.Err(err))
		maxOpenConns = 16
	}
	config.DBConfig.MaxIdleConns = maxIdleConns

	showLog, err := strconv.ParseBool(showLogStr)
	if err != nil {
		log.Error(ctx, "Can't parse show_log", log.Err(err))
		showLog = true
	}
	config.DBConfig.ShowLog = showLog

	showSQL, err := strconv.ParseBool(showSQLStr)
	if err != nil {
		log.Error(ctx, "Can't parse show_sql", log.Err(err))
		showLog = true
	}
	config.DBConfig.ShowSQL = showSQL

}

func loadLiveTokenEnvConfig(ctx context.Context) {
	//privateKeyPath := os.Getenv("live_token_private_key_path") //"./auth_private_key.pem"
	//content, err := ioutil.ReadFile(privateKeyPath)
	//if err != nil {
	//	log.Error(ctx, "loadAuthEnvConfig:load auth config error", log.Err(err), log.String("privateKeyPath", privateKeyPath))
	//	return
	//}
	content := `
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDAGN9KAcc61KBz8EQAH54bFwGK6PEQNVXXlsObwFd3Zos83bRm
+3grzP0pKWniZ6TL/y7ZgFh4OlUMh9qJjIt6Lpz9l4uDxkgDDrKHn8IrflBxjJKq
0OyXqwIYChnFoi/HGjcRtJhi8oTFToSvKMqIeUuLmWmLA8nXdDnMl7zwoQIDAQAB
AoGAHi0KDn8fA9/Y4L2SgQ52cLz5cg/LpocqV/aH/dSGKOyD3Oja6P6BzyehcTDf
QECVw7Hvcx1VSHWpXJGOw+K/Ggmt/+k+vxQKOuauFLPV72dKUChYQWXZnUWp7Ok2
wui1TbW3HIKQ3D5FujjQYxX3V9u8Y777F4icGSR3ie+OvZ0CQQDoqFxun6EBFVp+
sczV5wLKTjLRicBh+YEg4bMw28BWTlpVK1DA8kLTy9IEicxvj4/57fbyN20LiUW8
ne0kSWi/AkEA016+q5QGT0xljLiOxufvNNLwHIafPBKQ4CJ36u4yRKOfEvT4b9Kd
xE8Oh3WnW8vljB2pdQTyYuOAEqcgmUGenwJBALdRyZskzmEjKS4A/OxiXPF5ElPG
nb7VMOjuzhmmXXPjwwuu2K9fdEacJ/yJc3tH/GMrHNSX1aUsYbWQHoXkDdMCQCn2
jl4b9iC6JxMH9PiSRVA0bI0NQQG5IeANl8chYQN1hHhMACKbKs01cMn91qsH0Nu/
a8wanlB5oAyT94nVmDsCQQDctlxgTAyLwqHYdvsfZ+ao6xaTZkrhscU+PL81EXGQ
byUU2j8ZyKNaLnzwnHUOoolzxoaUryO+vdWT+Sy7y73D
-----END RSA PRIVATE KEY-----
`
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(content))
	if err != nil {
		log.Panic(ctx, "CreateJWT:create jwt error", log.Err(err))
	}
	config.LiveTokenConfig.PrivateKey = key
}

func Get() *Config {
	return config
}
