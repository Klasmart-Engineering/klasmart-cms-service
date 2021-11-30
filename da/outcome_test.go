package da

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/dbo"
)

func TestGetMaxConsecutiveShortcode(t *testing.T) {
	ctx := context.Background()
	result, err := GetOutcomeDA().GetMaxConsecutiveShortcode(ctx, dbo.MustGetDB(ctx), "355f4d5f-1c4c-421f-9fa8-08e1d1c03f7d")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}
