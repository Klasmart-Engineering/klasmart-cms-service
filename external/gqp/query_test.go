package gqp

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

type TestFilter struct {
	//Name string `gqls:"__name__"`
	//ID   *UUIDFilter   `gqls:"id,omitempty"`
	Name StringFilter `gqls:"name,omitempty"`
	AND  []TestFilter `gqls:"AND,omitempty"`
	//OR   []TestFilter  `gqls:"OR,omitempty"`
}

var TF = TestFilter{
	Name: StringFilter{
		Operator: OperatorTypeContains,
		Value:    "mmm",
	},
	AND: []TestFilter{
		//{Name: "xxx"},
		//{Name: "yyy"},
		{Name: StringFilter{OperatorTypeContains, "xxx", false}},
		{Name: StringFilter{OperatorTypeContains, "yyy", false}},
	},
}

func TestFilterMarshal(t *testing.T) {
	result, err := marshalFilter(pf)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(result)
}

var pf = ProgramFilter{
	//ID:   &UUIDFilter{Operator: OperatorTypeEq, Value: "id1"},
	Name: &StringFilter{Operator: OperatorTypeContains, Value: "Bada"},
	AND: []ProgramFilter{
		{Name: &StringFilter{Operator: OperatorTypeContains, Value: "Bada"}},
	},
	OR: []ProgramFilter{
		{Name: &StringFilter{Operator: OperatorTypeContains, Value: "Math"}},
	},
}

func TestNodeMarshal(t *testing.T) {
	ctx := context.Background()
	var result []ProgramsConnectionResponse
	values, err := marshalFiled(ctx, &result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(values)
}
func TestQuery(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		Token: token,
	}
	pf.FilterType()
	var result []ProgramsConnectionResponse
	err := Query(ctx, op, pf.FilterType(), pf, &result)
	if err != nil {
		t.Fatal(err)
	}
	for _, pageData := range result {
		for _, node := range pageData.Edges {
			fmt.Println(node.Node)
		}
	}
}

var token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImExZmFhNTc1LWVhMGMtNGEzMC04YmI2LTViYjM3M2MwYjA5NCIsImVtYWlsIjoiYWxsMTEyNEB5b3BtYWlsLmNvbSIsImV4cCI6Mjc0ODcyMDcwMSwiaXNzIjoiY2FsbWlkLWRlYnVnIn0.qVfuPzeQFKvHlOg3aPh45rQ878LrGif5I3yb3eZj7Z8"
