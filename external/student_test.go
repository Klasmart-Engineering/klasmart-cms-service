package external

import (
	"context"
	"fmt"
	"testing"
)

var stuToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MjE0NTk3MiwiaXNzIjoia2lkc2xvb3AifQ.aMTeQ32iIk7wokefzp92UYPy_eqva85S3q__wmHnaZN7pJsmnRTrrM3dZ2u8PbEFH1N1Z_8O-YPG4mfToEPnaal9FsuMcXADGknfQoUGwQA8TkdmB6z1KFG8gmrCZCWJW0dQrqnuJyQbk-2SqTb8JsWoTwQ5YrvCoG59tWx6ptXFqWYl8wu8LnXp9I4HeL-Ol2e0c9_h1xR1J9LDnmmhnWL3Ca9zxbwHW8igmyfSU5lR508mKglmHtrST1Bd_8l79dCXq2yTCWV88QuxowFc-juafUNXUG3Qy3UtU-q2YbrEGb_HQUBYQFrs5eOitZMj5zu87DpobmzGF8kdZK90RA"

func TestAmsStudentConnectionService_GetByClassIDs(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = stuToken
	IDs := []string{
		"0ff7769f-cc94-4a80-a780-dcb0947db18b",
		"00429737-f515-4348-b24f-919c2f82a2aa",
	}
	provider := AmsStudentConnectionService{}
	result1, err := provider.AmsStudentService.GetByClassIDs(ctx, testOperator, IDs)
	if err != nil {
		t.Fatal(err)
	}
	result2, err := provider.GetByClassIDs(ctx, testOperator, IDs)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("len:", len(result1) == len(result2))
}
