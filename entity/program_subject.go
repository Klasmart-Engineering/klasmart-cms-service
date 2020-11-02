package entity

type ProgramSubject struct {
	ID        string `gorm:"column:id;PRIMARY_KEY"`
	ProgramID string `gorm:"column:program_id;type:varchar(256);not null"`
	SubjectID string `gorm:"column:subject_id;type:varchar(256);not null"`
}

func (e ProgramSubject) TableName() string {
	return "programs_subjects"
}

func (e ProgramSubject) GetID() interface{} {
	return e.ID
}
