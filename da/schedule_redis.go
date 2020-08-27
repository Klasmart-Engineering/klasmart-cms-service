package da

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"sync"
	"time"
)

const (
	CacheScheduleIDKey = "kidsLoop2.schedule.id"
	//CacheScheduleConditionKey = "kidsLoop2.schedule.condition"
)

func (r *ScheduleRedisDA) BatchAddScheduleCache(ctx context.Context, schedules []*entity.ScheduleDetailsView) {
	if !config.Get().RedisConfig.OpenCache {
		return
	}
	go func() {
		for _, item := range schedules {
			key := r.scheduleIDKey(item.ID)
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
	}()
}
func (r *ScheduleRedisDA) GetScheduleCacheByIDs(ctx context.Context, ids []string) ([]*entity.ScheduleDetailsView, error) {
	if !config.Get().RedisConfig.OpenCache {
		return nil, errors.New("not open cache")
	}
	keys := make([]string, len(ids))
	for i := range ids {
		keys[i] = r.scheduleIDKey(ids[i])
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
func (r *ScheduleRedisDA) scheduleIDKey(id string) string {
	return fmt.Sprintf("%s.%v", CacheScheduleIDKey, id)
}

func (r *ScheduleRedisDA) Clean(ctx context.Context, ids []string) error {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = r.scheduleIDKey(id)
	}

	err := ro.MustGetRedis(ctx).Del(keys...).Err()
	if err != nil {
		log.Error(ctx, "redis del keys error", log.Err(err), log.Strings("ids", ids))
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
		_scheduleRedisDA = &ScheduleRedisDA{expiration: time.Minute * 3}
	})
	return _scheduleRedisDA
}
