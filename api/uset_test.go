package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/dgrijalva/jwt-go"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
)

func userSetup() {
	privateKeyPath := os.Getenv("kidsloop_cn_login_private_key_path")
	content, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		panic(err)
	}
	prv, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(content))
	if err != nil {
		panic(err)
	}

	publicKeyPath := os.Getenv("kidsloop_cn_login_public_key_path")
	content, err = ioutil.ReadFile(publicKeyPath)
	if err != nil {
		panic(err)
	}
	pub, err := jwt.ParseRSAPublicKeyFromPEM(content)
	if err != nil {
		panic(err)
	}

	config.Set(&config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: os.Getenv("connection_string"),
		},
		KidsLoopRegion: os.Getenv("kidsloop_region"),
		KidsloopCNLoginConfig: config.KidsloopCNLoginConfig{
			PrivateKey: prv,
			PublicKey:  pub,
		},
		RedisConfig: config.RedisConfig{
			OpenCache: true,
			Host:      os.Getenv("redis_host"),
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
		AMS: config.AMSConfig{
			EndPoint: os.Getenv("ams_endpoint"),
		},
	})
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		dbConf := config.Get().DBConfig
		c.ShowLog = dbConf.ShowLog
		c.ShowSQL = dbConf.ShowSQL
		c.MaxIdleConns = dbConf.MaxIdleConns
		c.MaxOpenConns = dbConf.MaxOpenConns
		c.ConnectionString = dbConf.ConnectionString
	})
	if err != nil {
		log.Error(context.TODO(), "create dbo failed", log.Err(err))
		panic(err)
	}
	dbo.ReplaceGlobal(dboHandler)
	initCache()
}

func TestUserLogin(t *testing.T) {
	req := LoginRequest{
		AuthTo:   "15221776376",
		AuthCode: "Bada1234",
		AuthType: constant.LoginByPassword,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttp(http.MethodPost, prefix+"/users/login", string(data))
	fmt.Println(res)
}

func TestUserRegister(t *testing.T) {
	// 15026743257
	// 15221776376
	req := RegisterRequest{
		Account:  "15221776376",
		AuthCode: "800036",
		Password: "Bada1234",
		ActType:  constant.AccountPhone,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	res := DoHttp(http.MethodPost, prefix+"/users/register", string(data))
	fmt.Println(res)
}

func TestUserSendCode(t *testing.T) {
	// 15026743257
	// 15221776376
	req := SendCodeRequest{
		Mobile: "15221776376",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	res := DoHttp(http.MethodPost, prefix+"/users/send_code", string(data))
	fmt.Println(res)
}

func TestUserForgetPassword(t *testing.T) {

}

func TestUserRestPassword(t *testing.T) {

}
