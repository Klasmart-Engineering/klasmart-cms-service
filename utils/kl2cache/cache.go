package kl2cache

import (
	"context"
	"strings"
)

type Provider interface {
	WithExpireStrategy(ctx context.Context, strategy ExpireStrategy) (provider Provider)
	Get(ctx context.Context, key Key, val interface{}, f func(ctx context.Context) (val interface{}, err error)) (err error)
	BatchGet(ctx context.Context, keys []Key, val interface{}, f func(ctx context.Context, keys []Key) (val map[string]interface{}, err error)) (err error)
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

//type Key interface {
//	Key() string
//	Arg(name string) interface{}
//}
//
//type stringKey struct {
//	key  []string
//	args map[string]interface{}
//}
//
//func (s *stringKey) Key() string {
//	return strings.Join(s.key, ":")
//}
//
//func (s *stringKey) Arg(name string) interface{} {
//	return s.args[name]
//}

//func NewStringKey(key []string, args map[string]interface{}) Key {
//	return &stringKey{
//		key:  key,
//		args: args,
//	}
//}

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"reflect"
//)
//
////
////type Key interface {
////	Key() string
////	Arg(name string) interface{}
////}
////type key struct {
////	key  string
////	args map[string]interface{}
////}
////
////func (k *key) Key() string {
////	return k.key
////}
////
////func (k *key) Arg(name string) interface{} {
////	return k.args[name]
////}
////
////type KeyOption func(key *key)
////
////func OptKeyArg(name string, val interface{}) KeyOption {
////	return func(key *key) {
////		key.args[name] = val
////	}
////}
////func NewKey(cacheKey string, opts ...KeyOption) Key {
////	return &key{
////		key:  cacheKey,
////		args: make(map[string]interface{}),
////	}
////}
//
////func Get(ctx context.Context, key Key, val interface{}, f func(ctx context.Context, key Key, val interface{}) (err error)) (err error) {
////	err = f(ctx, key, val)
////	return
////}
////
////func BatchGet(ctx context.Context, keys []Key, val interface{}, f func(ctx context.Context, keys []Key, val interface{}) (err error)) (err error) {
////	keys[0].Key() --> val[0]
////	err = f(ctx, keys, val)
////	return
////}
//
//type Key3333 struct {
//	CacheKey string
//	Args     map[string]interface{}
//}
//
//func Get1(ctx context.Context, key Key3333, fn func(ctx context.Context, key string) (val interface{}, err error)) (val interface{}, err error) {
//	typ := reflect.TypeOf(val)
//	val = reflect.New(typ).Interface()
//	err = json.Unmarshal([]byte(`{"name":"111"}`), val)
//	if err != nil {
//		return
//	}
//	val, err = fn(ctx, key)
//	if err != nil {
//		return
//	}
//
//	return
//}
//
//type User struct {
//	Name string
//}
//
//func init() {
//	user, err := Get1(context.Background(), []string{}, func(ctx context.Context, key string) (val interface{}, err error) {
//		val = &User{
//			Name: "dadwadwadw",
//		}
//		return
//	})
//	if err != nil {
//		panic(err)
//	}
//	u1 := user.(*User)
//	fmt.Println(u1.Name)
//}
//
//// val must be a slice
//func BatchGet1(ctx context.Context, keys []Key3333, fn func(ctx context.Context, keys [][]string) (val map[string]interface{}, err error)) (val interface{}, err error) {
//	typ := reflect.TypeOf(val)
//	val = reflect.New(typ).Interface()
//	err = json.Unmarshal([]byte(`{"name":"111"}`), val)
//	if err != nil {
//		return
//	}
//
//	// missed keys
//	vMap, err := fn(ctx, keys)
//	if err != nil {
//		return
//	}
//	for k, v := range vMap {
//		// cache k,v
//		fmt.Println(k, v)
//		reflect.Append(reflect.ValueOf(val), reflect.ValueOf(v))
//	}
//
//	return
//}
//
//type BatchKey struct {
//	Args map[string]interface{}
//	Keys []Key3333
//}
//
//func init() {
//	BatchGet1(context.Background(), []string{""}, func(ctx context.Context, keys BatchKey) (val map[string]interface{}, err error) {
//
//		users := []User{}
//		users = append(users, User{
//			Name: "1",
//		})
//		opID := keys.Args["op"].ID
//
//		users = append(users, User{
//			Name: "2",
//		})
//		for _, k := range keys {
//
//			k.Args["x"]
//		}
//
//		val[keys[0].CacheKey] = xx
//	})
//}
