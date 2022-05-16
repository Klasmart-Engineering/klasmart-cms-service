package external

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/KL-Engineering/chlorine"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cache/cache"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
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
	GetOnlyUnderOrgUsers(ctx context.Context, op *entity.Operator, orgID string) ([]*User, error)
	GetUserCount(ctx context.Context, op *entity.Operator, cond *entity.GetUserCountCondition) (count int, err error)
}

type User struct {
	ID         string `json:"id"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
}

func (u User) Name() string {
	return u.GivenName + " " + u.FamilyName
}

type NullableUser struct {
	Valid bool   `json:"valid"`
	StrID string `json:"str_id"`
	*User
}

func (n *NullableUser) StringID() string {
	return n.StrID
}
func (n *NullableUser) RelatedIDs() []*cache.RelatedEntity {
	return nil
}

var (
	_amsUserService UserServiceProvider
	_amsUserOnce    sync.Once
)

func GetUserServiceProvider() UserServiceProvider {
	_amsUserOnce.Do(func() {
		if config.Get().AMS.UseDeprecatedQuery {
			_amsUserService = &AmsUserService{}
		} else {
			_amsUserService = &AmsUserConnectionService{}
		}
	})
	return _amsUserService
}

type AmsUserService struct{}

func (s AmsUserService) GetUserCount(ctx context.Context, op *entity.Operator, cond *entity.GetUserCountCondition) (count int, err error) {
	mFilter := map[string]interface{}{}
	var condFilters []interface{}
	if cond.OrgID.Valid {
		condFilters = append(condFilters, map[string]interface{}{
			"organizationId": map[string]interface{}{
				"operator": "eq",
				"value":    cond.OrgID.String,
			},
		})
	}
	if cond.RoleID.Valid {
		condFilters = append(condFilters, map[string]interface{}{
			"roleId": map[string]interface{}{
				"operator": "eq",
				"value":    cond.RoleID.String,
			},
		})
	}
	if cond.SchoolIDs.Valid {
		var condIDs []interface{}
		for _, schoolID := range cond.SchoolIDs.Strings {
			condIDs = append(condIDs, map[string]interface{}{
				"schoolId": map[string]interface{}{
					"operator": "eq",
					"value":    schoolID,
				},
			})
		}
		condFilters = append(condFilters, map[string]interface{}{
			"OR": condIDs,
		})
	}
	if cond.ClassIDs.Valid {
		var condIDs []interface{}
		for _, id := range cond.ClassIDs.Strings {
			condIDs = append(condIDs, map[string]interface{}{
				"classId": map[string]interface{}{
					"operator": "eq",
					"value":    id,
				},
			})
		}
		condFilters = append(condFilters, map[string]interface{}{
			"OR": condIDs,
		})
	}

	mFilter["AND"] = condFilters
	log.Info(ctx, "GetUserCount", log.Any("filter", mFilter))
	q := `
query users($filter:UserFilter) {
  usersConnection(
    filter:$filter
    direction:FORWARD
  ){
    totalCount    
  }  
}
`
	request := chlorine.NewRequest(q, chlorine.ReqToken(op.Token))
	request.Var("filter", mFilter)
	data := &struct {
		UsersConnection struct {
			TotalCount int `json:"totalCount"`
		} `json:"usersConnection"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err = GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "GetRole failed",
			log.Err(err),
			log.Any("operator", op),
			log.Any("filter", mFilter))
		err = &entity.ExternalError{
			Err:  errors.New("response data contains err"),
			Type: constant.InternalErrorTypeAms,
		}
		return
	}
	count = data.UsersConnection.TotalCount
	return
}

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
	log.Info(ctx, "Doing BatchGet user",
		log.Strings("ids", ids))

	uuids := make([]string, 0, len(ids))
	for _, id := range ids {
		if utils.IsValidUUID(id) {
			uuids = append(uuids, id)
		} else {
			log.Warn(ctx, "invalid uuid type", log.String("id", id))
		}
	}

	res := make([]*NullableUser, 0, len(uuids))
	err := cache.GetPassiveCacheRefresher().BatchGet(ctx, s.Name(), uuids, &res, operator)
	if err != nil {
		return nil, err
	}
	log.Info(ctx, "BatchGet user success",
		log.Strings("ids", uuids),
		log.Any("res", res))

	resultMap := make(map[string]*NullableUser)
	for i := range res {
		resultMap[res[i].StringID()] = res[i]
	}
	newResult := make([]*NullableUser, 0, len(ids))
	for i := range ids {
		obj := resultMap[ids[i]]
		if obj == nil {
			obj = &NullableUser{
				Valid: false,
				StrID: ids[i],
			}
		}
		newResult = append(newResult, obj)
	}

	return newResult, nil
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
		fmt.Fprintf(sb, "q%d: user(user_id: $user_id_%d) {id:user_id given_name family_name email avatar}\n", index, index)
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
		if user == nil {
			continue
		}

		users = append(users, &NullableUser{
			Valid: user != nil,
			User:  user,
			StrID: _ids[indexMapping[index]],
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
		dict[user.ID] = user.Name()
	}
	log.Debug(ctx, "BatchGetNameMap:dict", log.Any("dict", dict))
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
