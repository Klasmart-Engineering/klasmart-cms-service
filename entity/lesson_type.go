package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type LessonType struct {
	ID   string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name string `json:"name" gorm:"column:name;type:varchar(255)"`
}

func (e LessonType) TableName() string {
	return constant.TableNameLessonType
}

func (e LessonType) GetID() interface{} {
	return e.ID
}
