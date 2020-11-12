package external

import (
	"context"
	"encoding/json"
	"fmt"
	cl "gitlab.badanamu.com.cn/calmisland/chlorine"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
)

func TestOrganizationService_BatchGet(t *testing.T) {
	config.LoadEnvConfig()
	orgs, err := GetOrganizationServiceProvider().BatchGet(context.Background(), []string{
		"3f135b91-a616-4c80-914a-e4463104dbac",
		"3f135b91-a616-4c80-914a-e4463104dbad",
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := range orgs {
		if orgs[i] != nil {
			fmt.Println(*(orgs[i]))
		} else {
			fmt.Println(i)
		}
	}
}

func TestClassService_BatchGet(t *testing.T) {
	classes, err := GetClassServiceProvider().BatchGet(context.Background(), []string{
		"f3d3cdf5-9ca8-44cf-a604-482e5d183049",
		"3b8074d5-893c-41c7-942f-d2115cc8bc32",
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := range classes {
		if classes[i] != nil {
			fmt.Println(*(classes[i]))
		} else {
			fmt.Println(i)
		}
	}
}

func TestXX(t *testing.T){
//	q := `{
//  "data": {
//    "organizations": null
//  }
//}
//`
	q1 := `{
  "data": {
    "organizations": [
      {
        "organization_id": "e236d102-5324-4740-8f36-629451557a2a",
        "organization_name": "Organization 1"
      }
    ]
  }
}`
	payload := make([]*Organization, 2)
	res := cl.Response{
		Data: &struct {
			Organizations []*Organization `json:"organizations"`
		}{},
	}

	json.Unmarshal([]byte(q1), &res)
	fmt.Printf("%#v\n", res.Data)
	fmt.Println(payload)
}

//func TestClassService_GetStudents(t *testing.T) {
//	students, err := GetClassServiceProvider().GetStudents(context.Background(), "f3d3cdf5-9ca8-44cf-a604-482e5d183049")
//	if err != nil {
//		t.Fatal(err)
//	}
//	for i := range students {
//		if students[i] != nil {
//			fmt.Println(*(students[i]))
//		} else {
//			fmt.Println(i)
//		}
//	}
//}
