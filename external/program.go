package external

import "context"

type ProgramServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Program, error)
}

type Program struct {
	ID   string
	Name string
}

func GetProgramServiceProvider() (ProgramServiceProvider, error) {
	return &mockProgramService{}, nil
}

type mockProgramService struct{}

func (s mockProgramService) BatchGet(ctx context.Context, ids []string) ([]*Program, error) {
	return GetMockData().Program, nil
}
