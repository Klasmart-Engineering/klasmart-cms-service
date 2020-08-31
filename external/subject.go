package external

import "context"

type SubjectServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Subject, error)
}

type Subject struct {
	ID   string
	Name string
}

func GetSubjectServiceProvider() (SubjectServiceProvider, error) {
	return &mockSubjectService{}, nil
}

type mockSubjectService struct{}

func (s mockSubjectService) BatchGet(ctx context.Context, ids []string) ([]*Subject, error) {
	return GetMockData().Subject, nil
}
