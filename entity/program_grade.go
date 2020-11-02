package entity

type ProgramGrade struct {
	ID        string `gorm:"column:id;PRIMARY_KEY"`
	ProgramID string `gorm:"column:program_id;type:varchar(256);not null"`
	GradeID   string `gorm:"column:grade_id;type:varchar(256);not null"`
}

func (e ProgramGrade) TableName() string {
	return "programs_grades"
}

func (e ProgramGrade) GetID() interface{} {
	return e.ID
}
