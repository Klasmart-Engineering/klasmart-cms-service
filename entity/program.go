package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type Program struct {
	ID   string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name string `json:"name" gorm:"column:name;type:varchar(255)"`
}

func (e Program) TableName() string {
	return constant.TableNameProgram
}

func (e Program) GetID() interface{} {
	return e.ID
}
