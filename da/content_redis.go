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

type IContentRedis interface {
	SaveContentCacheList(ctx context.Context, contents []*entity.ContentInfoWithDetails)
	SaveContentCacheListBySearchCondition(ctx context.Context, condition dbo.Conditions, c *ContentListWithKey)
	GetContentCacheByIdList(ctx context.Context, ids []string) ([]string, []*entity.ContentInfoWithDetails)
	GetContentCacheBySearchCondition(ctx context.Context, condition dbo.Conditions) *ContentListWithKey

	SaveContentCache(ctx context.Context, content *entity.ContentInfoWithDetails)
	GetContentCacheById(ctx context.Context, id string) *entity.ContentInfoWithDetails

	CleanContentCache(ctx context.Context, ids []string)

	SetExpiration(t time.Duration)
}

type ContentRedis struct {
	expiration time.Duration
}

type ContentListWithKey struct {
	Count       int                              `json:"count"`
	ContentList []*entity.ContentInfoWithDetails `json:"content_list"`
}

func (r *ContentRedis) contentKey(id string) string {
	return fmt.Sprintf("%v:%v", RedisKeyPrefixContentId, id)
}

func (r *ContentRedis) conditionHash(condition dbo.Conditions) string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%v", condition)))
	md5Hash := fmt.Sprintf("%x", h.Sum(nil))
	return fmt.Sprintf("%v", md5Hash)
}
func (r *ContentRedis) contentConditionKey(condition dbo.Conditions) string {
	md5Hash := r.conditionHash(condition)
	return fmt.Sprintf("%v:%v", RedisKeyPrefixContentCondition, md5Hash)
}

func (r *ContentRedis) SaveContentCacheList(ctx context.Context, contents []*entity.ContentInfoWithDetails) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}
	go func() {
		for i := range contents {
			key := r.contentKey(contents[i].ID)
			contentJSON, err := json.Marshal(contents[i])
			if err != nil {
				log.Error(ctx, "Can't parse content into json", log.Err(err), log.String("cid", contents[i].ID))
				continue
			}
			err = ro.MustGetRedis(ctx).SetNX(key, string(contentJSON), r.expiration).Err()
			if err != nil {
				log.Error(ctx, "Can't save content into cache", log.Err(err), log.String("key", key), log.String("data", string(contentJSON)))
				continue
			}
		}
	}()

}
func (r *ContentRedis) SaveContentCache(ctx context.Context, content *entity.ContentInfoWithDetails) {
	r.SaveContentCacheList(ctx, []*entity.ContentInfoWithDetails{
		content,
	})
}
func (r *ContentRedis) GetContentCacheById(ctx context.Context, id string) *entity.ContentInfoWithDetails {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}
	_, res := r.GetContentCacheByIdList(ctx, []string{id})
	if len(res) > 0 {
		return res[0]
	}
	return nil
}

func (r *ContentRedis) SaveContentCacheListBySearchCondition(ctx context.Context, condition dbo.Conditions, c *ContentListWithKey) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}
	go func() {
		key := r.contentConditionKey(condition)
		contentListJSON, err := json.Marshal(c)
		if err != nil {
			log.Error(ctx, "Can't parse content list into json", log.Err(err), log.String("key", key), log.String("data", string(contentListJSON)))
			return
		}
		ro.MustGetRedis(ctx).Expire(RedisKeyPrefixContentCondition, r.expiration)
		err = ro.MustGetRedis(ctx).HSetNX(RedisKeyPrefixContentCondition, r.conditionHash(condition), string(contentListJSON)).Err()
		//err = ro.MustGetRedis(ctx).SetNX(key, string(contentListJSON), r.expiration).Err()
		if err != nil {
			log.Error(ctx, "Can't save content list into cache", log.Err(err), log.String("key", key), log.String("data", string(contentListJSON)))
		}
	}()
}

func (r *ContentRedis) GetContentCacheByIdList(ctx context.Context, ids []string) ([]string, []*entity.ContentInfoWithDetails) {
	if !config.Get().RedisConfig.OpenCache {
		return ids, nil
	}
	keys := make([]string, len(ids))
	for i := range ids {
		keys[i] = r.contentKey(ids[i])
	}
	res, err := ro.MustGetRedis(ctx).MGet(keys...).Result()
	if err != nil {
		log.Error(ctx, "Can't get content list from cache", log.Err(err), log.Strings("keys", keys), log.Strings("ids", ids))
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

func (r *ContentRedis) GetContentCacheBySearchCondition(ctx context.Context, condition dbo.Conditions) *ContentListWithKey {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}
	key := r.contentConditionKey(condition)
	//res, err := ro.MustGetRedis(ctx).Get(key).Result()
	res, err := ro.MustGetRedis(ctx).HGet(RedisKeyPrefixContentCondition, r.conditionHash(condition)).Result()

	if err != nil {
		log.Error(ctx, "Can't get content condition from cache", log.Err(err), log.String("key", key), log.Any("condition", condition))
		return nil
	}
	contentLists := new(ContentListWithKey)
	err = json.Unmarshal([]byte(res), contentLists)
	if err != nil {
		log.Error(ctx, "Can't unmarshal content condition from cache", log.Err(err), log.String("key", key), log.String("json", res))
		err = ro.MustGetRedis(ctx).Del(key).Err()
		if err != nil {
			log.Error(ctx, "Can't delete content from cache", log.Err(err), log.String("key", key), log.String("json", res))
		}
		return nil
	}

	return contentLists
}

func (r *ContentRedis) CleanContentCache(ctx context.Context, ids []string) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}

	if len(ids) < 1 {
		return
	}
	//删除对应的id
	keys := make([]string, 0)
	for i := range ids {
		keys = append(keys, r.contentKey(ids[i]))
	}

	//删除所有condition cache
	//conditionKeys := ro.MustGetRedis(ctx).Keys(RedisKeyPrefixContentCondition + ":*").Val()
	//keys = append(keys, conditionKeys...)
	keys = append(keys, RedisKeyPrefixContentCondition)
	// go func() {
	err := ro.MustGetRedis(ctx).Del(keys...).Err()
	if err != nil {
		log.Error(ctx, "Can't clean content from cache", log.Err(err), log.Strings("keys", keys))
	}
	//time.Sleep(time.Second)
	//ro.MustGetRedis(ctx).Del(keys...)
	//if err != nil{
	//	log.Error(ctx, "Can't clean content again from cache", log.Err(err), log.Strings("keys", keys))
	//}
	// }()
}

func (r *ContentRedis) SetExpiration(t time.Duration) {
	r.expiration = t
}

var (
	_redisContentCache     *ContentRedis
	_redisContentCacheOnce sync.Once
)

func GetContentRedis() IContentRedis {
	_redisContentCacheOnce.Do(func() {
		_redisContentCache = &ContentRedis{expiration: time.Minute * 2}
	})
	return _redisContentCache
}
