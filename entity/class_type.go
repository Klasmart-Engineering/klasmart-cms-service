package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ClassType struct {
	ID       string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name     string `json:"name" gorm:"column:name;type:varchar(255)"`
	Number   int    `json:"number" gorm:"column:number;type:int"`
	CreateID string `gorm:"column:create_id;type:varchar(100)"`
	UpdateID string `gorm:"column:update_id;type:varchar(100)"`
	DeleteID string `gorm:"column:delete_id;type:varchar(100)"`
	CreateAt int64  `gorm:"column:create_at;type:bigint"`
	UpdateAt int64  `gorm:"column:update_at;type:bigint"`
	DeleteAt int64  `gorm:"column:delete_at;type:bigint"`
}

func (e ClassType) TableName() string {
	return constant.TableNameClassType
}

func (e ClassType) GetID() interface{} {
	return e.ID
}
