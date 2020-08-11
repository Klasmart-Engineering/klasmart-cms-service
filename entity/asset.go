package entity

import (
	"strings"
	"time"
)

type AssetObject struct {
	ID            string   `gorm:"type:varchar(50);PRIMARY_KEY;AUTO_INCREMENT;column:id" json:"id" dynamodbav:"id"`
	Name          string   `gorm:"type:char(256);NOT NULL;column:name json:"name" dynamodbav:"name"`
	Program       string    `gorm:"type:varchar(50);NOT NULL;column:name program:"program" dynamodbav:"program"`
	Subject       string    `gorm:"type:varchar(50);NOT NULL;column:name subject:"subject" dynamodbav:"subject`
	Developmental string    `gorm:"type:varchar(50);NOT NULL;column:developmental json:"developmental" dynamodbav:"developmental`
	Skills        string    `gorm:"type:varchar(50);NOT NULL;column:skills json:"skills" dynamodbav:"skills`
	Age           string    `gorm:"type:varchar(50);NOT NULL;column:age json:"age" dynamodbav:"age`
	Keywords      string `gorm:"type:text;NOT NULL;column:keywords json:"keywords" dynamodbav:"keywords"`
	Description   string   `gorm:"type:text;NOT NULL;column:description json:"description" dynamodbay: "description"`
	Thumbnail     string   `gorm:"type:text;NOT NULL;column:thumbnail json:"thumbnail" dynamodbav:"thumbnail"`

	Size     int64  `gorm:"type:bigint;NOT NULL;column:size json:"size" dynamodbav:"size"`
	Resource string `gorm:"type:text;NOT NULL;column:resource json:"resource" dynamodbav:"resource"`

	Author 		string `gorm:"type:varchar(50);NOT NULL;column:author json:"author" dynamodbav:"author"`
	AuthorName  string `gorm:"type:varchar(128);NOT NULL;column:author_name json:"author_name" dynamodbav:"author_name"`
	Org 		string `gorm:"type:varchar(50);NOT NULL;column:org json:"org" dynamodbav:"org"`

	CreatedAt *time.Time `gorm:"type:datetime;NOT NULL;column:created_at" json:"created_at" dynamodbav:"created_at"`
	UpdatedAt *time.Time `gorm:"type:datetime;NOT NULL;column:updated_at" json:"updated_at" dynamodbav:"updated_at"`
	DeletedAt *time.Time `gorm:"type:datetime;column:deleted_at" json:"deleted_at" dynamodbav:"deleted_at"`
}

func (a *AssetObject) ToAssetData() *AssetData{
	return &AssetData{
		ID:            a.ID,
		Name:          a.Name,
		Program:       a.Program,
		Subject:       a.Subject,
		Developmental: a.Developmental,
		Skills:        a.Skills,
		Age:           a.Age,
		Keywords:      strings.Split(a.Keywords, ","),
		Description:   a.Description,
		Thumbnail:     a.Thumbnail,
		Size:          a.Size,
		Resource:      a.Resource,
		Author:        a.Author,
		AuthorName:    a.AuthorName,
		Org:           a.Org,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		DeletedAt:     a.DeletedAt,
	}
}

type AssetData struct {
	ID            string
	Name          string
	Program       string
	Subject       string
	Developmental string
	Skills        string
	Age           string
	Keywords      []string
	Description   string
	Thumbnail     string

	Size     int64
	Resource string

	Author 		string
	AuthorName  string
	Org 		string

	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

type CreateAssetData struct {
	Name          string
	Program       string
	Subject       string
	Developmental string
	Skills        string
	Age           string
	Keywords      []string
	Description   string
	Thumbnail     string

	Resource string
}

func (a *CreateAssetData) ToAssetObject() *AssetObject{
	return &AssetObject{
		Name:          a.Name,
		Program:       a.Program,
		Subject:       a.Subject,
		Developmental: a.Developmental,
		Skills:        a.Skills,
		Age:           a.Age,
		Keywords:      strings.Join(a.Keywords, ","),
		Description:   a.Description,
		Thumbnail:     a.Thumbnail,
		Resource:      a.Resource,
	}
}

func (a *AssetData) ToAssetObject() *AssetObject{
	return &AssetObject{
		ID:            a.ID,
		Name:          a.Name,
		Program:       a.Program,
		Subject:       a.Subject,
		Developmental: a.Developmental,
		Skills:        a.Skills,
		Age:           a.Age,
		Keywords:      strings.Join(a.Keywords, ","),
		Description:   a.Description,
		Thumbnail:     a.Thumbnail,
		Size:          a.Size,
		Resource:      a.Resource,
		Author:        a.Author,
		AuthorName:    a.AuthorName,
		Org:           a.Org,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		DeletedAt:     a.DeletedAt,
	}
}

type UpdateAssetRequest struct {
	ID           string   `json:"id" dynamodbav:"id"`
	Name         string   `json:"name" dynamodbav:"name"`
}

func (a AssetObject) TableName() string {
	return "assets"
}

type SearchAssetCondition struct {
	ID        []string `json:"id"`

	SearchWords []string `json:"search_words"`
	FuzzyQuery string `json:"fuzzy_query"`
	IsSelf bool `json:"is_self"`

	OrderBy  string `json:"order_by"`
	PageSize int `json:"page_size"`
	Page     int `json:"page"`
}

type ResourcePath struct {
	Path string `json:"path"`
	Name string `json:"name"`
}
