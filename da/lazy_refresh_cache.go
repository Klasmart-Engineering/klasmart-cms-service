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

	return &LazyRefreshCache{
		name:            name,
		cache:           RedisKeyLazyRefreshCache,
		locker:          RedisKeyLazyRefreshCacheLocker,
		dataVersion:     RedisKeyLazyRefreshCacheDataVersion,
		refreshDuration: duration,
		refreshQuery:    query,
	}, nil
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
	if err != nil && err != redis.Nil {
		return err
	}

	log.Debug(ctx, "get cached data successfully",
		log.String("hash", hash),
		log.Any("requst", request),
		log.Any("data", data))

	// cache refresh trigger
	go func() {
		err := c.checkVersion(ctx, hash, request, data.Version)
		if err != nil {
			log.Warn(ctx, "check version failed", log.Err(err), log.String("hash", hash))
		}
	}()

	return nil
}

func (c LazyRefreshCache) checkVersion(ctx context.Context, hash string, request interface{}, cacheVersion int64) error {
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
		response, err := c.refreshQuery(ctx, request)
		if err != nil {
			return err
		}

		data := &lazyRefreshCacheData{
			Data:    response,
			Version: dataVersion,
		}

		// refresh data and version
		err = c.cache.Param(hash).SetObject(ctx, data, 0)
		if err != nil {
			return err
		}

		return nil
	})
}

func (c LazyRefreshCache) getDataVersion(ctx context.Context) (int64, error) {
	version, err := c.dataVersion.Param(c.name).GetInt64(ctx)
	if err == redis.Nil {
		log.Debug(ctx, "data version not exists", log.Any("name", c.name))
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return version, nil
}

func (c LazyRefreshCache) SetDataVersion(ctx context.Context, version int64) error {
	return c.dataVersion.Param(c.name).SetInt64(ctx, version, 0)
}

type lazyRefreshCacheData struct {
	Data    interface{} `json:"data"`
	Version int64       `json:"version"`
}
