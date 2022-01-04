package kl2cache

import (
	"context"
	"strings"
)

type Provider interface {
	WithExpireStrategy(ctx context.Context, strategy ExpireStrategy) (provider Provider)
	// Get set the value of key into val
	//
	// fetch data by fGetData if missed
	//
	// note: input val of Get must has the same type of output val of fGetDate
	// val must be a pointer
	Get(ctx context.Context, key Key, val interface{}, fGetData func(ctx context.Context, key Key) (val interface{}, err error)) (err error)
	// BatchGet set the values of keys into val
	//
	// get values of keys from cache
	// for missed keys call fGetData to get values, and then set it into cache
	// set all values(cached and fetched) into val
	//
	// note: val must be a pointer of slice, and it's item must has the same type of KeyValue.Val
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
