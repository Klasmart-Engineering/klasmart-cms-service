package entity

import (
	"calmisland/kidsloop2/constant"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type Tag struct {
	ID     string `dynamodbav:"id"`
	Name   string `dynamodbav:"name"`
	States int    `dynamodbav:"states"`

	CreatedAt int64 `dynamodbav:"createdAt"`
	UpdatedAt int64 `dynamodbav:"updated_at"`
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

func (t TagCondition) getConditions()[]expression.ConditionBuilder{
	//conditions := make([]expression.ConditionBuilder, 0)
	//if t.Name==""{
	//	condition :=
	//}
	return nil
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
