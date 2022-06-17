package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
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

func TestQueryAuthContent(t *testing.T) {
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

var token1 = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImNjOWYxNmY1LWJmMjctNDg0YS1hMzk3LWM5MzAyOWJjZGFlNyIsImVtYWlsIjoidGVjaDFAeW9wbWFpbC5jb20iLCJleHAiOjE2NDgwMTE1MjAsImlzcyI6ImtpZHNsb29wIn0.I1RD6SFn1U-mNl-hBNeNrN3xi44Asyxa4bCAJV6MKX6cxzRjFLvpiuNMS8R2wQ1Aw7aD4eFM3hNX-agKnbmSuBF3rMyAu7ciE3TpHyYVajTLj9ORfkl7SmRW2c-WyoMuWbKr2wQ8qxz3uvIHc3Wljx155e4XCRCiSY4Gpqd1NXA3Qs78WQ9maGOJPUtf8cBH2UAytICfYGHrZV9Ud8EINTkX7nx6kBd45aujdhlbzrHVkh-icvOJUcatm2B6llxUGY2IKHeKYWuYKcwirgtM0UFwjyLRojbMsibHcf-9yLHXJiUff315paLip2hrY-yUWIWwd8vPiMpTanXu4SikdQ"

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
	op := initOperator("a44da070-1907-46c4-bc4c-f26ced889439", "14494c07-0d4f-5141-9db2-15799993f448", "", "")
	url := "/v1/contents?publish_status=published&page_size=10&content_type=1&name=&content_name=&org_id=a44da070-1907-46c4-bc4c-f26ced889439"
	op.OrgID = "a44da070-1907-46c4-bc4c-f26ced889439"
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

func TestQueryContentsLessonPlan(t *testing.T) {
	op := initOperator("a44da070-1907-46c4-bc4c-f26ced889439", "14494c07-0d4f-5141-9db2-15799993f448", "", "")
	url := "/v1/contents_lesson_plans?org_id=a44da070-1907-46c4-bc4c-f26ced889439"
	op.OrgID = "a44da070-1907-46c4-bc4c-f26ced889439"
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

func TestQueryContentAboutLessonPlan(t *testing.T) {
	op := initOperator("60c064cc-bbd8-4724-b3f6-b886dce4774f", "afdfc0d9-ada9-4e66-b225-20f956d1a399", "org1119@yopmail.com", "Bada1234")
	url := "/v1/contents/61a8788ab0af9acacbcd456f/live/token?org_id=60c064cc-bbd8-4724-b3f6-b886dce4774f"
	op.OrgID = "60c064cc-bbd8-4724-b3f6-b886dce4774f"
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

func TestQueryContentsID(t *testing.T) {
	op := initOperator("60c064cc-bbd8-4724-b3f6-b886dce4774f", "0d3686a6-bf6a-4777-a716-31ce4aa0f516", "school1221a@yopmail.com", "Bada1234")
	url := "/v1/contents/61ad878f5ab1da1a6faf50c5?org_id=60c064cc-bbd8-4724-b3f6-b886dce4774f"
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

func TestSharedTooMany(t *testing.T) {
	data, err := ioutil.ReadFile("/home/blt/Downloads/kidsloop_alpha_cms/body.json")
	if err != nil {
		t.Error(err)
	}
	op := initOperator("a44da070-1907-46c4-bc4c-f26ced889439", "14494c07-0d4f-5141-9db2-15799993f448", "pj.williams@calmid.com", "LakersRBest2021")
	url := "/v1/folders/share?org_id=a44da070-1907-46c4-bc4c-f26ced889439"
	res := DoHttpWithOperator(http.MethodPut, op, url, string(data))
	fmt.Println(res)
}

var stmToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJraWRzbG9vcC1jbXMiLCJleHAiOjE2NTU0NzQwNDYsImlhdCI6MTY1NTM4NzY0NiwiaXNzIjoic3RtLWxhbWJkYSIsInN1YiI6ImF1dGhvcml6YXRpb24ifQ.PvnMGaAt6KxhCLbcSet6dIi02-n7-5glRA6luLaLGbv0t1TEQN6q4_Y78T_fmGnb0PvvKlIu87LYvlxFL5QZlcqk3uPqaLe60Jmy6cAbVZTMAaYlptG7sT0lGgw10fJEFvETD_SiKSdh3pbV_7DhY_0jbCjT7g8vTsUNnrVOIxaOi3FALimOTN9jIoUiJIoLKZn6X4DVXa3GMxdxbCTo_HaQ0jA1tIjY6_7tYXVp8ChWBUfwNG_S87JAOk_QTmP1Yp6BzBNMzjx8ru2_6IIl0kkGuTxhiZgevBK6g_E7niJSzUF8K1kWc4KAidk3MSFQutaS2NpJ3843qPbMpddQCA1y9jl-pAaiPZaWCXnw-QKF3zLnRhpe0ATWvzqVjdDKxauPkmcjfkk8Q5XT8phr3mxvp4TnK3CasY3Jo4zjRCzrauvAKj00nYnK5qrzrK6ybYPwU9lx2VW2PTnCaycbWlvs4H81SWsmT5Kpx1zzQBGnlFoM2rg4a1uvj1UzFuCB"

func TestGetSTMLessonPlans(t *testing.T) {
	op := entity.Operator{
		Token: stmToken,
	}
	IDs := []string{
		"628da79e552ba3b9994c9200",
		"6257868a9456ed3fc792b775",
		"624419aace8e2cbaa66ca0f8",
	}
	data, err := json.Marshal(IDs)
	if err != nil {
		t.Fatal(err)
	}
	url := "/v1/internal/stm/contents"
	res := DoHttpWithOperator(http.MethodPost, &op, url, string(data))
	fmt.Println(res)
}
