package external

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/KL-Engineering/kidsloop-cms-service/utils"
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

var prgToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MjE2NjE2NywiaXNzIjoia2lkc2xvb3AifQ.wm7rBpYLzGVH4egx6t00WPHfEx3BiGiRMfAXsSWF99uKuDebk529wkQykgAY1_NAk4qY_SKeDx8PJ5fvl12_Re7BjbU0ClUDlAQBQBpZ0sctuux9BhJVcqhCzd0658QWejGTmFoVDbMYWfwvls_Vpo5gm5S5bfCRzgFmbduZPbLYy_RGEyuhxc4-Hhd1ieSpyIZK1wAWZYL1nmixFWjT_HrrNo-C8sfTNHNRGWsfuNzjZAxUyQ9xVAbJ3TX3txtB3WMiJSErGBI5HtFRwdi0VlZ6lHpWkB5dZEQLX86kdj_LW_kqVMb1nbOaT5nmKjDuJ-TgyPLKnpJbdXBeN19QRw"

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

func TestQueryProgramText(t *testing.T) {
	ctx := context.Background()
	_ids := []string{"123"}
	sb := new(strings.Builder)
	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$program_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: program(id: $program_id_%d) {id name status system}\n", index, index)
	}
	sb.WriteString("}")
	t.Log(sb.String())
}
