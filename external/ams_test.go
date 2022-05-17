package external

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cache/cache"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils/kl2cache"
	"github.com/KL-Engineering/ro"
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

var token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImQ2NTNmODgwLWI5MjAtNDM1Zi04ZjJkLTk4YjVkNDYyMWViOCIsImVtYWlsIjoic2Nob29sXzAzMDMyOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MjY5NTE4NywiaXNzIjoia2lkc2xvb3AifQ.nr_skolxXDOiO4qy7DuUC28qr7sKUibAuEPFhg4HcPuu-dsKqcADQiYis4WFstVjwELd8CMWQkxlpPLEfDwslEfDCokXn6BAqfR6vxOU3ht8izHuljmCSXNC2nW7PFYfTBIF2GH471ElJFiCbAu_eSR6OJ6JtK7y3pIl9Z9hkVwztehvF8eKYpgPhmsyEr-pzUe9LH_e4YTMxTn6M3lw_eDDmoFIH9g8knR4Sb6wh99RDxtq-ePplpiwYo1rBubBLRyLifPACa9EnOyw98f4nYa0igRx8mAeGwHB1FoL4rqJuTc8amb0FonojOUnwv92EOkJFKxqslHvnkrupbvw5w"

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
