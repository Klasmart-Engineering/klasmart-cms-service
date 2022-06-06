package utils

import (
	"context"
	"github.com/gin-gonic/gin"
	"sync"

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

type ctxCacheKey struct{}

var _ctxCacheKey ctxCacheKey

func MaybeCacheIntoCtx(ctx context.Context, cacheKey string, value interface{}) {
	if cacheKey == "" {
		return
	}

	cache, ok := ctx.Value(_ctxCacheKey).(*sync.Map)
	if !ok {
		return
	}

	cache.Store(cacheKey, value)
}

func MaybeGetFromCtx[T any](ctx context.Context, cacheKey string, ret *T) (ok bool) {
	if cacheKey == "" {
		return
	}

	cache, ok := ctx.Value(_ctxCacheKey).(*sync.Map)
	if !ok {
		return
	}

	data, ok := cache.Load(cacheKey)
	if !ok {
		return
	}

	asserted, ok := data.(T)
	if !ok {
		return
	}
	*ret = asserted

	return true
}

func NewCachedCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, _ctxCacheKey, &sync.Map{})
}

func InsertCacheIntoGinCtx(c *gin.Context) {
	ctx := NewCachedCtx(c.Request.Context())
	c.Request = c.Request.WithContext(ctx)
}
