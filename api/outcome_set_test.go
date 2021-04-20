package api

import (
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
	"testing"
)

func TestCreateOutcomeSet(t *testing.T) {
	orgID := "8a31ebab-b879-4790-af99-ee4941a778b3"
	setupMilestone()
	op := initOperator(orgID, "", "")
	req := model.OutcomeSetCreateView{
		SetName: "math",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("%s/sets?org_id=%s", prefix, op.OrgID)
	res := DoHttpWithOperator(http.MethodPost, op, url, string(data))
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
		//OutcomeAncestorIDs: []string{"5f684175cc913614b8cd73d9", "5f684372cc913614b8cd7439"},
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
