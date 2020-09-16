package external

import "context"

type SubjectServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Subject, error)
}

type Subject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetSubjectServiceProvider() (SubjectServiceProvider, error) {
	return &mockSubjectService{}, nil
}

type mockSubjectService struct{}

func (s mockSubjectService) BatchGet(ctx context.Context, ids []string) ([]*Subject, error) {
	var subjects []*Subject
	for _, option := range GetMockData().Options {
		subjects = append(subjects, option.Subject...)
	}
	return subjects, nil
}
