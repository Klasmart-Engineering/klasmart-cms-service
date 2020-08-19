package cache

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"time"
)

type IContentCache interface {
	SaveContentCacheList(ctx context.Context, content []*entity.ContentInfoWithDetails) error
	SaveContentCacheListBySearchCondition(ctx context.Context, condition da.DyContentCondition, content []*entity.ContentInfoWithDetails) error
	GetContentCacheByIdList(ctx context.Context, ids []string)([]string, []*entity.ContentInfoWithDetails, error)
	GetContentCacheBySearchCondition(ctx context.Context, condition da.DyContentCondition) ([]*entity.ContentInfoWithDetails, error)

	CleanContentCache(ctx context.Context, ids []string) error

	SetExpiration(t time.Duration)
}

type RedisContentCache struct{
	expiration time.Duration
}

func (r *RedisContentCache) contentKey(id string) string {
	return fmt.Sprintf("kidsloop2.content.id.%v", id)
}
func (r *RedisContentCache) contentConditionKey(condition da.DyContentCondition) string {
	
}

func (r *RedisContentCache) SaveContentCacheList(ctx context.Context, content []*entity.ContentInfoWithDetails) error {
	ro.MustGetRedis(ctx).SetNX()
}

func (r *RedisContentCache) SaveContentCacheListBySearchCondition(ctx context.Context, condition da.DyContentCondition, content []*entity.ContentInfoWithDetails) error {
	panic("implement me")
}

func (r *RedisContentCache) GetContentCacheByIdList(ctx context.Context, ids []string) ([]string, []*entity.ContentInfoWithDetails, error) {
	panic("implement me")
}

func (r *RedisContentCache) GetContentCacheBySearchCondition(ctx context.Context, condition da.DyContentCondition) ([]*entity.ContentInfoWithDetails, error) {
	panic("implement me")
}

func (r *RedisContentCache) CleanContentCache(ctx context.Context, ids []string) error {
	panic("implement me")
}
func (r *RedisContentCache) SetExpiration(t time.Duration){

}