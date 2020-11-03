package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type Age struct {
	ID   string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name string `json:"name" gorm:"column:name;type:varchar(255)"`
}

func (e Age) TableName() string {
	return constant.TableNameAge
}

func (e Age) GetID() interface{} {
	return e.ID
}
