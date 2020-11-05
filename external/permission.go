package external

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type PermisionServiceProvider interface {
	HasPermision(ctx context.Context, operator *entity.Operator, permissionName PermissionName) (bool, error)
	GetHasPermissionOrganizations(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error)
}

var (
	_amsPermisionService *AmsPermisionService
	_amsPermisionOnce    sync.Once
)

func GetPermisionServiceProvider() PermisionServiceProvider {
	_amsPermisionOnce.Do(func() {
		_amsPermisionService = &AmsPermisionService{
			client: chlorine.NewClient(config.Get().AMS.EndPoint),
		}
	})

	return _amsPermisionService
}

type AmsPermisionService struct {
	client *chlorine.Client
}

func (s AmsPermisionService) HasPermision(ctx context.Context, operator *entity.Operator, permissionName PermissionName) (bool, error) {
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

func (s AmsPermisionService) GetHasPermissionOrganizations(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error) {
	// TODO: add implement
	return nil, nil
}
