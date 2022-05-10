package entity

import "github.com/KL-Engineering/kidsloop-cms-service/constant"

type FeedbackAssignment struct {
	ID                 string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	FeedbackID         string `json:"feedback_id" gorm:"column:feedback_id;type:varchar(100)"`
	AttachmentID       string `json:"attachment_id" gorm:"column:attachment_id;type:varchar(255)"`
	AttachmentName     string `json:"attachment_name" gorm:"column:attachment_name;type:varchar(255)"`
	Number             int    `json:"number" gorm:"column:number;type:int"`
	ReviewAttachmentID string `json:"review_attachment_id" gorm:"column:review_attachment_id"`

	CreateAt int64 `json:"create_at" gorm:"column:create_at;type:bigint"`
	UpdateAt int64 `json:"-" gorm:"column:update_at;type:bigint"`
	DeleteAt int64 `json:"-" gorm:"column:delete_at;type:bigint"`
}

func (e FeedbackAssignment) TableName() string {
	return constant.TableNameFeedbackAssignment
}

func (e FeedbackAssignment) GetID() interface{} {
	return e.ID
}
