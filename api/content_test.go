package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var server *Server

func TestMain(m *testing.M) {
	server = NewServer()
	code := m.Run()
	os.Exit(code)
}

const url = ""
const prefix = "/v1"

func DoHttp(method string, url string, body string) *http.Response {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, bytes.NewBufferString(body))
	req.Header.Add("Authorization", "")
	server.ServeHTTP(w, req)
	res := w.Result()
	return res
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