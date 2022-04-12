package da

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/ro"
)

var (
	ErrRedisDisabled = errors.New("redis disabled")
)

var (
	_scheduleRedisDA     IScheduleCacheDA
	_scheduleRedisDAOnce sync.Once
)

type IScheduleCacheDA interface {
	Set(ctx context.Context, orgID string, condition *ScheduleCacheCondition, data interface{}) error
	GetScheduleListView(ctx context.Context, orgID string, condition *ScheduleCondition) ([]*entity.ScheduleListView, error)
	GetScheduledDates(ctx context.Context, orgID string, condition *ScheduleCondition) ([]string, error)
	GetScheduleBasic(ctx context.Context, orgID string, scheduleID string) (*entity.ScheduleBasic, error)
	GetScheduleDetailView(ctx context.Context, orgID string, userID string, scheduleID string) (*entity.ScheduleDetailsView, error)
	GetScheduleFilterUndefinedClass(ctx context.Context, orgID string, permissionMap map[external.PermissionName]bool) (bool, error)
	Clean(ctx context.Context, orgID string) error
}

type ScheduleCacheDataType string

const (
	ScheduleListView             ScheduleCacheDataType = "ScheduleListView"
	ScheduleBasic                ScheduleCacheDataType = "ScheduleBasic"
	ScheduleDetailView           ScheduleCacheDataType = "ScheduleDetailView"
	ScheduledDates               ScheduleCacheDataType = "ScheduledDates"
	ScheduleFilterUndefinedClass ScheduleCacheDataType = "UndefinedClass"
)

type ScheduleCacheCondition struct {
	Condition  dbo.Conditions
	UserID     string
	ScheduleID string
	SchoolID   string
	// TODO map is unordered
	PermissionMap map[external.PermissionName]bool
	DataType      ScheduleCacheDataType
}

type ScheduleRedisDA struct {
	expiration time.Duration
}

func GetScheduleRedisDA() IScheduleCacheDA {
	_scheduleRedisDAOnce.Do(func() {
		_scheduleRedisDA = &ScheduleRedisDA{
			expiration: config.Get().Schedule.CacheExpiration,
		}
	})
	return _scheduleRedisDA
}

func (r *ScheduleRedisDA) Set(ctx context.Context, orgID string, condition *ScheduleCacheCondition, data interface{}) error {
	if !config.Get().RedisConfig.OpenCache {
		log.Warn(ctx, "redis config open_cache is false")
		return ErrRedisDisabled
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Error(ctx, "can't parse data into json",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("data", data),
		)
		return err
	}
	key := r.getScheduleConditionKey(orgID)
	field := r.getScheduleConditionField(condition)
	redisClient := ro.MustGetRedis(ctx)
	pipe := redisClient.TxPipeline()
	exist := pipe.Exists(ctx, key)
	pipe.HSet(ctx, key, field, string(b))
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Error(ctx, "failed to exec redis pipeline",
			log.Err(err),
			log.String("key", key),
			log.String("field", field),
		)
		return err
	}

	log.Debug(ctx, "HSet in redis",
		log.String("key", key),
		log.String("field", field),
		log.Any("exist", exist),
		log.Duration("expiration", r.expiration))

	// key not exist
	if exist.Val() == int64(0) {
		err = redisClient.Expire(ctx, key, r.expiration).Err()
		if err != nil {
			log.Error(ctx, "failed to set timeout",
				log.Err(err),
				log.String("key", key),
				log.Any("expiration", r.expiration),
			)
			return err
		}
	}

	return nil
}

func (r *ScheduleRedisDA) GetScheduleListView(ctx context.Context, orgID string, condition *ScheduleCondition) ([]*entity.ScheduleListView, error) {
	cacheCondition := &ScheduleCacheCondition{
		Condition: condition,
		DataType:  ScheduleListView,
	}
	data, err := r.get(ctx, orgID, cacheCondition)
	if err == redis.Nil {
		log.Info(ctx, "redis cache missed", log.Any("condition", cacheCondition))
		return nil, err
	} else if err != nil {
		log.Error(ctx, "get from redis cache error",
			log.Err(err),
			log.Any("condition", cacheCondition),
		)
		return nil, err
	}

	var result []*entity.ScheduleListView
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		log.Error(ctx, "unmarshal ScheduleListView error ",
			log.Err(err),
			log.Any("condition", cacheCondition),
			log.String("data", data),
		)
		return nil, err
	}
	return result, nil
}

func (r *ScheduleRedisDA) GetScheduledDates(ctx context.Context, orgID string, condition *ScheduleCondition) ([]string, error) {
	cacheCondition := &ScheduleCacheCondition{
		Condition: condition,
		DataType:  ScheduledDates,
	}
	data, err := r.get(ctx, orgID, cacheCondition)
	if err == redis.Nil {
		log.Info(ctx, "redis cache missed", log.Any("condition", cacheCondition))
		return nil, err
	} else if err != nil {
		log.Error(ctx, "get from redis cache error",
			log.Err(err),
			log.Any("condition", cacheCondition),
		)
		return nil, err
	}

	var result []string
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		log.Error(ctx, "unmarshal ScheduledDates error ",
			log.Err(err),
			log.Any("condition", cacheCondition),
			log.String("data", data),
		)
		return nil, err
	}
	return result, nil
}

func (r *ScheduleRedisDA) GetScheduleBasic(ctx context.Context, orgID string, scheduleID string) (*entity.ScheduleBasic, error) {
	cacheCondition := &ScheduleCacheCondition{
		ScheduleID: scheduleID,
		DataType:   ScheduleBasic,
	}
	data, err := r.get(ctx, orgID, cacheCondition)
	if err == redis.Nil {
		log.Info(ctx, "redis cache missed", log.Any("condition", cacheCondition))
		return nil, err
	} else if err != nil {
		log.Error(ctx, "get from redis cache error",
			log.Err(err),
			log.Any("condition", cacheCondition),
		)
		return nil, err
	}

	var result *entity.ScheduleBasic
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		log.Error(ctx, "unmarshal ScheduleBasic error ",
			log.Err(err),
			log.Any("condition", cacheCondition),
			log.String("data", data),
		)
		return nil, err
	}
	return result, nil
}

func (r *ScheduleRedisDA) GetScheduleDetailView(ctx context.Context, orgID string, userID string, scheduleID string) (*entity.ScheduleDetailsView, error) {
	cacheCondition := &ScheduleCacheCondition{
		UserID:     userID,
		ScheduleID: scheduleID,
		DataType:   ScheduleDetailView,
	}
	data, err := r.get(ctx, orgID, cacheCondition)
	if err == redis.Nil {
		log.Info(ctx, "redis cache missed", log.Any("condition", cacheCondition))
		return nil, err
	} else if err != nil {
		log.Error(ctx, "get from redis cache error",
			log.Err(err),
			log.Any("condition", cacheCondition),
		)
		return nil, err
	}

	var result *entity.ScheduleDetailsView
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		log.Error(ctx, "unmarshal ScheduleBasic error ",
			log.Err(err),
			log.Any("condition", cacheCondition),
			log.String("data", data),
		)
		return nil, err
	}

	return result, nil
}

func (r *ScheduleRedisDA) GetScheduleFilterUndefinedClass(ctx context.Context, orgID string, permissionMap map[external.PermissionName]bool) (bool, error) {
	cacheCondition := &ScheduleCacheCondition{
		PermissionMap: permissionMap,
		DataType:      ScheduleFilterUndefinedClass,
	}
	data, err := r.get(ctx, orgID, cacheCondition)
	if err == redis.Nil {
		log.Info(ctx, "redis cache missed", log.Any("condition", cacheCondition))
		return false, err
	} else if err != nil {
		log.Error(ctx, "get from redis cache error",
			log.Err(err),
			log.Any("condition", cacheCondition),
		)
		return false, err
	}

	var result bool
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		log.Error(ctx, "unmarshal ScheduleBasic error ",
			log.Err(err),
			log.Any("condition", cacheCondition),
			log.String("data", data),
		)
		return false, err
	}
	return result, nil
}

func (r *ScheduleRedisDA) Clean(ctx context.Context, orgID string) error {
	if !config.Get().RedisConfig.OpenCache {
		log.Warn(ctx, "redis config open_cache is false")
		return ErrRedisDisabled
	}

	key := r.getScheduleConditionKey(orgID)
	err := ro.MustGetRedis(ctx).Del(ctx, key).Err()
	if err != nil {
		log.Error(ctx, "redis del keys error",
			log.Err(err),
			log.String("key", key),
		)
		return err
	}

	return nil
}

func (r *ScheduleRedisDA) getScheduleConditionKey(orgID string) string {
	return fmt.Sprintf("%s:%s", RedisKeyPrefixScheduleCondition, orgID)
}

func (r *ScheduleRedisDA) getScheduleConditionField(condition *ScheduleCacheCondition) string {
	h := md5.New()
	b, _ := json.Marshal(condition)
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (r *ScheduleRedisDA) get(ctx context.Context, orgID string, condition *ScheduleCacheCondition) (string, error) {
	if !config.Get().RedisConfig.OpenCache {
		log.Warn(ctx, "redis config open_cache is false")
		return "", ErrRedisDisabled
	}

	key := r.getScheduleConditionKey(orgID)
	field := r.getScheduleConditionField(condition)
	log.Debug(ctx, "HGet in redis",
		log.String("key", key),
		log.String("field", field))

	return ro.MustGetRedis(ctx).HGet(ctx, key, field).Result()
}
