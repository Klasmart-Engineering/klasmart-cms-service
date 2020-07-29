package entity

import "time"

type AssetObject struct {
	Id       string   `json:"id" dynamodbav:"id"`
	Name     string   `json:"name" dynamodbav:"name"`
	Category string   `json:"category" dynamodbav:"category"`
	Size     int      `json:"size" dynamodbav:"size"`
	Tags     []string `json:"tags" dynamodbav:"tags"`
	URL      string   `json:"url" dynamodbav:"url"`

	Uploader   string     `json:"uploader" dynamodbav:"uploader"`

	CreatedAt *time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" dynamodbav:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" dynamodbav:"deleted_at"`
}


type UpdateAssetRequest struct {
	Id       string   `json:"id" dynamodbav:"id"`
	Name     string `json:"name" dynamodbav:"name"`
	Category string `json:"category" dynamodbav:"category"`
	Tag      []string `json:"tag" dynamodbav:"tag"`
	URL      string `json:"url" dynamodbav:"url"`
}

func (a AssetObject) TableName() string{
	return "assets"
}

type FileExtensionRequest struct {
	Extension string `json:"extension"`
}
type PathRequest struct {
	Path string `json:"path"`
}