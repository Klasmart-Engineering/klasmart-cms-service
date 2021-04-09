package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCreateOutcomeSet(t *testing.T) {
	req := OutcomeSetCreateView{
		SetName: "math",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	res := DoHttp(http.MethodPost, prefix+"/sets"+"?org_id=e85adc92-b91a-4794-a6c2-860885bd58be", string(data))
	//request, err := http.NewRequest("POST", "https://hub-test.kidsloop.cn/v1/ecards", bytes.NewBuffer(data))
	//if err != nil {
	//	t.Fatal(err)
	//}
	//res, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

func TestPullOutcomeSet(t *testing.T) {
	name := "math"
	res := DoHttp(http.MethodGet, prefix+"/sets"+"?org_id=e85adc92-b91a-4794-a6c2-860885bd58be&set_name="+name, "")
	fmt.Println(res)
}

func TestBulkBindOutcomeSet(t *testing.T) {
	req := BulkBindOutcomeSetRequest{
		//OutcomeIDs: []string{"5f684175cc913614b8cd73d9", "5f684372cc913614b8cd7439"},
		//SetIDs:     []string{"60583986fc229e722bead09a", "60583cfcf5808149b5cbe24c"},
		OutcomeIDs: []string{"5f684175cc913614b8cd73d9", "5f684372cc913614b8cd7439"},
		SetIDs:     []string{"60583986fc229e722bead09a", "60583cfcf5808149b5cbe24c"},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttp(http.MethodPost, prefix+"/sets/bulk_bind"+"?org_id=e85adc92-b91a-4794-a6c2-860885bd58be", string(data))
	fmt.Println(res)
}
