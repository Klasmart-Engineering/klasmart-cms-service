package entity

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type Tag struct {
	ID     string `json:"id" dynamodbav:"id"`
	Name   string `json:"name" dynamodbav:"name"`
	States int    `json:"states" dynamodbav:"states"`

	CreatedID string `json:"-" dynamodbav:"created_id"`
	UpdatedID string `json:"-" dynamodbav:"updated_id"`
	DeletedID string `json:"-" dynamodbav:"deleted_id"`

	CreatedAt int64 `json:"-" dynamodbav:"created_at"`
	UpdatedAt int64 `json:"-" dynamodbav:"updated_at"`
	DeletedAt int64 `json:"-" dynamodbav:"deleted_at"`
}

func (t Tag) TableName() string {
	return constant.TableNameTag
}

type TagAddView struct {
	Name string `json:"name"`
}
type TagUpdateView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	States   int    `json:"states"`
}
type TagView struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	States   int    `json:"states"`
	CreateAt int64  `json:"created_at"`
}