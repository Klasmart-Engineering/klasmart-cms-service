package mutex

import (
	"context"
	"fmt"
	"sync"
	"time"

	locker "github.com/KL-Engineering/distributed_lock"
	"github.com/KL-Engineering/distributed_lock/drivers"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
)

const (
	defaultLockTimeout = time.Minute * 3
)

func NewLock(ctx context.Context, key string, params ...interface{}) (sync.Locker, error) {
	lockKey := key
	for i := range params {
		lockKey = fmt.Sprintf("%v:%v", lockKey, params[i])
	}
	if config.Get().RedisConfig.Host == "" {
		//if redis hasn't config
		return &sync.Mutex{}, nil
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
