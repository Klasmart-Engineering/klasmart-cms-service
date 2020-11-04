package external

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type PermisionServiceProvider interface {
	HasPermision(ctx context.Context, operator *entity.Operator, permissionName PermissionName) (bool, error)
	GetHasPermissionOrganizations(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error)
}

var (
	_permisionService *AmsPermisionService
	_permisionOnce    sync.Once
)

func GetPermisionServiceProvider() PermisionServiceProvider {
	_permisionOnce.Do(func() {
		_permisionService = &AmsPermisionService{
			client: chlorine.NewClient(config.Get().AMS.EndPoint),
		}
	})

	return _permisionService
}

type AmsPermisionService struct {
	client *chlorine.Client
}

func (s AmsPermisionService) HasPermision(ctx context.Context, operator *entity.Operator, permissionName PermissionName) (bool, error) {
	// TODO: add implement
	return false, nil
}

func (s AmsPermisionService) GetHasPermissionOrganizations(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error) {
	// TODO: add implement
	return nil, nil
}
