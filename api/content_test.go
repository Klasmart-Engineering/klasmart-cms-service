package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/common-cn/common"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
)

var server *Server

func initDB() {
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
}
func initCache() {
	if config.Get().RedisConfig.OpenCache {
		ro.SetConfig(&redis.Options{
			Addr:     fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
			Password: config.Get().RedisConfig.Password,
		})
		initDataSource()
	}
}

func initDataSource() {
	//init querier
	ctx := context.Background()
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

func TestMain(m *testing.M) {
	config.Set(&config.Config{})
	if os.Getenv("env") == "HTTP" {
		common.Setenv(common.EnvHTTP)
	} else {
		common.Setenv(common.EnvLAMBDA)
	}

	log.Debug(context.TODO(), "init api server success")
	server = NewServer()
	code := m.Run()
	os.Exit(code)
}

const prefix = "/v1"

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

func TestApprove(t *testing.T) {
	res := DoHttp(http.MethodPut, prefix+"/contents_review/1/approve", "")
	fmt.Println(res)
}

func TestGetTimeLocation(t *testing.T) {
	//fmt.Println(time.LoadLocation("America/Los_Angeles"))
	loc := time.Local
	time.ParseInLocation("", "", loc)
	//fmt.Println(time.Local)
}

func TestPrint(t *testing.T) {
	t.Log("gitlab.badanamu.com.cn/calmisland/common-log/log.ZapLogger.Error\\n\\t/go/pkg/mod/gitlab.badanamu.com.cn/calmisland/common-log@v0.1.3/log/logger_zap.go:40\\ngitlab.badanamu.com.cn/calmisland/common-log/log.Error\\n\\t/go/pkg/mod/gitlab.badanamu.com.cn/calmisland/common-log@v0.1.3/log/logger.go:41\\ngitlab.badanamu.com.cn/calmisland/kidsloop2/api.ExtractSession\\n\\t/builds/calmisland/kidsloop2/api/middleware.go:17\\ngitlab.badanamu.com.cn/calmisland/kidsloop2/api.MustLogin\\n\\t/builds/calmisland/kidsloop2/api/middleware.go:34\\ngithub.com/gin-gonic/gin.(*Context).Next\\n\\t/go/pkg/mod/github.com/gin-gonic/gin@v1.6.3/context.go:161\\ngithub.com/gin-gonic/gin.(*Engine).handleHTTPRequest\\n\\t/go/pkg/mod/github.com/gin-gonic/gin@v1.6.3/gin.go:409\\ngithub.com/gin-gonic/gin.(*Engine).ServeHTTP\\n\\t/go/pkg/mod/github.com/gin-gonic/gin@v1.6.3/gin.go:367\\ngitlab.badanamu.com.cn/calmisland/kidsloop2/api.Server.ServeHTTP\\n\\t/builds/calmisland/kidsloop2/api/server.go:33\\ngitlab.badanamu.com.cn/calmisland/common-cn/common.(*decoratorHandler).ServeHTTP\\n\\t/go/pkg/mod/gitlab.badanamu.com.cn/calmisland/common-cn@v0.15.0/common/common.go:47\\ngitlab.badanamu.com.cn/calmisland/common-cn/common.RunWithHTTPHandler.func2\\n\\t/go/pkg/mod/gitlab.badanamu.com.cn/calmisland/common-cn@v0.15.0/common/common.go:66\\nreflect.Value.call\\n\\t/usr/local/go/src/reflect/value.go:460\\nreflect.Value.Call\\n\\t/usr/local/go/src/reflect/value.go:321\\ngithub.com/aws/aws-lambda-go/lambda.NewHandler.func1\\n\\t/go/pkg/mod/github.com/aws/aws-lambda-go@v1.18.0/lambda/handler.go:124\\ngithub.com/aws/aws-lambda-go/lambda.lambdaHandler.Invoke\\n\\t/go/pkg/mod/github.com/aws/aws-lambda-go@v1.18.0/lambda/handler.go:24\\ngithub.com/aws/aws-lambda-go/lambda.(*Function).Invoke\\n\\t/go/pkg/mod/github.com/aws/aws-lambda-go@v1.18.0/lambda/function.go:64\\nreflect.Value.call\\n\\t/usr/local/go/src/reflect/value.go:460\\nreflect.Value.Call\\n\\t/usr/local/go/src/reflect/value.go:321\\nnet/rpc.(*service).call\\n\\t/usr/local/go/src/net/rpc/server.go:377")
}

func TestQueryAuthContent(t *testing.T) {
	setupMilestone()
	op := initOperator("a44da070-1907-46c4-bc4c-f26ced889439", "8842b2ec-b903-46c6-b062-05920a3b7f79", "student112301@yopmail.com", "")
	url := "/v1/contents_authed?submenu=more+featured&program_group=More+Featured+Content&order_by=-update_at&page=1&page_size=20&org_id=a44da070-1907-46c4-bc4c-f26ced889439"
	op.OrgID = "73e33241-ecf2-40a6-a642-39b0e60fe820"
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

/*
SELECT `cms_contents`.`id`,`cms_contents`.`content_type`,`cms_contents`.`content_name`,`cms_contents`.`keywords`,`cms_contents`.`description`,`cms_contents`.`thumbnail`,`cms_contents`.`source_type`,`cms_contents`.`outcomes`,`cms_contents`.`data`,`cms_contents`.`extra`,`cms_contents`.`suggest_time`,`cms_contents`.`author`,`cms_contents`.`creator`,`cms_contents`.`org`,`cms_contents`.`self_study`,`cms_contents`.`draw_activity`,`cms_contents`.`lesson_type`,`cms_contents`.`publish_status`,`cms_contents`.`reject_reason`,`cms_contents`.`remark`,`cms_contents`.`version`,`cms_contents`.`locked_by`,`cms_contents`.`source_id`,`cms_contents`.`latest_id`,`cms_contents`.`copy_source_id`,`cms_contents`.`parent_folder`,`cms_contents`.`dir_path`,`cms_contents`.`create_at`,`cms_contents`.`update_at`,`cms_contents`.`delete_at` FROM `cms_contents` WHERE exists (select content_id from cms_authed_contents where cms_authed_contents.org_id in ('','{share_all}') and cms_authed_contents.content_id = cms_contents.id and delete_at = 0) and
 publish_status in ('published')  and delete_at=0 ORDER BY update_at desc\G

SELECT `cms_contents`.`id`,dir_path FROM `cms_contents` WHERE exists (select content_id from cms_authed_contents where cms_authed_contents.org_id in ('','{share_all}') and cms_authed_contents.content_id = cms_contents.id and delete_at = 0) and  publish_status in ('published')  and delete_at=0 ORDER BY update_at desc;
*/

/*
SELECT `cms_contents`.`id`,`cms_contents`.`content_type`,`cms_contents`.`content_name`,`cms_contents`.`keywords`,`cms_contents`.`description`,`cms_contents`.`thumbnail`,`cms_contents`.`source_type`,`cms_contents`.`outcomes`,`cms_contents`.`data`,`cms_contents`.`extra`,`cms_contents`.`suggest_time`,`cms_contents`.`author`,`cms_contents`.`creator`,`cms_contents`.`org`,`cms_contents`.`self_study`,`cms_contents`.`draw_activity`,`cms_contents`.`lesson_type`,`cms_contents`.`publish_status`,`cms_contents`.`reject_reason`,`cms_contents`.`remark`,`cms_contents`.`version`,`cms_contents`.`locked_by`,`cms_contents`.`source_id`,`cms_contents`.`latest_id`,`cms_contents`.`copy_source_id`,`cms_contents`.`parent_folder`,`cms_contents`.`dir_path`,`cms_contents`.`create_at`,`cms_contents`.`update_at`,`cms_contents`.`delete_at` FROM `cms_contents` WHERE (dir_path like '/61775c746cf950261f91d12c%' or dir_path like '/61ad84533ed7a7c32a8eeb46%') and  publish_status in ('published')  and delete_at=0 ORDER BY update_at desc LIMIT 20\G

SELECT `cms_contents`.`id`,dir_path FROM `cms_contents` WHERE (dir_path like '/61775c746cf950261f91d12c%' or dir_path like '/61ad84533ed7a7c32a8eeb46%') and  publish_status in ('published')  and delete_at=0 ORDER BY update_at desc
*/

func TestQueryContentsFolders(t *testing.T) {
	setupMilestone()
	op := initOperator("a44da070-1907-46c4-bc4c-f26ced889439", "14494c07-0d4f-5141-9db2-15799993f448", "", "")
	url := "/v1/contents_folders?publish_status=published&submenu=published&content_type=1,2,10&order_by=-update_at&page=1&page_size=20&path=&org_id=a44da070-1907-46c4-bc4c-f26ced889439"
	op.OrgID = "a44da070-1907-46c4-bc4c-f26ced889439"
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

func TestQueryContents(t *testing.T) {
	setupMilestone()
	op := initOperator("a44da070-1907-46c4-bc4c-f26ced889439", "14494c07-0d4f-5141-9db2-15799993f448", "", "")
	url := "/v1/contents?publish_status=published&page_size=10&content_type=1&name=&content_name=&org_id=a44da070-1907-46c4-bc4c-f26ced889439"
	op.OrgID = "a44da070-1907-46c4-bc4c-f26ced889439"
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

func TestQueryContentsLessonPlan(t *testing.T) {
	setupMilestone()
	op := initOperator("a44da070-1907-46c4-bc4c-f26ced889439", "14494c07-0d4f-5141-9db2-15799993f448", "", "")
	url := "/v1/contents_lesson_plans?org_id=a44da070-1907-46c4-bc4c-f26ced889439"
	op.OrgID = "a44da070-1907-46c4-bc4c-f26ced889439"
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

func TestQueryContentAboutLessonPlan(t *testing.T) {
	setupMilestone()
	op := initOperator("60c064cc-bbd8-4724-b3f6-b886dce4774f", "afdfc0d9-ada9-4e66-b225-20f956d1a399", "org1119@yopmail.com", "Bada1234")
	url := "/v1/contents/61a8788ab0af9acacbcd456f/live/token?org_id=60c064cc-bbd8-4724-b3f6-b886dce4774f"
	op.OrgID = "60c064cc-bbd8-4724-b3f6-b886dce4774f"
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

func TestQueryContentsID(t *testing.T) {
	setupMilestone()
	op := initOperator("60c064cc-bbd8-4724-b3f6-b886dce4774f", "0d3686a6-bf6a-4777-a716-31ce4aa0f516", "school1221a@yopmail.com", "Bada1234")
	url := "/v1/contents/61ad878f5ab1da1a6faf50c5?org_id=60c064cc-bbd8-4724-b3f6-b886dce4774f"
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

func TestSharedTooMany(t *testing.T) {
	setupMilestone()
	data, err := ioutil.ReadFile("/home/blt/Downloads/kidsloop_alpha_cms/body.json")
	if err != nil {
		t.Error(err)
	}
	op := initOperator("a44da070-1907-46c4-bc4c-f26ced889439", "14494c07-0d4f-5141-9db2-15799993f448", "pj.williams@calmid.com", "LakersRBest2021")
	url := "/v1/folders/share?org_id=a44da070-1907-46c4-bc4c-f26ced889439"
	res := DoHttpWithOperator(http.MethodPut, op, url, string(data))
	fmt.Println(res)
}
