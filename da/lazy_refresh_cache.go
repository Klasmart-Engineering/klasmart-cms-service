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

type LazyRefreshCache struct {
	name            string
	cache           *ro.StringParameterKey
	locker          *ro.StringParameterKey
	dataVersion     *ro.StringParameterKey
	refreshDuration time.Duration
	refreshQuery    func(ctx context.Context, request interface{}) (interface{}, error)
}

func NewLazyRefreshCache(name string, duration time.Duration, query func(ctx context.Context, request interface{}) (interface{}, error)) (*LazyRefreshCache, error) {
	if name == "" {
		return nil, constant.ErrInvalidArgs
	}

	cache := &LazyRefreshCache{
		name:            name,
		cache:           RedisKeyLazyRefreshCache,
		locker:          RedisKeyLazyRefreshCacheLocker,
		dataVersion:     RedisKeyLazyRefreshCacheDataVersion,
		refreshDuration: duration,
		refreshQuery:    query,
	}

	return cache, nil
}

func (c LazyRefreshCache) Get(ctx context.Context, request, response interface{}) error {
	if request == nil {
		return nil
	}

	hash := utils.Hash(request)
	log.Debug(ctx, "request hash", log.String("hash", hash), log.Any("request", request))

	// get data and version from cache
	data := &lazyRefreshCacheData{Data: response}
	err := c.cache.Param(hash).GetObject(ctx, data)
	if err == redis.Nil {
		log.Debug(ctx, "lazy refresh cache miss",
			log.String("hash", hash),
			log.Any("requst", request))

		// query cache for the first time, let's refresh it
		err = c.refreshCache(ctx, hash, request)
		if err != nil {
			return err
		}

		err = c.cache.Param(hash).GetObject(ctx, data)
	}
	if err != nil {
		return err
	}

	log.Debug(ctx, "lazy refresh cache hit",
		log.String("hash", hash),
		log.Any("requst", request),
		log.Any("data", data))

	// cache refresh trigger
	go func() {
		ctxClone := utils.CloneContextWithTrace(ctx)

		defer func() {
			if err1 := recover(); err1 != nil {
				log.Error(ctxClone, "async refresh cache panic", log.Any("recover error", err1))
			}
		}()

		err := c.asyncRefreshCache(ctxClone, hash, request, data.Version)
		if err != nil {
			log.Warn(ctxClone, "async refresh cache failed", log.Err(err), log.String("hash", hash))
		}
	}()

	return nil
}

func (c LazyRefreshCache) asyncRefreshCache(ctx context.Context, hash string, request interface{}, cacheVersion int64) error {
	dataVersion, err := c.getDataVersion(ctx)
	if err != nil {
		return err
	}

	if cacheVersion == dataVersion {
		log.Debug(ctx, "cache version equal data version", log.Int64("version", cacheVersion))
		return nil
	}

	log.Debug(ctx, "cache version and data version are not equal",
		log.Int64("cacheVersion", cacheVersion),
		log.Int64("dataVersion", dataVersion))

	// fresh cache is version is different
	return c.locker.Param(hash).GetLocker(ctx, c.refreshDuration, func(ctx context.Context) error {
		return c.refreshCache(ctx, hash, request)
	})
}

func (c LazyRefreshCache) refreshCache(ctx context.Context, hash string, request interface{}) error {
	response, err := c.refreshQuery(ctx, request)
	if err != nil {
		return err
	}

	dataVersion, err := c.getDataVersion(ctx)
	if err != nil {
		return err
	}

	data := &lazyRefreshCacheData{
		Data:    response,
		Version: dataVersion,
	}

	// refresh data and version
	return c.cache.Param(hash).SetObject(ctx, data, 0)
}

func (c LazyRefreshCache) getDataVersion(ctx context.Context) (int64, error) {
	version, err := c.dataVersion.Param(c.name).GetInt64(ctx)
	if err == redis.Nil {
		// use current timestamp as init data version
		now := time.Now().UnixNano()
		log.Debug(ctx, "data version not exists, use current timstamp instead",
			log.Any("name", c.name),
			log.Int64("now", now))
		return now, nil
	}

	return version, err
}

func (c LazyRefreshCache) SetDataVersion(ctx context.Context, version int64) error {
	return c.dataVersion.Param(c.name).SetInt64(ctx, version, 0)
}

type lazyRefreshCacheData struct {
	Data    interface{} `json:"data"`
	Version int64       `json:"version"`
}
