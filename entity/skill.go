package entity

type Skill struct {
	ID              string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name            string `json:"name" gorm:"column:name;type:varchar(255)"`
	DevelopmentalID string `json:"developmental_id" gorm:"column:developmental_id;type:varchar(100)"`
}
