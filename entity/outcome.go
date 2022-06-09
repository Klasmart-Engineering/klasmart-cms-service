package entity

import "github.com/KL-Engineering/kidsloop-cms-service/constant"

const (
	OutcomeTable = "learning_outcomes"
)

const (
	OutcomeStatusDraft     = "draft"
	OutcomeStatusPending   = "pending"
	OutcomeStatusPublished = "published"
	OutcomeStatusRejected  = "rejected"
	OutcomeStatusHidden    = "hidden"
)

const (
	JoinComma = ","
)

type OutcomeStatus string

type Outcome struct {
	ID           string `gorm:"type:varchar(50);column:id" dynamodbav:"outcome_id" json:"outcome_id" dynamoupdate:"-"`
	Name         string `gorm:"type:varchar(255);NOT NULL;column:name" dynamodbav:"outcome_name" json:"outcome_name" dynamoupdate:":n"`
	Shortcode    string `gorm:"type:char(8);DEFAULT NULL;column:shortcode" dynamodbav:"shortcode" json:"shortcode" dynamoupdate:":code"`
	ShortcodeNum int    `gorm:"type:int;NOT NULL;column:shortcode_num" dynamodbav:"shortcode_num" json:"shortcode_cum" dynamoupdate:"-"`
	AncestorID   string `gorm:"type:varchar(50);column:ancestor_id" dynamodbav:"ancestor_id" json:"ancestor_id" dynamoupdate:"-"`
	Keywords     string `gorm:"type:text;NOT NULL;column:keywords" dynamodbav:"keywords" json:"keywords" dynamoupdate:":ky"`
	Description  string `gorm:"type:text;NOT NULL;column:description" dynamodbav:"description" json:"description" dynamoupdate:":de"`

	EstimatedTime  int    `gorm:"type:int;NOT NULL;column:estimated_time" dynamodbav:"estimated_time" json:"extra" dynamoupdate:":est"`
	AuthorID       string `gorm:"type:varchar(50);NOT NULL;column:author_id" dynamodbav:"author_id" json:"author" dynamoupdate:":au"`
	AuthorName     string `gorm:"type:varchar(128);NOT NULL;column:author_name" dynamodbav:"author_name" json:"author_name" dynamoupdate:":aun"`
	OrganizationID string `gorm:"type:varchar(50);NOT NULL;column:organization_id" dynamodbav:"org_id" json:"organization_id" dynamoupdate:":og"`

	PublishScope  string        `gorm:"type:varchar(50);NOT NULL;column:publish_scope;index" dynamodbav:"publish_scope" json:"publish_scope" dynamoupdate:":ps"`
	PublishStatus OutcomeStatus `gorm:"type:varchar(16);NOT NULL;column:publish_status;index" dynamodbav:"publish_status" json:"publish_status" dynamoupdate:":pst"`

	RejectReason string `gorm:"type:varchar(255);NOT NULL;column:reject_reason" dynamodbav:"reject_reason" json:"reject_reason" dynamoupdate:":rr"`
	Version      int    `gorm:"type:int;NOT NULL;column:version" dynamodbav:"version" json:"version" dynamoupdate:":ve"`
	LockedBy     string `gorm:"type:varchar(50);NOT NULL;column:locked_by" dynamodbav:"locked_by" json:"locked_by" dynamoupdate:":lb"`
	SourceID     string `gorm:"type:varchar(255);NOT NULL;column:source_id" dynamodbav:"source_id" json:"source_id" dynamoupdate:":si"`
	LatestID     string `gorm:"type:varchar(255);NOT NULL;column:latest_id" dynamodbav:"latest_id" json:"latest_id" dynamoupdate:":lsi"`
	Assumed      bool   `gorm:"type:tinyint(255);NOT NULL;column:assumed" dynamodbav:"assumed" json:"assumed" dynamoupdate:":asum"`

	CreateAt int64 `gorm:"type:bigint;NOT NULL;column:create_at" dynamodbav:"created_at" json:"created_at" dynamoupdate:":ca"`
	UpdateAt int64 `gorm:"type:bigint;NOT NULL;column:update_at" dynamodbav:"updated_at" json:"updated_at" dynamoupdate:":ua"`
	DeleteAt int64 `gorm:"type:bigint;column:delete_at" dynamodbav:"deleted_at" json:"deleted_at" dynamoupdate:":da"`

	Sets           []*Set       `gorm:"-" json:"sets"`
	Programs       []string     `gorm:"-" json:"programs"`
	Subjects       []string     `gorm:"-" json:"subjects"`
	Categories     []string     `gorm:"-" json:"categories" `
	Subcategories  []string     `gorm:"-" json:"subcategories"`
	Grades         []string     `gorm:"-" json:"grades"`
	Ages           []string     `gorm:"-" json:"ages"`
	Milestones     []*Milestone `gorm:"-" json:"milestones"`
	EditingOutcome *Outcome     `gorm:"-" json:"-"`

	ScoreThreshold float32 `gorm:"score_threshold"`
}

func (Outcome) TableName() string {
	return OutcomeTable
}

func (oc Outcome) GetID() interface{} {
	return oc.ID
}

func (oc Outcome) HasLocked() bool {
	return oc.LockedBy != "" && oc.LockedBy != constant.LockedByNoBody
}

const (
	OutcomeOrderByName = "name"
)

type OutcomeCondition struct {
	IDs            []string `json:"ids" form:"ids"`
	OutcomeName    string   `json:"outcome_name" form:"outcome_name"`
	Description    string   `json:"description" form:"description"`
	Keywords       string   `json:"keywords" form:"keywords"`
	Shortcode      string   `json:"shortcode" form:"shortcode"`
	AuthorID       string   `json:"-" form:"-"`
	AuthorName     string   `json:"author_name" form:"author_name"`
	Page           int      `json:"page" form:"page"`
	PageSize       int      `json:"page_size" form:"page_size"`
	OrderBy        string   `json:"order_by" form:"order_by"`
	PublishStatus  string   `json:"publish_status" form:"publish_status"`
	FuzzyKey       string   `json:"search_key" form:"search_key"`
	AuthorIDs      []string `json:"-" form:"-"`
	Assumed        int      `json:"assumed" form:"assumed"`
	PublishScope   string   `json:"publish_scope" form:"publish_scope"`
	OrganizationID string   `json:"organization_id" form:"organization_id"`
	SetName        string   `json:"set_name" form:"set_name"`
	IsLocked       *bool    `json:"is_locked"`

	ProgramIDs     []string `json:"program_ids" form:"program_ids"`
	SubjectIDs     []string `json:"subject_ids" form:"subject_ids"`
	CategoryIDs    []string `json:"category_ids" form:"category_ids"`
	SubCategoryIDs []string `json:"sub_category_ids" form:"sub_category_ids"`
	AgeIDs         []string `json:"age_ids" form:"age_ids"`
	GradeIDs       []string `json:"grade_ids" form:"grade_ids"`
}

type ExportOutcomeRequest struct {
	// Maximum array length is 50
	OutcomeIDs []string `json:"outcome_ids" binding:"max=50"`
	IsLocked   *bool    `json:"is_locked"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
}

type ExportOutcomeResponse struct {
	Data       []*ExportOutcomeView `json:"data"`
	TotalCount int                  `json:"total_count"`
}

type ExportOutcomeView struct {
	OutcomeID      string   `json:"outcome_id"`
	OutcomeName    string   `json:"outcome_name"`
	Shortcode      string   `json:"shortcode"`
	Assumed        bool     `json:"assumed"`
	Description    string   `json:"description"`
	Author         string   `json:"author"`
	Keywords       []string `json:"keywords"`
	Program        []string `json:"program"`
	Subject        []string `json:"subject"`
	Category       []string `json:"category"`
	Subcategory    []string `json:"subcategory"`
	Age            []string `json:"age"`
	Grade          []string `json:"grade"`
	Sets           []string `json:"sets"`
	Milestones     []string `json:"milestones"`
	CreatedAt      int64    `json:"created_at"`
	UpdatedAt      int64    `json:"updated_at"`
	ScoreThreshold float32  `json:"score_threshold"`
}

type VerifyImportOutcomeRequest struct {
	Data []*ImportOutcomeView `json:"data" binding:"gt=0,max=200"`
}

type VerifyImportOutcomeResponse struct {
	CreateData []*VerifyImportOutcomeView `json:"create_data"`
	UpdateData []*VerifyImportOutcomeView `json:"update_data"`
}

type VerifyImportOutcomeView struct {
	OutcomeName    string                      `json:"outcome_name" binding:"required"`
	Shortcode      VerifyImportOutcomeResult   `json:"shortcode"`
	Assumed        bool                        `json:"assumed"`
	Description    string                      `json:"description"`
	Keywords       []string                    `json:"keywords"`
	Program        []string                    `json:"program" binding:"gt=0"`
	Subject        []string                    `json:"subject" binding:"gt=0"`
	Category       []string                    `json:"category" binding:"gt=0"`
	Subcategory    []string                    `json:"subcategory"`
	Age            []string                    `json:"age"`
	Grade          []string                    `json:"grade"`
	Sets           []VerifyImportOutcomeResult `json:"sets"`
	Milestones     []VerifyImportOutcomeResult `json:"milestones"`
	ScoreThreshold float32                     `json:"score_threshold"`
}

type VerifyImportOutcomeResult struct {
	Value string `json:"value"`
	Error string `json:"error,omitempty"`
}

type ImportOutcomeRequest struct {
	Data []*ImportOutcomeView `json:"data" binding:"gt=0,max=200"`
}

type ImportOutcomeView struct {
	OutcomeName    string   `json:"outcome_name" binding:"required"`
	Shortcode      string   `json:"shortcode"`
	Assumed        bool     `json:"assumed"`
	Description    string   `json:"description"`
	Keywords       []string `json:"keywords"`
	Program        []string `json:"program" binding:"gt=0"`
	Subject        []string `json:"subject" binding:"gt=0"`
	Category       []string `json:"category" binding:"gt=0"`
	Subcategory    []string `json:"subcategory"`
	Age            []string `json:"age"`
	Grade          []string `json:"grade"`
	Sets           []string `json:"sets"`
	Milestones     []string `json:"milestones"`
	ScoreThreshold float32  `json:"score_threshold"`
}
