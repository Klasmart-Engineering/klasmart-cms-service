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
	Schedule      ScheduleConfig `json:"schedule" yaml:"schedule"`
}

var config *Config

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

	CDNServicePath string
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
			config.CDNConfig.CDNServicePath = assertGetEnv("cdn_service_path")
		} else if config.CDNConfig.CDNMode == "key" {
			config.CDNConfig.CDNPath = assertGetEnv("cdn_path")
			config.CDNConfig.CDNKeyId = assertGetEnv("cdn_key_id")
			config.CDNConfig.CDNPrivateKey = assertGetEnv("cdn_private_key")
		} else {
			log.Panic(ctx, "Unsupported cdn_mode", log.String("CDNMode", config.CDNConfig.CDNMode))
		}
		config.CDNConfig.CDNMode = assertGetEnv("cdn_mode")
		if config.CDNConfig.CDNMode == "service" {
			config.CDNConfig.CDNServicePath = assertGetEnv("cdn_service_path")
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

func Get() *Config {
	return config
}
