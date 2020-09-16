package external

import "context"

type UserServiceProvider interface {
	GetUserInfoByID(ctx context.Context, userID string) (*UserInfo, error)
	//BatchGet(ctx context.Context, ids []string) ([]*Organization, error)
	//GetMine(ctx context.Context, userID string) ([]*Organization, error)
	//GetParents(ctx context.Context, orgID string) ([]*Organization, error)
	//GetChildren(ctx context.Context, orgID string) ([]*Organization, error)
}

type UserInfo struct {
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
	OrgID     string `json:"org_id"`
	OrgType   string `json:"org_type"`
}

func GetUserServiceProvider() UserServiceProvider {
	return &mockUserService{}
}

type mockUserService struct{}

func (s mockUserService) GetUserInfoByID(ctx context.Context, userID string) (*UserInfo, error) {
	return GetMockData().Users[0], nil
}
