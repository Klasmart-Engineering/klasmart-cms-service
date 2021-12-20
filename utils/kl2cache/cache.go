package kl2cache

import (
	"context"
	"strings"
)

type Provider interface {
	WithExpireStrategy(ctx context.Context, strategy ExpireStrategy) (provider Provider)
	Get(ctx context.Context, key Key, val interface{}, fGetData func(ctx context.Context) (val interface{}, err error)) (err error)
	BatchGet(ctx context.Context, keys []Key, val interface{}, fGetData func(ctx context.Context, keys []Key) (kvs []*KeyVal, err error)) (err error)
}

var DefaultProvider Provider

type Key interface {
	Key() string
}
type KeyByStrings []string

func (k KeyByStrings) Key() (key string) {
	key = strings.Join(k, ":")
	return
}

type KeyVal struct {
	Key Key
	Val interface{}
}
