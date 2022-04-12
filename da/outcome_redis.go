package da

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
)

type IOutcomeRedis interface {
	SaveOutcomeCacheList(ctx context.Context, outcomes []*entity.Outcome)
	SaveOutcomeCacheListBySearchCondition(ctx context.Context, op *entity.Operator, condition dbo.Conditions, c *OutcomeListWithKey)
	GetOutcomeCacheByIDList(ctx context.Context, IDs []string) ([]string, []*entity.Outcome)
	GetOutcomeCacheBySearchCondition(ctx context.Context, op *entity.Operator, condition dbo.Conditions) *OutcomeListWithKey

	SaveOutcomeCache(ctx context.Context, outcome *entity.Outcome)
	GetOutcomeCacheByID(ctx context.Context, ID string) *entity.Outcome

	CleanOutcomeCache(ctx context.Context, op *entity.Operator, IDs []string)
	CleanOutcomeConditionCache(ctx context.Context, op *entity.Operator, condition dbo.Conditions)

	SetExpiration(t time.Duration)
}

type OutcomeRedis struct {
	expiration time.Duration
}

type OutcomeListWithKey struct {
	Total       int               `json:"count"`
	OutcomeList []*entity.Outcome `json:"outcome_list"`
}

func (r *OutcomeRedis) outcomeKey(ID string) string {
	return fmt.Sprintf("%v:%v", RedisKeyPrefixOutcomeId, ID)
}

func (r *OutcomeRedis) conditionHash(condition dbo.Conditions) string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%v", condition)))
	md5Hash := fmt.Sprintf("%x", h.Sum(nil))
	return fmt.Sprintf("%v", md5Hash)
}
func (r *OutcomeRedis) outcomeConditionKey(orgID string, condition dbo.Conditions) string {
	md5Hash := r.conditionHash(condition)
	return fmt.Sprintf("%v:%v:%v", RedisKeyPrefixOutcomeCondition, orgID, md5Hash)
}

func (r *OutcomeRedis) outcomeOrgConditionKey(orgID string) string {
	return fmt.Sprintf("%v:%v:*", RedisKeyPrefixOutcomeCondition, orgID)
}

func (r *OutcomeRedis) SaveOutcomeCacheList(ctx context.Context, outcomes []*entity.Outcome) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}
	go func() {
		for i := range outcomes {
			key := r.outcomeKey(outcomes[i].ID)
			outcomeJSON, err := json.Marshal(outcomes[i])
			if err != nil {
				log.Error(ctx, "Can't parse outcome into json", log.Err(err), log.String("cid", outcomes[i].ID))
				continue
			}
			err = ro.MustGetRedis(ctx).SetNX(ctx, key, string(outcomeJSON), r.expiration).Err()
			if err != nil {
				log.Error(ctx, "Can't save outcome into cache", log.Err(err), log.String("key", key), log.String("data", string(outcomeJSON)))
				continue
			}
		}
	}()

}
func (r *OutcomeRedis) SaveOutcomeCache(ctx context.Context, outcome *entity.Outcome) {
	r.SaveOutcomeCacheList(ctx, []*entity.Outcome{
		outcome,
	})
}
func (r *OutcomeRedis) GetOutcomeCacheByID(ctx context.Context, ID string) *entity.Outcome {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}
	_, res := r.GetOutcomeCacheByIDList(ctx, []string{ID})
	if len(res) > 0 {
		return res[0]
	}
	return nil
}

func (r *OutcomeRedis) SaveOutcomeCacheListBySearchCondition(ctx context.Context, op *entity.Operator, condition dbo.Conditions, c *OutcomeListWithKey) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}
	//go func() {
	key := r.outcomeConditionKey(op.OrgID, condition)
	outcomeListJSON, err := json.Marshal(c)
	if err != nil {
		log.Error(ctx, "Can't parse outcome list into json", log.Err(err), log.String("key", key), log.String("data", string(outcomeListJSON)))
		return
	}
	log.Info(ctx, "save outcome into cache", log.String("cache", string(outcomeListJSON)), log.String("key", key), log.Any("condition", condition))
	//err = ro.MustGetRedis(ctx).HSetNX(RedisKeyPrefixOutcomeCondition, r.conditionHash(condition), string(outcomeListJSON)).Err()
	//ro.MustGetRedis(ctx).Expire(RedisKeyPrefixOutcomeCondition, r.expiration)
	err = ro.MustGetRedis(ctx).SetNX(ctx, key, string(outcomeListJSON), r.expiration).Err()
	if err != nil {
		log.Error(ctx, "Can't save outcome list into cache", log.Err(err), log.String("key", key), log.String("data", string(outcomeListJSON)))
	}
	//}()
}

func (r *OutcomeRedis) GetOutcomeCacheByIDList(ctx context.Context, IDs []string) ([]string, []*entity.Outcome) {
	if !config.Get().RedisConfig.OpenCache {
		return IDs, nil
	}
	keys := make([]string, len(IDs))
	for i := range IDs {
		keys[i] = r.outcomeKey(IDs[i])
	}
	res, err := ro.MustGetRedis(ctx).MGet(ctx, keys...).Result()
	if err != nil {
		log.Info(ctx, "Can't get outcome list from cache", log.Err(err), log.Strings("keys", keys), log.Strings("ids", IDs))
		return IDs, nil
	}

	// parse cachedOutcomes
	cachedOutcomes := make([]*entity.Outcome, 0)
	for i := range res {
		resJSON, ok := res[i].(string)
		if !ok {
			log.Error(ctx, "Get invalid data from cache", log.Any("data", res[i]))
			continue
		}
		outcome := new(entity.Outcome)
		err = json.Unmarshal([]byte(resJSON), outcome)
		if err != nil {
			log.Error(ctx, "Can't unmarshal outcome from cache", log.Err(err), log.String("JSON", resJSON))
			continue
		}
		cachedOutcomes = append(cachedOutcomes, outcome)
	}

	// mark id which need to be deleted
	deletedMarks := make([]bool, len(IDs))
	for i := range IDs {
		for j := range cachedOutcomes {
			if IDs[i] == cachedOutcomes[j].ID {
				deletedMarks[i] = true
			}
		}
	}

	//获取剩余ids
	restIds := make([]string, 0)
	for i := range IDs {
		if !deletedMarks[i] {
			restIds = append(restIds, IDs[i])
		}
	}

	return restIds, cachedOutcomes
}

func (r *OutcomeRedis) GetOutcomeCacheBySearchCondition(ctx context.Context, op *entity.Operator, condition dbo.Conditions) *OutcomeListWithKey {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}

	key := r.outcomeConditionKey(op.OrgID, condition)

	res, err := ro.MustGetRedis(ctx).Get(ctx, key).Result()
	//res, err := ro.MustGetRedis(ctx).HGet(RedisKeyPrefixOutcomeCondition, r.conditionHash(condition)).Result()

	if err != nil {
		log.Info(ctx, "Can't get outcome condition from cache", log.Err(err), log.String("key", key), log.Any("condition", condition))
		return nil
	}
	log.Info(ctx, "search outcome from cache", log.String("key", key), log.String("cache", res), log.Any("condition", condition))

	outcomeLists := new(OutcomeListWithKey)
	err = json.Unmarshal([]byte(res), outcomeLists)
	if err != nil {
		log.Error(ctx, "Can't unmarshal outcome condition from cache", log.Err(err), log.String("key", key), log.String("json", res))
		err = ro.MustGetRedis(ctx).Del(ctx, key).Err()
		if err != nil {
			log.Error(ctx, "Can't delete outcome from cache", log.Err(err), log.String("key", key), log.String("json", res))
		}
		return nil
	}
	log.Info(ctx, "search outcome from cache", log.String("key", key))

	return outcomeLists
}

func (r *OutcomeRedis) CleanOutcomeCache(ctx context.Context, op *entity.Operator, IDs []string) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}

	if len(IDs) < 1 {
		return
	}
	//delete related id
	keys := make([]string, 0)
	for i := range IDs {
		keys = append(keys, r.outcomeKey(IDs[i]))
	}

	//delete condition cache
	//conditionKeys := ro.MustGetRedis(ctx).Keys(RedisKeyPrefixOutcomeCondition + ":*").Val()
	//keys = append(keys, conditionKeys...)
	keys = append(keys, r.outcomeOrgConditionKey(op.OrgID))
	// go func() {
	err := ro.MustGetRedis(ctx).Del(ctx, keys...).Err()
	if err != nil {
		log.Error(ctx, "Can't clean outcome from cache", log.Err(err), log.Strings("keys", keys))
	}
	//time.Sleep(time.Second)
	//ro.MustGetRedis(ctx).Del(keys...)
	//if err != nil{
	//	log.Error(ctx, "Can't clean outcome again from cache", log.Err(err), log.Strings("keys", keys))
	//}
	// }()
}

func (r *OutcomeRedis) CleanOutcomeConditionCache(ctx context.Context, op *entity.Operator, condition dbo.Conditions) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}

	var keys []string
	if condition != nil {
		key := r.outcomeConditionKey(op.OrgID, condition)
		keys = append(keys, key)
	} else {
		var err error
		keys, err = ro.MustGetRedis(ctx).Keys(ctx, r.outcomeOrgConditionKey(op.OrgID)).Result()
		if err != nil {
			log.Error(ctx, "CleanOutcomeConditionCache: keys failed", log.Err(err), log.Strings("keys", keys))
			return
		}
	}

	if len(keys) == 0 {
		log.Debug(ctx, "CleanOutcomeConditionCache: empty", log.Any("condition", condition))
		return
	}

	err := ro.MustGetRedis(ctx).Del(ctx, keys...).Err()
	if err != nil {
		log.Error(ctx, "CleanOutcomeConditionCache: del failed", log.Err(err), log.Strings("keys", keys))
	}
}

func (r *OutcomeRedis) SetExpiration(t time.Duration) {
	r.expiration = t
}

var (
	_redisOutcomeCache     *OutcomeRedis
	_redisOutcomeCacheOnce sync.Once
)

func GetOutcomeRedis() IOutcomeRedis {
	_redisOutcomeCacheOnce.Do(func() {
		_redisOutcomeCache = &OutcomeRedis{expiration: time.Minute * 2}
	})
	return _redisOutcomeCache
}
