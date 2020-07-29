package config

import (
	"fmt"
	"os"
	"strconv"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/log"
)

type Config struct {
	StorageConfig StorageConfig
	CDNConfig     CDNConfig
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

func assertGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Environment %v is nil", key))
	}
	return value
}

func Init() {
	config = new(Config)
	config.StorageConfig.CloudEnv = assertGetEnv("cloud_env")
	config.StorageConfig.StorageBucket = assertGetEnv("storage_bucket")
	config.StorageConfig.StorageRegion = assertGetEnv("storage_region")

	accelerateStr := assertGetEnv("storage_accelerate")
	accelerate, err := strconv.ParseBool(accelerateStr)
	if err != nil {
		log.Get().Errorf("Can't parse storage_accelerate, value: %v", accelerateStr)
		panic(err)
	}
	config.StorageConfig.Accelerate = accelerate

	cdnOpenStr := assertGetEnv("cdn_open")
	cdnOpen, err := strconv.ParseBool(cdnOpenStr)
	if err != nil {
		log.Get().Errorf("Can't parse cdn_open, value: %v", cdnOpenStr)
		panic(err)
	}
	config.CDNConfig.CDNOpen = cdnOpen

	config.CDNConfig.CDNMode = assertGetEnv("cdn_mode")
	if config.CDNConfig.CDNMode == "service" {
		config.CDNConfig.CDNServicePath = assertGetEnv("cdn_service_path")
	} else if config.CDNConfig.CDNMode == "key" {
		config.CDNConfig.CDNPath = assertGetEnv("cdn_path")
		config.CDNConfig.CDNKeyId = assertGetEnv("cdn_key_id")
		config.CDNConfig.CDNPrivateKey = assertGetEnv("cdn_private_key")
	} else {
		log.Get().Errorf("Unsupported cdn_mode, value: %v", config.CDNConfig.CDNMode)
		panic("Unsupported cdn_mode")
	}

}

func Get() *Config {
	return config
}
