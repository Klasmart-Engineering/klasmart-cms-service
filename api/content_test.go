package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"gitlab.badanamu.com.cn/calmisland/common-cn/common"
	logger2 "gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
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
	}
}
func TestMain(m *testing.M) {
	log.Info(context.TODO(), "start kidsloop2 api service")
	defer func() {
		if err := recover(); err != nil {
			log.Info(context.TODO(), "kidsloop2 api service stopped", log.Any("err", err))
		} else {
			log.Info(context.TODO(), "kidsloop2 api service stopped")
		}
	}()

	// temp solution, will remove in next version
	logger2.SetLevel(logrus.DebugLevel)

	// read config
	config.LoadEnvConfig()

	log.Debug(context.TODO(), "load config success", log.Any("config", config.Get()))

	// init database connection
	initDB()

	log.Debug(context.TODO(), "init db success")
	initCache()

	log.Debug(context.TODO(), "init cache success")
	// init dynamodb connection
	storage.DefaultStorage()

	log.Debug(context.TODO(), "init storage success")

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

const url = ""
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
