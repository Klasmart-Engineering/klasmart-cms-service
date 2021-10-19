package entity

type ClassAttendanceRequest struct {
	ClassID                 string      `json:"class_id" binding:"required"`
	StudentID               string      `json:"student_id" binding:"required"`
	SelectedSubjectIDList   []string    `json:"selected_subject_id_list" binding:"gt=0"`
	UnSelectedSubjectIDList []string    `json:"un_selected_subject_id_list"`
	Durations               []TimeRange `json:"durations" binding:"gt=0"`
}

type ClassAttendanceResponse struct {
	AttendedCount  int                            `json:"attendedCount"`
	ScheduledCount int                            `json:"scheduledCount"`
	Items          []*ClassAttendanceResponseItem `json:"items"`
}

type ClassAttendanceResponseItem struct {
	Duration TimeRange `json:"duration"`

	AttendancePercentage                          float64 `json:"attendance_percentage"`
	ClassAverageAttendancePercentage              float64 `json:"class_average_attendance_percentage"`
	UnSelectedSubjectsAverageAttendancePercentage float64 `json:"un_selected_subjects_average_attendance_percentage"`
}
