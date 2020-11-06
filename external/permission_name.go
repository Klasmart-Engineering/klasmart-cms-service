package external

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type PermissionName string

func (p PermissionName) String() string {
	return string(p)
}

func (p PermissionName) Allow(ctx context.Context, operator *entity.Operator) (bool, error) {
	return GetPermissionServiceProvider().HasPermission(ctx, operator, p)
}

const (
	CreateContentPage201 PermissionName = "create_content_page_201"
)
