package da

import (
	"context"
	"testing"

	"github.com/KL-Engineering/dbo"
)

func TestDeleteTx(t *testing.T) {
	ctx := context.TODO()
	err := GetMilestoneRelationDA().DeleteTx(ctx, dbo.MustGetDB(ctx), []string{"607e7089eb753a919be411cf"})
	if err != nil {
		t.Fatal(err)
	}
}
