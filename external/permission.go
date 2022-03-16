package external

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"gitlab.badanamu.com.cn/calmisland/ro"
)

const (
	RedisKeyPrefixOrgPermission      = "org_permission"
	RedisKeyPrefixOrgPermissionMutex = "org_permission:lock"
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
	// get permission from cache
	permissionMap, err := s.getOrganizationPermission(ctx, operator, []PermissionName{permissionName})
	if err == nil {
		log.Debug(ctx, "permission cache hit",
			log.Any("operator", operator),
			log.Any("permissionMap", permissionMap))
		return permissionMap[permissionName], nil
	}
	log.Warn(ctx, "s.getOrganizationPermission error",
		log.Err(err),
		log.Any("operator", operator),
		log.Any("permissionName", permissionName))

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

	_, err = GetAmsClient().Run(ctx, request, response)
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

//TODO:No Test Program
func (s AmsPermissionService) HasOrganizationPermissions(ctx context.Context, operator *entity.Operator, permissionNames []PermissionName) (map[PermissionName]bool, error) {
	if len(permissionNames) == 0 {
		return map[PermissionName]bool{}, nil
	}

	// get permission from cache
	permissionMap, err := s.getOrganizationPermission(ctx, operator, permissionNames)
	if err == nil {
		log.Debug(ctx, "permission cache hit",
			log.Any("operator", operator),
			log.Any("permissionNames", permissionNames))
		return permissionMap, nil
	}
	log.Warn(ctx, "s.getOrganizationPermission error",
		log.Err(err),
		log.Any("operator", operator),
		log.Any("permissionNames", permissionNames))

	pns := make([]string, len(permissionNames))
	for index, permissionName := range permissionNames {
		pns[index] = permissionName.String()
	}

	_permissionNames, indexMapping := utils.SliceDeduplicationMap(pns)

	sb := new(strings.Builder)
	fmt.Fprintf(sb, "query($user_id: ID! $organization_id: ID! %s) {user(user_id: $user_id) {membership(organization_id: $organization_id) {",
		utils.StringCountRange(ctx, "$permission_name_", ": ID!", len(_permissionNames)))

	for index := range _permissionNames {
		fmt.Fprintf(sb, "q%d: checkAllowed(permission_name: $permission_name_%d)\n", index, index)
	}
	sb.WriteString("}}}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	request.Var("user_id", operator.UserID)
	request.Var("organization_id", operator.OrgID)
	for index, id := range _permissionNames {
		request.Var(fmt.Sprintf("permission_name_%d", index), id)
	}

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

	_, err = GetAmsClient().Run(ctx, request, response)
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
	buf := bytes.Buffer{}
	err = temp.Execute(&buf, utils.SliceDeduplication(orgIDs))
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
		if v.CheckAllowed {
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
	buf := bytes.Buffer{}

	err = temp.Execute(&buf, utils.SliceDeduplication(schoolIDs))
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
		if v.CheckAllowed {
			return true, nil
		}
	}
	return false, nil
}

func (s AmsPermissionService) hasOrganizationPermissions(ctx context.Context, operator *entity.Operator, permissionNames []PermissionName) (map[PermissionName]bool, error) {
	if len(permissionNames) == 0 {
		return map[PermissionName]bool{}, nil
	}

	pns := make([]string, len(permissionNames))
	for index, permissionName := range permissionNames {
		pns[index] = permissionName.String()
	}

	_permissionNames, indexMapping := utils.SliceDeduplicationMap(pns)

	sb := new(strings.Builder)
	fmt.Fprintf(sb, "query($user_id: ID! $organization_id: ID! %s) {user(user_id: $user_id) {membership(organization_id: $organization_id) {",
		utils.StringCountRange(ctx, "$permission_name_", ": ID!", len(_permissionNames)))

	for index := range _permissionNames {
		fmt.Fprintf(sb, "q%d: checkAllowed(permission_name: $permission_name_%d)\n", index, index)
	}
	sb.WriteString("}}}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	request.Var("user_id", operator.UserID)
	request.Var("organization_id", operator.OrgID)
	for index, id := range _permissionNames {
		request.Var(fmt.Sprintf("permission_name_%d", index), id)
	}

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

func (s AmsPermissionService) getOrganizationPermission(ctx context.Context, operator *entity.Operator, permissionNames []PermissionName) (map[PermissionName]bool, error) {
	// get permission from cache
	permissionMap, err := s.getOrganizationPermissionCache(ctx, operator.UserID, operator.OrgID, permissionNames)
	if err == nil {
		log.Debug(ctx, "permission cache hit",
			log.Any("operator", operator),
			log.Any("permissionMap", permissionMap))
		return permissionMap, nil
	}

	if err == ro.ErrKeyNotExist {
		// TODO maybe cache breakdown
		// query user all permissions
		permissionMap, err := s.hasOrganizationPermissions(ctx, operator, AllPermissionNames)
		if err != nil {
			log.Error(ctx, "s.hasOrganizationPermissions error",
				log.Any("operator", operator),
				log.Any("permissionNames", permissionNames),
				log.Any("allPermissionNames", AllPermissionNames))
			return nil, errors.New("query all permissions failed")
		}

		// save to redis cache
		err = s.setOrganizationPermissionCache(ctx, operator.UserID, operator.OrgID, permissionMap)
		log.Debug(ctx, "s.setOrganizationPermissionCache result", log.Err(err))
		return permissionMap, nil
	}

	log.Error(ctx, "s.getOrganizationPermissionCache error",
		log.Err(err),
		log.Any("operator", operator),
		log.Any("permissionNames", permissionNames))

	return nil, err
}

func (s AmsPermissionService) getOrganizationPermissionCache(ctx context.Context, userID, orgID string, permissionNames []PermissionName) (map[PermissionName]bool, error) {
	if !config.Get().RedisConfig.OpenCache {
		log.Error(ctx, "redis cache is not open")
		return nil, errors.New("redis cache is not open")
	}

	key := fmt.Sprintf("%s:%s:%s", RedisKeyPrefixOrgPermission, orgID, userID)
	fields := make([]string, len(permissionNames))
	for i, permissionName := range permissionNames {
		fields[i] = string(permissionName)
	}

	redisClient := ro.MustGetRedis(ctx)
	pipe := redisClient.TxPipeline()
	exist := pipe.Exists(key)
	r := pipe.HMGet(key, fields...)
	_, err := pipe.Exec()
	if err != nil {
		log.Error(ctx, "failed to exec redis pipeline",
			log.Err(err),
			log.String("key", key),
			log.Strings("fields", fields),
		)
		return nil, err
	}

	log.Debug(ctx, "redis pipeline exec result",
		log.Any("exist", exist.Val()),
		log.Any("result", r.Val()))

	// key not exist
	if exist.Val() == int64(0) {
		return nil, ro.ErrKeyNotExist
	}

	result := make(map[PermissionName]bool, len(permissionNames))
	res := r.Val()
	for i, v := range res {
		resStr, ok := v.(string)
		if !ok {
			log.Warn(ctx, "invalid data from cache", log.Any("res", res), log.Any("permissionNames", permissionNames))
		}

		allowed, err := strconv.ParseBool(resStr)
		if err != nil {
			log.Warn(ctx, "strconv.ParseBool error",
				log.Err(err),
				log.String("resStr", resStr))
		}

		result[permissionNames[i]] = allowed
	}

	return result, nil
}

func (s AmsPermissionService) setOrganizationPermissionCache(ctx context.Context, userID, orgID string, permissionMap map[PermissionName]bool) error {
	if !config.Get().RedisConfig.OpenCache {
		log.Error(ctx, "redis cache is not open")
		return errors.New("redis cache is not open")
	}

	key := fmt.Sprintf("%s:%s:%s", RedisKeyPrefixOrgPermission, orgID, userID)
	fields := make(map[string]interface{}, len(permissionMap))
	for k, v := range permissionMap {
		fields[string(k)] = v
	}

	redisClient := ro.MustGetRedis(ctx)
	pipe := redisClient.TxPipeline()

	hmsetResult := pipe.HMSet(key, fields)
	expireResult := pipe.Expire(key, config.Get().User.PermissionCacheExpiration)
	_, err := pipe.Exec()
	if err != nil {
		log.Error(ctx, "failed to exec redis pipeline",
			log.Err(err),
			log.String("key", key),
			log.Any("fields", fields),
			log.Duration("expiration", config.Get().User.PermissionCacheExpiration))
		return err
	}

	log.Debug(ctx, "redis pipeline exec result",
		log.Any("hmsetResult", hmsetResult.Val()),
		log.Any("expireResult", expireResult.Val()))

	return nil
}
