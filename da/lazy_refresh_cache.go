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
	RedisKeyPrefix  string
	Expiration      time.Duration
	RefreshDuration time.Duration
	RawQuery        func(ctx context.Context, request interface{}) (interface{}, error)
}

func (o LazyRefreshCacheOption) Validate() error {
	if o.RedisKeyPrefix == "" ||
		(o.Expiration > 0 && o.Expiration < o.RefreshDuration) ||
		o.RefreshDuration == 0 ||
		o.RawQuery == nil {
		return constant.ErrInvalidArgs
	}

	return nil
}

type LazyRefreshCache struct {
	option    *LazyRefreshCacheOption
	cacheKey  *ro.StringParameterKey
	lockerKey *ro.StringParameterKey
}

func NewLazyRefreshCache(option *LazyRefreshCacheOption) (*LazyRefreshCache, error) {
	err := option.Validate()
	if err != nil {
		return nil, err
	}

	return &LazyRefreshCache{
		option:    option,
		cacheKey:  ro.NewStringParameterKey(option.RedisKeyPrefix + ":cache:%s"),
		lockerKey: ro.NewStringParameterKey(option.RedisKeyPrefix + ":locker:%s"),
	}, nil
}

func (c LazyRefreshCache) Get(ctx context.Context, request, response interface{}) error {
	if request == nil {
		return nil
	}

	hash, err := utils.Hash(request)
	if err != nil {
		log.Warn(ctx, "caculate request hash failed", log.Err(err), log.Any("request", request))
		return err
	}
	log.Debug(ctx, "request hash", log.String("hash", hash), log.Any("request", request))

	// get data and version from cache
	err = c.cacheKey.Param(hash).GetObject(ctx, response)
	if err == redis.Nil {
		log.Debug(ctx, "lazy refresh cache miss",
			log.String("hash", hash),
			log.Any("requst", request))

		// query cache for the first time, let's refresh it
		err = c.refreshCache(ctx, hash, request)
		if err != nil {
			return err
		}

		err = c.cacheKey.Param(hash).GetObject(ctx, response)
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
	return c.lockerKey.Param(hash).GetLocker(ctx, c.option.RefreshDuration, func(ctx context.Context) error {
		response, err := c.option.RawQuery(ctx, request)
		if err != nil {
			return err
		}

		return c.cacheKey.Param(hash).SetObject(ctx, response, c.option.Expiration)
	})
}
