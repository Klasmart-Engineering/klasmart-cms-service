package external

import "context"

type SkillServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Skill, error)
}

type Skill struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetSkillServiceProvider() (SkillServiceProvider, error) {
	return &mockSkillService{}, nil
}

type mockSkillService struct{}

func (s mockSkillService) BatchGet(ctx context.Context, ids []string) ([]*Skill, error) {
	return GetMockData().Skills, nil
}
