package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"sync"
	"time"
)

type IContentCache interface {
	SaveContentCacheList(ctx context.Context, contents []*entity.ContentInfoWithDetails)
	SaveContentCacheListBySearchCondition(ctx context.Context, condition dbo.Conditions, c *ContentListWithKey)
	GetContentCacheByIdList(ctx context.Context, ids []string) ([]string, []*entity.ContentInfoWithDetails)
	GetContentCacheBySearchCondition(ctx context.Context, condition dbo.Conditions) *ContentListWithKey

	SaveContentCache(ctx context.Context, content *entity.ContentInfoWithDetails)
	GetContentCacheById(ctx context.Context, id string) *entity.ContentInfoWithDetails

	CleanContentCache(ctx context.Context, ids []string) error

	SetExpiration(t time.Duration)
}

type RedisContentCache struct {
	expiration time.Duration
}

type ContentListWithKey struct {
	Count         int                           `json:"count"`
	ContentList []*entity.ContentInfoWithDetails `json:"content_list"`
}

func (r *RedisContentCache) contentKey(id string) string {
	return fmt.Sprintf("kidsloop2.content.id.%v", id)
}
func (r *RedisContentCache) contentConditionKey(condition dbo.Conditions) string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%v", condition)))
	md5Hash := fmt.Sprintf("%x", h.Sum(nil))

	return fmt.Sprintf("kidsloop2.content.condition.%v", md5Hash)
}

func (r *RedisContentCache) SaveContentCacheList(ctx context.Context, contents []*entity.ContentInfoWithDetails) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}
	go func() {
		for i := range contents {
			key := r.contentKey(contents[i].ID)
			contentJSON, err := json.Marshal(contents[i])
			if err != nil {
				log.Error(ctx, "Can't parse content into json", log.Err(err))
				continue
			}
			err = ro.MustGetRedis(ctx).SetNX(key, string(contentJSON), r.expiration).Err()
			if err != nil {
				log.Error(ctx, "Can't save content into cache", log.Err(err))
				continue
			}
		}
	}()

}
func (r *RedisContentCache) SaveContentCache(ctx context.Context, content *entity.ContentInfoWithDetails) {
	r.SaveContentCacheList(ctx, []*entity.ContentInfoWithDetails{
		content,
	})
}
func (r *RedisContentCache) GetContentCacheById(ctx context.Context, id string) *entity.ContentInfoWithDetails {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}
	_, res := r.GetContentCacheByIdList(ctx, []string{id})
	if len(res) > 0 {
		return res[0]
	}
	return nil
}

func (r *RedisContentCache) SaveContentCacheListBySearchCondition(ctx context.Context, condition dbo.Conditions, c *ContentListWithKey) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}
	go func() {
		key := r.contentConditionKey(condition)
		contentListJSON, err := json.Marshal(c)
		if err != nil {
			log.Error(ctx, "Can't parse content list into json", log.Err(err))
			return
		}
		err = ro.MustGetRedis(ctx).SetNX(key, string(contentListJSON), r.expiration).Err()
		if err != nil {
			log.Error(ctx, "Can't save content list into cache", log.Err(err))
		}
	}()
}

func (r *RedisContentCache) GetContentCacheByIdList(ctx context.Context, ids []string) ([]string, []*entity.ContentInfoWithDetails) {
	if !config.Get().RedisConfig.OpenCache {
		return ids, nil
	}
	keys := make([]string, len(ids))
	for i := range ids {
		keys[i] = r.contentKey(ids[i])
	}
	res, err := ro.MustGetRedis(ctx).MGet(keys...).Result()
	if err != nil {
		log.Error(ctx, "Can't get content list from cache", log.Err(err))
		return ids, nil
	}

	//解析cachedContents
	cachedContents := make([]*entity.ContentInfoWithDetails, 0)
	for i := range res {
		resJSON, ok := res[i].(string)
		if !ok {
			log.Error(ctx, "Get invalid data from cache", log.Any("data", res[i]))
			continue
		}
		content := new(entity.ContentInfoWithDetails)
		err = json.Unmarshal([]byte(resJSON), content)
		if err != nil {
			log.Error(ctx, "Can't unmarshal content from cache", log.Err(err), log.String("JSON", resJSON))
			continue
		}
		cachedContents = append(cachedContents, content)
	}

	//标记需要被删掉的id
	deletedMarks := make([]bool, len(ids))
	for i := range ids {
		for j := range cachedContents {
			if ids[i] == cachedContents[j].ID {
				deletedMarks[i] = true
			}
		}
	}

	//获取剩余ids
	restIds := make([]string, 0)
	for i := range ids {
		if !deletedMarks[i] {
			restIds = append(restIds, ids[i])
		}
	}

	return restIds, cachedContents
}

func (r *RedisContentCache) GetContentCacheBySearchCondition(ctx context.Context, condition dbo.Conditions) *ContentListWithKey {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}
	key := r.contentConditionKey(condition)
	res, err := ro.MustGetRedis(ctx).Get(key).Result()
	if err != nil {
		log.Error(ctx, "Can't get content condition from cache", log.Err(err))
		return nil
	}
	contentLists := new(ContentListWithKey)
	err = json.Unmarshal([]byte(res), contentLists)
	if err != nil {
		log.Error(ctx, "Can't unmarshal content condition from cache", log.Err(err), log.String("json", res))
		ro.MustGetRedis(ctx).Del(key)
		return nil
	}

	return contentLists
}

func (r *RedisContentCache) CleanContentCache(ctx context.Context, ids []string) error {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}

	if len(ids) < 1 {
		return nil
	}
	//删除对应的id
	keys := make([]string, 0)
	for i := range ids {
		keys = append(keys, r.contentKey(ids[i]))
	}

	//删除所有condition cache
	conditionKeys := ro.MustGetRedis(ctx).Keys("kidsloop2.content.condition.*").Val()
	keys = append(keys, conditionKeys...)

	return ro.MustGetRedis(ctx).Del(keys...).Err()
}

func (r *RedisContentCache) SetExpiration(t time.Duration) {
	r.expiration = t
}

var (
	_redisContentCache     *RedisContentCache
	_redisContentCacheOnce sync.Once
)

func GetRedisContentCache() IContentCache {
	_redisContentCacheOnce.Do(func() {
		_redisContentCache = &RedisContentCache{expiration: time.Minute * 2}
	})
	return _redisContentCache
}
