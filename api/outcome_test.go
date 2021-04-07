package api

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"net/http"
	"testing"

	"github.com/dgrijalva/jwt-go"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/ro"
)

func TestCreateOutcome(t *testing.T) {
	createView := OutcomeCreateView{
		OutcomeName:   "TestOutcomeYY",
		Assumed:       true,
		Program:       []string{"prg001", "pr002"},
		Subject:       []string{"sbj001", "sbj002"},
		Developmental: []string{"dvt001", "dvt002"},
		Skills:        []string{"skl001", "skl002"},
		Age:           []string{"age001", "age002"},
		Grade:         []string{"grd001", "grd002"},
		Estimated:     30,
		Keywords:      []string{"kyd001", "kyd002"},
		Description:   "some description",
		Shortcode:     "003NZ",
		Sets: []*OutcomeSetCreateView{
			{SetID: "60583986fc229e722bead09a"},
			{SetID: "60583cfcf5808149b5cbe24c"},
		},
	}
	data, err := json.Marshal(createView)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	res := DoHttp(http.MethodPost, prefix+"/learning_outcomes", string(data))
	fmt.Println(res)
}

func TestGetOutcome(t *testing.T) {
	outcomeID := "5f63016bacc44d2ec014a4e9"
	res := DoHttp(http.MethodGet, prefix+"/learning_outcomes/"+outcomeID, "")
	fmt.Println(res)
}

func TestUpdateOutcome(t *testing.T) {
	createView := OutcomeCreateView{
		OutcomeName:   "TestModifyOutcomeXX",
		Assumed:       false,
		Program:       []string{"Modify_prg001", "pr002"},
		Subject:       []string{"Modify_sbj001", "sbj002"},
		Developmental: []string{"Modify_dvt001"},
		Skills:        []string{"Modify_skl001", "skl002"},
		Age:           []string{"Modify_age001", "age002"},
		Grade:         []string{"Modify_grd001", "grd002"},
		Estimated:     45,
		Keywords:      []string{"Modify_kyd001", "kyd002"},
		Description:   "some description",
		Shortcode:     "12345",
		Sets: []*OutcomeSetCreateView{
			{SetID: "60616c5c43c74caa2b16623c"},
			{SetID: "60583cfcf5808149b5cbe24c"},
		},
	}
	data, err := json.Marshal(createView)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	outcomeID := "60616800e7c9026bb00d1d6c"
	res := DoHttp(http.MethodPut, prefix+"/learning_outcomes/"+outcomeID, string(data))
	fmt.Println(res)
}

func TestDeleteOutcome(t *testing.T) {
	outcomeID := "605af5e5ad682b6f63aebb58"
	res := DoHttp(http.MethodDelete, prefix+"/learning_outcomes/"+outcomeID, "")
	fmt.Println(res)
}

func TestQueryOutcome(t *testing.T) {
	query := fmt.Sprintf("set_name=%s&assumed=%d", "math", 1)
	res := DoHttp(http.MethodGet, prefix+"/learning_outcomes"+"?"+query, "")
	fmt.Println(res)
}

func TestLockOutcome(t *testing.T) {
	// 5f603ac9c96f2c0decc55e56
	outcomeID := "5f603a90029dfdc992fee14a"
	res := DoHttp(http.MethodPut, prefix+"/learning_outcomes/"+outcomeID+"/lock", "")
	fmt.Println(res)
}

func TestPublishOutcome(t *testing.T) {
	outcomeID := "5f63016bacc44d2ec014a4e9"
	//data := "{\"a:1}"
	//data := ""
	req := PublishOutcomeReq{
		Scope: "1",
	}
	data, err := json.Marshal(&req)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttp(http.MethodPut, prefix+"/learning_outcomes/"+outcomeID+"/publish", string(data))
	fmt.Println(res)
}

func TestApproveOutcome(t *testing.T) {
	outcomeID := "5f603a90029dfdc992fee14a"
	res := DoHttp(http.MethodPut, prefix+"/learning_outcomes/"+outcomeID+"/approve", "")
	fmt.Println(res)
}

func TestRejectOutcome(t *testing.T) {
	outcomeID := "5f56def450275d71418a1d4b"
	body := OutcomeRejectReq{RejectReason: "refuse"}
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

func TestGetLearningOutcomesByIDs(t *testing.T) {
	op := &entity.Operator{
		UserID: "1",
		OrgID:  "1",
		Role:   "admin",
	}
	ctx := context.Background()
	ids := []string{"5f5726af0944d7c38e20696f"}
	outcomes, err := model.GetOutcomeModel().GetLearningOutcomesByIDs(ctx, op, dbo.MustGetDB(ctx), ids)
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
		Role:   "admin",
	}
	ctx := context.Background()
	//ids := []string{"5f5726af0944d7c38e20696f"}
	ids := []string{}
	outcomes, err := model.GetOutcomeModel().GetLatestOutcomesByIDs(ctx, op, dbo.MustGetDB(ctx), ids)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range outcomes {
		fmt.Printf("%+v\n", o)
	}
}

func TestRedis(t *testing.T) {

	redisKey := fmt.Sprintf("%s:%s", da.RedisKeyPrefixOutcomeShortcode, "1")
	num, err := ro.MustGetRedis(context.Background()).Incr(redisKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(utils.NumToBHex(context.TODO(), int(num), 36, constant.ShortcodeShowLength))
}

func TestNumToBHex(t *testing.T) {
	fmt.Println("")
}

func TestFindRoot(t *testing.T) {
	orgs := []*external.Organization{
		{ID: "1", ParentID: "3"},
		{ID: "2", ParentID: "1"},
		{ID: "3", ParentID: "4"},
		{ID: "4", ParentID: ""},
	}
	root := orgs[3]
	for i := 0; i < len(orgs); i++ {
		if root.ParentID != "" && root.ParentID != root.ID {
			for _, o := range orgs {
				if o.ID == root.ParentID {
					root = o
					break
				}
			}
		}
	}
	fmt.Printf("%+v", root)
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
