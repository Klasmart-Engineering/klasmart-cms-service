package entity

import "time"

type AssetObject struct {
	ID           string   `json:"id" dynamodbav:"id"`
	Name         string   `json:"name" dynamodbav:"name"`
	Category     string   `json:"category" dynamodbav:"category"`
	Size         int      `json:"size" dynamodbav:"size"`
	Tags         []string `json:"tags" dynamodbav:"tags"`
	ResourceName string   `json:"resource_name" dynamodbav:"resource_name"`

	Uploader   string     `json:"uploader" dynamodbav:"uploader"`

	CreatedAt *time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" dynamodbav:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" dynamodbav:"deleted_at"`
}


type UpdateAssetRequest struct {
	ID       string   `json:"id" dynamodbav:"id"`
	Name     string   `json:"name" dynamodbav:"name"`
	Category string   `json:"category" dynamodbav:"category"`
	Tag      []string `json:"tag" dynamodbav:"tag"`
	ResourceName     string   `json:"resource_name" dynamodbav:"resource_name"`
}

func (a AssetObject) TableName() string{
	return "assets"
}
type SearchAssetCondition struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	SizeMin  int    `json:"size_min"`
	SizeMax  int    `json:"size_max"`

	Tag 	string `json:"tag"`

	PageSize int `json:"page_size"`
	Page     int `json:"page"`
}

type ResourcePath struct{
	Path string `json:"path"`
	Name string `json:"name"`
}