package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-cn/common"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/kl2cache"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	if os.Getenv("env") == "HTTP" {
		common.Setenv(common.EnvHTTP)
	} else {
		common.Setenv(common.EnvLAMBDA)
	}

	config.LoadDBEnvConfig(ctx)
	da.InitMySQL(ctx)

	config.LoadRedisEnvConfig(ctx)
	da.InitRedis(ctx)

	initCache(ctx)
	initDataSource(ctx)
	initAms(ctx)
	server = NewServer()
	os.Exit(m.Run())
}

func initAms(ctx context.Context) {
	config.Get().AMS.EndPoint = os.Getenv("ams_endpoint")
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

func initDataSource(ctx context.Context) {
	//init querier
	engine := cache.GetCacheEngine()
	engine.SetExpire(ctx, constant.MaxCacheExpire)
	engine.OpenCache(ctx, config.Get().RedisConfig.OpenCache)
	cache.GetPassiveCacheRefresher().SetUpdateFrequency(constant.MaxCacheExpire, constant.MinCacheExpire)

	engine.AddDataSource(ctx, external.GetUserServiceProvider())
	engine.AddDataSource(ctx, external.GetTeacherServiceProvider())
	engine.AddDataSource(ctx, external.GetSubjectServiceProvider())
	engine.AddDataSource(ctx, external.GetSubCategoryServiceProvider())
	engine.AddDataSource(ctx, external.GetStudentServiceProvider())
	engine.AddDataSource(ctx, external.GetSchoolServiceProvider())
	engine.AddDataSource(ctx, external.GetProgramServiceProvider())
	engine.AddDataSource(ctx, external.GetOrganizationServiceProvider())
	engine.AddDataSource(ctx, external.GetGradeServiceProvider())
	engine.AddDataSource(ctx, external.GetClassServiceProvider())
	engine.AddDataSource(ctx, external.GetCategoryServiceProvider())
	engine.AddDataSource(ctx, external.GetAgeServiceProvider())
}
func DoHttp(method string, url string, body string) string {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, bytes.NewBufferString(body))
	req.Header.Add("Authorization", "")
	server.ServeHTTP(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		return fmt.Sprintf("StatusCode: %d", res.StatusCode)
	}
	data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		panic(err)
	}
	return string(data)
}
func DoHttpWithOperator(method string, op *entity.Operator, url string, body string) string {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, bytes.NewBufferString(body))
	req.Header.Add("Authorization", "")
	req.Header.Add("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "access", Value: op.Token})
	server.ServeHTTP(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		return fmt.Sprintf("StatusCode: %d", res.StatusCode)
	}
	data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		panic(err)
	}
	return string(data)
}
func initOperator(orgID string, userID string, authTo string, authCode string) *entity.Operator {
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
	//req.Header.Set("origin", "https://auth.kidsloop.net")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authority", "auth.kidsloop.net")
	req.Header.Set("accept", "application/json")
	//req.Header.Set("path", "/transfer")
	//req.Header.Set("scheme", "https")
	req.Header.Set("method", "POST")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36")
	req.Header.Set("referer", "https://auth.kidsloop.net/?continue=https%3A%2F%2Fbeta-hub.kidsloop.net%2F%23%2Fadmin%2Forganizations")
	req.Header.Set("cookie", "_ga=GA1.2.489381037.1617355818; locale=en; privacy=true")
	resp, err = (&http.Client{}).Do(req)
	op := &entity.Operator{
		OrgID: orgID,
	}
	if err != nil {
		op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6Ijc0MmIxNjI1LWVmYjctNWM1YS1iZDI3LWRjYzIwMmQ2YTEzNiIsImVtYWlsIjoib3JnMTIyMTAyQHlvcG1haWwuY29tIiwiZXhwIjoxNjE4OTA5OTE1LCJpc3MiOiJraWRzbG9vcCJ9.IRcPjpqH3AQfL_0i3rEPKXXLHbcjGvegv4iVseqSENzhr7X_iCckP2gLth4plN_mX-dNphQqJvV0-L5enTn1u8g3jbmXpR5VALV5Bf_5G-A6xZWUAwtxCyxVKlTqtOM5Pi-WEg8gPHgS9sGL2vT7eviOlcG3S3W0LV5QzYBBC55okNtHZLwS0N-eXzVT8oKOwyMTU8ftqTQ5f9slCUV7ennZrJ6FJX8oozlHixIg4NcTMpo_S0al4GTw2--BJU_DrEQZ80dgtBse1TE8QxY0_R8tbW6SNUJkSKkOZqVCUAcmUG_sY5rN5HGFzeuniNJpe179xdF8OUXCiH-9YmIBaw"
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
	for i := range resp.Cookies() {
		cookie := resp.Cookies()[i]
		if cookie.Name == "access" {
			switchData := struct {
				UserID string `json:"user_id"`
			}{UserID: userID}

			switchBody, err := json.Marshal(switchData)
			if err != nil {
				panic(err)
			}

			req, err = http.NewRequest("POST", transferUrl+"/switch", bytes.NewBuffer(switchBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("authority", "auth.kidsloop.net")
			req.Header.Set("accept", "application/json")
			req.Header.Set("method", "POST")
			req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36")
			req.Header.Set("referer", "https://auth.alpha.kidsloop.net/selectprofile")
			if err != nil {
				panic(err)
			}
			req.Header.Set("cookie", fmt.Sprintf("_ga=GA1.2.1518569125.1621408323; _gid=GA1.2.1255454809.1638756307; locale=zh-CN; privacy=true; access=%s", cookie.Value))
			resp1, err := (&http.Client{}).Do(req)
			if err != nil {
				panic(err)
			}
			for j := range resp1.Cookies() {
				cookie1 := resp1.Cookies()[j]
				if cookie1.Name == "access" {
					op.Token = cookie1.Value
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

		}
	}
	return nil
}
