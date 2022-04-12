package utils

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-cn/helper"
)

func CloneContextWithTrace(ctx context.Context) context.Context {
	newContext := context.Background()

	badaCtx, ok := helper.GetBadaCtx(ctx)
	if ok {
		newContext, _ = badaCtx.EmbedIntoContext(newContext)
	}

	return newContext
}
