package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ScheduleFeedback struct {
	ID         string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	ScheduleID string `json:"schedule_id" gorm:"column:schedule_id;type:varchar(100)"`
	UserID     string `json:"user_id" gorm:"column:user_id;type:varchar(100)"`
	Comment    string `json:"comment" gorm:"column:comment;type:text"`

	CreateAt int64 `json:"create_at" gorm:"column:create_at;type:bigint"`
	UpdateAt int64 `json:"-" gorm:"column:update_at;type:bigint"`
	DeleteAt int64 `json:"-" gorm:"column:delete_at;type:bigint"`
}

func (e ScheduleFeedback) TableName() string {
	return constant.TableNameScheduleFeedback
}

func (e ScheduleFeedback) GetID() interface{} {
	return e.ID
}

type ScheduleFeedbackAddInput struct {
	ScheduleID  string                    `json:"schedule_id"`
	Comment     string                    `json:"comment"`
	Assignments []*FeedbackAssignmentView `json:"assignments"`
}

type FeedbackAssignmentView struct {
	Url    string `json:"attachment_id"`
	Name   string `json:"attachment_name"`
	Number int    `json:"number"`
}
type ScheduleFeedbackView struct {
	ScheduleFeedback
	Assignments []*FeedbackAssignmentView `json:"assignments"`
}
