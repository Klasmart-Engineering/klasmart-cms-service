package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var server *Server

const prefix = "/v1"

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

var token1 = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImNjOWYxNmY1LWJmMjctNDg0YS1hMzk3LWM5MzAyOWJjZGFlNyIsImVtYWlsIjoidGVjaDFAeW9wbWFpbC5jb20iLCJleHAiOjE2NDgwMDgzNzksImlzcyI6ImtpZHNsb29wIn0.XkI6M7UdN4LhZwbMXoVMli6pO0JcI4PB9Oz6N4wRUSGHfh40UbJm2qYzVdxT3OaUpIG6UpJJRGRb-jHXK1aYcpvPdTou2AikMxqIwQuhwSVxHgL9zvDIcF0oRSdDEdYWisN_9dVS0bOOEvD1nRBQiO147eC99NvtT9hANgbe46C2irw5ysp0O-pA93CLLyQ6S6_gB3VGyXIXzW2tw2JO0kcJP965jDIDyCqcU7wZ1qnyTYeU1biDqUbDweh2p1sLI2RzoJ6y_taYJtUm7eFojQ98E2no2Airy1MB6NeHqLkld9y0xBnZja2bXQETnGnLqAqRTzarQAWCNXxCRAAoCA"

func TestQueryContentsFolders(t *testing.T) {
	op := &entity.Operator{
		OrgID:  "1d30ce69-fdaf-448c-9da4-b536e73ef8b9",
		UserID: "cc9f16f5-bf27-484a-a397-c93029bcdae7",
		Token:  token1,
	}
	//op := initOperator("a44da070-1907-46c4-bc4c-f26ced889439", "14494c07-0d4f-5141-9db2-15799993f448", "", "")
	url := "/v1/contents_folders?publish_status=published&submenu=published&content_type=1%2C2%2C10&order_by=-update_at&page=1&page_size=20&path=&org_id=1d30ce69-fdaf-448c-9da4-b536e73ef8b9"
	//op.OrgID = "1d30ce69-fdaf-448c-9da4-b536e73ef8b9"
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
