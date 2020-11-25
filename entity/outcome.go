package entity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

const (
	OutcomeStatusDraft     = "draft"
	OutcomeStatusPending   = "pending"
	OutcomeStatusPublished = "published"
	OutcomeStatusRejected  = "rejected"
	OutcomeStatusHidden    = "hidden"
)

type OutcomeStatus string
type Outcome struct {
	ID            string `gorm:"type:varchar(50);column:id" dynamodbav:"outcome_id" json:"outcome_id" dynamoupdate:"-"`
	Name          string `gorm:"type:varchar(255);NOT NULL;column:name" dynamodbav:"outcome_name" json:"outcome_name" dynamoupdate:":n"`
	Shortcode     string `gorm:"type:char(8);NOT NULL;column:shortcode" dynamodbav:"shortcode" json:"shortcode" dynamoupdate:":code"`
	AncestorID    string `gorm:"type:varchar(50);column:ancestor_id" dynamodbav:"ancestor_id" json:"ancestor_id" dynamoupdate:"-"`
	Program       string `gorm:"type:varchar(1024);NOT NULL;column:program" dynamodbav:"program" json:"program" dynamoupdate:":p"`
	Subject       string `gorm:"type:varchar(1024);NOT NULL;column:subject" dynamodbav:"subject" json:"subject" dynamoupdate:":su"`
	Developmental string `gorm:"type:varchar(1024);NOT NULL;column:developmental" dynamodbav:"developmental" json:"developmental" dynamoupdate:":dv"`
	Skills        string `gorm:"type:varchar(1024);NOT NULL;column:skills" dynamodbav:"skills" json:"skills" dynamoupdate:":sk"`
	Age           string `gorm:"type:varchar(1024);NOT NULL;column:age" dynamodbav:"age" json:"age" dynamoupdate:":a"`
	Grade         string `gorm:"type:varchar(1024);NOT NULL;column:grade" dynamodbav:"grade" json:"grade" dynamoupdate:":grd"`
	Keywords      string `gorm:"type:text;NOT NULL;column:keywords" dynamodbav:"keywords" json:"keywords" dynamoupdate:":ky"`
	Description   string `gorm:"type:text;NOT NULL;column:description" dynamodbav:"description" json:"description" dynamoupdate:":de"`

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
}

func (Outcome) TableName() string {
	return "learning_outcomes"
}

func (oc Outcome) GetID() interface{} {
	return oc.ID
}

func (oc *Outcome) Update(data *Outcome) {
	if data.Name != "" {
		oc.Name = data.Name
	}

	oc.Assumed = data.Assumed
	oc.Program = data.Program
	oc.Subject = data.Subject
	oc.Developmental = data.Developmental
	oc.Skills = data.Skills
	oc.Age = data.Age
	oc.Grade = data.Grade
	oc.EstimatedTime = data.EstimatedTime
	oc.Keywords = data.Keywords
	oc.Description = data.Description
	oc.PublishStatus = OutcomeStatusDraft
	oc.UpdateAt = time.Now().Unix()
}

func (oc *Outcome) Clone(op *Operator) Outcome {
	now := time.Now().Unix()
	return Outcome{
		ID:            utils.NewID(),
		AncestorID:    oc.AncestorID,
		Shortcode:     oc.Shortcode,
		Name:          oc.Name,
		Program:       oc.Program,
		Subject:       oc.Subject,
		Developmental: oc.Developmental,
		Skills:        oc.Skills,
		Age:           oc.Age,
		Grade:         oc.Grade,
		Keywords:      oc.Keywords,
		Description:   oc.Description,

		EstimatedTime:  oc.EstimatedTime,
		AuthorID:       op.UserID,
		AuthorName:     oc.AuthorName,
		OrganizationID: oc.OrganizationID,

		PublishStatus: OutcomeStatusDraft,
		PublishScope:  oc.PublishScope,
		LatestID:      oc.LatestID,

		Version:  1,
		SourceID: oc.ID,
		Assumed:  oc.Assumed,

		CreateAt: now,
		UpdateAt: now,
	}
}

func (oc *Outcome) SetStatus(ctx context.Context, status OutcomeStatus) error {
	switch status {
	case OutcomeStatusHidden:
		if oc.allowedToHidden() {
			oc.PublishStatus = OutcomeStatusHidden
			return nil
		}
	case OutcomeStatusPending:
		if oc.allowedToPending() {
			oc.PublishStatus = OutcomeStatusPending
			return nil
		}
	case OutcomeStatusPublished:
		if oc.allowedToBeReviewed() {
			oc.PublishStatus = OutcomeStatusPublished
			return nil
		}
	case OutcomeStatusRejected:
		if oc.allowedToBeReviewed() {
			oc.PublishStatus = OutcomeStatusRejected
			return nil
		}
	}
	err := errors.New(fmt.Sprintf("unsupported:[%s]", status))
	log.Error(ctx, "SetStatus failed",
		log.Err(err),
		log.String("status", string(status)))
	return err
}

func (oc Outcome) allowedToArchive() bool {
	switch oc.PublishStatus {
	case OutcomeStatusPublished:
		return true
	}
	return false
}

func (oc Outcome) allowedToAttachment() bool {
	// TODO
	return false
}

func (oc Outcome) allowedToPending() bool {
	switch oc.PublishStatus {
	case OutcomeStatusDraft, OutcomeStatusRejected:
		return true
	}
	return false
}

func (oc Outcome) allowedToBeReviewed() bool {
	switch oc.PublishStatus {
	case OutcomeStatusPending:
		return true
	}
	return false
}

func (oc Outcome) allowedToHidden() bool {
	switch oc.PublishStatus {
	case OutcomeStatusPublished:
		return true
	}
	return false
}

type OutcomeCondition struct {
	IDs            []string `json:"ids" form:"ids"`
	OutcomeName    string   `json:"outcome_name" form:"outcome_name"`
	Description    string   `json:"description" form:"description"`
	Keywords       string   `json:"keywords" form:"keywords"`
	Shortcode      string   `json:"shortcode" form:"shortcode"`
	AuthorID       string   `json:"author_id" form:"author_id"`
	AuthorName     string   `json:"author_name" form:"author_name"`
	Page           int      `json:"page" form:"page"`
	PageSize       int      `json:"page_size" form:"page_size"`
	OrderBy        string   `json:"order_by" form:"order_by"`
	PublishStatus  string   `json:"publish_status" form:"publish_status"`
	FuzzyKey       string   `json:"search_key" form:"search_key"`
	FuzzyAuthorIDs []string `json:"fuzzy_authors" form:"fuzzy_authors"`
	Assumed        int      `json:"assumed" form:"assumed"`
	PublishScope   string   `json:"publish_scope" form:"publish_scope"`
	OrganizationID string   `json:"organization_id" form:"organization_id"`
}
