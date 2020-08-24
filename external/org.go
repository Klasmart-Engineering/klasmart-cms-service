package external

import "context"

type OrganizationServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Organization, error)
	GetMine(ctx context.Context, userID string) ([]*Organization, error)
	GetParents(ctx context.Context, orgID string) ([]*Organization, error)
	GetChildren(ctx context.Context, orgID string) ([]*Organization, error)
}

type Organization struct {
	ID       string
	Name     string
	ParentID string
}

func GetOrganizationServiceProvider() (OrganizationServiceProvider, error) {
	return &mockOrganizationService{}, nil
}

type mockOrganizationService struct{}

func (s mockOrganizationService) BatchGet(ctx context.Context, ids []string) ([]*Organization, error) {
	return []*Organization{
		{
			ID:       "org-1",
			Name:     "org1",
			ParentID: "0",
		},
		{
			ID:       "org-2",
			Name:     "org2",
			ParentID: "0",
		},
	}, nil
}

func (s mockOrganizationService) GetMine(ctx context.Context, userID string) ([]*Organization, error) {
	return []*Organization{
		{
			ID:       "org-1",
			Name:     "org1",
			ParentID: "0",
		},
		{
			ID:       "org-2",
			Name:     "org2",
			ParentID: "0",
		},
	}, nil
}

func (s mockOrganizationService) GetParents(ctx context.Context, orgID string) ([]*Organization, error) {
	return []*Organization{
		{
			ID:       "org-1",
			Name:     "org1",
			ParentID: "0",
		},
		{
			ID:       "org-2",
			Name:     "org2",
			ParentID: "0",
		},
	}, nil
}

func (s mockOrganizationService) GetChildren(ctx context.Context, orgID string) ([]*Organization, error) {
	return []*Organization{
		{
			ID:       "org-1",
			Name:     "org1",
			ParentID: "0",
		},
		{
			ID:       "org-2",
			Name:     "org2",
			ParentID: "0",
		},
	}, nil
}
