package kl2cache

import (
	"context"
	"time"
)

type VisitedStat struct {
	Key          string
	CreatedAt    int64
	MaxExpiredAt int64
}
type ExpireStrategy func(stat *VisitedStat) (expiration time.Duration)

func ExpireStrategyFixed(ex time.Duration) (es ExpireStrategy) {
	return func(stat *VisitedStat) (expiration time.Duration) {
		expiration = ex
		return
	}
}

type config struct {
	Enable bool
	Redis  struct {
		Host     string
		Port     int
		Password string
	}

	ExpireStrategy ExpireStrategy
}

var conf = &config{
	ExpireStrategy: ExpireStrategyFixed(time.Minute * 10),
}

type Option func(c *config)

func OptStrategyFixed(ex time.Duration) Option {
	return func(c *config) {
		c.ExpireStrategy = ExpireStrategyFixed(ex)
	}
}

func OptEnable(enable bool) Option {
	return func(c *config) {
		c.Enable = enable
	}
}

func Init(ctx context.Context, opts ...Option) (err error) {
	for _, opt := range opts {
		opt(conf)
	}
	err = initRedis(ctx, conf)
	if err != nil {
		return
	}
	return
}
