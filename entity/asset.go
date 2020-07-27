package entity

import "time"

type AssetObject struct {
	Id string `json:"_id" dynamodbav:"_id"`
	Name string `json:"name" dynamodbav:"name"`
	Category string `json:"category" dynamodbav:"category"`
	Size int `json:"size" dynamodbav:"size"`
	Tag string `json:"tag" dynamodbav:"tag"`
	URL string `json:"url" dynamodbav:"url"`

	Uploader string `json:"uploader" dynamodbav:"uploader"`
	UploadedAt *time.Time `json:"uploadedAt" dynamodbav:"uploadedAt"`

	CreatedAt *time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt" dynamodbav:"deletedAt"`
}
