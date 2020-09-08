package external

import "context"

type DevelopmentalServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Developmental, error)
}

type Developmental struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetDevelopmentalServiceProvider() (DevelopmentalServiceProvider, error) {
	return &mockDevelopmentalService{}, nil
}

type mockDevelopmentalService struct{}

func (s mockDevelopmentalService) BatchGet(ctx context.Context, ids []string) ([]*Developmental, error) {
	return GetMockData().Developmentals, nil
}
