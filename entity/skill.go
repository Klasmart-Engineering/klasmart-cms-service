package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type Skill struct {
	ID              string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name            string `json:"name" gorm:"column:name;type:varchar(255)"`
	DevelopmentalID string `json:"developmental_id" gorm:"column:developmental_id;type:varchar(100)"`
}

func (e Skill) TableName() string {
	return constant.TableNameSkill
}

func (e Skill) GetID() interface{} {
	return e.ID
}
