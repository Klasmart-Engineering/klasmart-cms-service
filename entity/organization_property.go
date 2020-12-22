package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type OrganizationProperty struct {
	ID        string           `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Type      OrganizationType `json:"type" enums:"normal,headquarters" gorm:"column:type;type:varchar(200)"`
	CreatedID string           `json:"-" gorm:"column:created_id;type:varchar(100)"`
	UpdatedID string           `json:"-" gorm:"column:updated_id;type:varchar(100)"`
	DeletedID string           `json:"-" gorm:"column:deleted_id;type:varchar(100)"`
	CreatedAt int64            `json:"-" gorm:"column:created_at;type:bigint"`
	UpdatedAt int64            `json:"-" gorm:"column:updated_at;type:bigint"`
	DeleteAt  int64            `json:"-" gorm:"column:delete_at;type:bigint"`
}

func (OrganizationProperty) TableName() string {
	return constant.TableNameOrganizationProperty
}

type OrganizationType string

const (
	OrganizationTypeNormal       OrganizationType = "normal"
	OrganizationTypeHeadquarters OrganizationType = "headquarters"
)
