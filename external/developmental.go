package external

import "context"

type DevelopmentalServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Developmental, error)
}

type Developmental struct {
	ID   string
	Name string
}

func GetDevelopmentalServiceProvider() (DevelopmentalServiceProvider, error) {
	return &mockDevelopmentalService{}, nil
}

type mockDevelopmentalService struct{}

func (s mockDevelopmentalService) BatchGet(ctx context.Context, ids []string) ([]*Developmental, error) {
	return []*Developmental{
		{
			ID:   "Developmental-1",
			Name: "Developmental1",
		},
		{
			ID:   "Developmental-2",
			Name: "Developmental2",
		},
	}, nil
}
