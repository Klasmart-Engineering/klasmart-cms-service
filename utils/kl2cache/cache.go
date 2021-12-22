package kl2cache

import (
	"context"
	"strings"
)

type Provider interface {
	WithExpireStrategy(ctx context.Context, strategy ExpireStrategy) (provider Provider)
	Get(ctx context.Context, key Key, val interface{}, fGetData func(ctx context.Context, key Key) (val interface{}, err error)) (err error)
	BatchGet(ctx context.Context, keys []Key, val interface{}, fGetData func(ctx context.Context, keys []Key) (kvs []*KeyValue, err error)) (err error)
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

type KeyValue struct {
	Key Key
	Val interface{}
}
