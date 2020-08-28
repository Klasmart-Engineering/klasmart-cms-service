package mutex

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/distributed_lock/drivers"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"sync"
	locker "gitlab.badanamu.com.cn/calmisland/distributed_lock"
	"time"
)

func NewLock(ctx context.Context, key string, params ...interface{}) (sync.Locker, error){
	lockKey := key + ":"
	for i := range params{
		lockKey = lockKey + ":" + fmt.Sprintf("%v", params[i])
	}

	return locker.NewDistributedLock(locker.DistributedLockConfig{
		Driver:      "redis",
		Key:         fmt.Sprintf("%v", lockKey),
		Timeout:     time.Minute * 3,
		Ctx:         ctx,
		RedisConfig: drivers.RedisConfig{
			Host:     config.Get().RedisConfig.Host,
			Port:     config.Get().RedisConfig.Port,
			Password: config.Get().RedisConfig.Password,
		},
	})
}
