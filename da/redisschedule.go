package da

import (
	"context"
	"sync"
)

type RedisScheduleDA struct{}

func (r *RedisScheduleDA) CreateSchedule(ctx context.Context) {

}

var (
	_redisScheduleDA     *RedisScheduleDA
	_redisScheduleDAOnce sync.Once
)

func GetRedisScheduleDA() *RedisScheduleDA {
	_redisScheduleDAOnce.Do(func() {
		_redisScheduleDA = new(RedisScheduleDA)
	})
	return _redisScheduleDA
}
