package da

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
)

type IAssessmentRedisDA interface {
	GetAssessment(ctx context.Context, id string) (*entity.Assessment, error)
	SetAssessment(ctx context.Context, id string, result *entity.Assessment) error
	CleanAssessment(ctx context.Context, id string) error
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

func (da *assessmentRedisDA) GetAssessment(ctx context.Context, id string) (*entity.Assessment, error) {
	key := da.generateAssessmentCacheKey(id)
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

func (da *assessmentRedisDA) SetAssessment(ctx context.Context, id string, result *entity.Assessment) error {
	key := da.generateAssessmentCacheKey(id)
	value, err := json.Marshal(result)
	if err != nil {
		log.Error(ctx, "set assessment: json marshal failed",
			log.Err(err),
			log.String("id", id),
			log.Any("result", result),
		)
		return err
	}
	if err := da.setCache(ctx, key, string(value), da.getAssessmentCacheExpiration()); err != nil {
		log.Error(ctx, "set assessment: set cache failed",
			log.Err(err),
			log.String("id", id),
			log.Any("result", result),
			log.String("value", string(value)),
		)
		return err
	}
	return nil
}

func (da *assessmentRedisDA) CleanAssessment(ctx context.Context, id string) error {
	key := da.generateAssessmentCacheKey(id)
	if err := da.cleanCache(ctx, key); err != nil {
		log.Error(ctx, "clean assessment: clean cache failed",
			log.Err(err),
			log.String("id", id),
			log.String("key", key),
		)
		return err
	}
	return nil
}

func (da *assessmentRedisDA) generateAssessmentCacheKey(id string) string {
	return strings.Join([]string{RedisKeyPrefixAssessmentItem, id, constant.GitHash}, ":")
}

func (da *baseAssessmentRedisDA) getAssessmentCacheExpiration() time.Duration {
	return config.Get().Assessment.CacheExpiration
}

type baseAssessmentRedisDA struct{}

func (da *baseAssessmentRedisDA) getCache(ctx context.Context, key string) (string, error) {
	if !da.enableCache() {
		return "", nil
	}
	redisResult := ro.MustGetRedis(ctx).Get(ctx, key)
	if err := redisResult.Err(); err != nil {
		log.Error(ctx, "get cache: get failed from redis",
			log.Err(err),
			log.String("key", key),
		)
		return "", err
	}
	return redisResult.Val(), nil
}

func (da *baseAssessmentRedisDA) setCache(ctx context.Context, key string, value string, expiration time.Duration) error {
	if !da.enableCache() {
		return nil
	}
	if err := ro.MustGetRedis(ctx).Set(ctx, key, value, expiration).Err(); err != nil {
		log.Error(ctx, "set cache: set redis value failed",
			log.Err(err),
			log.String("key", key),
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
	if err := ro.MustGetRedis(ctx).Del(ctx, key).Err(); err != nil {
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
