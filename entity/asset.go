package entity

import "time"

type AssetObject struct {
	ID            int64   `gorm:"type:bigint;PRIMARY_KEY;AUTO_INCREMENT;column:id" json:"id" dynamodbav:"id"`
	Name          string   `gorm:"type:char(256);NOT NULL;column:name json:"name" dynamodbav:"name"`
	Program       int64    `gorm:"type:bigint;NOT NULL;column:name program:"program" dynamodbav:"program"`
	Subject       int64    `gorm:"type:bigint;NOT NULL;column:name subject:"subject" dynamodbav:"subject`
	Developmental int64    `gorm:"type:bigint;NOT NULL;column:developmental json:"developmental" dynamodbav:"developmental`
	Skills        int64    `gorm:"type:bigint;NOT NULL;column:skills json:"skills" dynamodbav:"skills`
	Age           int64    `gorm:"type:bigint;NOT NULL;column:age json:"age" dynamodbav:"age`
	Keywords      []string `gorm:"type:json;NOT NULL;column:keywords json:"keywords" dynamodbav:"keywords"`
	Description   string   `gorm:"type:text;NOT NULL;column:description json:"description" dynamodbay: "description"`
	Thumbnail     string   `gorm:"type:text;NOT NULL;column:thumbnail json:"thumbnail" dynamodbav:"thumbnail"`

	Size         int64  `gorm:"type:bigint;NOT NULL;column:size json:"size" dynamodbav:"size"`
	ResourceName string `gorm:"type:text;NOT NULL;column:resource_name json:"resource_name" dynamodbav:"resource_name"`

	Author string `gorm:"type:bigint;NOT NULL;column:author json:"author" dynamodbav:"author"`

	CreatedAt *time.Time `gorm:"type:datetime;NOT NULL;column:updated_at" json:"created_at" dynamodbav:"created_at"`
	UpdatedAt *time.Time `gorm:"type:datetime;NOT NULL;column:created_at" json:"updated_at" dynamodbav:"updated_at"`
	DeletedAt *time.Time `gorm:"type:datetime;column:deleted_at" json:"deleted_at" dynamodbav:"deleted_at"`
}

type UpdateAssetRequest struct {
	ID           int64   `json:"id" dynamodbav:"id"`
	Name         string   `json:"name" dynamodbav:"name"`
}

func (a AssetObject) TableName() string {
	return "assets"
}

type SearchAssetCondition struct {
	ID        []int64 `json:"id"`
	Name      string `json:"name"`

	SearchWords []string `json:"search_words"`
	Author      []int    `json:"author"`

	OrderBy  string `json:"order_by"`
	PageSize int `json:"page_size"`
	Page     int `json:"page"`
}

type ResourcePath struct {
	Path string `json:"path"`
	Name string `json:"name"`
}
