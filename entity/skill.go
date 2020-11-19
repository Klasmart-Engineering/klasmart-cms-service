package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type Skill struct {
	ID              string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name            string `json:"name" gorm:"column:name;type:varchar(255)"`
	DevelopmentalID string `json:"developmental_id" gorm:"column:developmental_id;type:varchar(100)"`
	CreateID        string `gorm:"column:create_id;type:varchar(100)"`
	UpdateID        string `gorm:"column:update_id;type:varchar(100)"`
	DeleteID        string `gorm:"column:delete_id;type:varchar(100)"`
	CreateAt        int64  `gorm:"column:create_at;type:bigint"`
	UpdateAt        int64  `gorm:"column:update_at;type:bigint"`
	DeleteAt        int64  `gorm:"column:delete_at;type:bigint"`
}

func (e Skill) TableName() string {
	return constant.TableNameSkill
}

func (e Skill) GetID() interface{} {
	return e.ID
}
