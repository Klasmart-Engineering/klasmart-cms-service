package external

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/kl2cache"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
)

var (
	testOperator *entity.Operator
)

func initOperator(orgID string, authTo string, authCode string) *entity.Operator {
	if authTo == "" {
		authTo = os.Getenv("auth_to")
	}
	if authCode == "" {
		authCode = os.Getenv("auth_code")
	}
	loginUrl := os.Getenv("ams_auth_endpoint")
	transferUrl := os.Getenv("auth_endpoint")
	data := struct {
		DeviceID   string `json:"deviceId"`
		DeviceName string `json:"deviceName"`
		Email      string `json:"email"`
		Password   string `json:"pw"`
	}{
		DeviceID:   "webpage",
		DeviceName: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36",
		Email:      authTo,
		Password:   authCode,
	}
	body, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", loginUrl+"/v1/login", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	loginResponse := struct {
		AccessToken string `json:"accessToken"`
	}{}
	err = json.Unmarshal(response, &loginResponse)
	if err != nil {
		panic(err)
	}

	data2 := struct {
		Token string `json:"token"`
	}{loginResponse.AccessToken}
	body, err = json.Marshal(data2)
	if err != nil {
		panic(err)
	}
	req, err = http.NewRequest("POST", transferUrl+"/transfer", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		panic(err)
	}
	op := &entity.Operator{
		OrgID: orgID,
	}
	for i := range resp.Cookies() {
		cookie := resp.Cookies()[i]
		if cookie.Name == "access" {
			op.Token = cookie.Value
			infos := strings.Split(op.Token, ".")
			info, err := base64.RawStdEncoding.DecodeString(infos[1])
			if err != nil {
				panic(err)
			}
			var user struct {
				ID string `json:"id"`
			}
			json.Unmarshal(info, &user)
			op.UserID = user.ID
			return op
		}
	}
	return nil
}

var token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTA1NzI4NCwiaXNzIjoia2lkc2xvb3AifQ.CF2DCRV1nRH_39uM82kLb3NhtsC1wt5sbzMXQRxJrXUYctK3PLa4WjGPJzM-qwMs5Q20ZwdwDAa9vf7x1hglUuCstO8aPJiOjt-9E4fU7h2IvGy-uLY93hP5Cmb9rlJ0BcM0aQfCJteSbs7TOsAnnInDRmQTaS-mD93whwDH_g2Om5hHMZ6dPOI7_IAPHbaV5q_ChLlGKiodYkN-n1JxyRWIubWZyxfFU9EwCgjQesTc7Pn-E6hSDdmkt4Y81ZmRrrnP40gCwa-pra1-H2zyv1iwld85mg4WIuC53D7p5Q_IpZKN_HCW4XGnzabIu2blLmETjxWeOTnVNelptTexDg"

func TestMain(m *testing.M) {
	ctx := context.Background()
	config.LoadAMSConfig(ctx)
	config.LoadH5PServiceConfig(ctx)
	config.LoadRedisEnvConfig(ctx)
	testOperator = &entity.Operator{
		UserID: "afdfc0d9-ada9-4e66-b225-20f956d1a399",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
		Token:  token,
	}
	ro.SetConfig(config.Get().RedisConfig.Option)
	initQuerier(ctx)
	initCache(ctx)
	os.Exit(m.Run())
}

func initQuerier(ctx context.Context) {
	engine := cache.GetCacheEngine()
	engine.AddDataSource(ctx, GetUserServiceProvider())
	engine.AddDataSource(ctx, GetTeacherServiceProvider())
	engine.AddDataSource(ctx, GetSubjectServiceProvider())
	engine.AddDataSource(ctx, GetSubCategoryServiceProvider())
	engine.AddDataSource(ctx, GetStudentServiceProvider())
	engine.AddDataSource(ctx, GetSchoolServiceProvider())
	engine.AddDataSource(ctx, GetProgramServiceProvider())
	engine.AddDataSource(ctx, GetOrganizationServiceProvider())
	engine.AddDataSource(ctx, GetGradeServiceProvider())
	engine.AddDataSource(ctx, GetClassServiceProvider())
	engine.AddDataSource(ctx, GetCategoryServiceProvider())
	engine.AddDataSource(ctx, GetAgeServiceProvider())
}

func initCache(ctx context.Context) {

	conf := config.Get()
	err := kl2cache.Init(ctx,
		kl2cache.OptEnable(conf.RedisConfig.OpenCache),
		kl2cache.OptRedis(conf.RedisConfig.Host, conf.RedisConfig.Port, conf.RedisConfig.Password),
		kl2cache.OptStrategyFixed(constant.MaxCacheExpire),
	)
	if err != nil {
		log.Panic(ctx, "kl2cache.Init failed", log.Err(err))
	}
}
func TestRegexp(t *testing.T) {
	r, _ := regexp.Compile("(\\S+=\\S+;)*access=\\S+(;\\S+=\\S+)*")
	t.Log(r.MatchString("abc=;access=123"))
}
