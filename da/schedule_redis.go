package da

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"sync"
	"time"
)

func (r *ScheduleRedisDA) BatchAddScheduleCache(ctx context.Context, schedules []*entity.ScheduleDetailsView) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}
	for _, item := range schedules {
		key := r.scheduleKey(item.ID)
		b, err := json.Marshal(item)
		if err != nil {
			log.Error(ctx, "Can't parse schedule into json",
				log.Err(err),
				log.Any("schedule", item),
			)
			continue
		}
		err = ro.MustGetRedis(ctx).Set(key, string(b), r.expiration).Err()
		if err != nil {
			log.Error(ctx, "Can't save schedule into cache",
				log.Err(err),
				log.Any("schedule", item),
			)
			continue
		}
	}
}

func (r *ScheduleRedisDA) addScheduleByCondition(ctx context.Context, condition dbo.Conditions, data interface{}) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Error(ctx, "Can't parse schedule list into json",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("data", data),
		)
		return
	}
	filed := r.conditionHash(condition)
	err = ro.MustGetRedis(ctx).Expire(RedisKeyPrefixScheduleCondition, r.expiration).Err()
	if err != nil {
		log.Error(ctx, "set schedule condition key expire error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("filed", filed),
			log.Any("data", data),
		)
		return
	}
	err = ro.MustGetRedis(ctx).HSet(RedisKeyPrefixScheduleCondition, filed, string(b)).Err()
	if err != nil {
		log.Error(ctx, "Can't save schedule into cache",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("filed", filed),
			log.Any("data", data),
		)
		return
	}
}

func (r *ScheduleRedisDA) AddScheduleListViewByCondition(ctx context.Context, condition dbo.Conditions, schedules []*entity.ScheduleListView) {
	r.addScheduleByCondition(ctx, condition, schedules)
}
func (r *ScheduleRedisDA) AddScheduleDatesByCondition(ctx context.Context, condition dbo.Conditions, dates []string) {
	r.addScheduleByCondition(ctx, condition, dates)
}

func (r *ScheduleRedisDA) GetScheduleCacheByIDs(ctx context.Context, ids []string) ([]*entity.ScheduleDetailsView, error) {
	if !config.Get().RedisConfig.OpenCache {
		return nil, errors.New("not open cache")
	}
	keys := make([]string, len(ids))
	for i := range ids {
		keys[i] = r.scheduleKey(ids[i])
	}
	res, err := ro.MustGetRedis(ctx).MGet(keys...).Result()
	if err != nil {
		log.Error(ctx, "Can't get schedule list from cache", log.Err(err))
		return nil, err
	}
	schedules := make([]*entity.ScheduleDetailsView, 0, len(res))
	for i := range res {
		resJSON, ok := res[i].(string)
		if !ok {
			log.Error(ctx, "Get invalid data from cache", log.Any("data", res[i]))
			return nil, errors.New("invalid cache")
		}
		schedule := new(entity.ScheduleDetailsView)
		err = json.Unmarshal([]byte(resJSON), schedule)
		if err != nil {
			log.Error(ctx, "Can't unmarshal schedule from cache",
				log.Err(err),
				log.String("JSON", resJSON))
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

func (r *ScheduleRedisDA) getScheduleByCondition(ctx context.Context, condition dbo.Conditions) (string, error) {
	if !config.Get().RedisConfig.OpenCache {
		return "", errors.New("not open cache ")
	}
	filed := r.conditionHash(condition)
	res, err := ro.MustGetRedis(ctx).HGet(RedisKeyPrefixScheduleCondition, filed).Result()
	if err != nil {
		log.Error(ctx, "Can't get schedule condition from cache",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("filed", filed),
		)
		return "", err
	}
	return res, nil
}

func (r *ScheduleRedisDA) GetScheduleListViewByCondition(ctx context.Context, condition dbo.Conditions) ([]*entity.ScheduleListView, error) {
	res, err := r.getScheduleByCondition(ctx, condition)
	if err != nil {
		return nil, err
	}
	var result []*entity.ScheduleListView
	err = json.Unmarshal([]byte(res), &result)
	if err != nil {
		log.Error(ctx, "unmarshal schedule error ",
			log.Err(err),
			log.Any("condition", condition),
			log.String("scheduleJson", res),
		)
		return nil, err
	}
	return result, nil
}
func (r *ScheduleRedisDA) GetScheduleDatesByCondition(ctx context.Context, condition dbo.Conditions) ([]string, error) {
	res, err := r.getScheduleByCondition(ctx, condition)
	if err != nil {
		return nil, err
	}
	var result []string
	err = json.Unmarshal([]byte(res), &result)
	if err != nil {
		log.Error(ctx, "unmarshal schedule error ",
			log.Err(err),
			log.Any("condition", condition),
			log.String("scheduleJson", res),
		)
		return nil, err
	}
	return result, nil
}

func (r *ScheduleRedisDA) scheduleKey(key string) string {
	return fmt.Sprintf("%s:%v", RedisKeyPrefixScheduleID, key)
}
func (r *ScheduleRedisDA) conditionHash(condition dbo.Conditions) string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%v", condition)))
	md5Hash := fmt.Sprintf("%x", h.Sum(nil))
	return fmt.Sprintf("%v", md5Hash)
}

func (r *ScheduleRedisDA) scheduleConditionKey(condition dbo.Conditions) string {
	md5Hash := r.conditionHash(condition)
	return fmt.Sprintf("%v:%v", RedisKeyPrefixScheduleCondition, md5Hash)
}
func (r *ScheduleRedisDA) Clean(ctx context.Context, ids []string) error {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = r.scheduleKey(id)
	}

	keys = append(keys, RedisKeyPrefixScheduleCondition)
	err := ro.MustGetRedis(ctx).Del(keys...).Err()
	if err != nil {
		log.Error(ctx, "redis del keys error",
			log.Err(err),
			log.Strings("ids", ids),
			log.Any("keys", keys),
		)
		return err
	}
	return nil
}

type ScheduleRedisDA struct {
	expiration time.Duration
}

var (
	_scheduleRedisDA     *ScheduleRedisDA
	_scheduleRedisDAOnce sync.Once
)

func GetScheduleRedisDA() *ScheduleRedisDA {
	_scheduleRedisDAOnce.Do(func() {
		_scheduleRedisDA = &ScheduleRedisDA{
			expiration: config.Get().Schedule.CacheExpiration,
		}
	})
	return _scheduleRedisDA
}
