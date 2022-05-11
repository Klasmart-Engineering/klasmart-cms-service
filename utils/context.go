package utils

import (
	"context"

	"github.com/KL-Engineering/tracecontext"
)

func CloneContextWithTrace(ctx context.Context) context.Context {
	newContext := context.Background()

	badaCtx, ok := tracecontext.GetTraceContext(ctx)
	if ok {
		newContext, _ = badaCtx.EmbedIntoContext(newContext)
	}

	return newContext
}
