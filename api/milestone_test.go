package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

func setupMilestone() {
	cfg := config.Get()
	if cfg == nil {
		cfg = &config.Config{}
	}
	cfg.DBConfig = config.DBConfig{
		ConnectionString: os.Getenv("connection_string"),
		MaxOpenConns:     8,
		MaxIdleConns:     8,
		ShowLog:          true,
		ShowSQL:          true,
	}
	cfg.RedisConfig = config.RedisConfig{
		OpenCache: true,
		Host:      os.Getenv("redis_host"),
		Port:      16379,
		Password:  "",
	}
	cfg.AMS = config.AMSConfig{
		EndPoint: os.Getenv("ams_endpoint"),
	}
	config.Set(cfg)
	initDB()
	initCache()
}

func TestCreateMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	req := model.MilestoneView{
		Name:           "mile02",
		Shortcode:      "Y0002",
		Organization:   &model.OrganizationView{OrganizationID: op.OrgID},
		ProgramIDs:     []string{"75004121-0c0d-486c-ba65-4c57deacb44b"},
		SubjectIDs:     []string{"5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", "36c4f793-9aa3-4fb8-84f0-68a2ab920d5a"},
		CategoryIDs:    []string{"b4cd42b8-a09b-4f66-a03a-b9f6b6f69895", "fa8ff09d-9062-4955-9b20-5fa20757f1d9"},
		SubcategoryIDs: []string{"d50cff7c-b0c7-43be-8ec7-877fa4c9a6fb", "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"},
		GradeIDs:       []string{"3ee3fd4c-6208-494f-9551-d48fabc4f42a"},
		AgeIDs:         []string{"bb7982cd-020f-4e1a-93fc-4a6874917f07"},
		OutcomeAncestorIDs: []string{
			"609b7ec4f060b597ab4782c7",
			"609b7fac691ad140891442cc",
		},
		WithPublish: true,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttpWithOperator(http.MethodPost, op, prefix+"/milestones"+"?org_id="+op.OrgID, string(data))
	fmt.Println(res)
}

func TestObtainMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	res := DoHttpWithOperator(http.MethodGet, op, prefix+"/milestones/"+"609b807f047581d7b0d46d17"+"?org_id="+op.OrgID, "")
	fmt.Println(res)
}

func TestUpdateMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	req := model.MilestoneView{
		Name:           "name07",
		Shortcode:      "00007",
		Organization:   &model.OrganizationView{OrganizationID: op.OrgID},
		ProgramIDs:     []string{"75004121-0c0d-486c-ba65-4c57deacb44b"},
		SubjectIDs:     []string{"5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", "36c4f793-9aa3-4fb8-84f0-68a2ab920d5a"},
		CategoryIDs:    []string{"b4cd42b8-a09b-4f66-a03a-b9f6b6f69895", "fa8ff09d-9062-4955-9b20-5fa20757f1d9"},
		SubcategoryIDs: []string{"d50cff7c-b0c7-43be-8ec7-877fa4c9a6fb", "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"},
		GradeIDs:       []string{"3ee3fd4c-6208-494f-9551-d48fabc4f42a"},
		OutcomeAncestorIDs: []string{
			"607e4a1e4225cb7dcdb55108",
		},
		Description: "Hello, Brilliant",
		WithPublish: true,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttpWithOperator(http.MethodPut, op, prefix+"/milestones/6099f4871551d2f217b77b7c"+"?org_id="+op.OrgID, string(data))
	fmt.Println(res)
}

func TestDeleteMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	req := model.MilestoneList{
		IDs: []string{},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttpWithOperator(http.MethodDelete, op, prefix+"/milestones"+"?org_id="+op.OrgID, string(data))
	fmt.Println(res)
}

func TestSearchMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	queryCondition := []string{
		//"search_key=name01",
		//"name=name01",
		//"description=math",
		//"shortcode=00001",
		"status=published",
		"page=1",
		"page_size=10",
		"order_by=created_at",
	}
	condition := "&" + strings.Join(queryCondition, "&")
	res := DoHttpWithOperator(http.MethodGet, op, prefix+"/milestones"+"?org_id="+op.OrgID+condition, "")
	obj := gjson.Parse(res)
	for _, v := range obj.Get("milestones").Array() {
		str := fmt.Sprintf("milestone_id: %v, locked_by: %v, last_edited_at: %v, last_edited_by: %v, locked_location: %v",
			v.Get("milestone_id"),
			v.Get("locked_by"),
			v.Get("last_edited_at"),
			v.Get("last_edited_by"),
			v.Get("locked_location"))
		t.Log(str)
	}
	t.Log(obj.Get("total"))
}

func TestSearchPrivateMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	queryCondition := []string{
		//"search_key=name01",
		//"name=name01",
		//"description=math",
		//"shortcode=00001",
		"status=published",
		"page=1",
		"page_size=10",
		"order_by=created_at",
	}
	op.UserID = "46790f5c-708f-59c2-9c2f-f24db63b86ad"
	condition := "&" + strings.Join(queryCondition, "&")
	res := DoHttpWithOperator(http.MethodGet, op, prefix+"/private_milestones"+"?org_id="+op.OrgID+condition, "")
	obj := gjson.Parse(res)
	for _, v := range obj.Get("milestones").Array() {
		str := fmt.Sprintf("milestone_id: %v, locked_by: %v, last_edited_at: %v, last_edited_by: %v, locked_location: %v",
			v.Get("milestone_id"),
			v.Get("locked_by"),
			v.Get("last_edited_at"),
			v.Get("last_edited_by"),
			v.Get("locked_location"))
		t.Log(str)
	}
	t.Log(obj.Get("total"))
}

func TestPublishMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	req := model.MilestoneList{
		IDs: []string{"609b9636b8f830a9402b0ba3"},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttpWithOperator(http.MethodPost, op, prefix+"/milestones/publish"+"?org_id="+op.OrgID, string(data))
	fmt.Println(res)
}

func TestOccupyMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	res := DoHttpWithOperator(http.MethodPut, op, prefix+"/milestones/"+"609b9636b8f830a9402b0ba3/occupy"+"?org_id="+op.OrgID, "")
	fmt.Println(res)
}

func TestCreateGeneral(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	res := DoHttpWithOperator(http.MethodPost, op, prefix+"/milestones/general"+"?org_id="+op.OrgID, "")
	fmt.Println(res)
}

func TestBulkPublishMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	req := model.MilestoneList{
		IDs: []string{"60ac9512480ebf7e27dd84db"},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttpWithOperator(http.MethodPut, op, prefix+"/bulk_publish/milestones"+"?org_id="+op.OrgID, string(data))
	fmt.Println(res)
}

func TestBulkApproveMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	req := model.MilestoneList{
		IDs: []string{"60ac9512480ebf7e27dd84db"},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttpWithOperator(http.MethodPut, op, prefix+"/bulk_approve/milestones"+"?org_id="+op.OrgID, string(data))
	fmt.Println(res)
}
