package external

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

type TestFilter struct {
	//Name string `gqls:"__name__"`
	//ID   *UUIDFilter   `gqls:"id,omitempty"`
	Name *StringFilter `gqls:"name,omitempty"`
	AND  []TestFilter  `gqls:"AND,omitempty"`
	//OR   []TestFilter  `gqls:"OR,omitempty"`
}

var TF = TestFilter{
	Name: &StringFilter{
		Operator: StringOperator(OperatorTypeContains),
		Value:    "mmm",
	},
	AND: []TestFilter{
		//{Name: "xxx"},
		//{Name: "yyy"},
		{Name: &StringFilter{StringOperator(OperatorTypeContains), "xxx", false}},
		{Name: &StringFilter{StringOperator(OperatorTypeContains), "yyy", false}},
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
	//ID:   &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: "id1"},
	Name: &StringFilter{Operator: StringOperator(OperatorTypeContains), Value: "Bada"},
	AND: []ProgramFilter{
		{Name: &StringFilter{Operator: StringOperator(OperatorTypeContains), Value: "Bada"}},
	},
	OR: []ProgramFilter{
		{Name: &StringFilter{Operator: StringOperator(OperatorTypeContains), Value: "Math"}},
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
		Token: "",
	}
	pf.FilterType()
	var result []ProgramsConnectionResponse
	err := pageQuery(ctx, op, pf.FilterType(), pf, &result)
	if err != nil {
		t.Fatal(err)
	}
	for _, pageData := range result {
		for _, node := range pageData.Edges {
			fmt.Println(node.Node)
		}
	}
}
