package da

import (
	"context"
	"encoding/json"
	"fmt"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/ro"
)

var (
	_userRedisDA     IUserRedisDA
	_userRedisDAOnce sync.Once
)

type IUserRedisDA interface {
	SetUsers(ctx context.Context, orgID string, users []*User) error
	GetUsersByOrg(ctx context.Context, orgID string) ([]*User, error)
}

type UserRedisDA struct {
	expiration time.Duration
}

func GetUserRedisDA() IUserRedisDA {
	_userRedisDAOnce.Do(func() {
		_userRedisDA = &UserRedisDA{
			expiration: config.Get().User.CacheExpiration,
		}
	})
	return _userRedisDA
}

type User struct {
	external.User
	Schools []*external.School
	Classes []*external.Class
}

func (u *UserRedisDA) SetUsers(ctx context.Context, orgID string, users []*User) error {
	if !config.Get().RedisConfig.OpenCache {
		log.Warn(ctx, "redis config open_cache is false")
		return ErrRedisDisabled
	}

	redisClient := ro.MustGetRedis(ctx)
	pipe := redisClient.Pipeline()
	for _, v := range users {
		key := u.getUserCacheKey(orgID, v.ID)
		b, err := json.Marshal(v)
		if err != nil {
			log.Error(ctx, "can't parse data into json",
				log.Err(err),
				log.Any("data", v),
			)
			return err
		}
		pipe.Set(ctx, key, string(b), u.expiration)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error(ctx, "failed to exec redis pipeline",
			log.Err(err),
			log.String("orgID", orgID),
			log.Any("users", users),
		)
		return err
	}

	return nil
}

func (u *UserRedisDA) GetUsersByOrg(ctx context.Context, orgID string) ([]*User, error) {
	redisClient := ro.MustGetRedis(ctx)
	matchPattern := u.getUserMatchPattern(orgID)
	var keys []string
	var cursor uint64
	iter := redisClient.Scan(ctx, cursor, matchPattern, 10).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		log.Error(ctx, "failed to scan redis key",
			log.Err(err),
			log.String("matchPattern", matchPattern),
		)
	}

	if len(keys) == 0 {
		log.Debug(ctx, "record not found in redis",
			log.String("matchPattern", matchPattern))
		return nil, nil
	}

	result, err := redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		log.Error(ctx, "failed to mget redis",
			log.Err(err),
			log.Strings("keys", keys),
		)
		return nil, err
	}

	users := make([]*User, len(result))
	for i, v := range result {
		var user *User
		err = json.Unmarshal([]byte(v.(string)), &user)
		if err != nil {
			log.Error(ctx, "unmarshal User error ",
				log.Err(err),
				log.Any("data", v),
			)
			return nil, err
		}
		users[i] = user
	}

	return users, nil
}

func (u *UserRedisDA) getUserCacheKey(orgID, userID string) string {
	return fmt.Sprintf("%s:%s:%s", RedisKeyPrefixUser, orgID, userID)
}

func (u *UserRedisDA) getUserMatchPattern(orgID string) string {
	return fmt.Sprintf("%s:%s:*", RedisKeyPrefixUser, orgID)
}
