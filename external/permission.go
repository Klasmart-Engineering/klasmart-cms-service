package external

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"go.uber.org/zap/buffer"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type PermissionServiceProvider interface {
	HasOrganizationPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) (bool, error)
	HasSchoolPermission(ctx context.Context, operator *entity.Operator, schoolID string, permissionName PermissionName) (bool, error)
	HasAnyOrganizationPermission(ctx context.Context, operator *entity.Operator, orgIDs []string, permissionName PermissionName) (bool, error)
	HasAnySchoolPermission(ctx context.Context, operator *entity.Operator, schoolIDs []string, permissionName PermissionName) (bool, error)
	HasOrganizationPermissions(ctx context.Context, operator *entity.Operator, permissions []PermissionName) (map[PermissionName]bool, error)
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
	}`, chlorine.ReqToken(operator.Token))
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

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "check user org permission failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("permissionName", permissionName.String()))
		return false, err
	}

	log.Info(ctx, "check org permission success",
		log.Any("operator", operator),
		log.String("permissionName", permissionName.String()),
		log.Bool("hasPermission", data.User.Membership.CheckAllowed))

	return data.User.Membership.CheckAllowed, nil
}

func (s AmsPermissionService) HasSchoolPermission(ctx context.Context, operator *entity.Operator, schoolID string, permissionName PermissionName) (bool, error) {
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
	}`, chlorine.ReqToken(operator.Token))
	request.Var("user_id", operator.UserID)
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

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "check user school permission failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("schoolID", schoolID),
			log.String("permissionName", permissionName.String()))
		return false, err
	}

	log.Info(ctx, "check school permissions success",
		log.Any("operator", operator),
		log.String("schoolID", schoolID),
		log.String("permissionName", permissionName.String()),
		log.Bool("hasPermission", data.User.SchoolMembership.CheckAllowed))

	return data.User.SchoolMembership.CheckAllowed, nil
}

func (s AmsPermissionService) HasOrganizationPermissions(ctx context.Context, operator *entity.Operator, permissionNames []PermissionName) (map[PermissionName]bool, error) {
	if len(permissionNames) == 0 {
		return map[PermissionName]bool{}, nil
	}

	pns := make([]string, len(permissionNames))
	for index, permissionName := range permissionNames {
		pns[index] = permissionName.String()
	}

	_permissionNames, indexMapping := utils.SliceDeduplicationMap(pns)

	sb := new(strings.Builder)
	sb.WriteString("query($user_id: ID! $organization_id: ID!) {user(user_id: $user_id) {membership(organization_id: $organization_id) {")
	for index, permissionName := range _permissionNames {
		fmt.Fprintf(sb, "q%d: checkAllowed(permission_name: \"%s\")\n", index, permissionName)
	}
	sb.WriteString("}}}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	request.Var("user_id", operator.UserID)
	request.Var("organization_id", operator.OrgID)

	data := make(map[PermissionName]bool, len(permissionNames))
	response := &chlorine.Response{
		Data: &struct {
			User struct {
				Membership map[PermissionName]bool `json:"membership"`
			} `json:"user"`
		}{struct {
			Membership map[PermissionName]bool `json:"membership"`
		}{Membership: data}},
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "check org permissions success failed", log.Err(err), log.Any("permissionNames", permissionNames))
		return nil, err
	}

	permissions := make(map[PermissionName]bool, len(data))
	for index, permissionName := range permissionNames {
		permissions[permissionName] = data[PermissionName(fmt.Sprintf("q%d", indexMapping[index]))]
	}

	log.Info(ctx, "check org permissions success",
		log.Any("operator", operator),
		log.Any("permissionNames", permissionNames),
		log.Any("permissions", permissions))

	return permissions, nil
}

func (s AmsPermissionService) HasAnyOrganizationPermission(ctx context.Context, operator *entity.Operator, orgIDs []string, permissionName PermissionName) (bool, error) {
	if len(orgIDs) == 0 {
		return false, nil
	}
	raw := `
query($user_id: ID!, $permission: ID!) {
	user(user_id: $user_id) {
		{{range $i, $e := .}}
		index_{{$i}}: membership(organization_id: "{{$e}}") {
			checkAllowed(permission_name: $permission)
		}
		{{end}}
	}
}
`

	temp, err := template.New("Permissions").Parse(raw)
	if err != nil {
		log.Error(ctx, "temp error", log.String("raw", raw), log.Err(err))
		return false, err
	}
	buf := buffer.Buffer{}
	err = temp.Execute(&buf, orgIDs)
	if err != nil {
		log.Error(ctx, "temp execute failed", log.String("raw", raw), log.Err(err))
		return false, err
	}
	req := chlorine.NewRequest(buf.String(), chlorine.ReqToken(operator.Token))
	req.Var("user_id", operator.UserID)
	req.Var("permission", permissionName)

	payload := make(map[string]struct {
		CheckAllowed bool `json:"checkAllowed"`
	}, len(orgIDs))
	res := chlorine.Response{
		Data: &struct {
			User map[string]struct {
				CheckAllowed bool `json:"checkAllowed"`
			} `json:"user"`
		}{User: payload},
	}

	_, err = GetAmsClient().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", buf.String()), log.Any("res", res), log.Err(err))
		return false, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", buf.String()), log.Any("res", res), log.Err(res.Errors))
		return false, res.Errors
	}

	log.Info(ctx, "check org permissions success",
		log.Any("operator", operator),
		log.Strings("org", orgIDs),
		log.Any("hasPermission", payload))

	for _, v := range payload {
		if v.CheckAllowed == true {
			return true, nil
		}
	}
	return false, nil
}

func (s AmsPermissionService) HasAnySchoolPermission(ctx context.Context, operator *entity.Operator, schoolIDs []string, permissionName PermissionName) (bool, error) {
	if len(schoolIDs) == 0 {
		return false, nil
	}
	raw := `
query($user_id: ID!, $permission: ID!) {
	user(user_id: $user_id) {
		{{range $i, $e := .}}
		index_{{$i}}: school_membership(school_id: "{{$e}}") {
			checkAllowed(permission_name: $permission)
		}
		{{end}}
	}
}
`

	temp, err := template.New("Permissions").Parse(raw)
	if err != nil {
		log.Error(ctx, "temp error", log.String("raw", raw), log.Err(err))
		return false, err
	}
	buf := buffer.Buffer{}
	err = temp.Execute(&buf, schoolIDs)
	if err != nil {
		log.Error(ctx, "temp execute failed", log.String("raw", raw), log.Err(err))
		return false, err
	}
	req := chlorine.NewRequest(buf.String(), chlorine.ReqToken(operator.Token))
	req.Var("user_id", operator.UserID)
	req.Var("permission", permissionName)

	payload := make(map[string]struct {
		CheckAllowed bool `json:"checkAllowed"`
	}, len(schoolIDs))
	res := chlorine.Response{
		Data: &struct {
			User map[string]struct {
				CheckAllowed bool `json:"checkAllowed"`
			} `json:"user"`
		}{User: payload},
	}

	_, err = GetAmsClient().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", buf.String()), log.Any("res", res), log.Err(err))
		return false, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", buf.String()), log.Any("res", res), log.Err(res.Errors))
		return false, res.Errors
	}

	log.Info(ctx, "check org permissions success",
		log.Any("operator", operator),
		log.Strings("org", schoolIDs),
		log.Any("hasPermission", payload))

	for _, v := range payload {
		if v.CheckAllowed == true {
			return true, nil
		}
	}
	return false, nil
}
