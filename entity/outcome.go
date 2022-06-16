package entity

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

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

var Shortcode3Validate = regexp.MustCompile(`^[A-Z0-9]{3}$`)
var Shortcode5Validate = regexp.MustCompile(`^[A-Z0-9]{5}$`)

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
	RowNumber      int      `json:"row_number"`
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
	ExistError bool                       `json:"exist_error"`
}

type VerifyImportOutcomeView struct {
	RowNumber      int                         `json:"row_number"`
	OutcomeName    string                      `json:"outcome_name" binding:"required"`
	Shortcode      VerifyImportOutcomeResults  `json:"shortcode"`
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
	ScoreThreshold float32                     `json:"score_threshold"`
}

type VerifyImportOutcomeResult struct {
	Value string `json:"value"`
	Error string `json:"error,omitempty"`
}

type VerifyImportOutcomeResults struct {
	Value  string   `json:"value"`
	Errors []string `json:"errors,omitempty"`
}

type ImportOutcomeRequest struct {
	CreateData []*ImportOutcomeView `json:"create_data"`
	UpdateData []*ImportOutcomeView `json:"update_data"`
}

type ImportOutcomeView struct {
	RowNumber      int      `json:"row_number" binding:"required"`
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
	ScoreThreshold float32  `json:"score_threshold"`
}

func (v ImportOutcomeView) ConvertToPendingOutcome(ctx context.Context, op *Operator) (*Outcome, error) {
	shortcodeNum, err := utils.BHexToNum(ctx, v.Shortcode)
	if err != nil {
		log.Error(ctx, "utils.BHexToNum error",
			log.Err(err),
			log.String("shortcode", v.Shortcode))
		return nil, errors.New("shortcode is invalid")
	}

	outcome := &Outcome{
		ID:             utils.NewID(),
		Name:           v.OutcomeName,
		Assumed:        v.Assumed,
		Description:    v.Description,
		Shortcode:      v.Shortcode,
		ShortcodeNum:   shortcodeNum,
		OrganizationID: op.OrgID,
		AuthorID:       op.UserID,
		PublishStatus:  OutcomeStatusPending,
		PublishScope:   op.OrgID,
	}

	// TODO if assumed is true, then scoreThreshold is 0
	outcome.ScoreThreshold = v.ScoreThreshold

	programIDs := utils.SliceDeduplicationExcludeEmpty(v.Program)
	subjectIDs := utils.SliceDeduplicationExcludeEmpty(v.Subject)
	categoryIDs := utils.SliceDeduplicationExcludeEmpty(v.Category)
	subCategoryIDs := utils.SliceDeduplicationExcludeEmpty(v.Subcategory)
	gradeIDs := utils.SliceDeduplicationExcludeEmpty(v.Grade)
	ageIDs := utils.SliceDeduplicationExcludeEmpty(v.Age)

	outcome.Keywords = strings.Join(v.Keywords, JoinComma)
	outcome.Programs = programIDs
	outcome.Subjects = subjectIDs
	outcome.Categories = categoryIDs
	outcome.Subcategories = subCategoryIDs
	outcome.Grades = gradeIDs
	outcome.Ages = ageIDs

	outcome.Sets = make([]*Set, len(v.Sets))
	for i := range v.Sets {
		set := &Set{
			ID: v.Sets[i],
		}
		outcome.Sets[i] = set
	}

	return outcome, nil
}

func (v ImportOutcomeView) ConvertToVerifyView(ctx context.Context) *VerifyImportOutcomeView {
	verifyView := &VerifyImportOutcomeView{
		RowNumber:   v.RowNumber,
		OutcomeName: v.OutcomeName,
		Shortcode: VerifyImportOutcomeResults{
			Value: v.Shortcode,
		},
		Assumed:        v.Assumed,
		Description:    v.Description,
		Keywords:       v.Keywords,
		Program:        v.Program,
		Subject:        v.Subject,
		Category:       v.Category,
		Subcategory:    v.Subcategory,
		Age:            v.Age,
		Grade:          v.Grade,
		ScoreThreshold: v.ScoreThreshold,
	}

	verifyView.Sets = make([]VerifyImportOutcomeResult, len(v.Sets))
	for i := range v.Sets {
		verifyView.Sets[i] = VerifyImportOutcomeResult{
			Value: v.Sets[i],
		}
	}

	return verifyView
}
