package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"testing"
)

func TestGetTree(t *testing.T) {
	ctx := context.Background()

	op := &entity.Operator{
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "f27efd10-000e-4542-bef2-0ccda39b93d3",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImQ2NTNmODgwLWI5MjAtNDM1Zi04ZjJkLTk4YjVkNDYyMWViOCIsImVtYWlsIjoic2Nob29sXzAzMDMyOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0ODc5NTU4MywiaXNzIjoia2lkc2xvb3AifQ.ACFGroiiaK-3P1BwY-UVu9W7s9iuwA5ueaV6qpz6YdiMw-oo93MPdJqhNF0KnsN7w8o1q2nlSBYqXMDLUd4WLM5eOtrk_cWztBgl_BPfsuM2RGg-szLMzmxY1D5srqY1rqx-9kU-onCm-nbKCB090Zo8RPWh39OH4smX5zzSWSdXbPmTkUHFEMEhmpKDU9Sqt-5kNcOyVJr5kiu22Ku6hdwRCYxPIxganE1Uoex8LJnGloQChyo02o2mYR7ybxiXOnNmgCabDK4MysKXT0MB9wGlmv_LMgC97O0yTFEWz8hI8FosJMZDt5QX0i-9sXTPnPygCAQZ0GIr58ntO0EuZQ"

	roomData, err := external.GetH5PRoomScoreServiceProvider().BatchGet(ctx, op, []string{""})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(roomData)
}
