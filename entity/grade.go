package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type Grade struct {
	ID   string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name string `json:"name" gorm:"column:name;type:varchar(255)"`
}

func (e Grade) TableName() string {
	return constant.TableNameGrade
}

func (e Grade) GetID() interface{} {
	return e.ID
}
