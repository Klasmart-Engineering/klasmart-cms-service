package kl2cache

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"

	"github.com/google/uuid"
)

func Test_redisProvider_Get(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, OptEnable(true), OptRedis("127.0.0.1", "6379", ""))
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
	err = DefaultProvider.Get(ctx, KeyByStrings{
		"HasPermission",
		id,
	}, r, func(ctx context.Context) (val interface{}, err error) {
		val = &User{
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
	err := Init(ctx, OptEnable(true), OptRedis("127.0.0.1", "6379", ""))
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
	op := &entity.Operator{}
	var keys []Key
	for _, id := range ids {
		keys = append(keys, KeyByStrings{
			"HasOrganizationPermission",
			op.OrgID,
			op.UserID,
			id,
		})
	}

	err = DefaultProvider.BatchGet(ctx, keys, rs, func(ctx context.Context, keys []Key) (val map[string]interface{}, err error) {
		val = map[string]interface{}{}
		for _, k := range keys {
			key := k.(KeyByStrings)
			orgID := key[1]
			userID := key[2]
			id := key[3]

			val[key.Key()] = &User{
				ID:   id,
				Name: orgID + userID,
			}
		}
		return
	})
	if err != nil {
		return
	}
	fmt.Println(*rs)
	return
}

func TestMultiExec(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, OptEnable(true), OptStrategyFixed(time.Minute), OptRedis("127.0.0.1", "6379", ""))
	if err != nil {
		panic(err)
	}

	c := DefaultProvider.(*redisProvider).Client
	tx := c.Pipeline()
	tx.Set("1", 1, time.Minute)
	tx.Set("3", 3, time.Minute)
	tx.Set("2", 2, time.Minute)
	cmds, err := tx.Exec()

	fmt.Println(cmds, err)
	for _, cmd := range cmds {
		fmt.Println(cmd.Err())
	}
}
