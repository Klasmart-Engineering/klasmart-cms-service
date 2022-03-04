package external

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/go-redis/redis"
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
		UserID: "afdfc0d9-ada9-4e66-b225-20f956d1a399",
		OrgID:  "60c064cc-bbd8-4724-b3f6-b886dce4774f", // Badanamu HQ
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0NjM3MTg1OSwiaXNzIjoia2lkc2xvb3AifQ.CyMOeEWfx0bwYJUre2EAnNoCte85mqi2gvKPJt2y2QimnOMt3Zg1uvS0iXGXmb_a5tdjywjDfNZXDmWS01ZgZLecfD9aDq45Y07gLJl9NDmVvoxnjNKrH4ley3l3bPhDRlbKCwgJn37fqzjiTSt0zbNKvBdpyYqjuT4_qlAAFRzmgIeg1vpSZCSVUyEzosdlgg8wLwaAmtUl0KEizj5uo2hJGN9aOeAe4ubnW-6fixgEkZFZNU1sEksFY6f7VRpxPW5Dr-p1PHRpkHmgSfIxSeDuC1rGIxYWcZsiGsBhU1umjCPjvNLMOlKUwgSO4G_NDfTR82XU7QlfmfJoD5FE2g",
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
