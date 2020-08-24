package external

import "context"

type ClassTypeServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*ClassType, error)
}

type ClassType struct {
	ID   string
	Name string
}

func GetClassTypeServiceProvider() (ClassTypeServiceProvider, error) {
	return &mockClassTypeService{}, nil
}

type mockClassTypeService struct{}

func (s mockClassTypeService) BatchGet(ctx context.Context, ids []string) ([]*ClassType, error) {
	return []*ClassType{
		{
			ID:   "ClassType-1",
			Name: "ClassType1",
		},
		{
			ID:   "ClassType-2",
			Name: "ClassType2",
		},
	}, nil
}
