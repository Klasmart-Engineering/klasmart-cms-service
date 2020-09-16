package external

import "context"

type DevelopmentalServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Developmental, error)
}

type Developmental struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Skill []*Skill `json:"skills"`
}

func GetDevelopmentalServiceProvider() DevelopmentalServiceProvider {
	return &mockDevelopmentalService{}
}

type mockDevelopmentalService struct{}

func (s mockDevelopmentalService) BatchGet(ctx context.Context, ids []string) ([]*Developmental, error) {
	var developments []*Developmental
	for _, option := range GetMockData().Options {
		developments = append(developments, option.Developmental...)
	}
	return developments, nil
}
