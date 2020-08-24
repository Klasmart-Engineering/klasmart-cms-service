package external

import "context"

type SkillServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Skill, error)
}

type Skill struct {
	ID   string
	Name string
}

func GetSkillServiceProvider() (SkillServiceProvider, error) {
	return &mockSkillService{}, nil
}

type mockSkillService struct{}

func (s mockSkillService) BatchGet(ctx context.Context, ids []string) ([]*Skill, error) {
	return []*Skill{
		{
			ID:   "Skill-1",
			Name: "Skill1",
		},
		{
			ID:   "Skill-2",
			Name: "Skill2",
		},
	}, nil
}
