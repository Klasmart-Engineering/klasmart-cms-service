package external

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type AMSService struct {
	client *chlorine.Client
}

var (
	_amsService     *AMSService
	_amsOnce        sync.Once
	_chlorineClient *chlorine.Client
	_chlorineOnce   sync.Once
)

func GetChlorine() *chlorine.Client {
	_chlorineOnce.Do(func() {
		_chlorineClient = chlorine.NewClient(config.Get().AMS.EndPoint)
	})
	return _chlorineClient
}

func GetAMSService() *AMSService {
	_amsOnce.Do(func() {
		_amsService = &AMSService{
			client: chlorine.NewClient(config.Get().AMS.EndPoint),
		}
	})

	return _amsService
}

func (s AMSService) HasPermision(ctx context.Context, operator *entity.Operator, permissionName PermissionName) (bool, error) {
	// TODO: add implement
	return false, nil
}

func (s AMSService) GetAccessibleOrganizations(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error) {
	// TODO: add implement
	return nil, nil
}

func (s AMSService) GetUserByID(ctx context.Context, id string) (*UserInfo, error) {
	// TODO: add implement
	return nil, nil
}
