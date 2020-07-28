package entity

import "time"

type CategoryObject struct {
	Id       string   `json:"id" dynamodbav:"id"`
	Name     string   `json:"name" dynamodbav:"name"`
	ParentID string  `json:"parent_id" dynamodbav:"parentID"`

	CreatedAt *time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" dynamodbav:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" dynamodbav:"deleted_at"`
}

func (CategoryObject) TableName() string{
	return "categories"
}
