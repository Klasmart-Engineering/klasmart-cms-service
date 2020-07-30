package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"


type Tag struct {
	ID     string `dynamodbav:"id"`
	Name   string `dynamodbav:"name"`
	States int    `dynamodbav:"states"`

	CreatedAt int64 `dynamodbav:"createdAt"`
	DeletedAt int64 `dynamodbav:"deletedAt"`
}

func (t Tag) TableName() string {
	return constant.TableNameTag
}

type TagCondition struct {
	Name     string `json:"name"`
	PageSize int64  `json:"page_size"`
	Page     int64  `json:"page"`
}

type TagAddView struct {
	Name string `json:"name"`
}
type TagUpdateView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type TagView struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	States   int    `json:"states"`
	CreateAt int64  `json:"created_at"`
}
