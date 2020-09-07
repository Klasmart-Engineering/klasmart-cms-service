package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestCreateOutcome(t *testing.T) {
	createView := OutcomeCreateView{
		OutcomeName:   "TestOutcomeXX",
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
	fmt.Println(data)
	res := DoHttp(http.MethodPost, prefix+"/learning_outcomes", string(data))
	fmt.Println(res)
}

func TestGetOutcome(t *testing.T) {
	outcomeID := "5f55d43f3695b7ca67729069"
	res := DoHttp(http.MethodGet, prefix+"/learning_outcomes/"+outcomeID, "")
	if res.StatusCode != http.StatusOK {
		t.Log(res.StatusCode)
		return
	}
	data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", string(data))
}

func TestUpdateOutcome(t *testing.T) {

}

func TestDeleteOutcome(t *testing.T) {

}

func TestQueryOutcome(t *testing.T) {

}

func TestLockOutcome(t *testing.T) {

}

func TestPublishOutcome(t *testing.T) {

}

func TestApproveOutcome(t *testing.T) {

}

func TestRejectOutcome(t *testing.T) {

}

func TestBulkPublishOutcome(t *testing.T) {

}

func TestBuldDeleteOutcome(t *testing.T) {

}

func TestQueryPrivateOutcome(t *testing.T) {

}

func TestQueryPendingOutcome(t *testing.T) {

}
