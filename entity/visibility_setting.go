package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type VisibilitySetting struct {
	ID   string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name string `json:"name" gorm:"column:name;type:varchar(255)"`
}

func (e VisibilitySetting) TableName() string {
	return constant.TableNameVisibilitySetting
}

func (e VisibilitySetting) GetID() interface{} {
	return e.ID
}
