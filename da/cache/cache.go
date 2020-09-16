package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"strconv"
	"time"
)
const (
	RedisKeyPrefixCachedCondition = ":condition"
	RedisKeyPrefixCachedId        = ":id"
	RedisKeyConditionKeys = "condition_keys"
)

var (
	ErrCacheObjectNotFound = "cache object not found"
)

type CachedObjectListWithTotal struct {
	Total      int            `json:"total"`
	ObjectList []CachedObject `json:"data"`
}

type CachedObjectListWithRestId struct {
	ObjectList 	[]CachedObject `json:"data"`
	RestId 		[]string `json:"rest_id"`
}

type ICache interface{
	SaveCacheObjectList(ctx context.Context, objectList []CachedObject)
	SaveCacheObjectListBySearchCondition(ctx context.Context, condition dbo.Conditions, objectList CachedObjectListWithTotal)
	GetCacheObjectByIdList(ctx context.Context, ids []string)*CachedObjectListWithRestId
	GetCacheObjectBySearchCondition(ctx context.Context, condition dbo.Conditions) *CachedObjectListWithTotal

	UpdateCacheObject(ctx context.Context, object CachedObject)

	SetExpiration(t time.Duration)
}

type CacheRedis struct {
	decoder Decoder
	expiration time.Duration
}

func (r *CacheRedis) GetCacheObjectByIdList(ctx context.Context, ids []string) *CachedObjectListWithRestId {
	if !config.Get().RedisConfig.OpenCache {
		return &CachedObjectListWithRestId{
			ObjectList: nil,
			RestId:     ids,
		}
	}
	keys := make([]string, len(ids))
	for i := range ids {
		keys[i] = r.contentKey(ids[i], r.decoder.ObjectName())
	}
	res, err := ro.MustGetRedis(ctx).MGet(keys...).Result()
	if err != nil {
		log.Error(ctx, "Can't get object list from cache", log.Err(err), log.Strings("keys", keys), log.Strings("ids", ids))
		return &CachedObjectListWithRestId{
			ObjectList: nil,
			RestId:     ids,
		}
	}

	//解析cachedContents
	cachedObjects := make([]CachedObject, 0)
	for i := range res {
		resJSON, ok := res[i].(string)
		if !ok {
			log.Error(ctx, "Get invalid data from cache", log.Any("data", res[i]))
			continue
		}
		obj, err := r.decoder.DecodeObject(resJSON)
		if err != nil {
			log.Error(ctx, "Can't unmarshal object from cache", log.Err(err), log.String("JSON", resJSON))
			continue
		}
		cachedObjects = append(cachedObjects, obj)
	}

	//标记需要被删掉的id
	deletedMarks := make([]bool, len(ids))
	for i := range ids {
		for j := range cachedObjects {
			if ids[i] == cachedObjects[j].ID() {
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

	return &CachedObjectListWithRestId{
		ObjectList: cachedObjects,
		RestId:     restIds,
	}
}

func (r *CacheRedis) GetCacheObjectBySearchCondition(ctx context.Context, condition dbo.Conditions) *CachedObjectListWithTotal {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}
	key := r.decoder.ObjectName() + RedisKeyPrefixCachedCondition
	field := r.conditionHash(condition)
	res, err := ro.MustGetRedis(ctx).HGet(key, field).Result()

	if err != nil {
		log.Error(ctx, "Can't get object condition from cache", log.Err(err), log.String("key", key), log.String("field", field), log.Any("condition", condition))
		return nil
	}
	cachedObjectList ,err := r.decoder.BatchDecodeObject(res)
	if err != nil {
		log.Error(ctx, "Can't unmarshal object condition from cache", log.Err(err), log.String("key", key), log.String("json", res))
		err = ro.MustGetRedis(ctx).HDel(key, field).Err()
		if err != nil {
			log.Error(ctx, "Can't delete object from cache", log.Err(err), log.String("key", key), log.String("json", res))
		}
		return nil
	}
	log.Info(ctx, "search object from cache", log.String("key", key))

	totalStr, err := ro.MustGetRedis(ctx).HGet(key, r.conditionTotalKey(field)).Result()
	if err != nil {
		log.Error(ctx, "Can't get object total condition from cache", log.Err(err), log.String("key", key), log.String("field", r.conditionTotalKey(field)), log.Any("condition", condition))
		return nil
	}
	total, err := strconv.Atoi(totalStr)
	if err != nil{
		log.Error(ctx, "Can't parse total from cache", log.Err(err), log.String("total", totalStr))
	}

	return &CachedObjectListWithTotal{
		Total:      total,
		ObjectList: cachedObjectList,
	}
}

func (r *CacheRedis) contentKey(id, name string) string {
	return fmt.Sprintf("%v:%v:%v", name, RedisKeyPrefixCachedId, id)
}

func (r *CacheRedis) conditionHash(condition dbo.Conditions) string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%v", condition)))
	md5Hash := fmt.Sprintf("%x", h.Sum(nil))
	return fmt.Sprintf("%v", md5Hash)
}

func (r *CacheRedis) conditionTotalKey(hash string) string {
	return hash + ":total"
}

func (r *CacheRedis) conditionKey(name string, condition dbo.Conditions) string {
	md5Hash := r.conditionHash(condition)
	return fmt.Sprintf("%v:%v:%v", RedisKeyPrefixCachedCondition, name, md5Hash)
}

func (r *CacheRedis) SaveCacheObjectList(ctx context.Context, objectList []CachedObject){
	if !config.Get().RedisConfig.OpenCache {
		return
	}
	go func() {
		for i := range objectList {
			key := r.contentKey(objectList[i].ID(), objectList[i].Name())
			contentJSON, err := json.Marshal(objectList[i])
			if err != nil {
				log.Error(ctx, "Can't parse object into json", log.Err(err), log.String("cid", objectList[i].ID()), log.String("name", objectList[i].Name()))
				continue
			}
			err = ro.MustGetRedis(ctx).SetNX(key, string(contentJSON), r.expiration).Err()
			if err != nil {
				log.Error(ctx, "Can't save object into cache", log.Err(err), log.String("key", key), log.String("data", string(contentJSON)))
				continue
			}
		}
	}()

}
func (r *CacheRedis) SaveContentCache(ctx context.Context, content CachedObject) {
	r.SaveCacheObjectList(ctx, []CachedObject{
		content,
	})
}

func (r *CacheRedis) SaveCacheObjectListBySearchCondition(ctx context.Context, condition dbo.Conditions, c CachedObjectListWithTotal) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}

	if len(c.ObjectList) < 1 {
		return
	}

	go func() {
		key := r.conditionKey(c.ObjectList[0].Name(), condition)
		log.Info(ctx, "save object into cache", log.Any("condition", condition), log.String("key", key))
		contentListJSON, err := json.Marshal(c.ObjectList)
		if err != nil {
			log.Error(ctx, "Can't parse object list into json", log.Err(err), log.String("key", key), log.String("data", string(contentListJSON)))
			return
		}
		conditionHash := r.conditionHash(condition)
		ro.MustGetRedis(ctx).Expire(RedisKeyPrefixCachedCondition, r.expiration)
		err = ro.MustGetRedis(ctx).HSetNX(RedisKeyPrefixCachedCondition, conditionHash, string(contentListJSON)).Err()
		if err != nil {
			log.Error(ctx, "Can't save object list into cache", log.Err(err), log.String("key", key), log.String("data", string(contentListJSON)))
			return
		}
		err = ro.MustGetRedis(ctx).HSetNX(RedisKeyPrefixCachedCondition, r.conditionTotalKey(conditionHash), strconv.Itoa(c.Total)).Err()
		if err != nil {
			log.Error(ctx, "Can't save object list total into cache", log.Err(err), log.String("key", key), log.String("data", string(contentListJSON)))
			return
		}

		conditionJSON, err := json.Marshal(condition)
		if err != nil {
			log.Error(ctx, "Can't marshal condition", log.Err(err), log.Any("condition", condition))
			return
		}
		err = ro.MustGetRedis(ctx).HSetNX(RedisKeyConditionKeys, conditionHash, string(conditionJSON)).Err()
		if err != nil {
			log.Error(ctx, "Can't save object key into cache", log.Err(err), log.String("key", key), log.String("data", string(contentListJSON)))
		}
	}()
}

func (r *CacheRedis) UpdateCacheObject(ctx context.Context, object CachedObject) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}
	//删除id
	key := r.contentKey(object.ID(), object.Name())
	err := ro.MustGetRedis(ctx).Del(key).Err()
	if err != nil{
		log.Info(ctx, "delete object from cache", log.String("key", key))
	}

	cds, err := ro.MustGetRedis(ctx).HGetAll(RedisKeyConditionKeys).Result()
	if err != nil{
		log.Info(ctx, "failed to get condition key from cache", log.Err(err))
	}

	conditionEntity := make([]*ConditionEntity, len(cds))
	i := 0
	for k, v := range cds{
		condition, err := r.decoder.DecodeCondition(v)
		if err != nil{
			log.Info(ctx, "failed to decode condition from cache", log.Err(err))
			continue
		}
		conditionEntity[i] = &ConditionEntity{
			ConditionKey: k,
			Conditions:   condition,
		}
		i ++
	}
	result := object.IsConditionMapping(conditionEntity)
	deletedKeys := make([]string ,0)
	for i := range result {
		if result[i].IsMapping {
			deletedKeys = append(deletedKeys, result[i].ConditionKey)
			deletedKeys = append(deletedKeys, r.conditionTotalKey(result[i].ConditionKey))
		}
	}

	err = ro.MustGetRedis(ctx).HDel(RedisKeyPrefixCachedCondition, deletedKeys...).Err()
	if err != nil{
		log.Info(ctx, "failed to get condition key from cache", log.Err(err))
	}
}

func (r *CacheRedis) SetExpiration(t time.Duration) {
	r.expiration = t
}

func NewCacheRedis(d Decoder) *CacheRedis {
	return &CacheRedis{
		decoder:    d,
	}
}