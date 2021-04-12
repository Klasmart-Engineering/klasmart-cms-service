package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

type AssessmentContent struct {
	ID             string                   `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	AssessmentID   string                   `gorm:"column:assessment_id;type:varchar(64);not null" json:"assessment_id"`
	ContentID      string                   `gorm:"column:content_id;type:varchar(64);not null" json:"content_id"`
	ContentName    string                   `gorm:"column:content_name;type:varchar(255);not null" json:"content_name"`
	ContentType    ContentType              `gorm:"column:content_type;type:int;not null" json:"content_type"`
	ContentComment string                   `gorm:"column:content_comment;type:text;not null" json:"content_comment"`
	Checked        bool                     `gorm:"column:checked;type:boolean;not null" json:"checked"`
	OutcomeIDs     utils.SQLJSONStringArray `gorm:"column:outcome_ids;type:json;not null" json:"outcome_ids"`
}

func (AssessmentContent) TableName() string {
	return "assessments_contents"
}
