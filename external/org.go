package external

import (
	"context"
	"fmt"

	cl "gitlab.badanamu.com.cn/calmisland/chlorine"
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
}

type mockOrganizationService struct{}

func (s mockOrganizationService) BatchGet(ctx context.Context, ids []string) ([]*Organization, error) {
	client := cl.NewClient(url)
	q := `query orgs($orgIDs: [ID!]){
	organizations(organization_ids: $orgIDs){
    	id: organization_id
    	Name: organization_name
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
	_, err := client.Run(ctx, req, &res)
	if err != nil {
		return nil, err
	}
	if len(res.Errors) > 0 {
		return nil, res.Errors
	}
	fmt.Println(payload)
	return payload, nil
	//return GetMockData().Organizations, nil
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
