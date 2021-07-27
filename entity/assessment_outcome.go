package entity

type AssessmentOutcome struct {
	ID           string `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	AssessmentID string `gorm:"column:assessment_id;type:varchar(64);not null" json:"assessment_id"`
	OutcomeID    string `gorm:"column:outcome_id;type:varchar(64);not null" json:"outcome_id"`
	Skip         bool   `gorm:"column:skip;type:boolean;not null" json:"skip"`                   // not attempted
	NoneAchieved bool   `gorm:"column:none_achieved;type:boolean;not null" json:"none_achieved"` // only for display
	Checked      bool   `gorm:"column:checked;type:boolean;not null" json:"checked"`
}

func (AssessmentOutcome) TableName() string {
	return "assessments_outcomes"
}
