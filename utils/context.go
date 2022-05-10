package utils

import (
	"context"

	gintrace "github.com/KL-Engineering/gin-trace"
)

func CloneContextWithTrace(ctx context.Context) context.Context {
	newContext := context.Background()

	badaCtx, ok := gintrace.GetBadaCtx(ctx)
	if ok {
		newContext, _ = badaCtx.EmbedIntoContext(newContext)
	}

	return newContext
}
