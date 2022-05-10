package entity

import "github.com/KL-Engineering/kidsloop-cms-service/constant"

type LessonType struct {
	ID       string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name     string `json:"name" gorm:"column:name;type:varchar(255)"`
	Number   int    `json:"number" gorm:"column:number;type:int"`
	CreateID string `gorm:"column:create_id;type:varchar(100)"`
	UpdateID string `gorm:"column:update_id;type:varchar(100)"`
	DeleteID string `gorm:"column:delete_id;type:varchar(100)"`
	CreateAt int64  `gorm:"column:create_at;type:bigint"`
	UpdateAt int64  `gorm:"column:update_at;type:bigint"`
	DeleteAt int64  `gorm:"column:delete_at;type:bigint"`
}

func (e LessonType) TableName() string {
	return constant.TableNameLessonType
}

func (e LessonType) GetID() interface{} {
	return e.ID
}
