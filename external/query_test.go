package external

import (
	"context"
	"fmt"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
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

var sToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTIzNTA1MSwiaXNzIjoia2lkc2xvb3AifQ.cZr9W66wxJKlmi8whUDoU9D6KwTuLu7_x0vPp1jkg6ZGfhxSAA6BQBb0yQTuNyxuPCpZtnxLvmX4_jvooMqCHS_ZwFDvV3BsrkXZrcjchTZN8KVpRkK9_BN-L8AHF0ZHO4O_N9mqt-5_lFpTpqxE5jmazNLMUat5zp2BUxm3ODy7bIL3B7rv3Of07us-j__ZBw6Nq_PI6PthplpHtgDCVmZiOo78e21jdODnTZ7Wq7QkGLi312njpC-trIvOmZF1LMbgL3oH7y3ds6XuPiGUkm8ZLSfeW0pFE5JFOk29Izj6jz0N-Tb9njqS7P0eG1jJKiso2h9h6rVh1HgTtqOMOQ"

func TestSubPageQuerying(t *testing.T) {
	ctx := context.Background()
	IDs := []string{
		"1dd1d0b4-df1a-4486-bd6c-a89f9b92f779",
		"1e200965-df57-461e-8af3-e255886e8e41",
	}
	testOperator.Token = sToken
	result := make(map[string][]SchoolMembershipsConnectionResponse)
	subPageQuery(ctx, testOperator, "userNode", SchoolMembershipFilter{}, IDs, result)
}
