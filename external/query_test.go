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

func TestQuery2(t *testing.T) {
	ctx := context.Background()
	var result []ProgramsConnectionResponse
	//op := &entity.Operator{
	//	Token: "",
	//}
	str, err := queryString(ctx, string(pf.FilterName()), string(pf.ConnectionName()), &result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(str)
}

func TestQuery(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		Token: "",
	}
	pf.FilterName()
	var result []ProgramsConnectionResponse
	err := pageQuery(ctx, op, pf, &result)
	if err != nil {
		t.Fatal(err)
	}
	for _, pageData := range result {
		for _, node := range pageData.Edges {
			fmt.Println(node.Node)
		}
	}
}
