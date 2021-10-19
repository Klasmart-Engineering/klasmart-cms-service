package external

import (
	"context"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrNoOperatorInOptions       = errors.New("no operator in options")
	ErrInvalidOpteratorInOptions = errors.New("invalid operator in options")
)

type UserServiceProvider interface {
	cache.IDataSource
	Get(ctx context.Context, operator *entity.Operator, id string) (*User, error)
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableUser, error)
	BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableUser, error)
	BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
	Query(ctx context.Context, operator *entity.Operator, organizationID, keyword string) ([]*User, error)
	GetByOrganization(ctx context.Context, operator *entity.Operator, organizationID string) ([]*User, error)
	NewUser(ctx context.Context, operator *entity.Operator, email string) (string, error)
	FilterByPermission(ctx context.Context, operator *entity.Operator, userIDs []string, permissionName PermissionName) ([]string, error)
	GetOnlyUnderOrgUsers(ctx context.Context, op *entity.Operator, orgID string) ([]*User, error)
}

type User struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
}

type NullableUser struct {
	Valid bool `json:"-"`
	*User
}

func (n *NullableUser) StringID() string {
	return n.User.ID
}
func (n *NullableUser) RelatedIDs() []*cache.RelatedEntity {
	return nil
}

var (
	_amsUserService *AmsUserService
	_amsUserOnce    sync.Once
)

func GetUserServiceProvider() UserServiceProvider {
	_amsUserOnce.Do(func() {
		_amsUserService = &AmsUserService{}
	})

	return _amsUserService
}

type AmsUserService struct{}

func (s AmsUserService) Get(ctx context.Context, operator *entity.Operator, id string) (*User, error) {
	users, err := s.BatchGet(ctx, operator, []string{id})
	if err != nil {
		return nil, err
	}

	if users[0].User == nil || !users[0].Valid {
		return nil, constant.ErrRecordNotFound
	}

	return users[0].User, nil
}

//TODO:No Test Program
func (s AmsUserService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableUser, error) {
	res := make([]*NullableUser, 0, len(ids))
	err := cache.GetPassiveCacheRefresher().BatchGet(ctx, s.Name(), ids, &res, operator)
	if err != nil {
		return nil, err
	}
	log.Info(ctx, "BatchGet user success",
		log.Strings("ids", ids),
		log.Any("res", res))
	return res, nil
}

func (s AmsUserService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)
	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$user_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: user(user_id: $user_id_%d) {id:user_id name:user_name given_name family_name email avatar}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("user_id_%d", index), id)
	}

	data := map[string]*User{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get users by ids failed", log.Err(err), log.Strings("ids", ids))
		return nil, err
	}

	users := make([]cache.Object, 0, len(data))
	for index := range ids {
		user := data[fmt.Sprintf("q%d", indexMapping[index])]
		users = append(users, &NullableUser{
			Valid: user != nil,
			User:  user,
		})
	}

	log.Info(ctx, "get users by ids success",
		log.Strings("ids", ids),
		log.Any("users", users))

	return users, nil
}

func (s AmsUserService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableUser, error) {
	users, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]*NullableUser{}, err
	}
	log.Info(ctx, "BatchGetMap: BatchGet user success",
		log.Strings("ids", ids),
		log.Any("users", users))

	dict := make(map[string]*NullableUser, len(users))
	for _, user := range users {
		if user.User == nil || !user.Valid {
			continue
		}
		dict[user.ID] = user
	}

	return dict, nil
}

func (s AmsUserService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
	users, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]string{}, err
	}

	dict := make(map[string]string, len(users))
	for _, user := range users {
		if user.User == nil || !user.Valid {
			continue
		}
		dict[user.ID] = user.Name
	}

	return dict, nil
}

func (s AmsUserService) Query(ctx context.Context, operator *entity.Operator, organizationID, keyword string) ([]*User, error) {
	request := chlorine.NewRequest(`
	query(
		$organization_id: ID!
		$keyword: String!
	) {
		organization(organization_id: $organization_id) {
			findMembers(search_query: $keyword) {
				user{
					id: user_id
					name: user_name
					given_name
					family_name
					email
					avatar
				}
			}
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("organization_id", organizationID)
	request.Var("keyword", keyword)

	data := &struct {
		Organization struct {
			FindMembers []struct {
				User *User `json:"user"`
			} `json:"findMembers"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query users by keyword failed",
			log.Err(err),
			log.String("organizationID", organizationID),
			log.String("keyword", keyword))
		return nil, err
	}

	users := make([]*User, 0, len(data.Organization.FindMembers))
	for _, member := range data.Organization.FindMembers {
		users = append(users, member.User)
	}

	log.Info(ctx, "query users by keyword success",
		log.String("organizationID", organizationID),
		log.String("keyword", keyword),
		log.Any("users", users))

	return users, nil
}

func (s AmsUserService) GetByOrganization(ctx context.Context, operator *entity.Operator, organizationID string) ([]*User, error) {
	request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
		  memberships{
			user{
			  id: user_id
			  name: user_name
			  given_name
			  family_name
			  email
			  avatar
			}
		  }
		}
	  }`, chlorine.ReqToken(operator.Token))
	request.Var("organization_id", organizationID)

	data := &struct {
		Organization struct {
			Memberships []struct {
				User *User `json:"user"`
			} `json:"memberships"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query users by org failed",
			log.Err(err),
			log.String("organizationID", organizationID))
		return nil, err
	}

	users := make([]*User, 0, len(data.Organization.Memberships))
	for _, member := range data.Organization.Memberships {
		users = append(users, member.User)
	}

	log.Info(ctx, "query users by org success",
		log.String("organizationID", organizationID),
		log.Any("users", users))

	return users, nil
}

func (s AmsUserService) NewUser(ctx context.Context, operator *entity.Operator, email string) (string, error) {
	request := chlorine.NewRequest(`
	mutation newUser($email: String){
		newUser(email:$email){
			user_id
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("$email", email)

	data := &struct {
		NewUser struct {
			UserID string `json:"user_id"`
		} `json:"newUser"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query users by keyword failed",
			log.Err(err),
			log.String("email", email))
		return "", err
	}

	return data.NewUser.UserID, nil
}

//TODO:Test Failed
func (s AmsUserService) FilterByPermission(ctx context.Context, operator *entity.Operator, userIDs []string, permissionName PermissionName) ([]string, error) {
	if len(userIDs) == 0 {
		return []string{}, nil
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(userIDs)

	sb := new(strings.Builder)
	fmt.Fprintf(sb, "query ($organization_id: ID! $permission_name: ID! %s) {", utils.StringCountRange(ctx, "$user_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: user(user_id: $user_id_%d) {membership(organization_id: $organization_id) {checkAllowed(permission_name: $permission_name)}}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	request.Var("organization_id", operator.OrgID)
	request.Var("permission_name", permissionName.String())
	for index, id := range _ids {
		request.Var(fmt.Sprintf("user_id_%d", index), id)
	}

	data := map[string]*struct {
		Membership struct {
			CheckAllowed bool `json:"checkAllowed"`
		} `json:"membership"`
	}{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "filter users by permission failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("userIDs", userIDs),
			log.String("permission", permissionName.String()))
		return nil, err
	}

	filtered := make([]string, 0, len(userIDs))
	appended := make(map[string]bool, len(_ids))
	for index, UserID := range userIDs {
		if appended[UserID] {
			continue
		}

		user := data[fmt.Sprintf("q%d", indexMapping[index])]
		if user == nil || !user.Membership.CheckAllowed {
			continue
		}

		filtered = append(filtered, UserID)
		appended[UserID] = true
	}

	log.Info(ctx, "filter users by permission success",
		log.Any("operator", operator),
		log.Strings("userIDs", userIDs),
		log.String("permission", permissionName.String()),
		log.Strings("result", filtered))

	return filtered, nil
}

func (s AmsUserService) GetOnlyUnderOrgUsers(ctx context.Context, op *entity.Operator, orgID string) ([]*User, error) {
	userInfos, err := GetUserServiceProvider().GetByOrganization(ctx, op, orgID)
	if err != nil {
		log.Error(ctx, "GetUserServiceProvider.GetByOrganization error",
			log.String("org_id", orgID),
			log.Any("op", op),
		)
		return nil, err
	}
	userIDs := make([]string, len(userInfos))
	for i, item := range userInfos {
		userIDs[i] = item.ID
	}
	userSchoolMap, err := GetSchoolServiceProvider().GetByUsers(ctx, op, orgID, userIDs)
	if err != nil {
		log.Error(ctx, "GetSchoolServiceProvider.GetByUsers error",
			log.Any("op", op),
			log.String("org_id", orgID),
			log.Strings("userIDs", userIDs),
		)
		return nil, err
	}
	userClassMap, err := GetClassServiceProvider().GetByUserIDs(ctx, op, userIDs)
	if err != nil {
		log.Error(ctx, "GetClassServiceProvider.GetByUserIDs error", log.Any("op", op), log.Strings("userIDs", userIDs))
		return nil, err
	}
	result := make([]*User, 0)
	for _, item := range userInfos {
		if len(userSchoolMap[item.ID]) > 0 {
			continue
		}
		if len(userClassMap[item.ID]) > 0 {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func (s AmsUserService) Name() string {
	return "ams_user_service"
}

func optionsWithOperator(ctx context.Context, options ...interface{}) (*entity.Operator, error) {
	if len(options) < 1 {
		return nil, ErrNoOperatorInOptions
	}
	operator, ok := options[0].(*entity.Operator)
	if !ok {
		log.Error(ctx, "invalid options",
			log.Any("options", options))
		return nil, ErrInvalidOpteratorInOptions
	}
	return operator, nil
}
