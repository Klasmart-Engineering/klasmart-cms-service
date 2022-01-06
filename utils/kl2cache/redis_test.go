package kl2cache_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/kl2cache"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"

	"github.com/google/uuid"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func Test_redisProvider_Get(t *testing.T) {
	ctx := context.Background()
	err := kl2cache.Init(
		ctx,
		kl2cache.OptEnable(true),
		kl2cache.OptRedis("127.0.0.1", 6379, ""),
		kl2cache.OptStrategyFixed(time.Minute),
	)
	if err != nil {
		panic(err)
	}
	id := uuid.NewString()
	err = getUser(ctx, id)
	if err != nil {
		panic(err)
	}
	err = getUser(ctx, id)
	if err != nil {
		panic(err)
	}

}
func getUser(ctx context.Context, id string) (err error) {
	r := &User{}
	err = kl2cache.DefaultProvider.Get(ctx, kl2cache.KeyByStrings{
		"HasPermission",
		id,
	}, r, func(ctx context.Context, key kl2cache.Key) (val interface{}, err error) {
		val = User{
			ID:   id,
			Name: "111",
		}
		return
	})
	fmt.Println(*r)
	return
}

func Test_redisProvider_BatchGet(t *testing.T) {
	ctx := context.Background()
	err := kl2cache.Init(ctx, kl2cache.OptEnable(true), kl2cache.OptRedis("127.0.0.1", 6379, ""))
	if err != nil {
		panic(err)
	}

	keys := []string{uuid.NewString(), uuid.NewString()}
	err = getUserByIds(ctx, keys)
	if err != nil {
		panic(err)
	}
	keys = append(keys, uuid.NewString())
	err = getUserByIds(ctx, keys)
	if err != nil {
		panic(err)
	}

}

type KeyHasPermission struct {
	Op   *entity.Operator
	Perm *entity.ContentPermission
}

func (k *KeyHasPermission) Key() (key string) {
	strs := []string{
		"HasOrganizationPerm",
		k.Op.OrgID,
		k.Op.UserID,
		k.Perm.ID,
	}
	key = strings.Join(strs, ":")
	return
}
func getUserByIds(ctx context.Context, ids []string) (err error) {
	rs := &[]*User{}
	op := &entity.Operator{
		OrgID:  "dawdawdwa",
		UserID: "xxxxxxxxxxxxxxxx",
	}
	var keys []kl2cache.Key
	for _, id := range ids {
		keys = append(keys, kl2cache.KeyByStrings{
			"HasOrganizationPermission",
			op.OrgID,
			op.UserID,
			id,
		})
	}

	fGetData := func(ctx context.Context, missedKeys []kl2cache.Key) (valArr []*kl2cache.KeyValue, err error) {
		for _, k := range missedKeys {
			key := k.(kl2cache.KeyByStrings)
			orgID := key[1]
			userID := key[2]
			id := key[3]

			valArr = append(valArr, &kl2cache.KeyValue{
				Key: key,
				Val: &User{
					ID:   id,
					Name: orgID + userID,
				},
			})
		}
		return
	}
	err = kl2cache.DefaultProvider.BatchGet(ctx, keys, rs, fGetData)
	if err != nil {
		return
	}
	fmt.Println(*rs)
	return
}

func Test_redisProvider_Get_NoCache(t *testing.T) {
	ctx := context.Background()
	err := kl2cache.Init(
		ctx,
		kl2cache.OptEnable(false),
	)
	if err != nil {
		fmt.Println(err)
	}
	id := uuid.NewString()
	err = getUser(ctx, id)
	if err != nil {
		fmt.Println(err)
	}
	err = getUser(ctx, id)
	if err != nil {
		fmt.Println(err)
	}

}

func Test_redisProvider_BatchGet_NoCache(t *testing.T) {
	ctx := context.Background()
	err := kl2cache.Init(ctx, kl2cache.OptEnable(
		false), kl2cache.OptRedis("127.0.0.1", 6379, ""))
	if err != nil {
		panic(err)
	}

	keys := []string{uuid.NewString(), uuid.NewString()}
	err = getUserByIds(ctx, keys)
	if err != nil {
		fmt.Println(err)
	}
	keys = append(keys, uuid.NewString())
	err = getUserByIds(ctx, keys)
	if err != nil {
		fmt.Println(err)
	}

}

func Test_redisProvider_Get_NoCache_Val_Nil(t *testing.T) {
	ctx := context.Background()
	err := kl2cache.Init(
		ctx,
		kl2cache.OptEnable(false),
	)
	if err != nil {
		fmt.Println(err)
	}
	id := uuid.NewString()
	err = getUser1(ctx, id)
	if err != nil {
		fmt.Println(err)
	}
	err = getUser1(ctx, id)
	if err != nil {
		fmt.Println(err)
	}

}

func getUser1(ctx context.Context, id string) (err error) {
	r := &User{}
	err = kl2cache.DefaultProvider.Get(ctx, kl2cache.KeyByStrings{
		"HasPermission",
		id,
	}, r, func(ctx context.Context, key kl2cache.Key) (val interface{}, err error) {
		return
	})
	fmt.Println(*r)
	return
}
