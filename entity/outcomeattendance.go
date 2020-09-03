package entity

type OutcomeAttendance struct {
	ID           string `json:"id"`
	OutcomeID    string `json:"outcome_id"`
	AttendanceID string `json:"attendance_id"`
}

func (OutcomeAttendance) TableName() string {
	return "outcomes_attendances"
}
