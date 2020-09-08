package entity

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
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
	ContentTypeLesson   = 2
	ContentTypeAssets   = 3

	ContentTypeAssetImage    = 10
	ContentTypeAssetVideo    = 11
	ContentTypeAssetAudio    = 12
	ContentTypeAssetDocument = 13
)

var (
	ErrRequireContentName  = errors.New("content name required")
	ErrRequirePublishScope = errors.New("publish scope required")
	ErrInvalidResourceId   = errors.New("invalid resource id")
	ErrInvalidContentType  = errors.New("invalid content type")
)

type ContentPublishStatus string
type ContentType int

func NewContentType(contentType int) ContentType {
	switch contentType {
	case ContentTypeMaterial:
		return ContentTypeMaterial
	case ContentTypeLesson:
		return ContentTypeLesson
	case ContentTypeAssetDocument:
		return ContentTypeAssetDocument
	case ContentTypeAssetAudio:
		return ContentTypeAssetAudio
	case ContentTypeAssetImage:
		return ContentTypeAssetImage
	case ContentTypeAssetVideo:
		return ContentTypeAssetVideo
	default:
		return ContentTypeAssetImage
	}
}

func (c ContentType) Validate() error {
	switch c {
	case ContentTypeMaterial:
		return nil
	case ContentTypeLesson:
		return nil
	case ContentTypeAssetDocument:
		return nil
	case ContentTypeAssetAudio:
		return nil
	case ContentTypeAssetImage:
		return nil
	case ContentTypeAssetVideo:
		return nil
	}
	return ErrInvalidContentType
}

func (c ContentType) IsAsset() bool {
	switch c {
	case ContentTypeAssets:
		fallthrough
	case ContentTypeAssetImage:
		fallthrough
	case ContentTypeAssetVideo:
		fallthrough
	case ContentTypeAssetAudio:
		fallthrough
	case ContentTypeAssetDocument:
		return true
	}
	return false
}

func (c ContentType) ContentTypeInt() []int {
	switch c {
	case ContentTypeLesson:
		return []int{ContentTypeLesson}
	case ContentTypeMaterial:
		return []int{ContentTypeMaterial}
	case ContentTypeAssets:
		return []int{ContentTypeAssetImage, ContentTypeAssetVideo, ContentTypeAssetAudio, ContentTypeAssetDocument}
	case ContentTypeAssetImage:
		return []int{ContentTypeAssetImage}
	case ContentTypeAssetVideo:
		return []int{ContentTypeAssetVideo}
	case ContentTypeAssetAudio:
		return []int{ContentTypeAssetAudio}
	case ContentTypeAssetDocument:
		return []int{ContentTypeAssetDocument}
	}
	return []int{ContentTypeLesson}
}

func (c ContentType) ContentType() []ContentType {
	switch c {
	case ContentTypeLesson:
		return []ContentType{ContentTypeLesson}
	case ContentTypeMaterial:
		return []ContentType{ContentTypeMaterial}
	case ContentTypeAssets:
		return []ContentType{ContentTypeAssetImage, ContentTypeAssetVideo, ContentTypeAssetAudio, ContentTypeAssetDocument}
	case ContentTypeAssetImage:
		return []ContentType{ContentTypeAssetImage}
	case ContentTypeAssetVideo:
		return []ContentType{ContentTypeAssetVideo}
	case ContentTypeAssetAudio:
		return []ContentType{ContentTypeAssetAudio}
	case ContentTypeAssetDocument:
		return []ContentType{ContentTypeAssetDocument}
	}
	return []ContentType{ContentTypeLesson}
}

func (c ContentType) Name() string {
	switch c {
	case ContentTypeLesson:
		return "Plan"
	case ContentTypeMaterial:
		return "Material"
	case ContentTypeAssets:
		fallthrough
	case ContentTypeAssetImage:
		fallthrough
	case ContentTypeAssetVideo:
		fallthrough
	case ContentTypeAssetAudio:
		fallthrough
	case ContentTypeAssetDocument:
		return "Assets"
	}
	return "Unknown"
}

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
	ID string `gorm:"type:varchar(50);PRIMARY_KEY;AUTO_INCREMENT" dynamodbav:"content_id" json:"content_id" dynamoupdate:"-"`
}

type ContentStatisticsInfo struct {
	SubContentCount int `json:"subcontent_count"`
	OutcomesCount   int `json:"outcomes_count"`
}

type Content struct {
	ID            string      `gorm:"type:varchar(50);PRIMARY_KEY;AUTO_INCREMENT" dynamodbav:"content_id" json:"content_id" dynamoupdate:"-"`
	ContentType   ContentType `gorm:"type:int;NOTNULL; column:content_type" dynamodbav:"content_type" json:"content_type" dynamoupdate:":ct"`
	Name          string      `gorm:"type:varchar(255);NOT NULL;column:content_name" dynamodbav:"content_name" json:"content_name" dynamoupdate:":n"`
	Program       string      `gorm:"type:varchar(1024);NOT NULL;column:program" dynamodbav:"program" json:"program" dynamoupdate:":p"`
	Subject       string      `gorm:"type:varchar(1024);NOT NULL;column:subject" dynamodbav:"subject" json:"subject" dynamoupdate:":su"`
	Developmental string      `gorm:"type:varchar(1024);NOT NULL;column:developmental" dynamodbav:"developmental" json:"developmental" dynamoupdate:":dv"`
	Skills        string      `gorm:"type:varchar(1024);NOT NULL;column:skills" dynamodbav:"skills" json:"skills" dynamoupdate:":sk"`
	Age           string      `gorm:"type:varchar(1024);NOT NULL;column:age" dynamodbav:"age" json:"age" dynamoupdate:":a"`
	Grade         string      `gorm:"type:varchar(1024);NOT NULL;column:grade" dynamodbav:"grade" json:"grade" dynamoupdate:":grd"`
	Keywords      string      `gorm:"type:text;NOT NULL;column:keywords" dynamodbav:"keywords" json:"keywords" dynamoupdate:":ky"`
	Description   string      `gorm:"type:text;NOT NULL;column:description" dynamodbav:"description" json:"description" dynamoupdate:":de"`
	Thumbnail     string      `gorm:"type:text;NOT NULL;column:thumbnail" dynamodbav:"thumbnail" json:"thumbnail" dynamoupdate:":th"`

	Outcomes string `gorm:"type:text;NOT NULL;column:outcomes"`
	Data     string `gorm:"type:json;NOT NULL;column:data" dynamodbav:"content_data" json:"content_data" dynamoupdate:":d"`
	Extra    string `gorm:"type:text;NOT NULL;column:extra" dynamodbav:"extra" json:"extra" dynamoupdate:":ex"`

	SuggestTime int    `gorm:"type:int;NOT NULL;column:suggest_time" dynamodbav:"suggest_time" json:"extra" dynamoupdate:":sut"`
	Author      string `gorm:"type:varchar(50);NOT NULL;column:author" dynamodbav:"author" json:"author" dynamoupdate:":au"`
	AuthorName  string `gorm:"type:varchar(128);NOT NULL;column:author_name" dynamodbav:"author_name" json:"author_name" dynamoupdate:":aun"`
	Org         string `gorm:"type:varchar(50);NOT NULL;column:org" dynamodbav:"org" json:"org" dynamoupdate:":og"`

	PublishScope  string               `gorm:"type:varchar(50);NOT NULL;column:publish_scope;index" dynamodbav:"publish_scope" json:"publish_scope" dynamoupdate:":ps"`
	PublishStatus ContentPublishStatus `gorm:"type:varchar(16);NOT NULL;column:publish_status;index" dynamodbav:"publish_status" json:"publish_status" dynamoupdate:":pst"`

	RejectReason string `gorm:"type:varchar(255);NOT NULL;column:reject_reason" dynamodbav:"reject_reason" json:"reject_reason" dynamoupdate:":rr"`
	Version      int64  `gorm:"type:int;NOT NULL;column:version" dynamodbav:"version" json:"version" dynamoupdate:":ve"`
	LockedBy     string `gorm:"type:varchar(50);NOT NULL;column:locked_by" dynamodbav:"locked_by" json:"locked_by" dynamoupdate:":lb"`
	SourceID     string `gorm:"type:varchar(255);NOT NULL;column:source_id" dynamodbav:"source_id" json:"source_id" dynamoupdate:":si"`
	LatestID     string `gorm:"type:varchar(255);NOT NULL;column:latest_id" dynamodbav:"latest_id" json:"latest_id" dynamoupdate:":lsi"`

	CreateAt int64 `gorm:"type:bigint;NOT NULL;column:create_at" dynamodbav:"created_at" json:"created_at" dynamoupdate:":ca"`
	UpdateAt int64 `gorm:"type:bigint;NOT NULL;column:update_at" dynamodbav:"updated_at" json:"updated_at" dynamoupdate:":ua"`
	DeleteAt int64 `gorm:"type:bigint;column:delete_at" dynamodbav:"deleted_at" json:"deleted_at" dynamoupdate:":da"`
}

func (u Content) UpdateExpress() string {
	tags := getDynamoTags(u)
	updateExpressParts := make([]string, 0)
	for i := range tags {
		updateExpressParts = append(updateExpressParts, tags[i].JSONTag+" = "+tags[i].DynamoTag)
	}
	updateExpress := strings.Join(updateExpressParts, ",")
	return "set " + updateExpress
}

type UpdateDyContent struct {
	ContentType   ContentType ` json:":ct"`
	Name          string      `json:":n"`
	Program       string      `json:":p"`
	Subject       string      `json:":su"`
	Developmental string      `json:":dv"`
	Skills        string      `json:":sk"`
	Age           string      `json:":a"`
	Keywords      string      `json:":ky"`
	Description   string      `json:":de"`
	Thumbnail     string      `json:":th"`
	LockedBy      string      `json:":lb"`

	Data  string `json:":d"`
	Extra string `json:":ex"`

	Author      string `json:":au"`
	AuthorName  string `json:":aun"`
	Org         string `json:":og"`
	SuggestTime int    `json:":sut"`

	PublishScope  string               `json:":ps"`
	PublishStatus ContentPublishStatus `json:":pst"`

	RejectReason string `json:":rr"`
	SourceID     string `json:":si"`
	LatestID     string `json:"lsi"`
	Version      int64  `json:":ve"`

	OrgUserId                     string `json:":ouid"`
	ContentTypeOrgIdPublishStatus string `json:":cps"`

	CreatedAt int64 `json:":ca"`
	UpdatedAt int64 `json:":ua"`
	DeletedAt int64 `json:":da"`
}

type TagValues struct {
	JSONTag   string
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
		if updateTag == "-" {
			continue
		}
		tagValues = append(tagValues, TagValues{
			JSONTag:   f.Tag.Get("json"),
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
	ContentType   ContentType `json:"content_type"`
	Name          string      `json:"name"`
	Program       []string    `json:"program"`
	Subject       []string    `json:"subject"`
	Developmental []string    `json:"developmental"`
	Skills        []string    `json:"skills"`
	Age           []string    `json:"age"`
	Grade         []string    `json:"grade"`
	Keywords      []string    `json:"keywords"`
	Description   string      `json:"description"`
	Thumbnail     string      `json:"thumbnail"`
	SuggestTime   int         `json:"suggest_time"`
	//RejectReason  string      `json:"reject_reason"`

	Outcomes []string `json:"outcomes"`

	PublishScope string `json:"publish_scope"`

	Data  string `json:"data"`
	Extra string `json:"extra"`
}

func (c CreateContentRequest) Validate() error {
	if c.Name == "" {
		return ErrRequireContentName
	}
	if c.PublishScope == "" {
		return ErrRequirePublishScope
	}
	if c.Thumbnail != "" {
		parts := strings.Split(c.Thumbnail, "-")
		if len(parts) != 2 {
			return ErrInvalidResourceId
		}
		// _, exist := storage.DefaultStorage().ExistFile(ctx, parts[0], parts[1])
		// if !exist {
		// 	return ErrResourceNotFound
		// }
	}
	return nil
}

type ContentInfoWithDetails struct {
	ContentInfo
	ContentTypeName   string   `json:"content_type_name"`
	ProgramName       []string `json:"program_name"`
	SubjectName       []string `json:"subject_name"`
	DevelopmentalName []string `json:"developmental_name"`
	SkillsName        []string `json:"skills_name"`
	AgeName           []string `json:"age_name"`
	GradeName         []string `json:"grade_name"`
	OrgName           string   `json:"org_name"`
}

type ContentName struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ContentInfo struct {
	ID            string      `json:"id"`
	ContentType   ContentType `json:"content_type"`
	Name          string      `json:"name"`
	Program       []string    `json:"program"`
	Subject       []string    `json:"subject"`
	Developmental []string    `json:"developmental"`
	Skills        []string    `json:"skills"`
	Age           []string    `json:"age"`
	Grade         []string    `json:"grade"`
	Keywords      []string    `json:"keywords"`
	Description   string      `json:"description"`
	Thumbnail     string      `json:"thumbnail"`
	Version       int64       `json:"version"`
	SuggestTime   int         `json:"suggest_time"`

	Outcomes []string `json:"outcomes"`

	SourceID     string `json:"source_id"`
	LockedBy     string `json:"locked_by"`
	RejectReason string `json:"reject_reason"`
	LatestID     string `json:"latest_id"`

	Data  string `json:"data"`
	Extra string `json:"extra"`

	Author     string `json:"author"`
	AuthorName string `json:"author_name"`
	Org        string `json:"org"`

	PublishScope  string               `json:"publish_scope"`
	PublishStatus ContentPublishStatus `json:"publish_status"`

	CreatedAt int64 `json:"created_at"`
}

type ContentData interface {
	Unmarshal(ctx context.Context, data string) error
	Marshal(ctx context.Context) (string, error)

	Validate(ctx context.Context, contentType ContentType) error
	PrepareResult(ctx context.Context) error
	SubContentIds(ctx context.Context) ([]string, error)
}

func (cInfo *ContentInfo) SetStatus(status ContentPublishStatus) error {
	switch status {
	case ContentStatusArchive:
		if cInfo.allowedToArchive() {
			cInfo.PublishStatus = ContentStatusArchive
			return nil
		}
	case ContentStatusAttachment:
		//TODO
		fmt.Println(cInfo.PublishStatus)
	case ContentStatusDraft:
		//TODO
		fmt.Println(cInfo.PublishStatus)
	case ContentStatusHidden:
		if cInfo.allowedToHidden() {
			cInfo.PublishStatus = ContentStatusHidden
			return nil
		}
	case ContentStatusPending:
		if cInfo.allowedToPending() {
			cInfo.PublishStatus = ContentStatusPending
			return nil
		}
	case ContentStatusPublished:
		if cInfo.allowedToBeReviewed() {
			cInfo.PublishStatus = ContentStatusPublished
			return nil
		}
		fmt.Println(cInfo.PublishStatus)
	case ContentStatusRejected:
		if cInfo.allowedToBeReviewed() {
			cInfo.PublishStatus = ContentStatusRejected
			return nil
		}
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
