package external

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestAmsProgramService_BatchGet(t *testing.T) {
	ids := []string{"04c630cc-fabe-4176-80f2-30a029907a33", "7565ae11-8130-4b7d-ac24-1d9dd6f792f2"}
	programs, err := GetProgramServiceProvider().BatchGet(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetProgramServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(programs) != len(ids) {
		t.Errorf("GetProgramServiceProvider().BatchGet() want %d results got %d", len(ids), len(programs))
		return
	}

	for _, program := range programs {
		if program == nil {
			t.Error("GetProgramServiceProvider().BatchGet() get null")
			return
		}
	}
	time.Sleep(time.Second)
}

var prgToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MTg4OTg4OSwiaXNzIjoia2lkc2xvb3AifQ.RQMWjHbKtkv5LsD478cQt_BY7ZHTxRo1sUEJN1ymaZyXwIZXpVcjNHW8jURw4U9GYgg4pzfnCqmaD1X7XoYSL7cWgzS8_d6XrcWmn5kVuC4BBoEdg9JfPxoxfQCZsmDKM6kds-cJu7HH2RM_U8a55Az4tNhZ0ZQ8xZKe4pEJBeQJcUsCbdnH0LF-Lak2OWoP3vQ9bI_sA3ak0gvfRpBpm8_1B8ZDiR2VrkOr_h2sl-iIdQBW7LbPd2EI4nr6RkZyXytTfgiRpCAmbn3OYZOTWzf5OMsubE2NeCyZKEIJZc0rmMTGwVTe2tpXudiE21W-pFnPk97PFQFhPADFfgbN3A"

func TestAmsProgramService_GetByOrganization(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = prgToken
	provider := AmsProgramConnectionService{}
	programs1, err := provider.AmsProgramService.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetProgramServiceProvider().GetByOrganization() error = %v", err)
		return
	}
	programs2, err := provider.GetByOrganization(ctx, testOperator)
	if err != nil {
		t.Errorf("GetProgramServiceProvider().GetByOrganization() error = %v", err)
		return
	}
	fmt.Println("lens:", len(programs1) == len(programs2))
}
