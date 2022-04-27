package external

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

var rtoken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTA2MTc4NiwiaXNzIjoia2lkc2xvb3AifQ.pgT5xd0DUKYJ7xrFVdsgyzWcyoHMJCemN2vV9H6NmFCcPpPjfsCI1CsvkuE3wimFlrDZ-itwg0tJr24YMfmuEKOn_ZPp7s_ZJDHSGxF9E2G9TAutX_VWPhLHcUYT4XAoHbXfkpJ-WDi-IHGrt2Ob2OxmYykUTB6f376t35gb2quDbRMzPUkNL-sBTJ0W2MrCB_92cDP7gkAaGrxZOv7rX4AKhhNOppaqMcliWiQDZ6myRmP1YQXsE-yLJBx4GznZCpryyv9C6CAh9vXJuf-9KyYJ5KXQ1VSQ2Lg7WhH81hodUXGTwuPSoFfsceAWKu2Q0c31NO-vdy876eU6D31H2Q"

func TestAmsRoleService_GetRole(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = rtoken
	role, err := GetRoleServiceProvider().GetRole(ctx, testOperator, "Organization Admin")
	if err != nil {
		t.Fatal(err)
	}
	bs, err := json.Marshal(role)
	fmt.Println(string(bs))
}
