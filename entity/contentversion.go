package entity

import "time"

type ContentVersionData struct {
	RecordId	int64
	ContentId string
	Version   int
}


type ContentVersion struct {
	Id        int64 `gorm:"PRIMARY_KEY;AUTO_INCREMENT;column:id"`
	ContentId string `gorm:"type:varchar(50);NOT NULL;column:content_id;index"`
	LastId    string `gorm:"type:varchar(50);NOT NULL;column:last_id;index"`
	MainId    string `gorm:"type:varchar(50);NOT NULL;column:main_id;index"`
	SourceId  string `gorm:"type:varchar(50);NOT NULL;column:source_id;index"`
	Version   int   `gorm:"type:int;NOT NULL;column:version"`

	UpdatedAt *time.Time `gorm:"type:datetime;NOT NULL;column:updated_at"`
	CreatedAt *time.Time `gorm:"type:datetime;NOT NULL;column:created_at"`
	DeletedAt *time.Time `gorm:"type:datetime;column:deleted_at"`
}
func (s ContentVersion) TableName() string {
	return "cms_content_versions"
}

func (s ContentVersion) GetID() interface{} {
	return s.Id
}