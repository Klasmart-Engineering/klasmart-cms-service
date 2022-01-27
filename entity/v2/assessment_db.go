package v2

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type Assessment struct {
	ID             string           `gorm:"column:id;PRIMARY_KEY"`
	OrgID          string           `gorm:"org_id"`
	ScheduleID     string           `gorm:"schedule_id"`
	AssessmentType AssessmentType   `gorm:"assessment_type"`
	Title          string           `gorm:"title"`
	Status         AssessmentStatus `gorm:"status"`
	CompleteAt     int64            `gorm:"complete_at"`
	ClassLength    int              `gorm:"class_length"`
	ClassEndAt     int64            `gorm:"class_end_at"`
	MigrateFlag    int              `gorm:"migrate_flag"`

	CreateAt int64 `gorm:"column:create_at;type:bigint"`
	UpdateAt int64 `gorm:"column:update_at;type:bigint"`
	DeleteAt int64 `gorm:"column:delete_at;type:bigint"`
}

func (Assessment) TableName() string {
	return constant.TableNameAssessmentV2
}

type AssessmentUser struct {
	ID             string               `gorm:"column:id;PRIMARY_KEY"`
	AssessmentID   string               `gorm:"assessment_id"`
	UserID         string               `gorm:"user_id"`
	UserType       AssessmentUserType   `gorm:"user_type"`
	StatusBySystem AssessmentUserStatus `gorm:"status_by_system"`
	StatusByUser   AssessmentUserStatus `gorm:"status_by_user"`

	CreateAt int64 `gorm:"column:create_at;type:bigint"`
	UpdateAt int64 `gorm:"column:update_at;type:bigint"`
	DeleteAt int64 `gorm:"column:delete_at;type:bigint"`
}

func (AssessmentUser) TableName() string {
	return constant.TableNameAssessmentsUsersV2
}

// AssessmentContent: from ContentLibrary and ScheduleAttachment
type AssessmentContent struct {
	ID              string                  `gorm:"column:id;PRIMARY_KEY"`
	AssessmentID    string                  `gorm:"assessment_id"`
	ContentID       string                  `gorm:"content_id"`
	ContentType     AssessmentContentType   `gorm:"content_type"`
	Status          AssessmentContentStatus `gorm:"status"`
	ReviewerComment string                  `gorm:"reviewer_comment"`

	CreateAt int64 `gorm:"column:create_at;type:bigint"`
	UpdateAt int64 `gorm:"column:update_at;type:bigint"`
	DeleteAt int64 `gorm:"column:delete_at;type:bigint"`
}

func (AssessmentContent) TableName() string {
	return constant.TableNameAssessmentsContentsV2
}

type AssessmentReviewerFeedback struct {
	ID                string                  `gorm:"column:id;PRIMARY_KEY"`
	AssessmentUserID  string                  `gorm:"assessment_user_id"`
	CompleteAt        int64                   `gorm:"complete_at"`
	Status            UserResultProcessStatus `gorm:"status"`
	ReviewerID        string                  `gorm:"reviewer_id"`
	StudentFeedbackID string                  `gorm:"student_feedback_id"`
	AssessScore       AssessmentUserAssess    `gorm:"assess_score"`
	ReviewerComment   string                  `gorm:"reviewer_comment"`

	CreateAt int64 `gorm:"column:create_at;type:bigint"`
	UpdateAt int64 `gorm:"column:update_at;type:bigint"`
	DeleteAt int64 `gorm:"column:delete_at;type:bigint"`
}

func (AssessmentReviewerFeedback) TableName() string {
	return constant.TableNameAssessmentReviewerFeedbackV2
}

type AssessmentUserOutcome struct {
	ID                  string                      `gorm:"column:id;PRIMARY_KEY"`
	AssessmentUserID    string                      `gorm:"assessment_user_id"`
	AssessmentContentID string                      `gorm:"assessment_content_id"`
	OutcomeID           string                      `gorm:"outcome_id"`
	Status              AssessmentUserOutcomeStatus `gorm:"status"`

	CreateAt int64 `gorm:"column:create_at;type:bigint"`
	UpdateAt int64 `gorm:"column:update_at;type:bigint"`
	DeleteAt int64 `gorm:"column:delete_at;type:bigint"`
}

func (AssessmentUserOutcome) TableName() string {
	return constant.TableNameAssessmentsUsersOutcomesV2
}

type AssessmentUserResultDBView struct {
	AssessmentReviewerFeedback
	ScheduleID   string `gorm:"schedule_id"`
	AssessmentID string `gorm:"assessment_id"`
	UserID       string `gorm:"user_id"`
	Title        string `gorm:"title"`
}
