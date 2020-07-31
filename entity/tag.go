package entity

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type Tag struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	States int    `json:"states"`

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
	DeletedAt int64 `json:"deleted_at"`
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
