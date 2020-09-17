package external

import "context"

type TeacherServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Teacher, error)
	Query(ctx context.Context, keyword string) ([]*Teacher, error)
}

type Teacher struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetTeacherServiceProvider() TeacherServiceProvider {
	return &mockTeacherService{}
}

type mockTeacherService struct{}

func (s mockTeacherService) BatchGet(ctx context.Context, ids []string) ([]*Teacher, error) {
	return GetMockData().Teachers, nil
}

func (s mockTeacherService) Query(ctx context.Context, keyword string) ([]*Teacher, error) {
	return GetMockData().Teachers, nil
}
