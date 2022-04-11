package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
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
	StorageConfig         StorageConfig         `yaml:"storage_config"`
	CDNConfig             CDNConfig             `yaml:"cdn_config"`
	Schedule              ScheduleConfig        `json:"schedule" yaml:"schedule"`
	DBConfig              DBConfig              `yaml:"db_config"`
	RedisConfig           RedisConfig           `yaml:"redis_config"`
	CryptoConfig          CryptoConfig          `yaml:"crypto_config"`
	LiveTokenConfig       LiveTokenConfig       `yaml:"live_token_config"`
	Assessment            AssessmentConfig      `yaml:"assessment_config"`
	AMS                   AMSConfig             `json:"ams" yaml:"ams"`
	H5P                   H5PServiceConfig      `json:"h5p" yaml:"h5p"`
	DataService           DataServiceConfig     `json:"data_service" yaml:"data_service"`
	KidsLoopRegion        string                `json:"kidsloop_region" yaml:"kidsloop_region"`
	TencentConfig         TencentConfig         `json:"tencent" yaml:"tencent"`
	KidsloopCNLoginConfig KidsloopCNLoginConfig `json:"kidsloop_cn" yaml:"kidsloop_cn"`
	CORS                  CORSConfig            `json:"cors" yaml:"cors"`
	ShowInternalErrorType bool                  `json:"show_internal_error_type"`
	User                  UserConfig            `json:"user" yaml:"user"`
	Report                ReportConfig          `json:"report" yaml:"report"`
	NewRelic              NewRelicConfig        `json:"new_relic" yaml:"new_relic"`
}

var config = &Config{}

type CORSConfig struct {
	AllowOrigins      []string `json:"allow_origins"`
	AllowFileProtocol bool     `json:"allow_file_protocol"`
}

type CryptoConfig struct {
	PrivateKeyPath string `yaml:"h5p_private_key_path"`
}

type RedisConfig struct {
	OpenCache bool `yaml:"open_cache"`
	// DEPRECATED
	// will remove in future, please use Option instead
	Host string `yaml:"host"`
	// DEPRECATED
	// will remove in future, please use Option instead
	Port int `yaml:"port"`
	// DEPRECATED
	// will remove in future, please use Option instead
	Password string         `yaml:"password" json:"-"`
	Option   *redis.Options `yaml:"option" json:"-"`
}

type DBConfig struct {
	DBMode string `yaml:"db_mode"`

	ConnectionString string        `yaml:"connection_string" json:"-"`
	MaxOpenConns     int           `yaml:"max_open_conns"`
	MaxIdleConns     int           `yaml:"max_idle_conns"`
	ConnMaxLifetime  time.Duration `yaml:"conn_max_life_time"`
	ConnMaxIdleTime  time.Duration `yaml:"conn_max_idle_time"`
	ShowLog          bool          `yaml:"show_log"`
	ShowSQL          bool          `yaml:"show_sql"`
	SlowThreshold    time.Duration `yaml:"slow_threshold"`

	DynamoEndPoint string `yaml:"dynamo_end_point"`
	DynamoRegion   string `yaml:"dynamo_region"`
}

type StorageConfig struct {
	Accelerate      bool   `yaml:"accelerate"`
	StorageProtocol string `yaml:"storage_protocol"`

	StorageEndPoint string `yaml:"storage_end_point"`
	StorageBucket   string `yaml:"storage_bucket"`
	StorageRegion   string `yaml:"storage_region"`

	StorageDownloadMode StorageDownloadMode `yaml:"storage_download_mode"`
	StorageSigMode      bool                `yaml:"storage_sig_mode"`
}

type CDNConfig struct {
	CDNRestrictedViewer bool   `yaml:"cdn_enable_restricted_viewer"`
	CDNPath             string `yaml:"cdn_path"`
	CDNKeyId            string `yaml:"cdn_key_id"`
	CDNPrivateKeyPath   string `yaml:"cdn_private_key_path"`
}

type ScheduleConfig struct {
	ReviewTypeEnabled bool          `json:"review_type_enabled" yaml:"review_type_enabled"`
	MaxRepeatYear     int           `json:"max_repeat_year" yaml:"max_repeat_year"`
	CacheExpiration   time.Duration `yaml:"cache_expiration"`
	ClassEventSecret  interface{}   `json:"class_event_secret"`
}

type LiveTokenConfig struct {
	PrivateKey interface{} `yaml:"private_key" json:"-"`
	//PublicKey  string      `yaml:"public_key"`
	AssetsUrlPrefix string `yaml:"assets_url_prefix"`
}

type AssessmentConfig struct {
	CacheExpiration      time.Duration `json:"cache_expiration" yaml:"cache_expiration"`
	AddAssessmentSecret  interface{}   `json:"-"`
	DefaultRemainingTime time.Duration `json:"default_remaining_time" yaml:"default_remaining_time"`
}

type AMSConfig struct {
	EndPoint              string      `json:"endpoint" yaml:"endpoint"`
	TokenVerifyKey        interface{} `json:"-" yaml:"token_verify_key"`
	AuthorizedKey         string      `json:"authorized_key"`
	ReplaceWithConnection bool        `json:"replace_with_connection"`
}

type DataServiceConfig struct {
	EndPoint            string `json:"endpoint" yaml:"endpoint"`
	AuthorizedKey       string `json:"authorized_key"`
	PublicAuthorizedKey string `json:"public_authorized_key"`
}

type H5PServiceConfig struct {
	EndPoint string `json:"endpoint" yaml:"endpoint"`
}

type KidsloopCNLoginConfig struct {
	Open         string      `json:"open" yaml:"open"`
	PrivateKey   interface{} `json:"-" yaml:"private_key"`
	PublicKey    interface{} `json:"-" yaml:"public_key"`
	CookieDomain string      `json:"cookie_domain" yaml:"cookie_domain"`
	InviteNotify string      `json:"invite_notify" yaml:"invite_notify"`
}

type TencentConfig struct {
	Sms TencentSmsConfig `json:"sms" yaml:"sms"`
}

type TencentSmsConfig struct {
	SDKAppID         string `json:"sdk_app_id" yaml:"sdk_app_id"`
	SecretID         string `json:"secret_id" yaml:"secret_id"`
	SecretKey        string `json:"-" yaml:"secret_key"`
	EndPoint         string `json:"endpoint" yaml:"endpoint"`
	Sign             string `json:"sign" yaml:"sign"`
	TemplateID       string `json:"template_id" yaml:"template_id"`
	TemplateParamSet string `json:"template_param_set" yaml:"template_param_set"`
	MobilePrefix     string `json:"mobile_prefix" yaml:"mobile_prefix"`
	OTPPeriod        string `json:"otp_period" yaml:"otp_period"`
}

type UserConfig struct {
	CacheExpiration           time.Duration `json:"cache_expiration" yaml:"cache_expiration"`
	PermissionCacheExpiration time.Duration `json:"permission_cache_expiration" yaml:"permission_cache_expiration"`
}

type ReportConfig struct {
	PublicKey string `json:"report_public_key" yaml:"report_public_key"`
}

type NewRelicConfig struct {
	NewRelicDistributedTracingEnabled bool
	NewRelicAppName                   string
	NewRelicLabels                    map[string]string
	NewRelicLicenseKey                string
}

func (c NewRelicConfig) Enable() bool {
	return strings.TrimSpace(c.NewRelicAppName) != "" &&
		strings.TrimSpace(c.NewRelicLicenseKey) != ""
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
	LoadDBEnvConfig(ctx)
	LoadRedisEnvConfig(ctx)
	loadScheduleEnvConfig(ctx)
	loadCryptoEnvConfig(ctx)
	loadLiveTokenEnvConfig(ctx)
	LoadAMSConfig(ctx)
	LoadH5PServiceConfig(ctx)
	loadDataServiceConfig(ctx)
	loadTencentConfig(ctx)
	loadKidsloopCNLoginConfig(ctx)
	loadAssessmentConfig(ctx)
	loadCORSConfig(ctx)
	loadShowInternalErrorTypeConfig(ctx)
	loadUserConfig(ctx)
	loadReportConfig(ctx)
	loadNewRelicConfig(ctx)
}

func loadShowInternalErrorTypeConfig(ctx context.Context) {
	var err error
	s := os.Getenv("show_internal_error_type")
	if s == "" {
		return
	}
	config.ShowInternalErrorType, err = strconv.ParseBool(s)
	if err != nil {
		log.Panic(ctx,
			"loadShowInternalErrorTypeConfig:load show_internal_error_type failed",
			log.Err(err),
			log.String("env_key", "show_internal_error_type"),
			log.Any("show_internal_error_type", s),
		)
	}
}
func loadKidsloopCNLoginConfig(ctx context.Context) {
	config.KidsLoopRegion = os.Getenv("kidsloop_region")
	if config.KidsLoopRegion != constant.KidsloopCN {
		return
	}
	privateKeyPath := os.Getenv("kidsloop_cn_login_private_key_path")
	content, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		log.Panic(ctx, "loadKidsloopCNLoginConfig:load auth config error", log.Err(err), log.String("privateKeyPath", privateKeyPath))
	}
	prv, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(content))
	if err != nil {
		log.Panic(ctx, "loadKidsloopCNLoginConfig:ParseRSAPrivateKeyFromPEM failed", log.Err(err))
	}
	config.KidsloopCNLoginConfig.PrivateKey = prv

	publicKeyPath := os.Getenv("kidsloop_cn_login_public_key_path")
	content, err = ioutil.ReadFile(publicKeyPath)
	if err != nil {
		log.Panic(ctx, "loadKidsloopCNLoginConfig:ReadFile failed", log.Err(err), log.String("publicKeyPath", publicKeyPath))
	}
	pub, err := jwt.ParseRSAPublicKeyFromPEM(content)
	if err != nil {
		log.Panic(ctx, "loadKidsloopCNLoginConfig:ParseRSAPublicKeyFromPEM failed", log.Err(err))
	}
	config.KidsloopCNLoginConfig.PublicKey = pub

	config.KidsloopCNLoginConfig.CookieDomain = os.Getenv("kidsloop_cn_login_cookie_domain")

	config.KidsloopCNLoginConfig.InviteNotify = os.Getenv("kidsloop_cn_invite_notify")
}

func loadTencentConfig(ctx context.Context) {
	config.KidsLoopRegion = os.Getenv("kidsloop_region")
	if config.KidsLoopRegion != constant.KidsloopCN {
		return
	}
	config.TencentConfig.Sms.SDKAppID = assertGetEnv("tc_sms_sdk_app_id")
	config.TencentConfig.Sms.SecretID = assertGetEnv("tc_sms_secret_id")
	config.TencentConfig.Sms.SecretKey = assertGetEnv("tc_sms_secret_key")
	config.TencentConfig.Sms.EndPoint = assertGetEnv("tc_sms_endpoint")
	config.TencentConfig.Sms.Sign = assertGetEnv("tc_sms_sign")
	config.TencentConfig.Sms.TemplateID = assertGetEnv("tc_sms_template_id")
	config.TencentConfig.Sms.TemplateParamSet = assertGetEnv("tc_sms_template_param_set")
	config.TencentConfig.Sms.MobilePrefix = assertGetEnv("tc_sms_mobile_prefix")
	config.TencentConfig.Sms.OTPPeriod = os.Getenv("OTP_PERIOD")
}

func loadCryptoEnvConfig(ctx context.Context) {
	config.CryptoConfig.PrivateKeyPath = os.Getenv("h5p_private_key_path")
}

func loadStorageEnvConfig(ctx context.Context) {
	config.StorageConfig.StorageProtocol = assertGetEnv("storage_protocol")
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

func LoadRedisEnvConfig(ctx context.Context) {
	openCache, _ := strconv.ParseBool(os.Getenv("open_cache"))
	config.RedisConfig.OpenCache = openCache
	redisURL := os.Getenv("redis_url")
	if redisURL != "" {
		// new config
		option, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Panic(ctx, "redis url invalid", log.Err(err), log.String("url", redisURL))
		}

		log.Debug(ctx, "redis url parsed", log.Any("option", option))
		config.RedisConfig.Option = option

		lastIndex := strings.LastIndex(option.Addr, ":")
		if lastIndex >= 0 {
			config.RedisConfig.Host = option.Addr[:lastIndex]
			port, err := strconv.Atoi(option.Addr[lastIndex+1:])
			if err != nil {
				config.RedisConfig.Port = 6379
			} else {
				config.RedisConfig.Port = port
			}
			config.RedisConfig.Password = option.Password
		}

		return
	}

	// old config
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

	config.RedisConfig.Option = &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
	}
}

func loadScheduleEnvConfig(ctx context.Context) {
	reviewTypeEnabledStr := os.Getenv("schedule_review_type_enabled")
	reviewTypeEnabled, err := strconv.ParseBool(reviewTypeEnabledStr)
	if err != nil {
		log.Warn(ctx, "parse env schedule_review_type_enabled failed",
			log.Err(err),
			log.String("schedule_review_type_enabled", reviewTypeEnabledStr))
	}
	config.Schedule.ReviewTypeEnabled = reviewTypeEnabled

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

	//publicKeyPath := os.Getenv("ams_class_event_jwt_public_key_path")
	//content, err := ioutil.ReadFile(publicKeyPath)
	//if err != nil {
	//	log.Panic(ctx, "load schedule config: load public key failed", log.Err(err), log.String("public_key_path", publicKeyPath))
	//}
	//key, err := jwt.ParseRSAPublicKeyFromPEM(content)
	//if err != nil {
	//	log.Panic(ctx, "load schedule config: ParseRSAPublicKeyFromPEM failed", log.Err(err))
	//}
	//config.Schedule.ClassEventSecret = key
}

func LoadDBEnvConfig(ctx context.Context) {
	config.DBConfig.ConnectionString = assertGetEnv("connection_string")
	maxOpenConnsStr := os.Getenv("max_open_conns")
	maxIdleConnsStr := os.Getenv("max_idle_conns")
	connMaxLifetimeStr := os.Getenv("conn_max_life_time")
	connMaxIdleTimeStr := os.Getenv("conn_max_idle_time")
	slowThresholdStr := os.Getenv("slow_threshold")

	maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
	if err != nil {
		log.Error(ctx, "Can't parse max_open_conns", log.Err(err))
		maxOpenConns = 16
	}
	config.DBConfig.MaxOpenConns = maxOpenConns

	maxIdleConns, err := strconv.Atoi(maxIdleConnsStr)
	if err != nil {
		log.Error(ctx, "Can't parse max_idle_conns", log.Err(err))
		maxIdleConns = 16
	}
	config.DBConfig.MaxIdleConns = maxIdleConns

	connMaxLifetime, err := time.ParseDuration(connMaxLifetimeStr)
	if err != nil {
		log.Warn(ctx, "invalid conn_max_life_time",
			log.String("conn_max_life_time", connMaxLifetimeStr),
			log.Err(err))
		config.DBConfig.ConnMaxLifetime = constant.DBDefaultConnMaxLifetime
	} else {
		config.DBConfig.ConnMaxLifetime = connMaxLifetime
	}

	connMaxIdleTime, err := time.ParseDuration(connMaxIdleTimeStr)
	if err != nil {
		log.Warn(ctx, "invalid conn_max_idle_time",
			log.String("conn_max_idle_time", connMaxIdleTimeStr),
			log.Err(err))
	} else {
		config.DBConfig.ConnMaxIdleTime = connMaxIdleTime
	}

	slowThreshold, err := time.ParseDuration(slowThresholdStr)
	if err != nil {
		log.Warn(ctx, "invalid slow_threshold",
			log.String("slow_threshold", slowThresholdStr),
			log.Err(err))
		config.DBConfig.SlowThreshold = constant.DBDefaultSlowThreshold
	} else {
		config.DBConfig.SlowThreshold = slowThreshold
	}
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

	assetsUrlPrefix := os.Getenv("live_assets_url_prefix")
	_, err = url.Parse(assetsUrlPrefix)
	if err != nil {
		log.Panic(ctx, "load LiveTokenEnvConfig:load live_assets_url_prefix config error",
			log.Err(err),
			log.String("assetsUrlPrefix", assetsUrlPrefix),
		)
	}
	config.LiveTokenConfig.AssetsUrlPrefix = assetsUrlPrefix
}

func loadAssessmentConfig(ctx context.Context) {
	cacheExpiration, err := time.ParseDuration(os.Getenv("assessment_cache_expiration"))
	if err != nil {
		config.Assessment.CacheExpiration = constant.ScheduleDefaultCacheExpiration
	} else {
		config.Assessment.CacheExpiration = cacheExpiration
	}

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

	defaultRemainingTime, err := time.ParseDuration(os.Getenv("assessment_default_remaining_time"))
	if err != nil {
		log.Debug(ctx, "set assessment_default_remaining_time failed",
			log.Err(err),
		)
		config.Assessment.DefaultRemainingTime = constant.AssessmentDefaultRemainingTime
	} else {
		config.Assessment.DefaultRemainingTime = defaultRemainingTime
	}
}

func LoadAMSConfig(ctx context.Context) {
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
	config.AMS.AuthorizedKey = os.Getenv("user_service_api_key")
	config.AMS.ReplaceWithConnection = !(strings.ToLower(os.Getenv("no_use_connection")) == "true")
}

func LoadH5PServiceConfig(ctx context.Context) {
	config.H5P.EndPoint = constant.H5PServiceDefaultEndpoint
	h5pEndpoint := os.Getenv("h5p_endpoint")
	if h5pEndpoint != "" {
		config.H5P.EndPoint = h5pEndpoint
	}
}

func loadDataServiceConfig(ctx context.Context) {
	// TODO assertGetEnv
	// config.DataService.EndPoint = os.Getenv("data_service_endpoint")
	// config.DataService.AuthorizedKey = os.Getenv("data_service_api_key")
	// config.DataService.PublicAuthorizedKey = os.Getenv("data_service_public_key")
	config.DataService.EndPoint = "https://dev-global-adaptive-review-api.data.kidsloop.net"
	config.DataService.AuthorizedKey = "uM72VB8WJl85tw66Ps4ri5uZJaBvxzsmF5sa0yg5"
}

func loadCORSConfig(ctx context.Context) {
	config.CORS.AllowOrigins = strings.Split(os.Getenv("cors_domain_list"), ",")
	config.CORS.AllowFileProtocol, _ = strconv.ParseBool(os.Getenv("cors_allow_file_protocol"))
}

func loadUserConfig(ctx context.Context) {
	cacheExpiration, err := time.ParseDuration(os.Getenv("user_cache_expiration"))
	if err != nil {
		config.User.CacheExpiration = constant.UserDefaultCacheExpiration
	} else {
		config.User.CacheExpiration = cacheExpiration
	}

	permissionCacheExpiration, err := time.ParseDuration(os.Getenv("user_permission_cache_expiration"))
	if err != nil {
		config.User.PermissionCacheExpiration = constant.UserPermissionDefaultCacheExpiration
	} else {
		config.User.PermissionCacheExpiration = permissionCacheExpiration
	}
}

func loadReportConfig(ctx context.Context) {
	// config.Report.PublicKey = assertGetEnv("student_usage_report_public_key")
}

func loadNewRelicConfig(ctx context.Context) {
	// all newRelic config are optional
	nr := &config.NewRelic
	nr.NewRelicAppName = os.Getenv("NEW_RELIC_APP_NAME")
	nr.NewRelicLicenseKey = os.Getenv("NEW_RELIC_LICENSE_KEY")
	nr.NewRelicLabels = make(map[string]string)
	nr.NewRelicDistributedTracingEnabled = strings.ToLower(
		os.Getenv("NEW_RELIC_DISTRIBUTED_TRACING_ENABLED")) == "true"

	// rawLabels format: "key1:val1;key2:val2;key3:val3"
	rawLabels := os.Getenv("NEW_RELIC_LABELS")
	for _, label := range strings.Split(rawLabels, ";") {
		kv := strings.Split(strings.TrimSpace(label), ":")
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		nr.NewRelicLabels[key] = val
	}
}

func Get() *Config {
	return config
}

func Set(c *Config) {
	config = c
}
