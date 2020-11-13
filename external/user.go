package external

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type UserServiceProvider interface {
	Get(ctx context.Context, id string) (*User, error)
	BatchGet(ctx context.Context, ids []string) ([]*NullableUser, error)
	Query(ctx context.Context, organizationID, keyword string) ([]*User, error)
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NullableUser struct {
	Valid bool `json:"-"`
	*User
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

func (s AmsUserService) Get(ctx context.Context, id string) (*User, error) {
	users, err := s.BatchGet(ctx, []string{id})
	if err != nil {
		return nil, err
	}

	if !users[0].Valid {
		return nil, constant.ErrRecordNotFound
	}

	return users[0].User, nil
}

func (s AmsUserService) BatchGet(ctx context.Context, ids []string) ([]*NullableUser, error) {
	if len(ids) == 0 {
		return []*NullableUser{}, nil
	}

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range ids {
		fmt.Fprintf(sb, "q%d: user(user_id: \"%s\") {id:user_id name:user_name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String())

	data := map[string]*User{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get users by ids failed", log.Err(err), log.Strings("ids", ids))
		return nil, err
	}

	users := make([]*NullableUser, 0, len(data))
	for index := range ids {
		user := data[fmt.Sprintf("q%d", index)]
		users = append(users, &NullableUser{
			Valid: user != nil,
			User:  user,
		})
	}

	return users, nil
}

func (s AmsUserService) Query(ctx context.Context, organizationID, keyword string) ([]*User, error) {
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
				}
			}
		}
	}`)
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

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get users by keyword failed",
			log.Err(err),
			log.String("organizationID", organizationID),
			log.String("keyword", keyword))
		return nil, err
	}

	users := make([]*User, 0, len(data.Organization.FindMembers))
	for _, member := range data.Organization.FindMembers {
		users = append(users, member.User)
	}

	return users, nil
}
