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
		CalcExpired: conf.ExpireStrategy,
		Client:      redisClient,
	}
	return
}

type redisProvider struct {
	CalcExpired ExpireStrategy
	Client      *redis.Client
}

func (r *redisProvider) BatchGet(ctx context.Context, keys []Key, val interface{}, fGetData func(ctx context.Context, keys []Key) (kvs []*KeyValue, err error)) (err error) {
	var keyStrArr []string
	for _, key := range keys {
		keyStrArr = append(keyStrArr, key.Key())
	}
	log.Info(ctx, "start mget from redis", log.Any("keys", keyStrArr))
	rsCached, err := r.Client.MGet(keyStrArr...).Result()
	if err == redis.Nil {
		log.Info(ctx, "redis mget got redis.Nil", log.Any("keys", keyStrArr), log.Err(err))
		err = nil
	}
	if err != nil {
		log.Error(ctx, "redis mget failed", log.Any("keys", keyStrArr))
		return
	}

	valStrArr := make([]string, 0, len(keys))
	missed := make([]Key, 0, len(keys))
	mMissed := map[string]bool{}
	for i := 0; i < len(keys); i++ {
		s, ok := rsCached[i].(string)
		if ok {
			valStrArr = append(valStrArr, s)
		} else {
			missed = append(missed, keys[i])
			mMissed[keys[i].Key()] = true
		}
	}
	if len(missed) > 0 {
		log.Info(ctx, "some key not in redis", log.Any("missed", missed))
		var rsGot []*KeyValue
		rsGot, err = fGetData(ctx, missed)
		if err != nil {
			return
		}
		log.Info(ctx, "got data by call fGetData", log.Any("keys", missed), log.Any("rsGot", rsGot))
		pipe := r.Client.Pipeline()
		for _, kv := range rsGot {
			if !mMissed[kv.Key.Key()] {
				log.Info(ctx, "key not in missed", log.Any("kv", kv))
				continue
			}
			var buf []byte
			buf, err = json.Marshal(kv.Val)
			if err != nil {
				log.Error(ctx, "json.Marshal val failed", log.Err(err), log.Any("val", kv.Val))
				return
			}
			s := string(buf)
			err = pipe.Set(kv.Key.Key(), s, r.CalcExpired(nil)).Err()
			if err != nil {
				log.Error(ctx, "pipe set failed", log.Any("key", kv.Key.Key()), log.Any("val", s), log.Err(err))
				return
			}
			valStrArr = append(valStrArr, s)
		}
		_, err = pipe.Exec()
		if err != nil {
			log.Error(ctx, "set redis caches failed", log.Any("rsGot", rsGot), log.Err(err))
			return
		}
	}

	valStr := "[" + strings.Join(valStrArr, ",") + "]"
	err = json.Unmarshal([]byte(valStr), val)
	if err != nil {
		log.Error(ctx, "json.Unmarshal failed", log.Any("from", valStr), log.Any("to", val))
		return
	}
	return
}
func (r *redisProvider) WithExpireStrategy(ctx context.Context, strategy ExpireStrategy) (provider Provider) {
	cloned := *r
	cloned.CalcExpired = strategy
	provider = &cloned
	return
}
func (r *redisProvider) Get(ctx context.Context, key Key, val interface{}, fGetData func(ctx context.Context, key Key) (val interface{}, err error)) (err error) {
	var buf []byte
	s, err := r.Client.Get(key.Key()).Result()
	switch err {
	case nil:
		buf = []byte(s)
	case redis.Nil:
		err = nil
		var val1 interface{}
		val1, err = fGetData(ctx, key)
		if err != nil {
			return
		}
		buf, err = json.Marshal(val1)
		if err != nil {
			log.Error(ctx, "marshal value failed", log.Any("value", val1))
			return
		}
		err = r.Client.Set(key.Key(), string(buf), r.CalcExpired(nil)).Err()
		if err != nil {
			log.Error(ctx, "redis set failed", log.Err(err), log.Any("key", key.Key()), log.Any("val", string(buf)))
			return
		}
	default:
		log.Error(ctx, "get value from redis failed", log.Any("key", key), log.Err(err))
		return
	}

	err = json.Unmarshal(buf, val)
	if err != nil {
		log.Error(ctx, "json.Unmarshal failed", log.Err(err), log.Any("buf", string(buf)))
		return
	}
	return
}
