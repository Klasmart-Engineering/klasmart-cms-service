package external

import "context"

type LessonTypeProvider interface {
	BatchGet(ctx context.Context, ids []int) ([]*LessonType, error)
}
type LessonType struct {
	ID   int `json:"id"`
	Name string `json:"name"`
}
type mockLessonTypeService struct{}

func (m *mockLessonTypeService) BatchGet(ctx context.Context, ids []int) ([]*LessonType, error) {
	//TODO: to add publish scope
	return []*LessonType{
		{
			ID:   0,
			Name: "No",
		},
		{
			ID:   1,
			Name: "Test",
		},
		{
			ID:   2,
			Name: "NoTest",
		},
	}, nil
}

func GetLessonTypeProvider() LessonTypeProvider {
	return &mockLessonTypeService{}
}
