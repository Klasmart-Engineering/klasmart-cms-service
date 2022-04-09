package connections

import (
	"context"
	"errors"
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
	ctx := context.Background()
	result, err := queryString(ctx, pf)
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
	var programEdges []ProgramsConnectionEdge
	err := Query[ProgramFilter, ProgramsConnectionResponse](ctx, op, pf, func(ctx context.Context, result interface{}) error {
		concrete, ok := result.(ProgramsConnectionResponse)
		if !ok {
			return errors.New("assert failed")
		}
		programEdges = append(programEdges, concrete.Edges...)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(programEdges)
}
