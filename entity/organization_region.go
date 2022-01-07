package entity

type OrganizationRegion struct {
	ID             string `gorm:"type:varchar(256);column:id;PRIMARY_KEY"`
	Headquarter    string `gorm:"type:varchar(255);NOT NULL;column:headquarter;index"`
	OrganizationID string `gorm:"type:varchar(255);NOT NULL;column:organization_id"`

	CreateAt int64 `gorm:"type:bigint;NOT NULL;column:create_at"`
	UpdateAt int64 `gorm:"type:bigint;NOT NULL;column:update_at"`
	DeleteAt int64 `gorm:"type:bigint;column:delete_at"`
}

func (OrganizationRegion) TableName() string {
	return "organizations_regions"
}
