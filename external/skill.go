package external

import "context"

type SkillServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Skill, error)
}

type Skill struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetSkillServiceProvider() SkillServiceProvider {
	return &mockSkillService{}
}

type mockSkillService struct{}

func (s mockSkillService) BatchGet(ctx context.Context, ids []string) ([]*Skill, error) {
	var skills []*Skill
	for _, option := range GetMockData().Options {
		for _, development := range option.Developmental {
			skills = append(skills, development.Skill...)
		}
	}
	return skills, nil
}
