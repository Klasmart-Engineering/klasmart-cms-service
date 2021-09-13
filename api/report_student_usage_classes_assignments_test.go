package api

import (
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"net/http"
	"testing"
)

func TestGetClassesAssignmentsOverview(t *testing.T) {
	setupMilestone()
	request := entity.ClassesAssignmentOverViewRequest{
		ClassIDs:  []string{"1", "2"},
		Durations: []entity.TimeRange{},
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("%s/reports/student_usage/classes_assignments_overview", prefix)
	//op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	//res := DoHttpWithOperator(http.MethodPost, op, url, string(data))
	res := DoHttp(http.MethodPost, url, string(data))
	fmt.Println(res)
}
