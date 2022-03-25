package da

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"gitlab.badanamu.com.cn/calmisland/ro"
)

type LazyRefreshCacheOption struct {
	CacheKey        *ro.StringParameterKey
	LockerKey       *ro.StringParameterKey
	RefreshDuration time.Duration
	RawQuery        func(ctx context.Context, request interface{}) (interface{}, error)
}

func (o LazyRefreshCacheOption) Validate() error {
	if o.CacheKey == nil ||
		o.LockerKey == nil ||
		o.RefreshDuration == 0 ||
		o.RawQuery == nil {
		return constant.ErrInvalidArgs
	}

	return nil
}

type LazyRefreshCache struct {
	option *LazyRefreshCacheOption
}

func NewLazyRefreshCache(option *LazyRefreshCacheOption) (*LazyRefreshCache, error) {
	err := option.Validate()
	if err != nil {
		return nil, err
	}

	return &LazyRefreshCache{option: option}, nil
}

func (c LazyRefreshCache) Get(ctx context.Context, request, response interface{}) error {
	if request == nil {
		return nil
	}

	hash := utils.Hash(request)
	log.Debug(ctx, "request hash", log.String("hash", hash), log.Any("request", request))

	// get data and version from cache
	err := c.option.CacheKey.Param(hash).GetObject(ctx, response)
	if err == redis.Nil {
		log.Debug(ctx, "lazy refresh cache miss",
			log.String("hash", hash),
			log.Any("requst", request))

		// query cache for the first time, let's refresh it
		err = c.refreshCache(ctx, hash, request)
		if err != nil {
			return err
		}

		err = c.option.CacheKey.Param(hash).GetObject(ctx, response)
	}
	if err != nil {
		return err
	}

	log.Debug(ctx, "lazy refresh cache hit",
		log.String("hash", hash),
		log.Any("requst", request),
		log.Any("response", response))

	// cache refresh trigger
	go func() {
		ctxClone := utils.CloneContextWithTrace(ctx)

		defer func() {
			if err1 := recover(); err1 != nil {
				log.Error(ctxClone, "async refresh cache panic", log.Any("recover error", err1))
			}
		}()

		err := c.refreshCache(ctxClone, hash, request)
		if err != nil {
			log.Warn(ctxClone, "async refresh cache failed", log.Err(err), log.String("hash", hash))
		}
	}()

	return nil
}

func (c LazyRefreshCache) refreshCache(ctx context.Context, hash string, request interface{}) error {
	// get locker before refresh cache
	return c.option.CacheKey.Param(hash).GetLocker(ctx, c.option.RefreshDuration, func(ctx context.Context) error {
		response, err := c.option.RawQuery(ctx, request)
		if err != nil {
			return err
		}

		return c.option.CacheKey.Param(hash).SetObject(ctx, response, 0)
	})
}
