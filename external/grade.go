package external

import "context"

type GradeServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Grade, error)
}

type Grade struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetGradeServiceProvider() (GradeServiceProvider, error) {
	return &mockGradeService{}, nil
}

type mockGradeService struct{}

func (s mockGradeService) BatchGet(ctx context.Context, ids []string) ([]*Grade, error) {
	return GetMockData().Grades, nil
}
