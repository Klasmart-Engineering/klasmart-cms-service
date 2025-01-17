package model

import (
	"context"
	"errors"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
)

var (
	ErrUnknownRegion = errors.New("unknown region")
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

	switch orgInfo.Region {
	case entity.Global:
		log.Info(ctx, "Check global org permission")
		permissions := []external.PermissionName{external.PublishFeaturedContentForAllHub, external.PublishFeaturedContentForAllOrgsHub}
		hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, permissions)
		if err != nil {
			log.Error(ctx, "get permission failed",
				log.Err(err))
			return false, err
		}
		//有permission，直接返回
		//has permission
		for k, v := range hasPermission {
			if !v {
				log.Warn(ctx, "No permission",
					log.String("Permission", string(k)))
				return false, nil
			}
		}
	case entity.VN:
		log.Info(ctx, "Check VN org permission")
		permission := external.PublishFeaturedContentForSpecificOrgsHub
		hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permission)
		if err != nil {
			log.Error(ctx, "get permission failed", log.Err(err))
			return false, err
		}
		//有permission，直接返回
		//has permission
		if !hasPermission {
			log.Warn(ctx, "No permission",
				log.String("Permission", string(external.PublishFeaturedContentForSpecificOrgsHub)))
			return false, nil
		}
	case entity.UnknownRegion:
		log.Warn(ctx, "unknown region", log.String("region", string(entity.UnknownRegion)))
		return false, ErrUnknownRegion
	}

	return true, nil
}

func (s *FolderPermissionModel) CheckFolderOperatorPermission(ctx context.Context, op *entity.Operator) (bool, error) {
	permission := external.CreateFolder289
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permission)
	if err != nil {
		log.Error(ctx, "get permission failed", log.Err(err))
		return false, err
	}
	//有permission，直接返回
	//has permission
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
