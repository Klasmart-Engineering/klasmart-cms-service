package external

import (
	"context"
	"fmt"
	"testing"
)

var stuToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTI5MjczNiwiaXNzIjoia2lkc2xvb3AifQ.G8XdvbOW25mdqtqklkdxYc0X1qO5Xk4WE2q4pouzAAHkrP6PDQjkIzOkbVUKTL6Sq26Ppriwuz4K5uRG7K6KlZv0C4e1EUimrthlts0H6CoeNkHtrRppo1W55Hmu2ERnLee334KkOzcGgIb_AKsh64355rwVEcfBPNwd_qSlP_WkBwapJHydQmuqDr0bFX-FCFexy0Y5ajI3ybOWbeJ7N6kgR-9gzxrnbWHBY9DeUO2elNdcMAVud6R0eGEFy1p6sycYD2s2ND-5yV9gsaOZFwbzIVVpLB0YHSUQrYCzNkkQwVeEXIRoFEVOVWFUBHXD5hxMlXPMHNQ6eXMUcj2p7w"

func TestAmsStudentConnectionService_GetByClassIDs(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = stuToken
	IDs := []string{
		"0ff7769f-cc94-4a80-a780-dcb0947db18b",
		"00429737-f515-4348-b24f-919c2f82a2aa",
	}
	//result, err := GetSchoolServiceProvider().GetByClasses(ctx, testOperator, IDs)
	result, err := GetStudentServiceProvider().GetByClassIDs(ctx, testOperator, IDs)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range result {
		fmt.Println(k, v)
	}
}
