package entity

import "github.com/KL-Engineering/kidsloop-cms-service/constant"

type OrganizationProperty struct {
	ID        string                     `json:"id" gorm:"column:id;PRIMARY_KEY"`
	Type      OrganizationType           `json:"type" enums:"normal,headquarters" gorm:"column:type;type:varchar(200)"`
	Region    OrganizationPropertyRegion `json:"region" enums:"global,vn" gorm:"column:region;type:varchar(100)"`
	CreatedID string                     `json:"-" gorm:"column:created_id;type:varchar(100)"`
	UpdatedID string                     `json:"-" gorm:"column:updated_id;type:varchar(100)"`
	DeletedID string                     `json:"-" gorm:"column:deleted_id;type:varchar(100)"`
	CreatedAt int64                      `json:"-" gorm:"column:created_at;type:bigint"`
	UpdatedAt int64                      `json:"-" gorm:"column:updated_at;type:bigint"`
	DeleteAt  int64                      `json:"-" gorm:"column:delete_at;type:bigint"`
}

func (OrganizationProperty) TableName() string {
	return constant.TableNameOrganizationProperty
}

type OrganizationPropertyRegion string

const (
	UnknownRegion OrganizationPropertyRegion = "unknown"
	Global        OrganizationPropertyRegion = "global"
	VN            OrganizationPropertyRegion = "vn"
)

type RegionOrganizationInfo struct {
	ID   string `json:"organization_id"`
	Name string `json:"organization_name"`
}

type OrganizationType string

const (
	OrganizationTypeNormal       OrganizationType = "normal"
	OrganizationTypeHeadquarters OrganizationType = "headquarters"
)

type OrganizationInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
