package da

import (
	"context"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"strings"
	"sync"
	"time"
)

type IAssessmentRedisDA interface {
	Item(ctx context.Context, id string) (*entity.Assessment, error)
	CacheItem(ctx context.Context, id string, result *entity.Assessment) error
	CleanItem(ctx context.Context, id string) error
}

var (
	assessmentRedisDAInstance     IAssessmentRedisDA
	assessmentRedisDAInstanceOnce = sync.Once{}
)

func GetAssessmentRedisDA() IAssessmentRedisDA {
	assessmentRedisDAInstanceOnce.Do(func() {
		assessmentRedisDAInstance = &assessmentRedisDA{}
	})
	return assessmentRedisDAInstance
}

type assessmentRedisDA struct {
	baseAssessmentRedisDA
}

func (da *assessmentRedisDA) Item(ctx context.Context, id string) (*entity.Assessment, error) {
	key := da.itemCacheKey(id)
	value, err := da.getCache(ctx, key)
	if err != nil {
		log.Error(ctx, "get assessment: get cache failed",
			log.Err(err),
			log.String("id", id),
			log.String("key", key),
		)
		return nil, err
	}
	result := &entity.Assessment{}
	if err := json.Unmarshal([]byte(value), result); err != nil {
		log.Error(ctx, "get assessment: json unmarshal failed",
			log.Err(err),
			log.String("id", id),
			log.String("key", key),
			log.String("value", value),
		)
		return nil, err
	}
	return result, nil
}

func (da *assessmentRedisDA) CacheItem(ctx context.Context, id string, result *entity.Assessment) error {
	key := da.itemCacheKey(id)
	value, err := json.Marshal(result)
	if err != nil {
		log.Error(ctx, "cache item: json marshal failed",
			log.Err(err),
			log.String("id", id),
			log.Any("result", result),
		)
		return err
	}
	if err := da.setCache(ctx, key, string(value)); err != nil {
		log.Error(ctx, "cache item: set cache failed",
			log.Err(err),
			log.String("id", id),
			log.Any("result", result),
			log.String("value", string(value)),
		)
		return err
	}
	return nil
}

func (da *assessmentRedisDA) CleanItem(ctx context.Context, id string) error {
	key := da.itemCacheKey(id)
	if err := da.cleanCache(ctx, key); err != nil {
		log.Error(ctx, "clean cache item: clean cache failed",
			log.Err(err),
			log.String("id", id),
			log.String("key", key),
		)
		return err
	}
	return nil
}

func (da *assessmentRedisDA) itemCacheKey(id string) string {
	return strings.Join([]string{RedisKeyPrefixAssessmentItem, id}, ":")
}

type baseAssessmentRedisDA struct{}

func (da *baseAssessmentRedisDA) getCache(ctx context.Context, key string) (string, error) {
	if !da.enableCache() {
		return "", nil
	}
	redisResult := ro.MustGetRedis(ctx).Get(key)
	if err := redisResult.Err(); err != nil {
		log.Error(ctx, "get cache: get failed from redis",
			log.Err(err),
			log.String("key", key),
		)
		return "", err
	}
	return redisResult.Val(), nil
}

func (da *baseAssessmentRedisDA) getHashCache(ctx context.Context, key string, field string) (string, error) {
	if !da.enableCache() {
		return "", nil
	}
	redisResult := ro.MustGetRedis(ctx).HGet(key, field)
	if err := redisResult.Err(); err != nil {
		log.Error(ctx, "get hash cache: get failed from redis",
			log.Err(err),
			log.String("key", key),
			log.String("field", field),
		)
		return "", err
	}
	return redisResult.Val(), nil
}

func (da *baseAssessmentRedisDA) setCache(ctx context.Context, key string, value string) error {
	if !da.enableCache() {
		return nil
	}
	if err := ro.MustGetRedis(ctx).Set(key, value, da.cacheExpiration()).Err(); err != nil {
		log.Error(ctx, "set cache: set redis value failed",
			log.Err(err),
			log.String("key", key),
			log.Any("value", value),
		)
		return err
	}
	return nil
}

func (da *baseAssessmentRedisDA) setHashCache(ctx context.Context, key string, field string, value string) error {
	if !da.enableCache() {
		return nil
	}
	ro.MustGetRedis(ctx).Expire(key, da.cacheExpiration())
	if err := ro.MustGetRedis(ctx).HSet(key, field, value).Err(); err != nil {
		log.Error(ctx, "set cache with key: set redis value failed",
			log.Err(err),
			log.String("key", key),
			log.String("field", field),
			log.Any("value", value),
		)
		return err
	}
	return nil
}

func (da *baseAssessmentRedisDA) cleanCache(ctx context.Context, key string) error {
	if !da.enableCache() {
		return nil
	}
	if err := ro.MustGetRedis(ctx).Del(key).Err(); err != nil {
		log.Error(ctx, "clean cache: del failed by redis",
			log.Err(err),
			log.String("key", key),
		)
		return err
	}
	return nil
}

func (da *baseAssessmentRedisDA) enableCache() bool {
	return config.Get().RedisConfig.OpenCache
}

func (da *baseAssessmentRedisDA) cacheExpiration() time.Duration {
	return config.Get().Assessment.CacheExpiration
}
