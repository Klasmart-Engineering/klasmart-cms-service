package mq

import (
	"context"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/imq"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
)

var (
	_mq     imq.IMessageQueue
	_mqOnce sync.Once
)

func GetMQ(ctx context.Context) (imq.IMessageQueue, error) {
	var err error
	_mqOnce.Do(func() {
		cfg := imq.Config{
			Drive:              "redis-list",
			RedisHost:          config.Get().RedisConfig.Host,
			RedisPort:          config.Get().RedisConfig.Port,
			RedisPassword:      config.Get().RedisConfig.Password,
			RedisHandlerThread: 1,
		}
		_mq, err = imq.CreateMessageQueue(cfg)
		if err != nil {
			log.Error(ctx, "create mq failed",
				log.Err(err),
				log.Any("config", cfg),
			)
		}
	})
	if err != nil {
		return nil, err
	}
	return _mq, nil
}

func MustGetMQ(ctx context.Context) imq.IMessageQueue {
	mq, err := GetMQ(ctx)
	if err != nil {
		panic(err)
	}
	return mq
}
