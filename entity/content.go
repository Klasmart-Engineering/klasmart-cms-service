package entity

import (
	"context"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
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

type Content struct {
	ID            string `gorm:"type:varchar(50);PRIMARY_KEY;AUTO_INCREMENT"`
	ContentType   int    `gorm:"type:int;NOTNULL; column: content_type"`
	Name          string `gorm:"type:char(256);NOT NULL;column:name"`
	Program       string `gorm:"type:varchar(50);NOT NULL;column:program"`
	Subject       string `gorm:"type:varchar(50);NOT NULL;column:subject"`
	Developmental string `gorm:"type:varchar(50);NOT NULL;column:developmental"`
	Skills        string `gorm:"type:varchar(50);NOT NULL;column:skills"`
	Age           string `gorm:"type:varchar(50);NOT NULL;column:age"`
	Keywords      string `gorm:"type:text;NOT NULL;column:keywords"`
	Description   string `gorm:"type:text;NOT NULL;column:description"`
	Thumbnail     string `gorm:"type:text;NOT NULL;column:thumbnail"`

	Data  string `gorm:"type:json;NOT NULL;column:data"`
	Extra string `gorm:"type:json;NOT NULL;column:extra"`

	Author     string `gorm:"type:varchar(50);NOT NULL;column:author"`
	AuthorName string `gorm:"type:varchar(128);NOT NULL;column:author_name"`
	Org        string `gorm:"type:varchar(50);NOT NULL;column:org"`

	PublishScope  string               `gorm:"type:varchar(50);NOT NULL;column:publish_scope;index"`
	PublishStatus ContentPublishStatus `gorm:"type:varchar(16);NOT NULL;column:publish_status;index"`

	Version int64 `gorm:"type:int;NOT NULL;column:version"`

	CreatedAt *time.Time `gorm:"type:datetime;NOT NULL;column:created_at"`
	UpdatedAt *time.Time `gorm:"type:datetime;NOT NULL;column:updated_at"`
	DeletedAt *time.Time `gorm:"type:datetime;column:deleted_at"`
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

	DoPublish bool `json:"do_publish"`

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

	Validate(ctx context.Context, contentType int, tx *dbo.DBContext) error
	PrepareResult(ctx context.Context) error
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
