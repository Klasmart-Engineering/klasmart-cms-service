package external

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type RoleServiceProvider interface {
	GetRole(ctx context.Context, op *entity.Operator, roleName entity.RoleName) (role *entity.Role, err error)
}

var _roleOnce = sync.Once{}
var _amsRoleService RoleServiceProvider

func GetRoleServiceProvider() RoleServiceProvider {
	_roleOnce.Do(func() {
		if config.Get().AMS.UseDeprecatedQuery {
			_amsRoleService = &AmsRoleService{
				client: chlorine.NewClient(config.Get().AMS.EndPoint),
			}
		} else {
			_amsRoleService = &AmsRoleConnectionService{}
		}
	})

	return _amsRoleService
}

type AmsRoleService struct {
	client *chlorine.Client
}

func (a *AmsRoleService) GetRole(ctx context.Context, op *entity.Operator, roleName entity.RoleName) (role *entity.Role, err error) {
	q := fmt.Sprintf(`
query roles($name: String!){
  rolesConnection(  
    filter:{ 
      name:{
        operator:eq
        value: $name
      }
    }
    direction:FORWARD    
  ){
    totalCount
    edges{
      node{
        id
        name
        status
      }
    }
  }  
}
`)
	request := chlorine.NewRequest(q, chlorine.ReqToken(op.Token))
	request.Var("name", roleName)
	data := &struct {
		RolesConnection struct {
			Edges []struct {
				Node entity.Role `json:"node"`
			} `json:"edges"`
		} `json:"rolesConnection"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err = GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "GetRole failed",
			log.Err(err),
			log.Any("operator", op),
			log.Any("roleName", roleName))
		err = &entity.ExternalError{
			Err:  errors.New("response data contains err"),
			Type: constant.InternalErrorTypeAms,
		}
		return
	}
	if len(data.RolesConnection.Edges) == 0 {
		err = &entity.ExternalError{
			Err:  fmt.Errorf("role not found: %s", roleName),
			Type: constant.InternalErrorTypeAms,
		}
		log.Error(ctx, "getRole failed", log.Any("roleName", roleName), log.Any("op", op), log.Err(err))
		return
	}
	role = &data.RolesConnection.Edges[0].Node
	return
}
