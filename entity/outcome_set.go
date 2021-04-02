package entity

type Set struct {
	ID             string `gorm:"column:id;primary_key" json:"set_id"`
	Name           string `gorm:"column:name" json:"set_name"`
	OrganizationID string `gorm:"column:organization_id" json:"organization_id"`
	CreateAt       int64  `gorm:"column:create_at" json:"created_at"`
	UpdateAt       int64  `gorm:"column:update_at" json:"updated_at"`
	DeleteAt       int64  `gorm:"column:delete_at" json:"deleted_at"`
}

func (Set) TableName() string {
	return "sets"
}

type OutcomeSet struct {
	ID        int    `gorm:"column:id;primary_key"`
	OutcomeID string `gorm:"column:outcome_id" json:"outcome_id"`
	SetID     string `gorm:"column:set_id" json:"set_id"`
	CreateAt  int64  `gorm:"column:create_at" json:"created_at"`
	UpdateAt  int64  `gorm:"column:update_at" json:"updated_at"`
	DeleteAt  int64  `gorm:"column:delete_at" json:"deleted_at"`
}

func (OutcomeSet) TableName() string {
	return "outcomes_sets"
}
