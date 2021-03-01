package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ScheduleFeedback struct {
	ID            string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	ParentID      string `json:"parent_id" gorm:"column:parent_id;type:varchar(100)"`
	ScheduleID    string `json:"schedule_id" gorm:"column:schedule_id;type:varchar(100)"`
	UserID        string `json:"user_id" gorm:"column:user_id;type:varchar(100)"`
	AssignmentUrl string `json:"assignment_url" gorm:"column:assignment_url;type:text"`
	Comment       string `json:"comment" gorm:"column:comment;type:text"`

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
	ScheduleFeedback
}
