package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"sync"
)
type IFolderPermissionModel interface {
	CheckFolderOperatorPermission(ctx context.Context, op *entity.Operator) (bool, error)
	CheckShareFolderOperatorPermission(ctx context.Context, op *entity.Operator) (bool, error)
}

type FolderPermissionModel struct {

}

//PublishFeaturedContentForAllHub

func (s *FolderPermissionModel) CheckShareFolderOperatorPermission(ctx context.Context, op *entity.Operator) (bool, error) {
	orgInfo, err := GetOrganizationPropertyModel().GetOrDefault(ctx, op.OrgID)
	if err != nil {
		log.Warn(ctx, "parse get folder shared records params failed",
			log.Err(err),
			log.String("orgID", op.OrgID))
		return false, err
	}
	if orgInfo.Type != entity.OrganizationTypeHeadquarters {
		log.Info(ctx, "org is not in head quarter",
			log.Any("orgInfo", orgInfo))
		return false, nil
	}

	permission := external.PublishFeaturedContentForAllHub
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permission)
	if err != nil {
		log.Error(ctx, "get permission failed", log.Err(err))
		return false, err
	}
	//有permission，直接返回
	if hasPermission {
		return true, nil
	}
	return false, nil
}

func (s *FolderPermissionModel) CheckFolderOperatorPermission(ctx context.Context, op *entity.Operator) (bool, error) {
	permission := external.CreateFolder289
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permission)
	if err != nil {
		log.Error(ctx, "get permission failed", log.Err(err))
		return false, err
	}
	//有permission，直接返回
	if hasPermission {
		return true, nil
	}
	return false, nil
}

var (
	_folderPermissionModel     IFolderPermissionModel
	_folderPermissionModelOnce sync.Once
)

func GetFolderPermissionModel() IFolderPermissionModel {
	_folderPermissionModelOnce.Do(func() {
		_folderPermissionModel = new(FolderPermissionModel)
	})
	return _folderPermissionModel
}
