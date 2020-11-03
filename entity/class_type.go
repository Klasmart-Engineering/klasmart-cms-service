package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ClassType struct {
	ID   string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name string `json:"name" gorm:"column:name;type:varchar(255)"`
}

func (e ClassType) TableName() string {
	return constant.TableNameClassType
}

func (e ClassType) GetID() interface{} {
	return e.ID
}
