package entity

import (
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"time"
)

type Outcome struct {
	ID            string `gorm:"type:varchar(50);PRIMARY_KEY;AUTO_INCREMENT" dynamodbav:"outcome_id" json:"outcome_id" dynamoupdate:"-"`
	AncestorID    string `gorm:"type:varchar(50);column:ancestor_id" dynamodbav:"ancestor_id" json:"outcome_id" dynamoupdate:"-"`
	Shortcode     string `gorm:"type:char(8);NOT NULL;column:shortcode" dynamodbav:"shortcode" json:"shortcode" dynamoupdate:":code"`
	Name          string `gorm:"type:varchar(255);NOT NULL;column:outcome_name" dynamodbav:"outcome_name" json:"outcome_name" dynamoupdate:":n"`
	Program       string `gorm:"type:varchar(1024);NOT NULL;column:program" dynamodbav:"program" json:"program" dynamoupdate:":p"`
	Subject       string `gorm:"type:varchar(1024);NOT NULL;column:subject" dynamodbav:"subject" json:"subject" dynamoupdate:":su"`
	Developmental string `gorm:"type:varchar(1024);NOT NULL;column:developmental" dynamodbav:"developmental" json:"developmental" dynamoupdate:":dv"`
	Skills        string `gorm:"type:varchar(1024);NOT NULL;column:skills" dynamodbav:"skills" json:"skills" dynamoupdate:":sk"`
	Age           string `gorm:"type:varchar(1024);NOT NULL;column:age" dynamodbav:"age" json:"age" dynamoupdate:":a"`
	Grade         string `gorm:"type:varchar(1024);NOT NULL;column:grade" dynamodbav:"grade" json:"grade" dynamoupdate:":grd"`
	Keywords      string `gorm:"type:text;NOT NULL;column:keywords" dynamodbav:"keywords" json:"keywords" dynamoupdate:":ky"`
	Description   string `gorm:"type:text;NOT NULL;column:description" dynamodbav:"description" json:"description" dynamoupdate:":de"`

	EstimatedTime  int64  `gorm:"type:int;NOT NULL;column:estimated_time" dynamodbav:"estimated_time" json:"extra" dynamoupdate:":est"`
	AuthorID       string `gorm:"type:varchar(50);NOT NULL;column:author_id" dynamodbav:"author_id" json:"author" dynamoupdate:":au"`
	AuthorName     string `gorm:"type:varchar(128);NOT NULL;column:author_name" dynamodbav:"author_name" json:"author_name" dynamoupdate:":aun"`
	OrganizationID string `gorm:"type:varchar(50);NOT NULL;column:org_id" dynamodbav:"org_id" json:"organization_id" dynamoupdate:":og"`

	PublishScope  string               `gorm:"type:varchar(50);NOT NULL;column:publish_scope;index" dynamodbav:"publish_scope" json:"publish_scope" dynamoupdate:":ps"`
	PublishStatus ContentPublishStatus `gorm:"type:varchar(16);NOT NULL;column:publish_status;index" dynamodbav:"publish_status" json:"publish_status" dynamoupdate:":pst"`

	RejectReason string `gorm:"type:varchar(255);NOT NULL;column:reject_reason" dynamodbav:"reject_reason" json:"reject_reason" dynamoupdate:":rr"`
	Version      int64  `gorm:"type:int;NOT NULL;column:version" dynamodbav:"version" json:"version" dynamoupdate:":ve"`
	LockedBy     string `gorm:"type:varchar(50);NOT NULL;column:locked_by" dynamodbav:"locked_by" json:"locked_by" dynamoupdate:":lb"`
	SourceID     string `gorm:"type:varchar(255);NOT NULL;column:source_id" dynamodbav:"source_id" json:"source_id" dynamoupdate:":si"`
	LatestID     string `gorm:"type:varchar(255);NOT NULL;column:latest_id" dynamodbav:"latest_id" json:"latest_id" dynamoupdate:":lsi"`
	Assumed      bool   `gorm:"type:tinyint(255);NOT NULL;column:assumed" dynamodbav:"assumed" json:"latest_id" dynamoupdate:":asum"`

	CreateAt int64 `gorm:"type:bigint;NOT NULL;column:create_at" dynamodbav:"created_at" json:"created_at" dynamoupdate:":ca"`
	UpdateAt int64 `gorm:"type:bigint;NOT NULL;column:update_at" dynamodbav:"updated_at" json:"updated_at" dynamoupdate:":ua"`
	DeleteAt int64 `gorm:"type:bigint;column:delete_at" dynamodbav:"deleted_at" json:"deleted_at" dynamoupdate:":da"`
}

func (oc *Outcome) Update(data *Outcome) {
	if data.Name != "" {
		oc.Name = data.Name
	}

	oc.Program = data.Program
	oc.Subject = data.Subject
	oc.Developmental = data.Subject
	oc.Skills = data.Skills
	oc.Age = data.Age
	oc.Grade = data.Grade
	oc.EstimatedTime = data.EstimatedTime
	oc.Keywords = data.Keywords
	oc.Description = data.Description
}

func (oc *Outcome) Clone() Outcome {
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
		AuthorID:       oc.AuthorID,
		AuthorName:     oc.AuthorName,
		OrganizationID: oc.OrganizationID,

		PublishStatus: ContentStatusDraft,

		Version:  1,
		SourceID: oc.ID,
		Assumed:  oc.Assumed,

		CreateAt: now,
		UpdateAt: now,
	}
}

func (oc *Outcome) SetStatus(status ContentPublishStatus) error {
	switch status {
	//case ContentStatusArchive:
	//	if oc.allowedToArchive() {
	//		oc.PublishStatus = ContentStatusArchive
	//	}
	//	return nil
	//case ContentStatusDraft:
	//	//TODO
	case ContentStatusHidden:
		if oc.allowedToHidden() {
			oc.PublishStatus = ContentStatusHidden
		}
		return nil
	case ContentStatusPending:
		if oc.allowedToPending() {
			oc.PublishStatus = ContentStatusPending
		}
		return nil
	case ContentStatusPublished:
		if oc.allowedToBeReviewed() {
			oc.PublishStatus = ContentStatusPublished
		}
		return nil
	case ContentStatusRejected:
		if oc.allowedToBeReviewed() {
			oc.PublishStatus = ContentStatusRejected
		}
		return nil
	}
	return errors.New(fmt.Sprintf("unsupported:[%s]", status))
}

func (oc Outcome) allowedToArchive() bool {
	switch oc.PublishStatus {
	case ContentStatusPublished:
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
	case ContentStatusDraft:
		return true
	}
	return false
}

func (oc Outcome) allowedToBeReviewed() bool {
	switch oc.PublishStatus {
	case ContentStatusPending:
		return true
	}
	return false
}

func (oc Outcome) allowedToHidden() bool {
	switch oc.PublishStatus {
	case ContentStatusPublished:
		return true
	}
	return false
}

func (oc Outcome) CanBeCancelled() bool {
	if oc.PublishStatus == ContentStatusDraft {
		return true
	}
	return false
}

func (oc Outcome) CanBeDeleted() bool {
	if oc.PublishStatus == ContentStatusArchive {
		return true
	}
	return false
}
