package entity

import (
	"context"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
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

	MaterialInputSourceH5p    = 1
	MaterialInputSourceDisk   = 2
	MaterialInputSourceAssets = 3

	FileTypeImage = 1
	FileTypeVideo = 2
	FileTypeAudio = 3
	FileTypeDocument = 4
	FileTypeH5p = 5

	FileTypeAssetsTypeOffset = 9

	SelfStudyTrue     = 1
	SelfStudyFalse    = 2
	DrawActivityTrue  = 1
	DrawActivityFalse = 2
	LessonTypeTest    = 1
	LessonTypeNotTest = 2
)

var (
	ErrRequireContentName  = errors.New("content name required")
	ErrRequirePublishScope = errors.New("publish scope required")
	ErrInvalidResourceId   = errors.New("invalid resource id")
	ErrInvalidContentType  = errors.New("invalid content type")
	ErrInvalidLessonType   = errors.New("invalid lesson type")
)

type ContentPublishStatus string
type ContentType int

type FileType int

func NewFileType(fileType int) FileType {
	switch fileType {
	case FileTypeVideo:
		return FileTypeVideo
	case FileTypeImage:
		return FileTypeImage
	case FileTypeAudio:
		return FileTypeAudio
	case FileTypeDocument:
		return FileTypeDocument
	case FileTypeH5p:
		return FileTypeH5p
	default:
		return FileTypeH5p
	}
}
func (f FileType) String() string {
	switch f {
	case FileTypeVideo:
		return "video"
	case FileTypeImage:
		return "image"
	case FileTypeAudio:
		return "audio"
	case FileTypeDocument:
		return "document"
	case FileTypeH5p:
		return "h5p"
	default:
		return "h5p"
	}
}


func NewContentType(contentType int) ContentType {
	switch contentType {
	case ContentTypeMaterial:
		return ContentTypeMaterial
	case ContentTypeLesson:
		return ContentTypeLesson
	case ContentTypeAssets:
		return ContentTypeAssets
	default:
		return ContentTypeAssets
	}
}

func (c ContentType) Validate() error {
	switch c {
	case ContentTypeMaterial:
		return nil
	case ContentTypeLesson:
		return nil
	case ContentTypeAssets:
		return nil
	}
	return ErrInvalidContentType
}

func (c ContentType) IsAsset() bool {
	switch c {
	case ContentTypeAssets:
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
		return []int{ContentTypeAssets}
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
		return []ContentType{ContentTypeAssets}
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
	ID            string      `gorm:"type:varchar(50);PRIMARY_KEY;AUTO_INCREMENT"`
	ContentType   ContentType `gorm:"type:int;NOT NULL; column:content_type"`
	Name          string      `gorm:"type:varchar(255);NOT NULL;column:content_name"`
	Program       string      `gorm:"type:varchar(1024);NOT NULL;column:program"`
	Subject       string      `gorm:"type:varchar(1024);NOT NULL;column:subject"`
	Developmental string      `gorm:"type:varchar(1024);NOT NULL;column:developmental"`
	Skills        string      `gorm:"type:varchar(1024);NOT NULL;column:skills"`
	Age           string      `gorm:"type:varchar(1024);NOT NULL;column:age"`
	Grade         string      `gorm:"type:varchar(1024);NOT NULL;column:grade"`
	Keywords      string      `gorm:"type:text;NOT NULL;column:keywords"`
	Description   string      `gorm:"type:text;NOT NULL;column:description"`
	Thumbnail     string      `gorm:"type:text;NOT NULL;column:thumbnail"`

	SourceType   string `gorm:"type:varchar(256);NOT NULL; column:source_type"`

	Outcomes string `gorm:"type:text;NOT NULL;column:outcomes"`
	Data     string `gorm:"type:json;NOT NULL;column:data"`
	Extra    string `gorm:"type:text;NOT NULL;column:extra"`

	SuggestTime int    `gorm:"type:int;NOT NULL;column:suggest_time"`
	Author      string `gorm:"type:varchar(50);NOT NULL;column:author"`
	AuthorName  string `gorm:"type:varchar(128);NOT NULL;column:author_name"`
	Org         string `gorm:"type:varchar(50);NOT NULL;column:org"`

	SelfStudy BoolTinyInt    `gorm:"type:tinyint;NOT NULL;column:self_study"`
	DrawActivity BoolTinyInt    `gorm:"type:tinyint;NOT NULL;column:draw_activity"`
	LessonType int    `gorm:"type:tinyint;NOT NULL;column:lesson_type"`

	PublishScope  string               `gorm:"type:varchar(50);NOT NULL;column:publish_scope;index"`
	PublishStatus ContentPublishStatus `gorm:"type:varchar(16);NOT NULL;column:publish_status;index"`

	RejectReason string `gorm:"type:varchar(255);NOT NULL;column:reject_reason"`
	Remark string `gorm:"type:varchar(255);NOT NULL;column:remark"`
	Version      int64  `gorm:"type:int;NOT NULL;column:version"`
	LockedBy     string `gorm:"type:varchar(50);NOT NULL;column:locked_by"`
	SourceID     string `gorm:"type:varchar(255);NOT NULL;column:source_id"`
	LatestID     string `gorm:"type:varchar(255);NOT NULL;column:latest_id"`

	CreateAt int64 `gorm:"type:bigint;NOT NULL;column:create_at"`
	UpdateAt int64 `gorm:"type:bigint;NOT NULL;column:update_at"`
	DeleteAt int64 `gorm:"type:bigint;column:delete_at"`
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
		if updateTag == constant.LockedByNoBody {
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

type TinyIntBool bool

func (t TinyIntBool) Int()BoolTinyInt{
	if t{
		return 1
	}else{
		return 2
	}
}

type BoolTinyInt int
func (b BoolTinyInt) Bool() TinyIntBool {
	if b == 1 {
		return true
	}
	return false
}

type CreateContentRequest struct {
	ContentType   ContentType `json:"content_type"`
	SourceType 	  string `json:"source_type"`
	Name          string      `json:"name"`
	Program       string    `json:"program"`
	Subject       []string    `json:"subject"`
	Developmental []string    `json:"developmental"`
	Skills        []string    `json:"skills"`
	Age           []string    `json:"age"`
	Grade         []string    `json:"grade"`
	Keywords      []string    `json:"keywords"`
	Description   string      `json:"description"`
	Thumbnail     string      `json:"thumbnail"`
	SuggestTime   int         `json:"suggest_time"`

	SelfStudy TinyIntBool `json:"self_study"`
	DrawActivity TinyIntBool `json:"draw_activity"`
	LessonType int `json:"lesson_type"`

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
	if c.LessonType > 0 {
		if c.LessonType != LessonTypeTest && c.LessonType != LessonTypeNotTest {
			return ErrInvalidLessonType
		}
	}

	return nil
}

type ContentInfoWithDetailsResponse struct {
	Total int `json:"total"`
	ContentList []*ContentInfoWithDetails `json:"list"`
}

type ContentInfoWithDetails struct {
	ContentInfo
	ContentTypeName   string   `json:"content_type_name"`
	ProgramName       string `json:"program_name"`
	SubjectName       []string `json:"subject_name"`
	DevelopmentalName []string `json:"developmental_name"`
	SkillsName        []string `json:"skills_name"`
	AgeName           []string `json:"age_name"`
	GradeName         []string `json:"grade_name"`
	OrgName           string   `json:"org_name"`
	PublishScopeName string `json:"publish_scope_name"`
	LessonTypeName string `json:"lesson_type_name"`

	OutcomeEntities	 []*Outcome `json:"outcome_entities"`
}

type ContentName struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	ContentType ContentType `json:"content_type"`
}

type SubContentsWithName struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Data ContentData `json:"data"`
}

type ContentInfo struct {
	ID            string      `json:"id"`
	ContentType   ContentType `json:"content_type"`
	Name          string      `json:"name"`
	Program       string    `json:"program"`
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
	SourceType	string `json:"source_type"`

	SelfStudy TinyIntBool `json:"self_study"`
	DrawActivity TinyIntBool `json:"draw_activity"`
	LessonType int `json:"lesson_type"`

	Outcomes []string `json:"outcomes"`

	SourceID     string `json:"source_id"`
	LockedBy     string `json:"locked_by"`
	RejectReason []string `json:"reject_reason"`
	Remark string `json:"remark"`
	LatestID     string `json:"latest_id"`

	Data  string `json:"data"`
	Extra string `json:"extra"`

	Author     string `json:"author"`
	AuthorName string `json:"author_name"`
	Org        string `json:"org"`

	PublishScope  string               `json:"publish_scope"`
	PublishStatus ContentPublishStatus `json:"publish_status"`

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

type ContentData interface {
	Unmarshal(ctx context.Context, data string) error
	Marshal(ctx context.Context) (string, error)

	Validate(ctx context.Context, contentType ContentType) error
	PrepareResult(ctx context.Context) error
	PrepareSave(ctx context.Context) error
	SubContentIds(ctx context.Context) []string
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
		fmt.Println(cInfo.PublishStatus)
		if cInfo.allowedToBeReviewed() {
			cInfo.PublishStatus = ContentStatusPublished
			return nil
		}
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
