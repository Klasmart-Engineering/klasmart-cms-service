package external

import (
	"context"
	"fmt"
	"testing"
)

var tchToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTMzMDMwNiwiaXNzIjoia2lkc2xvb3AifQ.k0aekIypdHRFhb8rGs0sMqW2oYwdGE1NSTuah2oiWEma1I19x0zm9mi-U41-LzhT8ikPtSOlclHtMciK7dLL7jnGiJLW_g6RzUTJcfrCqAEWDem6rFTnwmBhhaRcIzxVq43kvvww2StRuu5jnppSp8vfGkwVxDfD3MzZmvrwIqKPEaxxaCwJQXfWPfAccgV0Z-l2ODaAnK8kr6nSDQ6WwdjS5RXwF4KGKJpM7y9sUAw7GJ_FSoAzVy1eSWUsgX_RdzyU1TSxRQlBe9QIq5Etxc9wjAfL5NU3MC1bnD-zUvoCvTEdAmxLf60XR2S1f-vQD0RB2pub9h4iD1FiOrY-8w"

func TestAmsTeacherService_GetBySchools(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = tchToken
	IDs := []string{
		"0adee5ec-9454-44a9-b894-05ca1768b01e",
		"0bf25570-337d-42fd-a594-09821f0d59fb",
	}
	result, err := GetTeacherServiceProvider().GetBySchools(ctx, testOperator, IDs)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(result)
}
