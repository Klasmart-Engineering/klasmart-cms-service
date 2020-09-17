package external

import "context"

type GradeServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Grade, error)
}

type Grade struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetGradeServiceProvider() GradeServiceProvider {
	return &mockGradeService{}
}

type mockGradeService struct{}

func (s mockGradeService) BatchGet(ctx context.Context, ids []string) ([]*Grade, error) {
	var grades []*Grade
	for _, option := range GetMockData().Options {
		grades = append(grades, option.Grade...)
	}
	return grades, nil
}
