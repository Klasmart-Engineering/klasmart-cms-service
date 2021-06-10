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
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/ro"
)

type ScheduleCacheFlag string

const (
	ScheduleFilterUnDefineClass ScheduleCacheFlag = "UnDefineClass"
)

type ScheduleCacheCondition struct {
	Condition     dbo.Conditions
	UserID        string
	ScheduleID    string
	SchoolID      string
	CacheFlag     ScheduleCacheFlag
	PermissionMap map[external.PermissionName]bool
}

type IScheduleCacheDA interface {
	Add(ctx context.Context, orgID string, condition *ScheduleCacheCondition, data interface{}) error
	SearchToListView(ctx context.Context, orgID string, condition *ScheduleCondition) ([]*entity.ScheduleListView, error)
	SearchToStrings(ctx context.Context, orgID string, condition *ScheduleCondition) ([]string, error)
	SearchToBasicData(ctx context.Context, orgID string, scheduleID string) (*entity.ScheduleBasic, error)
	SearchToScheduleDetails(ctx context.Context, orgID string, userID string, scheduleID string) (*entity.ScheduleDetailsView, error)
	SearchToBool(ctx context.Context, orgID string, condition *ScheduleCacheCondition) (bool, error)
	Clean(ctx context.Context, orgID string) error

	SaveScheduleListView(ctx context.Context, orgID string, condition *ScheduleCacheCondition, data []*entity.ScheduleListView) error
	GetScheduleListView(ctx context.Context, orgID string, condition *ScheduleCacheCondition) ([]*entity.ScheduleListView, error)
}

func (r *ScheduleRedisDA) Add(ctx context.Context, orgID string, condition *ScheduleCacheCondition, data interface{}) error {
	if !config.Get().RedisConfig.OpenCache {
		log.Info(ctx, "redis disabled")
		return nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Error(ctx, "Can't parse schedule list into json",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("data", data),
		)
		return err
	}
	key := r.getHSetKey(orgID)
	filed := r.conditionHash(condition)
	err = ro.MustGetRedis(ctx).Expire(key, r.expiration).Err()
	if err != nil {
		log.Error(ctx, "set schedule condition key expire error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("filed", filed),
			log.Any("data", data),
		)
		return err
	}
	err = ro.MustGetRedis(ctx).HSet(key, filed, string(b)).Err()
	if err != nil {
		log.Error(ctx, "Can't save schedule into cache",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("filed", filed),
			log.Any("data", data),
		)
		return err
	}
	return nil
}

func (r *ScheduleRedisDA) SearchToListView(ctx context.Context, orgID string, condition *ScheduleCondition) ([]*entity.ScheduleListView, error) {
	res, err := r.search(ctx, orgID, &ScheduleCacheCondition{Condition: condition})
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

func (r *ScheduleRedisDA) SearchToStrings(ctx context.Context, orgID string, condition *ScheduleCondition) ([]string, error) {
	res, err := r.search(ctx, orgID, &ScheduleCacheCondition{Condition: condition})
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

func (r *ScheduleRedisDA) SearchToBasicData(ctx context.Context, orgID string, scheduleID string) (*entity.ScheduleBasic, error) {
	res, err := r.search(ctx, orgID, &ScheduleCacheCondition{ScheduleID: scheduleID})
	if err != nil {
		return nil, err
	}
	var result *entity.ScheduleBasic
	err = json.Unmarshal([]byte(res), &result)
	if err != nil {
		log.Error(ctx, "unmarshal schedule error ",
			log.Err(err),
			log.String("scheduleID", scheduleID),
			log.String("scheduleJson", res),
		)
		return nil, err
	}
	return result, nil
}

func (r *ScheduleRedisDA) SearchToScheduleDetails(ctx context.Context, orgID string, userID string, scheduleID string) (*entity.ScheduleDetailsView, error) {
	res, err := r.search(ctx, orgID, &ScheduleCacheCondition{ScheduleID: scheduleID, UserID: userID})
	if err != nil {
		return nil, err
	}
	var result *entity.ScheduleDetailsView
	err = json.Unmarshal([]byte(res), &result)
	if err != nil {
		log.Error(ctx, "unmarshal schedule error ",
			log.Err(err),
			log.String("userID", userID),
			log.String("scheduleID", scheduleID),
			log.String("scheduleJson", res),
		)
		return nil, err
	}
	return result, nil
}

func (r *ScheduleRedisDA) SearchToBool(ctx context.Context, orgID string, condition *ScheduleCacheCondition) (bool, error) {
	res, err := r.search(ctx, orgID, condition)
	if err != nil {
		return false, err
	}
	var result bool
	err = json.Unmarshal([]byte(res), &result)
	if err != nil {
		log.Error(ctx, "unmarshal schedule error ",
			log.Err(err),
			log.String("orgID", orgID),
			log.Any("condition", condition),
			log.String("scheduleJson", res),
		)
		return false, err
	}
	return result, nil
}

func (r *ScheduleRedisDA) Clean(ctx context.Context, orgID string) error {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}
	key := r.getHSetKey(orgID)
	err := ro.MustGetRedis(ctx).Del(key).Err()
	if err != nil {
		log.Error(ctx, "redis del keys error",
			log.Err(err),
			log.String("key", key),
		)
		return err
	}
	return nil
}

func (r *ScheduleRedisDA) SaveScheduleListView(ctx context.Context, orgID string, condition *ScheduleCacheCondition, data []*entity.ScheduleListView) error {
	if !config.Get().RedisConfig.OpenCache {
		return nil
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Error(ctx, "Can't parse ScheduleListView into json",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("data", data),
		)
		return err
	}

	key := r.getScheduleListViewKey(orgID, condition)
	err = ro.MustGetRedis(ctx).Set(key, string(b), r.expiration).Err()
	if err != nil {
		log.Error(ctx, "Save ScheduleListView redis error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("key", key),
			log.Any("data", data),
		)
		return err
	}

	return nil
}

func (r *ScheduleRedisDA) GetScheduleListView(ctx context.Context, orgID string, condition *ScheduleCacheCondition) ([]*entity.ScheduleListView, error) {
	if !config.Get().RedisConfig.OpenCache {
		return nil, nil
	}

	key := r.getScheduleListViewKey(orgID, condition)
	data, err := ro.MustGetRedis(ctx).Get(key).Result()
	if err != nil {
		log.Error(ctx, "Get ScheduleListView redis error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("key", key),
		)
		return nil, err
	}

	var result []*entity.ScheduleListView
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		log.Error(ctx, "Unmarshal ScheduleListView error ",
			log.Err(err),
			log.Any("condition", condition),
			log.String("scheduleJson", data),
		)
		return nil, err
	}

	return result, nil
}

func (r *ScheduleRedisDA) getHSetKey(orgID string) string {
	return fmt.Sprintf("%s:%s", RedisKeyPrefixScheduleCondition, orgID)
}

func (r *ScheduleRedisDA) search(ctx context.Context, orgID string, condition *ScheduleCacheCondition) (string, error) {
	if !config.Get().RedisConfig.OpenCache {
		return "", errors.New("not open cache ")
	}
	key := r.getHSetKey(orgID)
	filed := r.conditionHash(condition)
	res, err := ro.MustGetRedis(ctx).HGet(key, filed).Result()
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

func (r *ScheduleRedisDA) conditionHash(condition *ScheduleCacheCondition) string {
	h := md5.New()
	b, _ := json.Marshal(condition)
	h.Write(b)
	md5Hash := fmt.Sprintf("%x", h.Sum(nil))
	return md5Hash
}

func (r *ScheduleRedisDA) getScheduleListViewKey(orgID string, condition *ScheduleCacheCondition) string {
	md5Hash := r.conditionHash(condition)
	return fmt.Sprintf("%s:%s:%s", RedisKeyPrefixScheduleCondition, orgID, md5Hash)
}

//func (r *ContentRedis) conditionHash(condition dbo.Conditions) string {
//	h := md5.New()
//	h.Write([]byte(fmt.Sprintf("%v", condition)))
//	md5Hash := fmt.Sprintf("%x", h.Sum(nil))
//	return fmt.Sprintf("%v", md5Hash)
//}
//func (r *ContentRedis) contentConditionKey(condition dbo.Conditions) string {
//	md5Hash := r.conditionHash(condition)
//	return fmt.Sprintf("%v:%v", RedisKeyPrefixContentCondition, md5Hash)
//}
//func (r *ScheduleRedisDA) BatchAdd(ctx context.Context, op *entity.Operator, schedules []*entity.ScheduleDetailsView) error {
//	if !config.Get().RedisConfig.OpenCache {
//		log.Info(ctx, "redis disabled")
//		return nil
//	}
//	for _, item := range schedules {
//		key := r.scheduleKey(fmt.Sprintf("%s_%s", op.UserID, item.ID))
//		b, err := json.Marshal(item)
//		if err != nil {
//			log.Error(ctx, "Can't parse schedule into json",
//				log.Err(err),
//				log.Any("schedule", item),
//			)
//			return err
//		}
//		err = ro.MustGetRedis(ctx).Set(key, string(b), r.expiration).Err()
//		if err != nil {
//			log.Error(ctx, "Can't save schedule into cache",
//				log.Err(err),
//				log.Any("schedule", item),
//			)
//			return err
//		}
//	}
//	return nil
//}

//func (r *ScheduleRedisDA) GetByIDs(ctx context.Context, op *entity.Operator, ids []string) ([]*entity.ScheduleDetailsView, error) {
//	if !config.Get().RedisConfig.OpenCache {
//		return nil, errors.New("not open cache")
//	}
//	keys := make([]string, len(ids))
//	for i, id := range ids {
//		keys[i] = r.scheduleKey(fmt.Sprintf("%s_%s", op.UserID, id))
//	}
//	res, err := ro.MustGetRedis(ctx).MGet(keys...).Result()
//	if err != nil {
//		log.Error(ctx, "Can't get schedule list from cache", log.Err(err))
//		return nil, err
//	}
//	schedules := make([]*entity.ScheduleDetailsView, 0, len(res))
//	for i := range res {
//		resJSON, ok := res[i].(string)
//		if !ok {
//			log.Error(ctx, "Get invalid data from cache", log.Any("data", res[i]))
//			return nil, errors.New("invalid cache")
//		}
//		schedule := new(entity.ScheduleDetailsView)
//		err = json.Unmarshal([]byte(resJSON), schedule)
//		if err != nil {
//			log.Error(ctx, "Can't unmarshal schedule from cache",
//				log.Err(err),
//				log.String("JSON", resJSON))
//			return nil, err
//		}
//		schedules = append(schedules, schedule)
//	}
//	return schedules, nil
//}

//func (r *ScheduleRedisDA) scheduleKey(key string) string {
//	return fmt.Sprintf("%s:%v", RedisKeyPrefixScheduleID, key)
//}

//func (r *ScheduleRedisDA) scheduleConditionKey(condition *ScheduleCacheCondition) string {
//	md5Hash := r.conditionHash(condition)
//	return fmt.Sprintf("%v:%v", RedisKeyPrefixScheduleCondition, md5Hash)
//}

type ScheduleRedisDA struct {
	expiration time.Duration
}

var (
	_scheduleRedisDA     IScheduleCacheDA
	_scheduleRedisDAOnce sync.Once
)

func GetScheduleRedisDA() IScheduleCacheDA {
	_scheduleRedisDAOnce.Do(func() {
		_scheduleRedisDA = &ScheduleRedisDA{
			expiration: config.Get().Schedule.CacheExpiration,
		}
	})
	return _scheduleRedisDA
}
