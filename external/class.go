package external

import "context"

type ClassServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Class, error)
}

type Class struct {
	ID   string
	Name string
}

func GetClassServiceProvider() (ClassServiceProvider, error) {
	return &mockClassService{}, nil
}

type mockClassService struct{}

func (s mockClassService) BatchGet(ctx context.Context, ids []string) ([]*Class, error) {
	return []*Class{
		{
			ID:   "Class-1",
			Name: "Class1",
		},
		{
			ID:   "Class-2",
			Name: "Class2",
		},
	}, nil
}
