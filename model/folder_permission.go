package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)
type IFolderPermissionModel interface {
	CheckFolderOperatorPermission(ctx context.Context, op *entity.Operator) (bool, error)
}

type FolderPermissionModel struct {

}

func (s *FolderPermissionModel) CheckFolderOperatorPermission(ctx context.Context, op *entity.Operator) (bool, error) {
	// TODO:Waiting for folder operate permission name
	// var permission external.PermissionName
	// hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permission)
	// if err != nil {
	// 	log.Error(ctx, "get permission failed", log.Err(err))
	// 	return false, err
	// }
	// //有permission，直接返回
	// if hasPermission {
	// 	return true, nil
	// }
	// return false, nil

	return true, nil
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
