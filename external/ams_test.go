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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYzMTY4NTU3OSwiaXNzIjoia2lkc2xvb3AifQ.RBjU31IYIZHKSW36qDpvaH2fNBPBUeGVvtN-kApFGXLPqebSru9PZvv9D52LQgQq_ocE9g3DxsdMv0WF2IHVVz0KFHUhpJc-QDuMhFt4sm_sPeWASMgk-SMhGR_7ucNqVV6F97EATumIhkMuqTCqNpuJBUbuxCnRssTud76MATANOC90EQGcr8vfmzEVuv1TVCUmuTCvejGGbU-V-MKj5JeB32WIQeWqYwUkmhmR8sqxPbhR-ZRXIiVSni1OJzF35rcEGIAuEPYtRvuRypSF96LFVpZpPjxMIyT6ZltYWIsUw0b78xrRuc9pATVrT1dOpK-xulqKeVd6205aKqkrgw",
	}
	ro.SetConfig(&redis.Options{Addr: "127.0.0.1:6379"})
	initQuerier(context.Background())

	os.Exit(m.Run())
}

func initQuerier(ctx context.Context) {
	engine := cache.GetCacheEngine()
	engine.AddQuerier(ctx, GetUserServiceProvider())
	engine.AddQuerier(ctx, GetTeacherServiceProvider())
	engine.AddQuerier(ctx, GetSubjectServiceProvider())
	engine.AddQuerier(ctx, GetSubCategoryServiceProvider())
	engine.AddQuerier(ctx, GetStudentServiceProvider())
	engine.AddQuerier(ctx, GetSchoolServiceProvider())
	engine.AddQuerier(ctx, GetProgramServiceProvider())
	engine.AddQuerier(ctx, GetOrganizationServiceProvider())
	engine.AddQuerier(ctx, GetGradeServiceProvider())
	engine.AddQuerier(ctx, GetClassServiceProvider())
	engine.AddQuerier(ctx, GetCategoryServiceProvider())
	engine.AddQuerier(ctx, GetAgeServiceProvider())
}

func TestRegexp(t *testing.T) {
	r, _ := regexp.Compile("(\\S+=\\S+;)*access=\\S+(;\\S+=\\S+)*")
	t.Log(r.MatchString("abc=;access=123"))
}
