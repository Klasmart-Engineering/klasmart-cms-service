package entity

type BaseField struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Thumbnail   string `json:"thumbnail"`
	Description string `json:"description"`
}

type LessonPlan struct {
	BaseField
	Materials []*Material `json:"materials"`
}

type Material struct {
	BaseField
	Data string `json:"data"`
}
