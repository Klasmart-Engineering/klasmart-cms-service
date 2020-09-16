package external

import "context"

type ClassServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Class, error)
}

type Class struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetClassServiceProvider() ClassServiceProvider {
	return &mockClassService{}
}

type mockClassService struct{}

func (s mockClassService) BatchGet(ctx context.Context, ids []string) ([]*Class, error) {
	return GetMockData().Classes, nil
}
