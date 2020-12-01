package model

import (
	"context"
	"fmt"
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
				SDKAppID:         "1400392114",
				SecretID:         "AKIDGL9HBpFNo20BxSuq8BxTrb3RYP0UKKJ7",
				SecretKey:        "jALI5brFMGsDjyOYj9YbNTGxc4btsdpn",
				EndPoint:         "sms.tencentcloudapi.com",
				Sign:             "预元岛",
				TemplateID:       "650697",
				TemplateParamSet: "2",
				MobilePrefix:     "+86",
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
