package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type FeedbackAssignment struct {
	ID             string `json:"id" gorm:"column:id;PRIMARY_KEY"`
	FeedbackID     string `json:"feedback_id" gorm:"column:feedback_id;type:varchar(100)"`
	AssignmentUrl  string `json:"attachment_id" gorm:"column:assignment_url;type:varchar(500)"`
	AssignmentName string `json:"attachment_name" gorm:"column:assignment_name;type:varchar(100)"`
	Number         int    `json:"number" gorm:"column:number;type:int"`

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
