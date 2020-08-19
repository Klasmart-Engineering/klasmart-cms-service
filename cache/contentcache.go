package cache

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IContentCache interface {
	SaveContentCacheList(ctx context.Context, content []*entity.ContentInfoWithDetails) error
	SaveContentCacheListBySearchCondition(ctx context.Context)
	GetContentCacheByIdList(ctx context.Context, ids []string)([]string, []*entity.ContentInfoWithDetails, error)
	GetContentCacheBySearchCondition(ctx context.Context, condition da.DyContentCondition) ([]*entity.ContentInfoWithDetails, error)

	CleanContentCache(ctx context.Context, ids []string) error
}
