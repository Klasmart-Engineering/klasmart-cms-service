package mutex

import (
	"context"
	"fmt"
	"sync"
	"time"

	locker "gitlab.badanamu.com.cn/calmisland/distributed_lock"
	"gitlab.badanamu.com.cn/calmisland/distributed_lock/drivers"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
)

const (
	defaultLockTimeout = time.Minute * 3
)

func NewLock(ctx context.Context, key string, params ...interface{}) (sync.Locker, error) {
	lockKey := key
	for i := range params {
		lockKey = fmt.Sprintf("%v:%v", lockKey, params[i])
	}

	return locker.NewDistributedLock(locker.DistributedLockConfig{
		Driver:  "redis",
		Key:     lockKey,
		Timeout: defaultLockTimeout,
		Ctx:     ctx,
		RedisConfig: drivers.RedisConfig{
			Host:     config.Get().RedisConfig.Host,
			Port:     config.Get().RedisConfig.Port,
			Password: config.Get().RedisConfig.Password,
		},
	})
}
