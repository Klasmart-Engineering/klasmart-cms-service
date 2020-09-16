package external

import "context"

type ProgramServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Program, error)
}

type Program struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetProgramServiceProvider() (ProgramServiceProvider, error) {
	return &mockProgramService{}, nil
}

type mockProgramService struct{}

func (s mockProgramService) BatchGet(ctx context.Context, ids []string) ([]*Program, error) {
	var programs []*Program
	for _, option := range GetMockData().Options {
		programs = append(programs, option.Program)
	}
	return programs, nil
}
