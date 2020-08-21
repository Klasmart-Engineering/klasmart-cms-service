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
	return []*Grade{
		{
			ID:   "Grade-1",
			Name: "Grade1",
		},
		{
			ID:   "Grade-2",
			Name: "Grade2",
		},
	}, nil
}
