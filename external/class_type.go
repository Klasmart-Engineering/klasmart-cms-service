package external

import "context"

type ClassTypeServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*ClassType, error)
}

type ClassType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetClassTypeServiceProvider() ClassTypeServiceProvider {
	return &mockClassTypeService{}
}

type mockClassTypeService struct{}

func (s mockClassTypeService) BatchGet(ctx context.Context, ids []string) ([]*ClassType, error) {
	return GetMockData().ClassTypes, nil
}
