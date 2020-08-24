package external

import "context"

type TeacherServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Teacher, error)
	Query(ctx context.Context, keyword string) ([]*Teacher, error)
}

type Teacher struct {
	ID   string
	Name string
}

func GetTeacherServiceProvider() (TeacherServiceProvider, error) {
	return &mockTeacherService{}, nil
}

type mockTeacherService struct{}

func (s mockTeacherService) BatchGet(ctx context.Context, ids []string) ([]*Teacher, error) {
	return []*Teacher{
		{
			ID:   "Teacher-1",
			Name: "Teacher1",
		},
		{
			ID:   "Teacher-2",
			Name: "Teacher2",
		},
	}, nil
}

func (s mockTeacherService) Query(ctx context.Context, keyword string) ([]*Teacher, error) {
	return []*Teacher{
		{
			ID:   "Teacher-1",
			Name: "Teacher1",
		},
		{
			ID:   "Teacher-2",
			Name: "Teacher2",
		},
	}, nil
}
