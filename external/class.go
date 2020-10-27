package external

import "context"

type ClassServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Class, error)
	GetStudents(ctx context.Context, classID string) ([]*Student, error)
}

type Class struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Students []*Student `json:"students"`
}

func GetClassServiceProvider() ClassServiceProvider {
	return &mockClassService{}
}

type mockClassService struct{}

func (s mockClassService) BatchGet(ctx context.Context, ids []string) ([]*Class, error) {
	return GetMockData().Classes, nil
}

func (s mockClassService) GetStudents(ctx context.Context, classID string) ([]*Student, error) {
	classes := GetMockData().Classes
	for _, class := range classes {
		if class.ID == classID {
			return class.Students, nil
		}
	}
	return nil, nil
}
