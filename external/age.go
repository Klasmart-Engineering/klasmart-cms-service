package external

import "context"

type AgeServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Age, error)
}

type Age struct {
	ID   string
	Name string
}

func GetAgeServiceProvider() (AgeServiceProvider, error) {
	return &mockAgeService{}, nil
}

type mockAgeService struct{}

func (s mockAgeService) BatchGet(ctx context.Context, ids []string) ([]*Age, error) {
	return []*Age{
		{
			ID:   "Age-1",
			Name: "Age1",
		},
		{
			ID:   "Age-2",
			Name: "Age2",
		},
	}, nil
}
