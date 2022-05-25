package mutex

import (
	"context"
	"fmt"
	"github.com/newrelic/go-agent/v3/newrelic"
	"sync"
	"time"

	locker "github.com/KL-Engineering/distributed_lock"
	"github.com/KL-Engineering/distributed_lock/drivers"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
)

const (
	defaultLockTimeout = time.Minute * 3
)

// lockerDecorator adds trace between call lock and unlock
// user should guarantee that lock and unlock are called in pair to end the trace context
type lockerDecorator struct {
	inner locker.LockDriver
	ctx   context.Context
	seg   *newrelic.Segment
}

func (l *lockerDecorator) Lock() {
	txn := newrelic.FromContext(l.ctx)
	if txn != nil {
		l.seg = txn.StartSegment("distributed_lock")
	}
	l.inner.Lock()
}

func (l *lockerDecorator) Unlock() {
	if l.seg != nil {
		l.seg.End()
	}
	l.inner.Unlock()
}

func NewLock(ctx context.Context, key string, params ...interface{}) (sync.Locker, error) {
	lockKey := key
	for i := range params {
		lockKey = fmt.Sprintf("%v:%v", lockKey, params[i])
	}
	if config.Get().RedisConfig.Host == "" {
		//if redis hasn't config
		return &sync.Mutex{}, nil
	}

	inner, err := locker.NewDistributedLock(locker.DistributedLockConfig{
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
	return &lockerDecorator{inner: inner, ctx: ctx}, err
}
