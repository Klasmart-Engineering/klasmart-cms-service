package external

import (
	"context"
	"go.uber.org/zap/buffer"
	"sync"
	"text/template"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type PermissionServiceProvider interface {
	HasOrganizationPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) (bool, error)
	HasSchoolPermission(ctx context.Context, userID, schoolID string, permissionName PermissionName) (bool, error)
	HasPermissions(ctx context.Context, operator *entity.Operator, permissions []PermissionName)(map[PermissionName]bool, error)
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

func (s AmsPermissionService) HasOrganizationPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) (bool, error) {
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
			log.Err(err),
			log.Any("operator", operator),
			log.String("permissionName", permissionName.String()))
		return false, err
	}

	log.Debug(ctx, "check permission success",
		log.Any("operator", operator),
		log.String("permissionName", permissionName.String()),
		log.Bool("hasPermission", data.User.Membership.CheckAllowed))

	return data.User.Membership.CheckAllowed, nil
}

func (s AmsPermissionService) HasSchoolPermission(ctx context.Context, userID, schoolID string, permissionName PermissionName) (bool, error) {
	request := chlorine.NewRequest(`
	query(
		$user_id: ID! 
		$school_id: ID!
		$permission_name: ID!
	) {
		user(user_id: $user_id) {
			school_membership(school_id: $school_id) {
				checkAllowed(permission_name: $permission_name)
			}
		}
	}`)
	request.Var("user_id", userID)
	request.Var("school_id", schoolID)
	request.Var("permission_name", permissionName.String())

	data := &struct {
		User struct {
			SchoolMembership struct {
				CheckAllowed bool `json:"checkAllowed"`
			} `json:"school_membership"`
		} `json:"user"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "check user permission failed",
			log.Err(err),
			log.String("userID", userID),
			log.String("schoolID", schoolID),
			log.String("permissionName", permissionName.String()))
		return false, err
	}

	log.Debug(ctx, "check permission success",
		log.String("userID", userID),
		log.String("schoolID", schoolID),
		log.String("permissionName", permissionName.String()),
		log.Bool("hasPermission", data.User.SchoolMembership.CheckAllowed))

	return data.User.SchoolMembership.CheckAllowed, nil
}

func (s AmsPermissionService) HasPermissions(ctx context.Context, operator *entity.Operator, permissions []PermissionName)(map[PermissionName]bool, error){
	raw := `
query($user_id: ID! $organization_id: ID!) {
	user(user_id: $user_id) {
		membership(organization_id: $organization_id) {
			{{range $i, $e := .}}
			{{$e}}: checkAllowed(permission_name: "{{$e}}")
			{{end}}
		}
	}
}
`

	temp, err := template.New("Permissions").Parse(raw)
	if err != nil {
		log.Error(ctx, "temp error", log.String("raw", raw), log.Err(err))
		return nil, err
	}
	buf := buffer.Buffer{}
	err = temp.Execute(&buf, permissions)
	if err != nil {
		log.Error(ctx, "temp execute failed", log.String("raw", raw), log.Err(err))
		return nil, err
	}
	req := chlorine.NewRequest(buf.String())
	req.Var("user_id", operator.UserID)
	req.Var("organization_id", operator.OrgID)
	payload := make(map[PermissionName]bool, len(permissions))
	res := chlorine.Response{
		Data: &struct{
			User struct {Membership map[PermissionName]bool `json:"membership"`} `json:"user"`
		}{struct {Membership map[PermissionName]bool `json:"membership"`}{Membership: payload}},
	}

	_, err = GetChlorine().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", buf.String()), log.Any("res", res), log.Err(err))
		return nil, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", buf.String()), log.Any("res", res), log.Err(res.Errors))
		return nil, res.Errors
	}
	return payload, nil
}
