package external

import (
	"context"

	cl "gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

const url = "https://api.beta.kidsloop.net/user/"

type OrganizationServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Organization, error)
	GetMine(ctx context.Context, userID string) ([]*Organization, error)
	GetParents(ctx context.Context, orgID string) ([]*Organization, error)
	GetChildren(ctx context.Context, orgID string) ([]*Organization, error)
}

type Organization struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
}

func GetOrganizationServiceProvider() OrganizationServiceProvider {
	return &mockOrganizationService{}
	//return &graphqlOrganizationService{}
}

type graphqlOrganizationService struct{}

func (s graphqlOrganizationService) BatchGet(ctx context.Context, ids []string) ([]*Organization, error) {
	q := `query orgs($orgIDs: [ID!]){
	organizations(organization_ids: $orgIDs){
    	id: organization_id
    	name: organization_name
  	}
}`
	req := cl.NewRequest(q)
	req.Var("orgIDs", ids)
	payload := make([]*Organization, len(ids))
	res := cl.Response{
		Data: &struct {
			Organizations []*Organization
		}{Organizations: payload},
	}
	_, err := GetChlorine().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", q), log.Any("res", res), log.Err(err))
		return nil, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", q), log.Any("res", res), log.Err(res.Errors))
		return nil, res.Errors
	}
	return payload, nil
}

func (s graphqlOrganizationService) GetMine(ctx context.Context, userID string) ([]*Organization, error) {
	// Maybe don't need
	return GetMockData().Organizations, nil
}

func (s graphqlOrganizationService) GetParents(ctx context.Context, orgID string) ([]*Organization, error) {
	return []*Organization{}, nil
}

func (s graphqlOrganizationService) GetChildren(ctx context.Context, orgID string) ([]*Organization, error) {
	return []*Organization{}, nil
}

type mockOrganizationService struct{}

func (s mockOrganizationService) BatchGet(ctx context.Context, ids []string) ([]*Organization, error) {
	return GetMockData().Organizations, nil
}

func (s mockOrganizationService) GetMine(ctx context.Context, userID string) ([]*Organization, error) {
	return GetMockData().Organizations, nil
}

func (s mockOrganizationService) GetParents(ctx context.Context, orgID string) ([]*Organization, error) {
	return GetMockData().Organizations, nil
}

func (s mockOrganizationService) GetChildren(ctx context.Context, orgID string) ([]*Organization, error) {
	return GetMockData().Organizations, nil
}
