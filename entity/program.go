package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type GroupType string

const (
	GroupTypeBadaESL   = "BadaESL"
	GroupTypeBadaSTEAM = "BadaSTEAM"
)

func (t GroupType) Valid() bool {
	switch t {
	case GroupTypeBadaESL, GroupTypeBadaSTEAM:
		return true
	default:
		return false
	}
}

type Program struct {
	ID        string    `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Name      string    `json:"name" gorm:"column:name;type:varchar(255)"`
	Number    int       `json:"number" gorm:"column:number;type:int"`
	OrgType   string    `json:"org_type" gorm:"column:org_type;type:varchar(100)"`
	GroupName GroupType `json:"group_name" gorm:"column:group_name;type:varchar(100)"`
	CreateID  string    `gorm:"column:create_id;type:varchar(100)"`
	UpdateID  string    `gorm:"column:update_id;type:varchar(100)"`
	DeleteID  string    `gorm:"column:delete_id;type:varchar(100)"`
	CreateAt  int64     `gorm:"column:create_at;type:bigint"`
	UpdateAt  int64     `gorm:"column:update_at;type:bigint"`
	DeleteAt  int64     `gorm:"column:delete_at;type:bigint"`
}

func (e Program) TableName() string {
	return constant.TableNameProgram
}

func (e Program) GetID() interface{} {
	return e.ID
}
