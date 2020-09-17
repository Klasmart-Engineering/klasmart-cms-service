package external

import "context"

type AgeServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Age, error)
}

type Age struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetAgeServiceProvider() AgeServiceProvider {
	return &mockAgeService{}
}

type mockAgeService struct{}

func (s mockAgeService) BatchGet(ctx context.Context, ids []string) ([]*Age, error) {
	var ages []*Age
	for _, option := range GetMockData().Options {
		ages = append(ages, option.Age...)
	}
	return ages, nil
}
