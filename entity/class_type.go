package entity

type ClassType struct {
	ID   string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name string `json:"name" gorm:"column:name;type:varchar(255)"`
}
