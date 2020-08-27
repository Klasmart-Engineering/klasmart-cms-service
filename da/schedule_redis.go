package da

import (
	"context"
	"sync"
)

type ScheduleRedisDA struct{}

func (r *ScheduleRedisDA) CreateSchedule(ctx context.Context) {

}

var (
	_scheduleRedisDA     *ScheduleRedisDA
	_scheduleRedisDAOnce sync.Once
)

func GetRedisScheduleDA() *ScheduleRedisDA {
	_scheduleRedisDAOnce.Do(func() {
		_scheduleRedisDA = new(ScheduleRedisDA)
	})
	return _scheduleRedisDA
}
