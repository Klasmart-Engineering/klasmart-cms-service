package da

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"strings"
	"sync"
	"time"
)

type IAssessmentRedisDA interface {
	Detail(ctx context.Context, id string) (*entity.AssessmentDetailView, error)
	CacheDetail(ctx context.Context, id string, result *entity.AssessmentDetailView) error
	CleanDetail(ctx context.Context, id string) error
	List(ctx context.Context, cmd entity.ListAssessmentsCommand) (*entity.ListAssessmentsResult, error)
	CacheList(ctx context.Context, cmd entity.ListAssessmentsCommand, result *entity.ListAssessmentsResult) error
	CleanList(ctx context.Context) error
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

func (da *assessmentRedisDA) Detail(ctx context.Context, id string) (*entity.AssessmentDetailView, error) {
	result := &entity.AssessmentDetailView{}
	key := da.detailCacheKey(id)
	if err := da.getAndUnmarshalJSON(ctx, key, result); err != nil {
		log.Error(ctx, "get assessment detail cache: get failed from redis",
			log.Err(err),
			log.String("id", id),
		)
		return nil, err
	}
	return result, nil
}

func (da *assessmentRedisDA) CacheDetail(ctx context.Context, id string, result *entity.AssessmentDetailView) error {
	key := da.detailCacheKey(id)
	bs, err := json.Marshal(result)
	if err != nil {
		log.Error(ctx, "cache assessment detail: json marshal failed",
			log.Err(err),
			log.String("id", id),
			log.Any("result", result),
		)
		return err
	}
	value := string(bs)
	if err := da.cache(ctx, key, value); err != nil {
		log.Error(ctx, "cache assessment detail: cache failed",
			log.Err(err),
			log.String("id", id),
			log.Any("result", result),
			log.String("key", key),
			log.String("value", value),
		)
		return err
	}
	return nil
}

func (da *assessmentRedisDA) CleanDetail(ctx context.Context, id string) error {
	key := da.detailCacheKey(id)
	if err := da.clean(ctx, key); err != nil {
		log.Error(ctx, "clean assessment cache: del failed by redis",
			log.Err(err),
			log.String("id", id),
		)
		return err
	}
	return nil
}

func (da *assessmentRedisDA) detailCacheKey(id string) string {
	return strings.Join([]string{RedisKeyPrefixAssessmentDetail, id}, ":")
}

func (da *assessmentRedisDA) List(ctx context.Context, cmd entity.ListAssessmentsCommand) (*entity.ListAssessmentsResult, error) {
	result := &entity.ListAssessmentsResult{}
	field, err := da.listCacheField(cmd)
	if err != nil {
		log.Error(ctx, "get assessment list cache: get list cache field failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
	}
	if err := da.getHashAndUnmarshalJSON(ctx, RedisKeyPrefixAssessmentList, field, result); err != nil {
		log.Error(ctx, "get assessment list cache: get or unmarshal redis value failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.String("field", field),
		)
		return nil, err
	}
	return result, nil
}

func (da *assessmentRedisDA) CacheList(ctx context.Context, cmd entity.ListAssessmentsCommand, result *entity.ListAssessmentsResult) error {
	field, err := da.listCacheField(cmd)
	if err != nil {
		log.Error(ctx, "cache assessment list: get list cache field failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("result", result),
		)
	}
	bs, err := json.Marshal(result)
	if err != nil {
		log.Error(ctx, "cache assessment list: json marshal failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("result", result),
			log.String("field", field),
		)
		return err
	}
	value := string(bs)
	log.Debug(ctx, "cache list",
		log.String("key", RedisKeyPrefixAssessmentList),
		log.String("field", field),
		log.String("value", value),
	)
	if err := da.cacheHash(ctx, RedisKeyPrefixAssessmentList, field, value); err != nil {
		log.Error(ctx, "cache assessment list: cache hash failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("result", result),
			log.String("field", field),
			log.String("value", value),
		)
		return err
	}
	return nil
}

func (da *assessmentRedisDA) CleanList(ctx context.Context) error {
	key := RedisKeyPrefixAssessmentList
	if err := da.clean(ctx, key); err != nil {
		log.Error(ctx, "clean assessment list: clean failed",
			log.Err(err),
			log.String("key", key),
		)
		return err
	}
	return nil
}

func (da *assessmentRedisDA) listCacheField(cmd entity.ListAssessmentsCommand) (string, error) {
	bs, err := json.Marshal(cmd)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	hash.Write(bs)
	result := hash.Sum(nil)
	return fmt.Sprintf("%x", result), nil
}

type baseAssessmentRedisDA struct{}

func (da *baseAssessmentRedisDA) get(ctx context.Context, key string) (string, error) {
	if !da.enable() {
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

func (da *baseAssessmentRedisDA) getHash(ctx context.Context, key string, field string) (string, error) {
	if !da.enable() {
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

func (da *baseAssessmentRedisDA) getAndUnmarshalJSON(ctx context.Context, key string, result interface{}) error {
	value, err := da.get(ctx, key)
	if err != nil {
		log.Error(ctx, "get cache and unmarshal json: get failed",
			log.Err(err),
			log.String("key", key),
		)
		return err
	}
	if err := json.Unmarshal([]byte(value), result); err != nil {
		log.Error(ctx, "get cache and unmarshal json: json unmarshal value failed",
			log.Err(err),
			log.String("key", key),
			log.Any("value", value),
		)
		return err
	}
	return nil
}

func (da *baseAssessmentRedisDA) getHashAndUnmarshalJSON(ctx context.Context, key string, field string, result interface{}) error {
	value, err := da.getHash(ctx, key, field)
	if err != nil {
		log.Error(ctx, "get hash cache and unmarshal json: get failed",
			log.Err(err),
			log.String("key", key),
			log.String("field", field),
		)
		return err
	}
	if err := json.Unmarshal([]byte(value), result); err != nil {
		log.Error(ctx, "get hash cache and unmarshal json: json unmarshal value failed",
			log.Err(err),
			log.String("key", key),
			log.String("field", field),
			log.Any("value", value),
		)
		return err
	}
	return nil
}

func (da *baseAssessmentRedisDA) cache(ctx context.Context, key string, value string) error {
	if !da.enable() {
		return nil
	}
	if err := ro.MustGetRedis(ctx).Set(key, value, da.expiration()).Err(); err != nil {
		log.Error(ctx, "set cache: set redis value failed",
			log.Err(err),
			log.String("key", key),
			log.Any("value", value),
		)
		return err
	}
	return nil
}

func (da *baseAssessmentRedisDA) cacheHash(ctx context.Context, key string, field string, value string) error {
	if !da.enable() {
		return nil
	}
	ro.MustGetRedis(ctx).Expire(key, da.expiration())
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

func (da *baseAssessmentRedisDA) clean(ctx context.Context, key string) error {
	if !da.enable() {
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

func (da *baseAssessmentRedisDA) enable() bool {
	return config.Get().RedisConfig.OpenCache
}

func (da *baseAssessmentRedisDA) expiration() time.Duration {
	return config.Get().Assessment.CacheExpiration
}
