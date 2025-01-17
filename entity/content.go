package entity

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/KL-Engineering/kidsloop-cms-service/utils"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
)

const (
	ContentStatusDraft      = "draft"
	ContentStatusPending    = "pending"
	ContentStatusPublished  = "published"
	ContentStatusRejected   = "rejected"
	ContentStatusAttachment = "attachment"
	ContentStatusHidden     = "hidden"
	ContentStatusArchive    = "archive"

	AliasContentTypeFolder = 10

	MaterialInputSourceH5p              = 1
	MaterialInputSourceDisk             = 2
	MaterialInputSourceAssets           = 3
	MaterialInputSourceBadanamuAppToWeb = 100

	FileTypeImage            = 1
	FileTypeVideo            = 2
	FileTypeAudio            = 3
	FileTypeDocument         = 4
	FileTypeH5p              = 5
	FileTypeH5pExtend        = 6
	FileTypeBadanamuAppToWeb = 100

	FileTypeAssetsTypeOffset = 9

	SelfStudyTrue     = 1
	SelfStudyFalse    = 2
	DrawActivityTrue  = 1
	DrawActivityFalse = 2
	//LessonTypeTest    = "1"
	//LessonTypeNotTest = "2"

	ContentAuthed   ContentAuth = 1
	ContentUnauthed ContentAuth = 2

	ContentPropertyTypeProgram     ContentPropertyType = 1
	ContentPropertyTypeSubject     ContentPropertyType = 2
	ContentPropertyTypeCategory    ContentPropertyType = 3
	ContentPropertyTypeAge         ContentPropertyType = 4
	ContentPropertyTypeGrade       ContentPropertyType = 5
	ContentPropertyTypeSubCategory ContentPropertyType = 6

	PublishedQueryModeOnlyOwner  PublishedQueryMode = "query only owner"
	PublishedQueryModeAll        PublishedQueryMode = "query all"
	PublishedQueryModeOnlyOthers PublishedQueryMode = "query only others"
	PublishedQueryModeNone       PublishedQueryMode = "query none"
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

const (
	ContentTypeMaterial = 1
	ContentTypePlan     = 2
	ContentTypeAssets   = 3
)

type ContentPropertyType int

type FileType int

type ContentAuth int

type PublishedQueryMode string

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
	case FileTypeH5pExtend:
		return FileTypeH5pExtend
	case FileTypeBadanamuAppToWeb:
		return FileTypeBadanamuAppToWeb
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
	case FileTypeH5pExtend:
		return "extend h5p"
	case FileTypeBadanamuAppToWeb:
		return "Badanamu App to Web"
	default:
		return "unknown"
	}
}

func NewContentType(contentType int) ContentType {
	switch contentType {
	case ContentTypeMaterial:
		return ContentTypeMaterial
	case ContentTypePlan:
		return ContentTypePlan
	case ContentTypeAssets:
		return ContentTypeAssets
	case AliasContentTypeFolder:
		return AliasContentTypeFolder
	default:
		return ContentTypeAssets
	}
}

func (c ContentType) Validate() error {
	switch c {
	case ContentTypeMaterial:
		return nil
	case ContentTypePlan:
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
	case ContentTypePlan:
		return []int{ContentTypePlan}
	case ContentTypeMaterial:
		return []int{ContentTypeMaterial}
	case ContentTypeAssets:
		return []int{ContentTypeAssets}
	case AliasContentTypeFolder:
		return []int{AliasContentTypeFolder}
	}
	return []int{ContentTypePlan}
}

func (c ContentType) ContentType() []ContentType {
	switch c {
	case ContentTypePlan:
		return []ContentType{ContentTypePlan}
	case ContentTypeMaterial:
		return []ContentType{ContentTypeMaterial}
	case ContentTypeAssets:
		return []ContentType{ContentTypeAssets}
	}
	return []ContentType{ContentTypePlan}
}

func (c ContentType) Name() string {
	switch c {
	case ContentTypePlan:
		return "Plan"
	case ContentTypeMaterial:
		return "Material"
	case ContentTypeAssets:
		return "Assets"
	}
	return "Unknown"
}

type NullContentType struct {
	Value ContentType
	Valid bool
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

type ContentStatisticsInfo struct {
	SubContentCount int `json:"subcontent_count"`
	OutcomesCount   int `json:"outcomes_count"`
}

func ContentLink(id string) string {
	return string(FolderFileTypeContent) + "-" + id
}

type ContentVisibilitySetting struct {
	ContentID         string `gorm:"type:char(50);NOT NULL;column:content_id"`
	VisibilitySetting string `gorm:"type:char(50);NOT NULL;column:visibility_setting;index"`
}

func (ContentVisibilitySetting) TableName() string {
	return "cms_content_visibility_settings"
}

type ContentProperty struct {
	PropertyType ContentPropertyType `gorm:"type:int;column:property_type"`
	ContentID    string              `gorm:"type: varchar(50);column:content_id"`
	PropertyID   string              `gorm:"type: varchar(50);column:property_id"`
	Sequence     int                 `gorm:"type: index;column:sequence"`
}

func (ContentProperty) TableName() string {
	return "cms_content_properties"
}

type ContentWithVisibilitySettings struct {
	Content
	VisibilitySettings []string
}

type Content struct {
	ID          string      `gorm:"type:varchar(50);PRIMARY_KEY"`
	ContentType ContentType `gorm:"type:int;NOT NULL; column:content_type"`
	Name        string      `gorm:"type:varchar(255);NOT NULL;column:content_name"`
	Keywords    string      `gorm:"type:text;NOT NULL;column:keywords"`
	Description string      `gorm:"type:text;NOT NULL;column:description"`
	Thumbnail   string      `gorm:"type:text;NOT NULL;column:thumbnail"`

	SourceType string `gorm:"type:varchar(256); column:source_type"`

	Outcomes string `gorm:"type:text;NOT NULL;column:outcomes"`
	Data     string `gorm:"type:json;NOT NULL;column:data"`
	Extra    string `gorm:"type:text;NOT NULL;column:extra"`

	SuggestTime int    `gorm:"type:int;NOT NULL;column:suggest_time"`
	Author      string `gorm:"type:varchar(50);NOT NULL;column:author"`
	Creator     string `gorm:"type:varchar(50);NOT NULL;column:creator"`
	Org         string `gorm:"type:varchar(50);NOT NULL;column:org"`

	SelfStudy    BoolTinyInt `gorm:"type:tinyint;NOT NULL;column:self_study"`
	DrawActivity BoolTinyInt `gorm:"type:tinyint;NOT NULL;column:draw_activity"`
	LessonType   string      `gorm:"type:varchar(100);column:lesson_type"`

	PublishStatus ContentPublishStatus `gorm:"type:varchar(16);NOT NULL;column:publish_status;index"`

	RejectReason string `gorm:"type:varchar(255);NOT NULL;column:reject_reason"`
	Remark       string `gorm:"type:varchar(255);NOT NULL;column:remark"`
	Version      int64  `gorm:"type:int;NOT NULL;column:version"`
	LockedBy     string `gorm:"type:varchar(50);NOT NULL;column:locked_by"`
	SourceID     string `gorm:"type:varchar(50);NOT NULL;column:source_id"`
	LatestID     string `gorm:"type:varchar(50);NOT NULL;column:latest_id"`

	CopySourceID string `gorm:"type:varchar(50);column:copy_source_id"`

	ParentFolder string `gorm:"type:varchar(50);NOT NULL;column:parent_folder"`
	DirPath      Path   `gorm:"type:varchar(2048);column:dir_path"`

	CreateAt int64 `gorm:"type:bigint;NOT NULL;column:create_at"`
	UpdateAt int64 `gorm:"type:bigint;NOT NULL;column:update_at"`
	DeleteAt int64 `gorm:"type:bigint;column:delete_at"`
}

func (c *Content) ToContentSimplified() *ContentSimplified {
	return &ContentSimplified{
		ID:            c.ID,
		ContentName:   c.Name,
		ContentType:   c.ContentType,
		AuthorID:      c.Author,
		Data:          c.Data,
		CreateAt:      c.CreateAt,
		PublishStatus: c.PublishStatus,
	}
}
func (u Content) UpdateExpress() string {
	tags := getDynamoTags(u)
	updateExpressParts := make([]string, 0)
	for i := range tags {
		updateExpressParts = append(updateExpressParts, tags[i].JSONTag+" = "+tags[i].DynamoTag)
	}
	updateExpress := strings.Join(updateExpressParts, constant.StringArraySeparator)
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

func (t TinyIntBool) Int() BoolTinyInt {
	if t {
		return 1
	} else {
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

type CopyContentRequest struct {
	ContentID string `json:"content_id"`
	Deep      bool   `json:"deep"`
}

type ContentProperties struct {
	ContentID   string   `json:"content_id"`
	Program     string   `json:"program"`
	Subject     []string `json:"subject"`
	Category    []string `json:"developmental"`
	SubCategory []string `json:"skills"`
	Age         []string `json:"age"`
	Grade       []string `json:"grade"`
}
type ContentVisibilitySettings struct {
	ContentID          string   `json:"content_id"`
	VisibilitySettings []string `json:"visibility_settings"`
}

type ContentInternalConditionRequest struct {
	IDs          []string `json:"ids"`
	OrgID        string   `json:"org_id"`
	ContentType  int      `json:"content_type"`
	CreateAtLe   int      `json:"create_at_le"`
	CreateAtGe   int      `json:"create_at_ge"`
	PlanID       string   `json:"plan_id"`
	DataSourceID string   `json:"source_id"`
	ScheduleID   string   `json:"schedule_id"`
}

type ContentConditionRequest struct {
	Name               string   `json:"name"`
	ContentType        []int    `json:"content_type"`
	VisibilitySettings []string `json:"visibility_settings"`
	PublishStatus      []string `json:"publish_status"`
	Author             string   `json:"author"`
	Org                string   `json:"org"`
	Program            []string `json:"program"`
	SourceType         string   `json:"source_type"`
	DirPath            string   `json:"dir_path"`
	ParentID           string   `json:"parent_id"`
	ContentName        string   `json:"content_name"`
	DataSourceID       string   `json:"data_source_id"`

	//AuthedContentFlag bool           `json:"authed_content"`

	ContentIDs  NullStrings `json:"content_ids"`
	AuthedOrgID NullStrings `json:"authed_org_ids"`
	ParentsPath NullStrings `json:"parents_path"`
	OrderBy     string      `json:"order_by"`
	Pager       utils.Pager

	JoinUserIDList []string `json:"join_user_id_list"`

	PublishedQueryMode PublishedQueryMode `json:"published_query_mode"`
	GroupNames         []string           `json:"group_names"`
	ProgramIDs         []string           `json:"program_ids"`
	SubjectIDs         []string           `json:"subject_ids"`
	CategoryIDs        []string           `json:"category_ids"`
	SubCategoryIDs     []string           `json:"sub_category_ids"`
	AgeIDs             []string           `json:"age_ids"`
	GradeIDs           []string           `json:"grade_ids"`
	LessonPlanName     string             `json:"lesson_plan_name"`
}
type OrganizationOrSchool struct {
	ID    string
	Name  string
	Group string
}

type ContentPermission struct {
	ID             string `json:"id"`
	AllowEdit      bool   `json:"allow_edit"`
	AllowDelete    bool   `json:"allow_delete"`
	AllowApprove   bool   `json:"allow_approve"`
	AllowReject    bool   `json:"allow_reject"`
	AllowRepublish bool   `json:"allow_republish"`
}

type CreateContentRequest struct {
	ContentType ContentType `json:"content_type"`
	SourceType  string      `json:"source_type"`
	Name        string      `json:"name"`
	Program     string      `json:"program"`
	Subject     []string    `json:"subject"`
	Category    []string    `json:"developmental"`
	SubCategory []string    `json:"skills"`
	Age         []string    `json:"age"`
	Grade       []string    `json:"grade"`
	Keywords    []string    `json:"keywords"`
	Description string      `json:"description"`
	Thumbnail   string      `json:"thumbnail"`
	SuggestTime int         `json:"suggest_time"`

	SelfStudy    TinyIntBool `json:"self_study"`
	DrawActivity TinyIntBool `json:"draw_activity"`
	LessonType   string      `json:"lesson_type"`

	Outcomes []string `json:"outcomes"`

	PublishScope []string `json:"publish_scope"`

	Data  string `json:"data"`
	Extra string `json:"extra"`

	//TeacherManual     string `json:"teacher_manual"`
	//Name string `json:"teacher_manual_name"`
	TeacherManualBatch []*TeacherManualFile `json:"teacher_manual_batch"`

	ParentFolder string `json:"parent_folder"`
}

type TeacherManualFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *CreateContentRequest) Trim() {
	if c.Name != "" {
		c.Name = strings.TrimSpace(c.Name)
	}
}

func (c CreateContentRequest) Validate() error {
	if c.Name == "" {
		return ErrRequireContentName
	}
	if len(c.PublishScope) == 0 {
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

type ContentInfoWithDetailsResponse struct {
	Total       int                       `json:"total"`
	ContentList []*ContentInfoWithDetails `json:"list"`
}
type QueryContentResponse struct {
	Total int                 `json:"total"`
	List  []*QueryContentItem `json:"list"`
}

type QueryContentItem struct {
	ID              string               `json:"id"`
	ContentType     ContentType          `json:"content_type"`
	Name            string               `json:"name"`
	Thumbnail       string               `json:"thumbnail"`
	AuthorName      string               `json:"author_name"`
	Data            string               `json:"data"`
	Author          string               `json:"author"`
	PublishStatus   ContentPublishStatus `json:"publish_status"`
	ContentTypeName string               `json:"content_type_name"`
	Permission      ContentPermission    `json:"permission"`
	SuggestTime     int                  `json:"suggest_time"`
}

type QuerySharedContentV2Response struct {
	Total int                         `json:"total"`
	Items []*QuerySharedContentV2Item `json:"items"`
}
type QuerySharedContentV2Item struct {
	ID            string               `json:"id"`
	DirPath       string               `json:"dir_path"`
	ContentType   ContentType          `json:"content_type"`
	Name          string               `json:"name"`
	Thumbnail     string               `json:"thumbnail"`
	AuthorName    string               `json:"author_name"`
	Author        string               `json:"author"`
	PublishStatus ContentPublishStatus `json:"publish_status"`
}

type FolderContentInfoWithDetailsResponse struct {
	Total       int                  `json:"total"`
	ContentList []*FolderContentData `json:"list"`
}

type ContentSimplifiedList struct {
	Total             int                       `json:"total"`
	ContentList       []*ContentSimplified      `json:"list"`
	StudentContentMap []*ScheduleStudentContent `json:"student_content_map"`
}

type ScheduleStudentContent struct {
	StudentID  string   `json:"student_id"`
	ContentIDs []string `json:"content_ids"`
}

type ContentSimplified struct {
	ID          string `json:"id"`
	ContentName string `json:"content_name"`

	ContentType   ContentType          `json:"content_type"`
	AuthorID      string               `json:"author_id"`
	Data          string               `json:"data"`
	CreateAt      int64                `json:"create_at"`
	PublishStatus ContentPublishStatus `json:"publish_status"`
}

type ContentInfoWithDetails struct {
	ContentInfo
	ContentTypeName  string   `json:"content_type_name"`
	ProgramName      string   `json:"program_name"`
	SubjectName      []string `json:"subject_name"`
	CategoryName     []string `json:"developmental_name"`
	SubCategoryName  []string `json:"skills_name"`
	AgeName          []string `json:"age_name"`
	GradeName        []string `json:"grade_name"`
	OrgName          string   `json:"org_name"`
	PublishScopeName []string `json:"publish_scope_name"`
	LessonTypeName   string   `json:"lesson_type_name"`

	//AuthorName string `json:"author_name"`
	CreatorName string `json:"creator_name"`

	OutcomeEntities []*Outcome `json:"outcome_entities"`

	IsMine bool `json:"is_mine"`

	Permission ContentPermission `json:"permission"`
}

type ContentName struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	ContentType ContentType `json:"content_type"`
	LatestID    string      `json:"latest_id"`
	OutcomeIDs  []string    `json:"outcome_ids"`
}

type ContentInfoInternal struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	ContentType ContentType `json:"content_type"`
	LatestID    string      `json:"latest_id"`
	OutcomeIDs  []string    `json:"outcome_ids"`
	FileType    FileType    `json:"file_type"`
}

//Content in folder
type FolderContent struct {
	ID              string      `json:"id"`
	ContentName     string      `json:"name"`
	ContentType     ContentType `json:"content_type"`
	Description     string      `json:"description"`
	Keywords        string      `json:"keywords"`
	Author          string      `json:"author"`
	ItemsCount      int         `json:"items_count"`
	PublishStatus   string      `json:"publish_status"`
	Thumbnail       string      `json:"thumbnail"`
	Data            string      `json:"data"`
	AuthorName      string      `json:"author_name"`
	DirPath         string      `json:"dir_path"`
	ContentTypeName string      `json:"content_type_name"`
	CreateAt        int         `json:"create_at"`
	UpdateAt        int         `json:"update_at"`
}

//Content in folder
type FolderContentData struct {
	ID              string            `json:"id"`
	ContentName     string            `json:"name"`
	ContentType     ContentType       `json:"content_type"`
	Description     string            `json:"description"`
	Keywords        []string          `json:"keywords"`
	Author          string            `json:"author"`
	ItemsCount      int               `json:"items_count"`
	PublishStatus   string            `json:"publish_status"`
	Thumbnail       string            `json:"thumbnail"`
	Data            string            `json:"data"`
	AuthorName      string            `json:"author_name"`
	DirPath         string            `json:"dir_path"`
	ContentTypeName string            `json:"content_type_name"`
	CreateAt        int               `json:"create_at"`
	UpdateAt        int               `json:"update_at"`
	Permission      ContentPermission `json:"permission"`
}

type ContentInfo struct {
	ID          string      `json:"id"`
	ContentType ContentType `json:"content_type"`
	Name        string      `json:"name"`
	Program     string      `json:"program"`
	Subject     []string    `json:"subject"`
	Category    []string    `json:"developmental"`
	SubCategory []string    `json:"skills"`
	Age         []string    `json:"age"`
	Grade       []string    `json:"grade"`
	Keywords    []string    `json:"keywords"`
	Description string      `json:"description"`
	Thumbnail   string      `json:"thumbnail"`
	Version     int64       `json:"version"`
	SuggestTime int         `json:"suggest_time"`
	SourceType  string      `json:"source_type"`
	AuthorName  string      `json:"author_name"`

	SelfStudy    TinyIntBool `json:"self_study"`
	DrawActivity TinyIntBool `json:"draw_activity"`
	LessonType   string      `json:"lesson_type"`

	Outcomes []string `json:"outcomes"`

	SourceID     string   `json:"source_id"`
	LockedBy     string   `json:"locked_by"`
	RejectReason []string `json:"reject_reason"`
	Remark       string   `json:"remark"`
	LatestID     string   `json:"latest_id"`

	Data  string `json:"data"`
	Extra string `json:"extra"`

	//TeacherManual     []string `json:"teacher_manual"`
	//Name []string `json:"teacher_manual_name"`
	TeacherManualBatch []*TeacherManualFile `json:"teacher_manual_batch"`
	Author             string               `json:"author"`
	Creator            string               `json:"creator"`
	Org                string               `json:"org"`

	PublishScope  []string             `json:"publish_scope"`
	PublishStatus ContentPublishStatus `json:"publish_status"`

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

type ExtraDataInRequest struct {
	TeacherManualBatch []*TeacherManualFile `json:"teacher_manual_batch"`
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

type LessonPlanForSchedule struct {
	ID        string              `json:"id" gorm:"column:id" `
	Name      string              `json:"name" gorm:"column:name" `
	GroupName LessonPlanGroupName `json:"group_name" gorm:"column:group_name" `
	ProgramID string              `json:"-" gorm:"column:program_id" `
}

type GetLessonPlansCanScheduleResponse struct {
	Total int                      `json:"total"`
	Data  []*LessonPlanForSchedule `json:"data"`
}
