package da

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/dbo"
)

func TestDeleteTx(t *testing.T) {
	ctx := context.TODO()
	err := GetMilestoneRelationDA().DeleteTx(ctx, dbo.MustGetDB(ctx), []string{"607e7089eb753a919be411cf"})
	if err != nil {
		t.Fatal(err)
	}
}
