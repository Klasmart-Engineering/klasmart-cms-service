package api

import (
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
	"os"
	"strings"
	"testing"
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
		Name:           "name01",
		Shortcode:      "00002",
		Organization:   &model.OrganizationView{OrganizationID: op.OrgID},
		ProgramIDs:     []string{"75004121-0c0d-486c-ba65-4c57deacb44b"},
		SubjectIDs:     []string{"5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", "36c4f793-9aa3-4fb8-84f0-68a2ab920d5a"},
		CategoryIDs:    []string{"b4cd42b8-a09b-4f66-a03a-b9f6b6f69895", "fa8ff09d-9062-4955-9b20-5fa20757f1d9"},
		SubcategoryIDs: []string{"d50cff7c-b0c7-43be-8ec7-877fa4c9a6fb", "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"},
		GradeIDs:       []string{"3ee3fd4c-6208-494f-9551-d48fabc4f42a"},
		AgeIDs:         []string{"bb7982cd-020f-4e1a-93fc-4a6874917f07"},
		OutcomeAncestorIDs: []string{
			"607905030e4404103a3f595d",
		},
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
	res := DoHttpWithOperator(http.MethodGet, op, prefix+"/milestones/"+"607905db23b82681052ad541"+"?org_id="+op.OrgID, "")
	fmt.Println(res)
}

func TestUpdateMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	req := model.MilestoneView{
		Name:           "name02",
		Shortcode:      "00001",
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
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttpWithOperator(http.MethodPut, op, prefix+"/milestones/607e4f4c9a752f785251dcef"+"?org_id="+op.OrgID, string(data))
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
		"search_key=name01",
		//"name=name01",
		//"description=math",
		//"shortcode=00001",
		"status=draft",
		"page=0",
		"page_size=0",
		"order_by=created_at",
	}
	condition := "&" + strings.Join(queryCondition, "&")
	res := DoHttpWithOperator(http.MethodGet, op, prefix+"/milestones"+"?org_id="+op.OrgID+condition, "")
	fmt.Println(res)
}

func TestPublishMilestone(t *testing.T) {
	setupMilestone()
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	req := model.MilestoneList{
		IDs: []string{"6076aede7fe2f93d2b6851af"},
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
	res := DoHttpWithOperator(http.MethodPut, op, prefix+"/milestones/"+"60769e209a3cad38f7ae4d0a/occupy"+"?org_id="+op.OrgID, "")
	fmt.Println(res)
}

func f(attach interface{}) {
	//fmt.Println(attach.TableName())
	fmt.Println(attach)
}

func TestAttach(t *testing.T) {
	var milestoneAttach entity.MilestoneAttach
	f(milestoneAttach)
	fmt.Println(milestoneAttach.TableName())
}
