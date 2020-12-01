package config

import (
	"context"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

var (
	StorageDownloadNativeMode     StorageDownloadMode = "native"
	StorageDownloadCloudFrontMode StorageDownloadMode = "cloudfront"
)

type StorageDownloadMode string

func NewStorageDownloadMode(mode string) StorageDownloadMode {
	if mode == "native" {
		return StorageDownloadNativeMode
	} else {
		return StorageDownloadCloudFrontMode
	}
}

type Config struct {
	StorageConfig   StorageConfig    `yaml:"storage_config"`
	CDNConfig       CDNConfig        `yaml:"cdn_config"`
	Schedule        ScheduleConfig   `json:"schedule" yaml:"schedule"`
	DBConfig        DBConfig         `yaml:"db_config"`
	RedisConfig     RedisConfig      `yaml:"redis_config"`
	CryptoConfig    CryptoConfig     `yaml:"crypto_config"`
	LiveTokenConfig LiveTokenConfig  `yaml:"live_token_config"`
	Assessment      AssessmentConfig `yaml:"assessment_config"`
	AMS             AMSConfig        `json:"ams" yaml:"ams"`
}

var config *Config

type CryptoConfig struct {
	PrivateKeyPath string `yaml:"h5p_private_key_path"`
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
	Accelerate bool   `yaml:"accelerate"`
	CloudEnv   string `yaml:"cloud_env"`

	StorageEndPoint string `yaml:"storage_end_point"`
	StorageBucket   string `yaml:"storage_bucket"`
	StorageRegion   string `yaml:"storage_region"`

	StorageDownloadMode StorageDownloadMode `yaml:"storage_download_mode"`
	StorageSigMode      bool                `yaml:"storage_sig_mode"`
}

type CDNConfig struct {
	CDNPath           string `yaml:"cdn_path"`
	CDNKeyId          string `yaml:"cdn_key_id"`
	CDNPrivateKeyPath string `yaml:"cdn_private_key_path"`
}

type ScheduleConfig struct {
	MaxRepeatYear   int           `json:"max_repeat_year" yaml:"max_repeat_year"`
	CacheExpiration time.Duration `yaml:"cache_expiration"`
}

type LiveTokenConfig struct {
	PrivateKey interface{} `yaml:"private_key"`
	//PublicKey  string      `yaml:"public_key"`
}

type AssessmentConfig struct {
	//CacheExpiration     time.Duration `yaml:"cache_expiration"`
	AddAssessmentSecret interface{} `json:"add_assessment_secret"`
}

type AMSConfig struct {
	EndPoint       string      `json:"endpoint" yaml:"endpoint"`
	TokenVerifyKey interface{} `json:"token_verify_key" yaml:"token_verify_key"`
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
	loadAMSConfig(ctx)
	loadAssessmentConfig(ctx)
}

func loadCryptoEnvConfig(ctx context.Context) {
	config.CryptoConfig.PrivateKeyPath = os.Getenv("h5p_private_key_path")
}

func loadStorageEnvConfig(ctx context.Context) {
	config.StorageConfig.CloudEnv = assertGetEnv("cloud_env")
	config.StorageConfig.StorageBucket = assertGetEnv("storage_bucket")
	config.StorageConfig.StorageRegion = assertGetEnv("storage_region")
	config.StorageConfig.StorageEndPoint = os.Getenv("storage_endpoint")
	config.StorageConfig.StorageDownloadMode = NewStorageDownloadMode(assertGetEnv("storage_download_mode"))
	storageSigMode := assertGetEnv("storage_sig_mode") == "true"
	config.StorageConfig.StorageSigMode = storageSigMode

	accelerateStr := assertGetEnv("storage_accelerate")
	accelerate, err := strconv.ParseBool(accelerateStr)
	if err != nil {
		log.Panic(ctx, "Can't parse storage_accelerate",
			log.Err(err),
			log.String("accelerateStr", accelerateStr))
	}
	config.StorageConfig.Accelerate = accelerate

	if config.StorageConfig.StorageDownloadMode == StorageDownloadCloudFrontMode {
		config.CDNConfig.CDNPath = assertGetEnv("cdn_path")
		if config.StorageConfig.StorageSigMode {
			config.CDNConfig.CDNKeyId = assertGetEnv("cdn_key_id")
			config.CDNConfig.CDNPrivateKeyPath = assertGetEnv("cdn_private_key_path")
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
		log.Error(ctx, "Can't parse redis_port", log.Err(err), log.String("portStr", portStr))
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
	privateKeyPath := os.Getenv("live_token_private_key_path") //"./live_token_private_key.pem"
	content, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		log.Panic(ctx, "loadAuthEnvConfig:load auth config error", log.Err(err), log.String("privateKeyPath", privateKeyPath))
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(content))
	if err != nil {
		log.Panic(ctx, "CreateJWT:create jwt error", log.Err(err))
	}
	config.LiveTokenConfig.PrivateKey = key
}

func loadAssessmentConfig(ctx context.Context) {
	//cacheExpiration, err := time.ParseDuration(os.Getenv("assessment_cache_expiration"))
	//if err != nil {
	//	config.Assessment.CacheExpiration = constant.ScheduleDefaultCacheExpiration
	//} else {
	//	config.Assessment.CacheExpiration = cacheExpiration
	//}

	publicKeyPath := os.Getenv("ams_assessment_jwt_public_key_path")
	content, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		log.Panic(ctx, "load assessment config: load public key failed", log.Err(err), log.String("public_key_path", publicKeyPath))
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(content)
	if err != nil {
		log.Panic(ctx, "load assessment config: ParseRSAPublicKeyFromPEM failed", log.Err(err))
	}
	config.Assessment.AddAssessmentSecret = key
}

func loadAMSConfig(ctx context.Context) {
	config.AMS.EndPoint = assertGetEnv("ams_endpoint")
	publicKeyPath := os.Getenv("jwt_public_key_path") //"./jwt_public_key.pem"
	content, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		log.Panic(ctx, "loadAMSConfig:load public key failed", log.Err(err), log.String("publicKeyPath", publicKeyPath))
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(content)
	if err != nil {
		log.Panic(ctx, "loadAMSConfig:ParseRSAPublicKeyFromPEM failed", log.Err(err))
	}
	config.AMS.TokenVerifyKey = key
}

func Get() *Config {
	return config
}

func Set(c *Config) {
	config = c
}
