package external

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type PermissionServiceProvider interface {
	HasPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) (bool, error)
	GetHasPermissionOrganizations(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error)
}

var (
	_amsPermissionService *AmsPermissionService
	_amsPermissionOnce    sync.Once
)

func GetPermissionServiceProvider() PermissionServiceProvider {
	_amsPermissionOnce.Do(func() {
		_amsPermissionService = &AmsPermissionService{
			client: chlorine.NewClient(config.Get().AMS.EndPoint),
		}
	})

	return _amsPermissionService
}

type AmsPermissionService struct {
	client *chlorine.Client
}

func (s AmsPermissionService) HasPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) (bool, error) {
	request := chlorine.NewRequest(`
	query(
		$user_id: ID! 
		$organization_id: ID!
		$permission_name: ID!
	) {
		user(user_id: $user_id) {
			membership(organization_id: $organization_id) {
				checkAllowed(permission_name: $permission_name)
			}
		}
	}`)
	request.Var("user_id", operator.UserID)
	request.Var("organization_id", operator.OrgID)
	request.Var("permission_name", permissionName.String())

	data := &struct {
		User struct {
			Membership struct {
				CheckAllowed bool `json:"checkAllowed"`
			} `json:"membership"`
		} `json:"user"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "check user permission failed",
			log.Any("operator", operator),
			log.String("permissionName", permissionName.String()))
		return false, err
	}

	return data.User.Membership.CheckAllowed, nil
}

func (s AmsPermissionService) GetHasPermissionOrganizations(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error) {
	request := chlorine.NewRequest(`
	query(
		$user_id: ID!
		$permission_name: ID!
	){
		user(user_id: $user_id) {
			memberships {
				organization{
					organization_id
					organization_name        
				}
				checkAllowed(permission_name: $permission_name)
			}
		}
	}`)
	request.Var("user_id", operator.UserID)
	request.Var("permission_name", permissionName.String())

	data := &struct {
		Memberships []struct {
			Organization struct {
				OrganizationID   string `json:"organization_id"`
				OrganizationName string `json:"organization_name"`
			} `json:"organization"`
			CheckAllowed bool `json:"checkAllowed"`
		} `json:"memberships"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "check user permission failed",
			log.Any("operator", operator),
			log.String("permissionName", permissionName.String()))
		return nil, err
	}

	orgs := make([]*Organization, 0, len(data.Memberships))
	for _, membership := range data.Memberships {
		if !membership.CheckAllowed {
			continue
		}

		orgs = append(orgs, &Organization{
			ID:   membership.Organization.OrganizationID,
			Name: membership.Organization.OrganizationName,
		})
	}

	return orgs, nil
}
