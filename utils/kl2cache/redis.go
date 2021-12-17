package kl2cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

func OptRedis(host string, port int, password string) Option {
	return func(c *config) {
		c.Redis.Host = host
		c.Redis.Port = port
		c.Redis.Password = password
	}
}

var redisClient = &redis.Client{}

func initRedis(ctx context.Context, conf *config) (err error) {
	if !conf.Enable {
		return
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%v", conf.Redis.Host, conf.Redis.Port),
		Password: conf.Redis.Password,
	})
	err = redisClient.Ping().Err()
	if err != nil {
		log.Error(ctx, "ping redis failed", log.Err(err))
		return
	}

	DefaultProvider = &redisProvider{
		Client: redisClient,
	}
	return
}

type redisProvider struct {
	ExpireStrategy ExpireStrategy
	Client         *redis.Client
}

type User struct {
	ID   string
	Name string
}

func (r *redisProvider) BatchGet(ctx context.Context, keys []Key, val interface{}, f func(ctx context.Context, keys []Key) (val map[string]interface{}, err error)) (err error) {
	keyStrArr := []string{}
	for _, key := range keys {
		keyStrArr = append(keyStrArr, key.Key())
	}
	rs, err := r.Client.MGet(keyStrArr...).Result()
	if err != nil {
		return
	}
	valStrArr := []string{}
	var missed []Key
	for i := 0; i < len(keys); i++ {
		s, ok := rs[i].(string)
		if ok {
			valStrArr = append(valStrArr, s)
		} else {
			missed = append(missed, keys[i])
		}
	}
	if len(missed) > 0 {
		rsGot := map[string]interface{}{}
		rsGot, err = f(ctx, missed)
		if err != nil {
			return
		}
		pipe := r.Client.Pipeline()
		for _, mk := range missed {
			k := mk.Key()
			if v, ok := rsGot[k]; ok {
				var buf []byte
				buf, err = json.Marshal(v)
				if err != nil {
					return
				}
				s := string(buf)
				pipe.Set(mk.Key(), s, r.ExpireStrategy(nil))
				valStrArr = append(valStrArr, s)
			}
		}
		_, err = pipe.Exec()
		if err != nil {
			return
		}
	}

	err = json.Unmarshal([]byte("["+strings.Join(valStrArr, ",")+"]"), val)
	if err != nil {
		return
	}
	return
}
func (r *redisProvider) WithExpireStrategy(ctx context.Context, strategy ExpireStrategy) (provider Provider) {
	cloned := *r
	cloned.ExpireStrategy = strategy
	provider = &cloned
	return
}
func (r *redisProvider) Get(ctx context.Context, key Key, val interface{}, f func(ctx context.Context) (val interface{}, err error)) (err error) {
	var buf []byte
	s, err := r.Client.Get(key.Key()).Result()
	switch err {
	case nil:
		buf = []byte(s)
	case redis.Nil:
		err = nil
		var val1 interface{}
		val1, err = f(ctx)
		if err != nil {
			return
		}
		buf, err = json.Marshal(val1)
		if err != nil {
			log.Error(ctx, "marshal value failed", log.Any("value", val1))
			return
		}
		err = r.Client.Set(key.Key(), string(buf), r.ExpireStrategy(nil)).Err()
		if err != nil {
			return
		}
	default:
		log.Error(ctx, "get value from redis failed", log.Any("key", key), log.Err(err))
		return
	}

	err = json.Unmarshal(buf, val)
	if err != nil {
		return
	}
	return
}
