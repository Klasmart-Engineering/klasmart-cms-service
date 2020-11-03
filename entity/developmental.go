package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type Developmental struct {
	ID   string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name string `json:"name" gorm:"column:name;type:varchar(255)"`
}

func (e Developmental) TableName() string {
	return constant.TableNameDevelopmental
}

func (e Developmental) GetID() interface{} {
	return e.ID
}
