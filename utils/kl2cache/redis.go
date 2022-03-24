package kl2cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

func OptRedis(host string, port int, password string) Option {
	return func(c *config) {
		c.Redis.Host = host
		c.Redis.Port = port
		c.Redis.Password = password
	}
}
func initRedis(ctx context.Context, conf *config) (err error) {
	rProvider := &redisProvider{
		Enable: conf.Enable,
	}
	DefaultProvider = rProvider
	if !conf.Enable {
		return
	}
	rProvider.CalcExpired = conf.ExpireStrategy

	rProvider.Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%v", conf.Redis.Host, conf.Redis.Port),
		Password: conf.Redis.Password,
	})
	err = rProvider.Client.Ping(ctx).Err()
	if err != nil {
		log.Error(ctx, "ping redis failed", log.Err(err))
		return
	}
	return
}

type redisProvider struct {
	Enable      bool
	CalcExpired ExpireStrategy
	Client      *redis.Client
}

func (r *redisProvider) WithExpireStrategy(ctx context.Context, strategy ExpireStrategy) (provider Provider) {
	cloned := *r
	cloned.CalcExpired = strategy
	provider = &cloned
	return
}

func (r *redisProvider) Get(ctx context.Context, key Key, val interface{}, fGetData FuncGet) (err error) {
	innerPanic := false
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if innerPanic {
			panic(rec)
		}
		err = fmt.Errorf("kl2cache(redisProvider).Get panic: %+v", rec)
		log.Error(ctx, "Get panic", log.Err(err))
		return
	}()

	log.Info(ctx,
		"start Get",
		log.Any("cache", "kl2cache"),
		log.Any("provider", "redisProvider"),
		log.Any("method", "redisProvider.Get"),
		log.Any("key", key.Key()),
	)
	innerFnGetData := fGetData.wrapPanic()
	if !r.Enable {
		innerPanic, err = r.getWithOutCache(ctx, key, val, innerFnGetData)
		if err != nil {
			return
		}
		return
	}
	innerPanic, err = r.getWithCache(ctx, key, val, innerFnGetData)
	if err != nil {
		return
	}
	return
}

func (r *redisProvider) getWithOutCache(ctx context.Context, key Key, val interface{}, fGetData innerFuncGet) (innerPanic bool, err error) {
	vRes := reflect.ValueOf(val)
	if vRes.Kind() != reflect.Ptr {
		err = errors.New("val must be a pointer")
		log.Error(ctx, "getWithOutCache invalid params", log.Err(err))
		return
	}
	vRes = vRes.Elem()

	var got interface{}
	innerPanic, got, err = fGetData(ctx, key)
	if err != nil {
		return
	}
	if got == nil {
		return
	}
	vGot := reflect.ValueOf(got)
	if vGot.Kind() == reflect.Ptr {
		vGot = vGot.Elem()
	}
	if !vGot.IsZero() {
		vRes.Set(vGot)
	}
	return
}

func (r *redisProvider) getWithCache(ctx context.Context, key Key, val interface{}, fGetData innerFuncGet) (innerPanic bool, err error) {
	start := time.Now()
	var buf []byte
	s, err := r.Client.Get(ctx, key.Key()).Result()
	switch err {
	case nil:
		log.Info(ctx,
			"got data from redis",
			log.Any("key", key.Key()),
			log.Any("val", s),
			log.Duration("duration", time.Since(start)),
		)
		buf = []byte(s)
	case redis.Nil:
		log.Info(ctx,
			"miss cache from redis,call fGetData",
			log.Any("key", key.Key()),
			log.Duration("duration", time.Since(start)),
		)
		err = nil
		var val1 interface{}
		innerPanic, val1, err = fGetData(ctx, key)
		log.Info(ctx, "fGetData result", log.Any("val", val1), log.Err(err))
		if err != nil {
			return
		}

		buf, err = json.Marshal(val1)
		if err != nil {
			log.Error(ctx, "marshal value failed", log.Err(err), log.Any("value", val1))
			return
		}
		err = r.Client.Set(ctx, key.Key(), string(buf), r.CalcExpired(nil)).Err()
		if err != nil {
			log.Error(ctx, "redis set failed", log.Err(err), log.Any("key", key.Key()), log.Any("val", string(buf)))
			return
		}
	default:
		log.Error(ctx, "get value from redis failed",
			log.Err(err),
			log.Any("key", key),
			log.Duration("duration", time.Since(start)))
		return
	}

	err = json.Unmarshal(buf, val)
	if err != nil {
		log.Error(ctx,
			"json.Unmarshal failed",
			log.Err(err),
			log.Any("from", string(buf)),
			log.Any("to", val),
		)
		return
	}
	return
}

func (r *redisProvider) BatchGet(ctx context.Context, keys []Key, val interface{}, fGetData FuncBatchGet) (err error) {
	innerPanic := false
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if innerPanic {
			panic(rec)
		}
		err = fmt.Errorf("kl2cache(redisProvider).BatchGet panic: %+v", rec)
		log.Error(ctx, "BatchGet panic", log.Err(err))
		return
	}()

	var keyStrArr []string
	for _, key := range keys {
		keyStrArr = append(keyStrArr, key.Key())
	}
	log.Info(ctx,
		"start BatchGet",
		log.Any("cache", "kl2cache"),
		log.Any("provider", "redisProvider"),
		log.Any("method", "redisProvider.Get"),
		log.Any("keys", keyStrArr),
	)
	innerFnGetData := fGetData.wrapPanic()
	if !r.Enable {
		innerPanic, err = r.batchGetWithOutCache(ctx, keys, val, innerFnGetData)
		if err != nil {
			return
		}
		return
	}
	innerPanic, err = r.batchGetWithCache(ctx, keys, val, innerFnGetData)
	if err != nil {
		return
	}
	return
}

func (r *redisProvider) batchGetWithOutCache(ctx context.Context, keys []Key, val interface{}, fGetData innerFuncBatchGet) (innerPanic bool, err error) {
	vRes := reflect.ValueOf(val)
	if vRes.Kind() != reflect.Ptr {
		log.Error(ctx, "val for BatchGet must be a pointer", log.Any("vRes", vRes))
		err = constant.ErrBadUsageOfKl2Cache
		return
	}
	vRes = vRes.Elem()
	if vRes.Kind() != reflect.Slice {
		log.Error(ctx, "val must be a pointer of slice")
		err = constant.ErrBadUsageOfKl2Cache
		return
	}

	var kvs []*KeyValue
	innerPanic, kvs, err = fGetData(ctx, keys)
	if err != nil {
		return
	}
	for _, kv := range kvs {
		vGot := reflect.ValueOf(kv.Val)
		if vGot.Type() != reflect.TypeOf(val).Elem().Elem() {
			log.Error(ctx, "the type of item in val for BatchGet must equal the type of KeyVal.Val return by fGetData")
			err = constant.ErrBadUsageOfKl2Cache
			return
		}
		vRes = reflect.Append(vRes, vGot)
	}
	if len(kvs) > 0 {
		reflect.ValueOf(val).Elem().Set(vRes)
	}
	return

}

func (r *redisProvider) batchGetWithCache(ctx context.Context, keys []Key, val interface{}, fGetData innerFuncBatchGet) (innerPanic bool, err error) {
	var keyStrArr []string
	for _, key := range keys {
		keyStrArr = append(keyStrArr, key.Key())
	}

	start := time.Now()
	rsCached, err := r.Client.MGet(ctx, keyStrArr...).Result()
	if err == redis.Nil {
		log.Info(ctx, "redis mget got redis.Nil",
			log.Any("keys", keyStrArr),
			log.Err(err),
			log.Duration("duration", time.Since(start)))
		err = nil
	}
	if err != nil {
		log.Error(ctx, "redis mget failed",
			log.Err(err),
			log.Any("keys", keyStrArr),
			log.Duration("duration", time.Since(start)))
		return
	}

	log.Debug(ctx, "redis mget successfully",
		log.Any("keys", keyStrArr),
		log.Duration("duration", time.Since(start)))

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
	log.Info(ctx, "missed keys", log.Any("keys", keyStrArr), log.Any("missed", missed))
	if len(missed) > 0 {
		var rsGot []*KeyValue
		innerPanic, rsGot, err = fGetData(ctx, missed)
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
			err = pipe.Set(ctx, kv.Key.Key(), s, r.CalcExpired(nil)).Err()
			if err != nil {
				log.Error(ctx,
					"pipe set failed",
					log.Err(err),
					log.Any("key", kv.Key.Key()),
					log.Any("val", s),
				)
				return
			}
			valStrArr = append(valStrArr, s)
		}

		start := time.Now()
		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Error(ctx,
				"set redis caches failed",
				log.Err(err),
				log.Any("rsGot", rsGot),
				log.Duration("duration", time.Since(start)),
			)
			return
		}

		log.Debug(ctx,
			"set redis caches successfully",
			log.Any("rsGot", rsGot),
			log.Duration("duration", time.Since(start)))
	}

	valStr := "[" + strings.Join(valStrArr, ",") + "]"
	err = json.Unmarshal([]byte(valStr), val)
	if err != nil {
		log.Error(ctx,
			"json.Unmarshal failed",
			log.Err(err),
			log.Any("from", valStr),
			log.Any("to", val),
		)
		return
	}
	return

}
