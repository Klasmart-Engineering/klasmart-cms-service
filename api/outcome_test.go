package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/utils"

	"github.com/golang-jwt/jwt"

	"github.com/KL-Engineering/kidsloop-cms-service/config"

	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/KL-Engineering/ro"
)

func TestCreateOutcome(t *testing.T) {
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "", "")
	createView := model.OutcomeCreateView{
		OutcomeName:   "outcome001",
		Shortcode:     "X0001",
		Assumed:       true,
		Program:       []string{"75004121-0c0d-486c-ba65-4c57deacb44b"},
		Subject:       []string{"36c4f7939aa3-4fb8-84f0-68a2ab920d5a"},
		Developmental: []string{"b4cd42b8-a09b-4f66-a03a-b9f6b6f69895", "fa8ff09d-9062-4955-9b20-5fa20757f1d9"},
		Skills:        []string{"d50cff7c-b0c7-43be-8ec7-877fa4c9a6fb", "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"},
		Grade:         []string{"3ee3fd4c-6208-494f-9551-d48fabc4f42a"},
		Age:           []string{"bb7982cd-020f-4e1a-93fc-4a6874917f07"},
		Estimated:     30,
		Keywords:      []string{"kyd001", "kyd002"},
		Description:   "Hello, Brilliant",
		Sets: []*model.OutcomeSetCreateView{
			{SetID: "6077b3a9809f75d51b5cb023"},
		},
	}
	data, err := json.Marshal(createView)
	if err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("%s/learning_outcomes?org_id=%s", prefix, op.OrgID)
	res := DoHttpWithOperator(http.MethodPost, op, url, string(data))
	fmt.Println(res)
}

func TestGetOutcome(t *testing.T) {
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "", "")
	outcomeID := "609b7ec4f060b597ab4782c7"
	url := fmt.Sprintf("%s/learning_outcomes/%s?org_id=%s", prefix, outcomeID, op.OrgID)
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

func TestUpdateOutcome(t *testing.T) {
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "", "")
	outcomeID := "607e4a1e4225cb7dcdb55108"
	createView := model.OutcomeCreateView{
		OutcomeName:   "TestModifyOutcomeXX",
		Assumed:       false,
		Program:       []string{"75004121-0c0d-486c-ba65-4c57deacb44b"},
		Subject:       []string{"5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", "36c4f793-9aa3-4fb8-84f0-68a2ab920d5a"},
		Developmental: []string{"b4cd42b8-a09b-4f66-a03a-b9f6b6f69895", "fa8ff09d-9062-4955-9b20-5fa20757f1d9"},
		Skills:        []string{"d50cff7c-b0c7-43be-8ec7-877fa4c9a6fb", "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"},
		Grade:         []string{"3ee3fd4c-6208-494f-9551-d48fabc4f42a"},
		Age:           []string{"bb7982cd-020f-4e1a-93fc-4a6874917f07"},
		Estimated:     45,
		Keywords:      []string{"Modify_kyd001", "kyd002"},
		Description:   "some description",
		Shortcode:     "12345",
		Sets:          []*model.OutcomeSetCreateView{},
	}
	data, err := json.Marshal(createView)
	if err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("%s/learning_outcomes/%s?org_id=%s", prefix, outcomeID, op.OrgID)
	res := DoHttpWithOperator(http.MethodPut, op, url, string(data))
	fmt.Println(res)
}

func TestDeleteOutcome(t *testing.T) {
	outcomeID := "605af5e5ad682b6f63aebb58"
	res := DoHttp(http.MethodDelete, prefix+"/learning_outcomes/"+outcomeID, "")
	fmt.Println(res)
}

func TestQueryOutcome(t *testing.T) {
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "", "")
	//query := fmt.Sprintf("set_name=%s&assumed=%d", "math", 1)
	query := fmt.Sprintf("search_key=%s&publish_status=%s", "12345", "draft")
	url := fmt.Sprintf("%s/learning_outcomes?org_id=%s&%s", prefix, op.OrgID, query)
	res := DoHttpWithOperator(http.MethodGet, op, url, "")
	fmt.Println(res)
}

func TestLockOutcome(t *testing.T) {
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "", "")
	outcomeID := "607905030e4404103a3f595d"
	url := fmt.Sprintf("%s/learning_outcomes/%s/lock?org_id=%s", prefix, outcomeID, op.OrgID)
	res := DoHttpWithOperator(http.MethodPut, op, url, "")
	fmt.Println(res)
}

func TestPublishOutcome(t *testing.T) {
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "", "")
	outcomeID := "609cbafd6b534d371d08c083"
	req := model.PublishOutcomeReq{
		Scope: "1",
	}
	data, err := json.Marshal(&req)
	if err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("%s/learning_outcomes/%s/publish?org_id=%s", prefix, outcomeID, op.OrgID)
	res := DoHttpWithOperator(http.MethodPut, op, url, string(data))
	fmt.Println(res)
}

func TestApproveOutcome(t *testing.T) {
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "", "")
	outcomeID := "609cbafd6b534d371d08c083"
	url := fmt.Sprintf("%s/learning_outcomes/%s/approve?org_id=%s", prefix, outcomeID, op.OrgID)
	res := DoHttpWithOperator(http.MethodPut, op, url, "")
	//res := DoHttp(http.MethodPut, prefix+"/learning_outcomes/"+outcomeID+"/approve", "")
	fmt.Println(res)
}

func TestRejectOutcome(t *testing.T) {
	outcomeID := "5f56def450275d71418a1d4b"
	body := model.OutcomeRejectReq{RejectReason: "refuse"}
	data, err := json.Marshal(&body)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttp(http.MethodPut, prefix+"/learning_outcomes/"+outcomeID+"/reject", string(data))
	fmt.Println(res)
}

func TestBulkPublishOutcome(t *testing.T) {

}

func TestBulkDeleteOutcome(t *testing.T) {

}

func TestQueryPrivateOutcome(t *testing.T) {

}

func TestQueryPendingOutcome(t *testing.T) {

}

func TestGenerateShortcode(t *testing.T) {
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "", "")
	data, err := json.Marshal(ShortcodeRequest{
		Kind: entity.KindOutcome,
	})
	if err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("%s/shortcode?org_id=%s", prefix, op.OrgID)
	res := DoHttpWithOperator(http.MethodPost, op, url, string(data))
	fmt.Println(res)
}

func TestGetLearningOutcomesByIDs(t *testing.T) {
	op := &entity.Operator{
		UserID: "1",
		OrgID:  "1",
	}
	ctx := context.Background()
	ids := []string{"5f5726af0944d7c38e20696f"}
	outcomes, err := model.GetOutcomeModel().GetByIDs(ctx, op, dbo.MustGetDB(ctx), ids)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range outcomes {
		fmt.Printf("%+v\n", o)
	}
}

func TestGetLatestOutcomesByIDs(t *testing.T) {
	op := &entity.Operator{
		UserID: "1",
		OrgID:  "1",
	}
	ctx := context.Background()
	//ids := []string{"5f5726af0944d7c38e20696f"}
	ids := []string{}
	outcomes, err := model.GetOutcomeModel().GetLatestByIDsMapResult(ctx, op, dbo.MustGetDB(ctx), ids)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range outcomes {
		fmt.Printf("%+v\n", o)
	}
}

func TestRedis(t *testing.T) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", da.RedisKeyPrefixOutcomeShortcode, "1")
	num, err := ro.MustGetRedis(ctx).Incr(ctx, redisKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(utils.NumToBHex(context.TODO(), int(num), 36, constant.ShortcodeShowLength))
}

func TestNumToBHex(t *testing.T) {
	fmt.Println("")
}

var token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImI0MjI3NDNlLTllYmMtNWVkNy1iNzI1LTA2Mjk5NGVjNzdmMiIsImVtYWlsIjoiYnJpbGxpYW50LnlhbmdAYmFkYW5hbXUuY29tLmNuIiwiZXhwIjoxNjA1MTc4Nzk2LCJpc3MiOiJraWRzbG9vcCJ9.sDkGFTIWm-NgEDfNJoMS_3KoKcZs0smnR7whqWY0AMnYLFYX3j_Saj6gHjXHpmZMVewbnaNfv9lYfhSokFBZaCcYyeVBXQo6DHL6nppsMUFwmcTjl-NjqSGwYUvjpV7cmkmL33H8KojEuBUDP8kOK-cF5Km28PC6sV2nFRVBNFBXlcNsdB-CIQEeycCzRhw078GAP64Bpugay8W-77keldN-C1Qnrc6spbSCOKnxMpT94pBRzgB8D-vHdcnvB3zlfPj8RYWFlGE_uufHfPTSgS-nTzrz8vRhiJdOAYdPys90w87jGfmopm1AT-qDSqa4Qf8hMW4bj_UDAa4-1bI-yQ"

func TestVerifyToken(t *testing.T) {
	config.LoadEnvConfig()
	claims := &struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		*jwt.StandardClaims
	}{}
	_, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) {
		return config.Get().AMS.TokenVerifyKey, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(claims)
}
