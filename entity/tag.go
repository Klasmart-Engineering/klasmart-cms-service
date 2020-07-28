package entity

import "time"

type Tag struct {
	ID   string `json:"id" dynamodbav:"id"`
	Name string `json:"name" dynamodbav:"name"`

	CreatedAt *time.Time `json:"createdAt" dynamodbav:"createdAt"`
	DeletedAt *time.Time `json:"-" dynamodbav:"deletedAt"`
}

type TagCondition struct{
	Name string
}

type TagAddView struct{
	Name string
}
type TagUpdateView struct{
	ID string
	Name string
}
type TagView struct{
	ID string
	Name string
	CreateAt *time.Time
}