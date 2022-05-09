package external

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

var rolToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MjAyMDcwMCwiaXNzIjoia2lkc2xvb3AifQ.kdH8d5mqXjBcGWXDETF3SdFUFVSGNfuQk1Txr_zylhp8L1MMqLf6Kb6Y1_L5SrQw0BDfT2wCshKWRD8iptUsiY0-XzekY5PHVdwd0IKMCBQx2Jc1tEKep5Muf-CN-Zimc3eaD0HdDfWQx0XAcK9Q3LGZsmEDizgoEeQHFpQJu70XKkmH4PhYzTHl9egmY0kXSkZS50V8Wv5UrVGJ11nLbYzyrf6Sel2M4pHE2SnB9GYB8n7b8gqWZ8niICs1uW1VkmWQA0VqUcBZ77jVIf69OgrD7WkxZ_y9_mwcYN5En04q6-eVPUluUy0b69qTRz84q8sKdvQr_HbY-o1a1gQQDA"

func TestAmsRoleService_GetRole(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = rolToken
	role, err := GetRoleServiceProvider().GetRole(ctx, testOperator, "Organization Admin")
	if err != nil {
		t.Fatal(err)
	}
	bs, err := json.Marshal(role)
	fmt.Println(string(bs))
}
