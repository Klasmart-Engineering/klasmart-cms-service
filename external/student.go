package external

import "context"

type StudentServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Student, error)
}

type Student struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetStudentServiceProvider() StudentServiceProvider {
	return &mockStudentService{}
}

type mockStudentService struct{}

func (s mockStudentService) BatchGet(ctx context.Context, ids []string) ([]*Student, error) {
	return GetMockData().Students, nil
}
