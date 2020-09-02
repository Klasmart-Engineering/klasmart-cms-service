package external

import "context"

type GradeServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Grade, error)
}

type Grade struct {
	ID   string
	Name string
}

func GetGradeServiceProvider() (GradeServiceProvider, error) {
	return &mockGradeService{}, nil
}

type mockGradeService struct{}

func (s mockGradeService) BatchGet(ctx context.Context, ids []string) ([]*Grade, error) {
	return GetMockData().Grade, nil
}
