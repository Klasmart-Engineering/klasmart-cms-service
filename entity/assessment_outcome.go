package entity

type AssessmentOutcome struct {
	ID           string `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	AssessmentID string `gorm:"column:assessment_id;type:varchar(64);not null" json:"assessment_id"`
	OutcomeID    string `gorm:"column:outcome_id;type:varchar(64);not null" json:"outcome_id"`
	Skip         bool   `gorm:"column:skip;type:boolean;not null" json:"skip"`
	NoneAchieved bool   `gorm:"column:none_achieved;type:boolean;not null" json:"none_achieved"`
}

func (AssessmentOutcome) TableName() string {
	return "assessments_outcomes"
}
