package external

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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

func TestMain(m *testing.M) {
	config.Set(&config.Config{
		AMS: config.AMSConfig{
			EndPoint: "https://api.alpha.kidsloop.net/user/",
		},
		H5P: config.H5PServiceConfig{
			EndPoint: "https://api.alpha.kidsloop.net/assessment/graphql/",
		},
	})
	testOperator = &entity.Operator{
		UserID: "14494c07-0d4f-5141-9db2-15799993f448",
		OrgID:  "a44da070-1907-46c4-bc4c-f26ced889439", // Badanamu HQ
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjdiYmQzYTc4LTJkM2QtNGM1MC1hZWYyLTViY2M2Njc1NmRmMSIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYzMTg0Nzc3OCwiaXNzIjoia2lkc2xvb3AifQ.fiJMWtJW8PuoC2NynMPXoBtGx0c-P8gzQiS3_ZQpofkb2OLIIxbBbnvT2WBBDILHshs1RHYRm0nimBLHUU4iu2U4ZAaOdDYtnna4ONeby3WKh8VWWlD_aCZZzlKeGW1qqimrSrv7d0m-lZmT9GGovuk0hwlGyFLwsGl2scsxyJ3sS7amXZWjGKKOZyrMyPoD42X6OWPtjUtehiId5mzzUvruF2SbbZ1DW0TCNbZ0ifEDnd11aCl1GEeBm5RcEyAxKSYoqTESuF28PoWdK-mSxvaJXPU7rFdGmqWeJLgTRh6BgOYfxO7_V4VcHfnZaySuEhZ0C_BTHflfSaVUfJ98pg",
	}
	ro.SetConfig(&redis.Options{Addr: "127.0.0.1:6379"})
	initQuerier(context.Background())

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

func TestRegexp(t *testing.T) {
	r, _ := regexp.Compile("(\\S+=\\S+;)*access=\\S+(;\\S+=\\S+)*")
	t.Log(r.MatchString("abc=;access=123"))
}
