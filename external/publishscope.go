package external

import "context"

type PublishScopeProvider interface {
	DefaultPublishScope(ctx context.Context) string
}

type mockPublishScopeService struct{}

func (m *mockPublishScopeService) DefaultPublishScope(ctx context.Context) string{
	return "default"
}

func GetPublishScopeProvider() (PublishScopeProvider, error) {
	return &mockPublishScopeService{}, nil
}