package entity

type AssessmentAttendance struct {
	ID           string `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	AssessmentID string `gorm:"column:assessment_id;type:varchar(64);not null" json:"assessment_id"`
	AttendanceID string `gorm:"column:attendance_id;type:varchar(64);not null" json:"attendance_id"`
}

func (AssessmentAttendance) TableName() string {
	return "assessments_attendances"
}
