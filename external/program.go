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
	return []*Program{
		{
			ID:   "Program-1",
			Name: "Program1",
		},
		{
			ID:   "Program-2",
			Name: "Program2",
		},
	}, nil
}
