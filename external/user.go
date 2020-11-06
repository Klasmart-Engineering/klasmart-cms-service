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
	GetUserInfoByID(ctx context.Context, userID string) (*User, error)
	BatchGet(ctx context.Context, ids []string) ([]*User, error)
}

type User struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	OrgID    string `json:"org_id"`
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

func (s AmsUserService) GetUserInfoByID(ctx context.Context, userID string) (*User, error) {
	request := chlorine.NewRequest(`
	query user($userID: ID!){
		user(user_id:$userID){
			user_id
			user_name
			my_organization {
		  		organization_id
			}
		}
	}`)
	request.Var("userID", userID)

	user := &struct {
		User struct {
			UserID         string `json:"user_id"`
			UserName       string `json:"user_name"`
			MyOrganization struct {
				OrganizationID string `json:"organization_id"`
			} `json:"my_organization"`
		} `json:"user"`
	}{}

	response := &chlorine.Response{
		Data: user,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get user by id failed", log.String("userID", userID))
		return nil, err
	}

	return &User{
		UserID:   user.User.UserID,
		UserName: user.User.UserName,
		OrgID:    user.User.MyOrganization.OrganizationID,
	}, nil
}

func (s AmsUserService) BatchGet(ctx context.Context, ids []string) ([]*User, error) {
	if len(ids) == 0 {
		return []*User{}, nil
	}

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range ids {
		fmt.Fprintf(sb, "u%d: user(user_id: \"%s\") {user_id user_name my_organization { organization_id } }\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String())

	data := map[string]struct {
		UserID         string `json:"user_id"`
		UserName       string `json:"user_name"`
		MyOrganization struct {
			OrganizationID string `json:"organization_id"`
		} `json:"my_organization"`
	}{}

	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get users by ids failed", log.Strings("ids", ids))
		return nil, err
	}

	var queryAlias string
	users := make([]*User, 0, len(data))
	for index := range ids {
		queryAlias = fmt.Sprintf("u%d", index)
		user, found := data[queryAlias]
		if !found {
			return nil, constant.ErrRecordNotFound
		}

		users = append(users, &User{
			UserID:   user.UserID,
			UserName: user.UserName,
		})
	}

	return users, nil
}
