package entity

type VisibilitySetting struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Number   int    `json:"number"`
	Group    string `json:"group"`
	CreateID string
	UpdateID string
	DeleteID string
	CreateAt int64
	UpdateAt int64
	DeleteAt int64
}
