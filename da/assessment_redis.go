package da

import (
	"context"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"strconv"
	"strings"
	"sync"
	"time"
)

type IAssessmentRedisDA interface {
	GetAssessment(ctx context.Context, id string) (*entity.Assessment, error)
	SetAssessment(ctx context.Context, id string, result *entity.Assessment) error
	CleanAssessment(ctx context.Context, id string) error

	GetQueryLearningSummaryTimeFilterResult(ctx context.Context, args *entity.QueryLearningSummaryTimeFilterArgs) ([]*entity.LearningSummaryFilterYear, error)
	SetQueryLearningSummaryTimeFilterResult(ctx context.Context, args *entity.QueryLearningSummaryTimeFilterArgs, result []*entity.LearningSummaryFilterYear) error
	CleanQueryLearningSummaryTimeFilterResult(ctx context.Context, args *entity.QueryLearningSummaryTimeFilterArgs) error
}

var (
	assessmentRedisDAInstance     IAssessmentRedisDA
	assessmentRedisDAInstanceOnce = sync.Once{}
)

func GetAssessmentRedisDA() IAssessmentRedisDA {
	assessmentRedisDAInstanceOnce.Do(func() {
		assessmentRedisDAInstance = &assessmentRedisDA{
			nonce: utils.NewID(),
		}
	})
	return assessmentRedisDAInstance
}

type assessmentRedisDA struct {
	baseAssessmentRedisDA
	nonce string
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
	return strings.Join([]string{RedisKeyPrefixAssessmentItem, id, da.nonce}, ":")
}

func (da *baseAssessmentRedisDA) getAssessmentCacheExpiration() time.Duration {
	return config.Get().Assessment.CacheExpiration
}

func (da *assessmentRedisDA) GetQueryLearningSummaryTimeFilterResult(ctx context.Context, args *entity.QueryLearningSummaryTimeFilterArgs) ([]*entity.LearningSummaryFilterYear, error) {
	key := da.generateQueryLearningSummaryTimeFilterResultCacheKey(args)
	value, err := da.getCache(ctx, key)
	if err != nil {
		log.Error(ctx, "get query learning summary time filter result: get cache failed",
			log.Err(err),
			log.Any("args", args),
			log.String("key", key),
		)
		return nil, err
	}
	var result []*entity.LearningSummaryFilterYear
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		log.Error(ctx, "get query learning summary time filter result: json unmarshal failed",
			log.Err(err),
			log.Any("args", args),
			log.String("key", key),
			log.String("value", value),
		)
		return nil, err
	}
	return result, nil
}

func (da *assessmentRedisDA) SetQueryLearningSummaryTimeFilterResult(ctx context.Context, args *entity.QueryLearningSummaryTimeFilterArgs, result []*entity.LearningSummaryFilterYear) error {
	key := da.generateQueryLearningSummaryTimeFilterResultCacheKey(args)
	value, err := json.Marshal(result)
	if err != nil {
		log.Error(ctx, "set query learning summary time filter result: json marshal failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("result", result),
		)
		return err
	}
	if err := da.setCache(ctx, key, string(value), constant.AssessmentQueryLearningSummaryTimeFilterCacheExpiration); err != nil {
		log.Error(ctx, "set query learning summary time filter result: set cache failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("result", result),
			log.String("value", string(value)),
		)
		return err
	}
	return nil
}

func (da *assessmentRedisDA) CleanQueryLearningSummaryTimeFilterResult(ctx context.Context, args *entity.QueryLearningSummaryTimeFilterArgs) error {
	key := da.generateQueryLearningSummaryTimeFilterResultCacheKey(args)
	if err := da.cleanCache(ctx, key); err != nil {
		log.Error(ctx, "clean query learning summary time filter result: clean cache failed",
			log.Err(err),
			log.Any("args", args),
			log.String("key", key),
		)
		return err
	}
	return nil
}

func (da *assessmentRedisDA) generateQueryLearningSummaryTimeFilterResultCacheKey(args *entity.QueryLearningSummaryTimeFilterArgs) string {
	return strings.Join([]string{
		RedisKeyPrefixAssessmentQueryLearningSummaryTimeFilter,
		args.OrgID,
		string(args.SummaryType),
		strconv.Itoa(args.TimeOffset),
		strings.Join(args.SchoolIDs, "-"),
		args.TeacherID,
		args.StudentID,
		da.nonce,
	}, ":")
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

func (da *baseAssessmentRedisDA) setCache(ctx context.Context, key string, value string, expiration time.Duration) error {
	if !da.enableCache() {
		return nil
	}
	if err := ro.MustGetRedis(ctx).Set(key, value, expiration).Err(); err != nil {
		log.Error(ctx, "set cache: set redis value failed",
			log.Err(err),
			log.String("key", key),
			log.Any("value", value),
		)
		return err
	}
	return nil
}

func (da *baseAssessmentRedisDA) setHashCache(ctx context.Context, key string, field string, value string, expiration time.Duration) error {
	if !da.enableCache() {
		return nil
	}
	ro.MustGetRedis(ctx).Expire(key, expiration)
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
