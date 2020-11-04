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
	// TODO: add implement
	return false, nil
}

func (s AmsPermisionService) GetHasPermissionOrganizations(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error) {
	// TODO: add implement
	return nil, nil
}
