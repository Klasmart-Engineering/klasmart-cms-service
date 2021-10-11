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
			//EndPoint: "https://api.beta.kidsloop.net/user/",
			EndPoint: "https://api.alpha.kidsloop.net/user/",
		},
		H5P: config.H5PServiceConfig{
			EndPoint: "https://api.alpha.kidsloop.net/assessment/graphql/",
		},
	})
	testOperator = &entity.Operator{
		UserID: "130ca9b7-32e3-4436-ab8b-8d492a033633",
		OrgID:  "92db7ddd-1f23-4f64-bd47-94f6d34a50c0", // Badanamu HQ
		//Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjI1MjJlYWUwLTVmNzItNDVkMS05OGY2LTM1ODI3YWI4MTZhNyIsImVtYWlsIjoib3Jna2lkc2xvb3AxQHlvcG1haWwuY29tIiwiZXhwIjoxNjIyNjIxNTU2LCJpc3MiOiJraWRzbG9vcCJ9.h70Xq63TvRLJxZZ8-5sxxPTChjRwzNcG9KJ5rnRg66iDZ1HLfouI4W1ED6eJ6lzJqJdIdkL3nKGdXM-ePtXfctd8vGBmm2TBuk4C2fq_zX9R5N7MzHVS1wzbRzD3D-U_tLLY-_JmEV2ECgoajoFpHP2DkUuVA-qu1AubdFDLc9VS9ETbdNtEFQw0eF3x4eEiuT7WVKx3VX-FQLw0IJF0CMlNYSoWAWPxIqG7M5VwzTX1OPwG2eZXpwQ7RozPQWE7Sft1dQt9TvHPkj40xC7UgNB637wTB5-MUxJRNNIrNfcsJHOtro_ZT8JR7GnQWAf24Xmt2oOkFDDT68KVJ0yH2g",
		Token: "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjEzMGNhOWI3LTMyZTMtNDQzNi1hYjhiLThkNDkyYTAzMzYzMyIsImVtYWlsIjoibW95YW5AeW9wbWFpbC5jb20iLCJleHAiOjE2MzI3MTM3MDIsImlzcyI6ImtpZHNsb29wIn0.AUSifGVRSxns41Ok7hIG15iouO7bHAPaiWIXUHz0KeA_E5UwvGlb229eNOgnvkBbQxqShoGNDQoZxRVt5k-L3_OHoiIpS9hwnp8Hx5S6HXMe1YRg8BxqSjl2YGtZhjF6nzcRO477qqalBAPsdKQvXCcyTIGMGyMomp5EyercdrRXCXJ2wg-YFysJ_vuInVmYL1leaf3B80D8qwS6QuHn_jJ5OEOPnzxNn5mlgqtzMyIbB7YeYxalfhWX-Pn2r1YF-n2D4XIcKNsMNbQUlyqxVoTldVUoxMdujyERX6Qi4kxNql3IXxnPGg42zlACFYtC5wr5kTNIiDw9LOSJk2ZB8g",
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
