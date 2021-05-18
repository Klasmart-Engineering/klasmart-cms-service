package entity

type AssessmentContent struct {
	ID             string      `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	AssessmentID   string      `gorm:"column:assessment_id;type:varchar(64);not null" json:"assessment_id"`
	ContentID      string      `gorm:"column:content_id;type:varchar(64);not null" json:"content_id"`
	ContentName    string      `gorm:"column:content_name;type:varchar(255);not null" json:"content_name"`
	ContentType    ContentType `gorm:"column:content_type;type:int;not null" json:"content_type"`
	ContentComment string      `gorm:"column:content_comment;type:text;not null" json:"content_comment"`
	Checked        bool        `gorm:"column:checked;type:boolean;not null" json:"checked"`
}

func (AssessmentContent) TableName() string {
	return "assessments_contents"
}

type UpdateAssessmentContentArgs struct {
	ID      string `json:"id"`
	Checked bool   `json:"checked"`
	Comment string `json:"comment"`
}
