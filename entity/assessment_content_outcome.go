package entity

type AssessmentContentOutcome struct {
	ID           string `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	AssessmentID string `gorm:"column:assessment_id;type:varchar(64);not null" json:"assessment_id"`
	ContentID    string `gorm:"column:content_id;type:varchar(64);not null" json:"content_id"`
	OutcomeID    string `gorm:"column:outcome_id;type:varchar(64);not null" json:"outcome_id"`
}

func (AssessmentContentOutcome) TableName() string {
	return "assessment_contents_outcomes"
}
