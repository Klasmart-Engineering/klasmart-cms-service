package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type UserSetting struct {
	ID          string `json:"-" gorm:"column:id;PRIMARY_KEY"`
	UserID      string `json:"-" gorm:"column:user_id;type:varchar(100)"`
	SettingJson string `json:"setting_json" binding:"required" gorm:"column:setting_json;type:json;"`
}

func (e UserSetting) TableName() string {
	return constant.TableNameUserSetting
}

func (e UserSetting) GetID() interface{} {
	return e.ID
}

const (
	DefaultUserSettingID = "default_setting_0"
)

type UserSettingJsonContent struct {
	CMSPageSize int `json:"cms_page_size" binding:"required,min=1"`
}
