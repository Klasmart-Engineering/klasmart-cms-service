package connections

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

var pf = ProgramFilter{
	//ID: &UUIDFilter{Operator: OperatorTypeEq, Value: "id1"},
	Name: &StringFilter{Operator: OperatorTypeContains, Value: "Bada"},
	AND: []ProgramFilter{
		{Name: &StringFilter{Operator: OperatorTypeContains, Value: "Bada"}},
	},
	OR: []ProgramFilter{
		{Name: &StringFilter{Operator: OperatorTypeContains, Value: "Math"}},
	},
}

func TestQueryString(t *testing.T) {
	result, err := queryString(pf)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(result)
}

var token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImExZmFhNTc1LWVhMGMtNGEzMC04YmI2LTViYjM3M2MwYjA5NCIsImVtYWlsIjoiYWxsMTEyNEB5b3BtYWlsLmNvbSIsImV4cCI6Mjc0ODcyMDcwMSwiaXNzIjoiY2FsbWlkLWRlYnVnIn0.qVfuPzeQFKvHlOg3aPh45rQ878LrGif5I3yb3eZj7Z8"

func TestQuery(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		Token: token,
	}
	Query[ProgramFilter, ProgramsConnectionResponse, ProgramsConnectionEdge](ctx, op, pf)
}
