package api

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"net/http"
	"testing"
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
	outcomeID := "5f55d43f3695b7ca67729069"
	res := DoHttp(http.MethodGet, prefix+"/learning_outcomes/"+outcomeID, "")
	fmt.Println(res)
}

func TestUpdateOutcome(t *testing.T) {
	createView := OutcomeCreateView{
		OutcomeName:   "TestModifyOutcomeXX",
		Assumed:       false,
		Program:       []string{"Modify_prg001", "pr002"},
		Subject:       []string{"Modify_sbj001", "sbj002"},
		Developmental: []string{"Modify_dvt001", "dvt002"},
		Skills:        []string{"Modify_skl001", "skl002"},
		Age:           []string{"Modify_age001", "age002"},
		Grade:         []string{"Modify_grd001", "grd002"},
		Estimated:     45,
		Keywords:      []string{"Modify_kyd001", "kyd002"},
		Description:   "some description",
	}
	data, err := json.Marshal(createView)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	outcomeID := "5f55d43f3695b7ca67729069"
	res := DoHttp(http.MethodPut, prefix+"/learning_outcomes/"+outcomeID, string(data))
	fmt.Println(res)
}

func TestDeleteOutcome(t *testing.T) {
	outcomeID := "5f55d43f3695b7ca67729069"
	res := DoHttp(http.MethodDelete, prefix+"/learning_outcomes/"+outcomeID, "")
	fmt.Println(res)
}

func TestQueryOutcome(t *testing.T) {
	query := fmt.Sprintf("outcome_name=%s&keywords=%s", "TestOutcomeYY", "A01")
	res := DoHttp(http.MethodGet, prefix+"/learning_outcomes"+"?"+query, "")
	fmt.Println(res)
}

func TestLockOutcome(t *testing.T) {
	outcomeID := "5f5726af0944d7c38e20696f"
	res := DoHttp(http.MethodPut, prefix+"/learning_outcomes/"+outcomeID+"/lock", "")
	fmt.Println(res)
}

func TestPublishOutcome(t *testing.T) {
	outcomeID := "5f5728fc14a58b1efaa90ade"
	res := DoHttp(http.MethodPut, prefix+"/learning_outcomes/"+outcomeID+"/publish", "")
	fmt.Println(res)
}

func TestApproveOutcome(t *testing.T) {
	outcomeID := "5f5728fc14a58b1efaa90ade"
	res := DoHttp(http.MethodPut, prefix+"/learning_outcomes/"+outcomeID+"/approve", "")
	fmt.Println(res)
}

func TestRejectOutcome(t *testing.T) {
	outcomeID := "5f56def450275d71418a1d4b"
	body := struct {
		RejectReason string `json:"reject_reason"`
	}{RejectReason: "refuse"}
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

func TestRedis(t *testing.T) {

	redisKey := fmt.Sprintf("%s:%s", da.RedisKeyPrefixOutcomeShortcode, "1")
	num, err := ro.MustGetRedis(context.Background()).Incr(redisKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(model.NumToBHex(int(num), 36))
}

func TestNumToBHex(t *testing.T) {
	fmt.Println(model.PaddingStr(model.NumToBHex(900, 36), 3))
}
