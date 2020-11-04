package external

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

type UserServiceProvider interface {
	GetUserInfoByID(ctx context.Context, userID string) (*User, error)
}

type User struct {
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
	OrgID     string `json:"org_id"`
	OrgType   string `json:"org_type"`
}

func GetUserServiceProvider() UserServiceProvider {
	return &AmsUserService{}
}

type AmsUserService struct {
	client *chlorine.Client
}

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

	_, err := s.client.Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get user by id failed", log.String("userID", userID))
		return nil, err
	}

	return &User{
		UserID: user.User.UserID,
		Name:   user.User.UserName,
		OrgID:  user.User.MyOrganization.OrganizationID,
	}, nil
}
