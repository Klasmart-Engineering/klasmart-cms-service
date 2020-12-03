package model

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/ro"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
)

func Setup() {
	config.Set(&config.Config{
		RedisConfig: config.RedisConfig{
			OpenCache: false,
			Host:      "172.28.145.52",
			Port:      16379,
			Password:  "",
		},
		TencentConfig: config.TencentConfig{
			Sms: config.TencentSmsConfig{
				SDKAppID:         os.Getenv("tc_sms_sdk_app_id"),
				SecretID:         os.Getenv("tc_sms_secret_id"),
				SecretKey:        os.Getenv("tc_sms_secret_key"),
				EndPoint:         os.Getenv("tc_sms_endpoint"),
				Sign:             os.Getenv("tc_sms_sign"),
				TemplateID:       os.Getenv("tc_sms_template_id"),
				TemplateParamSet: os.Getenv("tc_sms_template_param_set"),
				MobilePrefix:     os.Getenv("tc_sms_mobile_prefix"),
			},
		},
	})
	ro.SetConfig(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
		Password: config.Get().RedisConfig.Password,
	})
}

func TestBubble_Launch(t *testing.T) {
	Setup()
	code, err := GetBubbleMachine("").Launch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(code)
}

func TestVerifyCode(t *testing.T) {
	Setup()
	pass, err := VerifyCode(context.Background(), "", "746367")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(pass)
}

func TestMakeSecretAndSalt(t *testing.T) {
	salt, secret := MakeSecretAndSalt(context.Background(), "12345")
	fmt.Println(salt, secret)
}

var st = "ul4XnTedkSW0zSfumb4IwooIF63T89+l"
var secret = "mLRw1yMmOYs+f7FaQ//MgzVMonaMnrDgyv0ePFTaw4X4ZEAEUOmOKmgdVsQexZIqT9xO5Zq9ylHbcc0yWvBPnA=="

func TestVerifySecretWithSalt(t *testing.T) {
	pass := VerifySecretWithSalt(context.Background(), "12345", secret, st)
	fmt.Println(pass)
}
