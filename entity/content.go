package entity

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	ContentStatusDraft      = "draft"
	ContentStatusPending    = "pending"
	ContentStatusPublished  = "published"
	ContentStatusRejected   = "rejected"
	ContentStatusAttachment = "attachment"
	ContentStatusHidden     = "hidden"
	ContentStatusArchive    = "archive"

	ContentTypeMaterial = 1
	ContentTypeLesson = 2
)


type ContentPublishStatus string

func NewContentPublishStatus(status string) ContentPublishStatus {
	switch status {
	case ContentStatusDraft:
		return ContentStatusDraft
	case ContentStatusPending:
		return ContentStatusPending
	case ContentStatusPublished:
		return ContentStatusPublished
	case ContentStatusRejected:
		return ContentStatusRejected
	case ContentStatusAttachment:
		return ContentStatusAttachment
	case ContentStatusHidden:
		return ContentStatusHidden
	case ContentStatusArchive:
		return ContentStatusArchive
	default:
		return ContentStatusDraft
	}
}

type ContentID struct {
	ID            string   `gorm:"type:varchar(50);PRIMARY_KEY;AUTO_INCREMENT" dynamodbav:"content_id" json:"content_id" dynamoupdate:"-"`
}

type Content struct {
	ID            string   `gorm:"type:varchar(50);PRIMARY_KEY;AUTO_INCREMENT" dynamodbav:"content_id" json:"content_id" dynamoupdate:"-"`
	ContentType   int 		`gorm:"type:int;NOTNULL; column: content_type" dynamodbav:"content_type" json:"content_type" dynamoupdate:":ct"`
	Name          string   `gorm:"type:char(256);NOT NULL;column:name" dynamodbav:"content_name" json:"content_name" dynamoupdate:":n"`
	Program       string    `gorm:"type:varchar(50);NOT NULL;column:program" dynamodbav:"program" json:"program" dynamoupdate:":p"`
	Subject       string    `gorm:"type:varchar(50);NOT NULL;column:subject" dynamodbav:"subject" json:"subject" dynamoupdate:":su"`
	Developmental string    `gorm:"type:varchar(50);NOT NULL;column:developmental" dynamodbav:"developmental" json:"developmental" dynamoupdate:":dv"`
	Skills        string    `gorm:"type:varchar(50);NOT NULL;column:skills" dynamodbav:"skills" json:"skills" dynamoupdate:":sk"`
	Age           string    `gorm:"type:varchar(50);NOT NULL;column:age" dynamodbav:"age" json:"age" dynamoupdate:":a"`
	Keywords      string `gorm:"type:text;NOT NULL;column:keywords" dynamodbav:"keywords" json:"keywords" dynamoupdate:":ky"`
	Description   string   `gorm:"type:text;NOT NULL;column:description" dynamodbav:"description" json:"description" dynamoupdate:":de"`
	Thumbnail     string   `gorm:"type:text;NOT NULL;column:thumbnail" dynamodbav:"thumbnail" json:"thumbnail" dynamoupdate:":th"`

	Data string `gorm:"type:json;NOT NULL;column:data" dynamodbav:"content_data" json:"content_data" dynamoupdate:":d"`
	Extra        string     `gorm:"type:json;NOT NULL;column:extra" dynamodbav:"extra" json:"extra" dynamoupdate:":ex"`

	Author 		string `gorm:"type:varchar(50);NOT NULL;column:author" dynamodbav:"author" json:"author" dynamoupdate:":au"`
	AuthorName  string `gorm:"type:varchar(128);NOT NULL;column:author_name" dynamodbav:"author_name" json:"author_name" dynamoupdate:":aun"`
	Org 		string `gorm:"type:varchar(50);NOT NULL;column:org" dynamodbav:"org" json:"org" dynamoupdate:":og"`

	PublishScope  string                       `gorm:"type:varchar(50);NOT NULL;column:publish_scope;index" dynamodbav:"publish_scope" json:"publish_scope" dynamoupdate:":ps"`
	PublishStatus ContentPublishStatus `gorm:"type:varchar(16);NOT NULL;column:publish_status;index" dynamodbav:"publish_status" json:"publish_status" dynamoupdate:":pst"`

	RejectReason string 				`gorm:"type:varchar(255);NOT NULL;column:reject_reason" dynamodbav:"reject_reason" json:"reject_reason" dynamoupdate:":rr"`
	Version  int64                       `gorm:"type:int;NOT NULL;column:version" dynamodbav:"version" json:"version" dynamoupdate:":ve"`
	LockedBy string 			 `gorm:"type:varchar(50);NOT NULL;column:locked_by" dynamodbav:"locked_by" json:"locked_by" dynamoupdate:":lb"`
	SourceId string 				`gorm:"type:varchar(255);NOT NULL;column:source_id" dynamodbav:"source_id" json:"source_id" dynamoupdate:":si"`
	LatestId string 			`gorm:"type:varchar(255);NOT NULL;column:latest_id" dynamodbav:"latest_id" json:"latest_id" dynamoupdate:":lsi"`

	CreatedAt *time.Time `gorm:"type:datetime;NOT NULL;column:created_at" dynamodbav:"created_at" json:"created_at" dynamoupdate:":ca"`
	UpdatedAt *time.Time `gorm:"type:datetime;NOT NULL;column:updated_at" dynamodbav:"updated_at" json:"updated_at" dynamoupdate:":ua"`
	DeletedAt *time.Time `gorm:"type:datetime;column:deleted_at" dynamodbav:"deleted_at" json:"deleted_at" dynamoupdate:":da"`
}

func (u Content)UpdateExpress() string{
	tags := getDynamoTags(u)
	updateExpressParts := make([]string, 0)
	for i := range tags {
		updateExpressParts = append(updateExpressParts, tags[i].JSONTag + " = " + tags[i].DynamoTag)
	}
	updateExpress := strings.Join(updateExpressParts, ",")
	return "set " + updateExpress
}

type UpdateDyContent struct {
	ContentType   int ` json:":ct"`
	Name          string `json:":n"`
	Program       string `json:":p"`
	Subject       string `json:":su"`
	Developmental string `json:":dv"`
	Skills        string `json:":sk"`
	Age           string `json:":a"`
	Keywords      string `json:":ky"`
	Description   string `json:":de"`
	Thumbnail     string `json:":th"`
	LockedBy string `json:":lb"`

	Data string `json:":d"`
	Extra        string `json:":ex"`

	Author 		string `json:":au"`
	AuthorName  string `json:":aun"`
	Org 		string `json:":og"`

	PublishScope  string `json:":ps"`
	PublishStatus ContentPublishStatus `json:":pst"`

	RejectReason string `json:":rr"`
	SourceId string `json:":si"`
	LatestId string `json:"lsi"`
	Version  int64 `json:":ve"`

	CreatedAt *time.Time `json:":ca"`
	UpdatedAt *time.Time `json:":ua"`
	DeletedAt *time.Time `json:":da"`
}

type TagValues struct {
	JSONTag string
	DynamoTag string
}

func getDynamoTags(s interface{}) []TagValues {
	rt := reflect.TypeOf(s)
	if rt.Kind() != reflect.Struct {
		return nil
	}
	tagValues := make([]TagValues, 0)
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		updateTag := f.Tag.Get("dynamoupdate")
		if updateTag == "-"{
			continue
		}
		tagValues = append(tagValues, TagValues{
			JSONTag: f.Tag.Get("json"),
			DynamoTag: updateTag,
		})
	}
	return tagValues
}
func (s Content) TableName() string {
	return "cms_contents"
}

func (s Content) GetID() interface{} {
	return s.ID
}

type CreateContentRequest struct {
	ContentType   int      `json:"content_type"`
	Name          string   `json:"name"`
	Program       string   `json:"program"`
	Subject       string   `json:"subject"`
	Developmental string   `json:"developmental"`
	Skills        string   `json:"skills"`
	Age           string   `json:"age"`
	Keywords      []string `json:"keywords"`
	Description   string   `json:"description"`
	Thumbnail     string   `json:"thumbnail"`

	DoPublish    bool   `json:"do_publish"`
	PublishScope string `json:"publish_scope"`

	Data  ContentData `json:"data"`
	Extra string      `json:"extra"`
}

type ContentInfoWithDetails struct {
	ContentInfo
	ContentTypeName   string `json:"content_type_name"`
	ProgramName       string `json:"program_name"`
	SubjectName       string `json:"subject_name"`
	DevelopmentalName string `json:"developmental_name"`
	SkillsName        string `json:"skills_name"`
	AgeName           string `json:"age_name"`
	OrgName           string `json:"org_name"`
}

type ContentInfo struct {
	ID            string   `json:"id"`
	ContentType   int      `json:"content_type"`
	Name          string   `json:"name"`
	Program       string   `json:"program"`
	Subject       string   `json:"subject"`
	Developmental string   `json:"developmental"`
	Skills        string   `json:"skills"`
	Age           string   `json:"age"`
	Keywords      []string `json:"keywords"`
	Description   string   `json:"description"`
	Thumbnail     string   `json:"thumbnail"`
	Version       int64    `json:"version"`

	SourceID string `json:"source_id"`
	LockedBy string `json:"locked_by"`
	RejectReason string `json:":rr"`
	LatestID string `json:"latest_id"`

	Data  ContentData `json:"data"`
	Extra string      `json:"extra"`

	Author     string `json:"author"`
	AuthorName string `json:"author_name"`
	Org        string `json:"org"`

	PublishScope  string               `json:"publish_scope"`
	PublishStatus ContentPublishStatus `json:"publish_status"`
}

type ContentData interface {
	Unmarshal(ctx context.Context, data string) error
	Marshal(ctx context.Context) (string, error)

	Validate(ctx context.Context, contentType int) error
	PrepareResult(ctx context.Context) error
	SubContentIds(ctx context.Context) ([]string ,error)
}

func (cInfo *ContentInfo) SetStatus(status ContentPublishStatus) error {
	switch status {
	case ContentStatusArchive:
		if cInfo.allowedToArchive() {
			cInfo.PublishStatus = ContentStatusArchive
		}
		return nil
	case ContentStatusAttachment:
		//TODO
		fmt.Println(cInfo.PublishStatus)
	case ContentStatusDraft:
		//TODO
		fmt.Println(cInfo.PublishStatus)
	case ContentStatusHidden:
		if cInfo.allowedToHidden() {
			cInfo.PublishStatus = ContentStatusHidden
		}
		return nil
	case ContentStatusPending:
		if cInfo.allowedToPending() {
			cInfo.PublishStatus = ContentStatusPending
		}
		return nil
	case ContentStatusPublished:
		if cInfo.allowedToBeReviewed() {
			cInfo.PublishStatus = ContentStatusPublished
		}
		return nil
		fmt.Println(cInfo.PublishStatus)
	case ContentStatusRejected:
		if cInfo.allowedToBeReviewed() {
			cInfo.PublishStatus = ContentStatusRejected
		}
		return nil
	}
	return errors.New(fmt.Sprintf("unsupported:[%s]", status))
}

func (cInfo ContentInfo) allowedToArchive() bool {
	switch cInfo.PublishStatus {
	case ContentStatusPublished:
		return true
	}
	return false
}

func (cInfo ContentInfo) allowedToAttachment() bool {
	// TODO
	return false
}

func (cInfo ContentInfo) allowedToPending() bool {
	switch cInfo.PublishStatus {
	case ContentStatusDraft:
		return true
	}
	return false
}

func (cInfo ContentInfo) allowedToBeReviewed() bool {
	switch cInfo.PublishStatus {
	case ContentStatusPending:
		return true
	}
	return false
}

func (cInfo ContentInfo) allowedToHidden() bool {
	switch cInfo.PublishStatus {
	case ContentStatusPublished:
		return true
	}
	return false
}

func (cInfo ContentInfo) CanBeCancelled() bool {
	if cInfo.PublishStatus == ContentStatusDraft {
		return true
	}
	return false
}

func (cInfo ContentInfo) CanBeDeleted() bool {
	if cInfo.PublishStatus == ContentStatusArchive {
		return true
	}
	return false
}
