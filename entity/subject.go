package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type Subject struct {
	ID   string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name string `json:"name" gorm:"column:name;type:varchar(255)"`
}

func (e Subject) TableName() string {
	return constant.TableNameSubject
}

func (e Subject) GetID() interface{} {
	return e.ID
}
