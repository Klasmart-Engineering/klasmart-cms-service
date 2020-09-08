package external

import "context"

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

func GetOrganizationServiceProvider() (OrganizationServiceProvider, error) {
	return &mockOrganizationService{}, nil
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
