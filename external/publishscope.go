package external

import "context"

type PublishScopeProvider interface {
	DefaultPublishScope(ctx context.Context) string
	BatchGet(ctx context.Context, ids []string) ([]*PublishScope, error)
}
type PublishScope struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type mockPublishScopeService struct{}

func (m *mockPublishScopeService) DefaultPublishScope(ctx context.Context) string {
	return "default"
}

func (m *mockPublishScopeService) BatchGet(ctx context.Context, ids []string) ([]*PublishScope, error) {
	//TODO: to add publish scope
	return []*PublishScope{
		{
			ID:   "visibility settings1",
			Name: "School",
		},
		{
			ID:   "visibility settings2",
			Name: "Org",
		},
		{
			ID:   "default",
			Name: "Default",
		},
		{
			ID:   "all",
			Name: "All",
		},
	}, nil
}

func GetPublishScopeProvider() PublishScopeProvider {
	return &mockPublishScopeService{}
}
