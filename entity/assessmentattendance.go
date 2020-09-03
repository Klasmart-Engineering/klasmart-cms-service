package entity

type AssessmentAttendance struct {
	ID           string `json:"id"`
	AssessmentID string `json:"assessment_id"`
	AttendanceID string `json:"attendance_id"`
}

func (AssessmentAttendance) TableName() string {
	return "assessments_attendances"
}
