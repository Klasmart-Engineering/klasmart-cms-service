package da

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestLazyRefreshCache_Get(t *testing.T) {
	ctx := context.Background()
	type request struct {
		A string
		B sql.NullInt64
	}

	type response struct {
		AAA string
		BBB int64
	}

	cast := time.Second * 1

	cache, _ := NewLazyRefreshCache(&LazyRefreshCacheOption{
		RedisKeyPrefix:  RedisKeyPrefixContentFolderQuery,
		RefreshDuration: time.Second * 5,
		RawQuery: func(ctx context.Context, input interface{}) (interface{}, error) {
			condition := input.(*request)

			// time cast
			time.Sleep(cast)
			return &response{
				AAA: condition.A,
				BBB: condition.B.Int64,
			}, nil
		}})

	request1 := &request{A: "abc", B: sql.NullInt64{Int64: 3, Valid: true}}
	response1 := &response{}
	err := cache.Get(ctx, request1, response1)
	if err != nil && err != redis.Nil {
		t.Errorf("get cache data failed due to %v", err)
	}

	time.Sleep(cast)

	response2 := &response{}
	err = cache.Get(ctx, request1, response2)
	if err != nil {
		t.Errorf("get cache data failed due to %v", err)
	}

	want := &response{
		AAA: request1.A,
		BBB: request1.B.Int64,
	}

	if !reflect.DeepEqual(response2, want) {
		t.Errorf("get invalid cache data, want %+v, get %+v", want, response2)
	}
}

func BenchmarkLazyRefreshCache_Get(b *testing.B) {
	ctx := context.Background()
	type request struct {
		A string
		B sql.NullInt64
	}

	type response struct {
		AAA string
		BBB int64
	}

	cast := time.Second * 1

	cache, _ := NewLazyRefreshCache(&LazyRefreshCacheOption{
		RedisKeyPrefix:  RedisKeyPrefixContentFolderQuery,
		RefreshDuration: time.Second * 5,
		RawQuery: func(ctx context.Context, input interface{}) (interface{}, error) {
			condition := input.(*request)

			// time cast
			time.Sleep(cast)
			return &response{
				AAA: condition.A,
				BBB: condition.B.Int64,
			}, nil
		}})

	request1 := &request{A: "abc", B: sql.NullInt64{Int64: 3, Valid: true}}
	response1 := &response{}
	// err := cache.Get(ctx, request1, response1)
	// if err != nil && err != redis.Nil {
	// 	b.Errorf("get cache data failed due to %v", err)
	// }

	want := &response{
		AAA: request1.A,
		BBB: request1.B.Int64,
	}

	for n := 0; n < b.N; n++ {
		err := cache.Get(ctx, request1, response1)
		if err != nil {
			b.Errorf("get cache data failed due to %v", err)
		}

		if !reflect.DeepEqual(response1, want) {
			b.Errorf("get invalid cache data, want %+v, get %+v", want, response1)
		}
	}
}
