package entity

type ClassAttendanceRequest struct {
	ClassID                 string      `json:"class_id" binding:"required"`
	StudentID               string      `json:"student_id" binding:"required"`
	SelectedSubjectIDList   []string    `json:"selected_subject_id_list" binding:"gt=0"`
	UnSelectedSubjectIDList []string    `json:"un_selected_subject_id_list"`
	Durations               []TimeRange `json:"durations" binding:"gt=0"`
}

type ClassAttendanceResponse struct {
	RequestStudentID string                         `json:"request_student_id"`
	Items            []*ClassAttendanceResponseItem `json:"items"`
	LabelID          string                         `json:"label_id"`
	LabelParams      AttedanceLabelParams           `json:"label_params"`
}

type AttedanceLabelParams struct {
	AttendedCount          int     `json:"attended_count"`
	ScheduledCount         int     `json:"scheduled_count"`
	LOCompareClass3week    float64 `json:"lo_compare_class_3_week"`
	AttendCompareLastWeek  float64 `json:"attend_compare_last_week"`
	AttendCompareLast3Week float64 `json:"attend_compare_last_3_week"`
	LOCompareClass         float64 `json:"lo_compare_class"`
}

type ClassAttendanceResponseItem struct {
	Duration                                      TimeRange `json:"duration"`
	AttendancePercentage                          float64   `json:"attendance_percentage"`
	ClassAverageAttendancePercentage              float64   `json:"class_average_attendance_percentage"`
	UnSelectedSubjectsAverageAttendancePercentage float64   `json:"un_selected_subjects_average_attendance_percentage"`
	AttendedCount                                 int       `json:"attended_count"`
	ScheduledCount                                int       `json:"scheduled_count"`
}

type ClassAttendance struct {
	ClassID      string `gorm:"column:class_id" json:"class_id"`
	SubjectID    string `gorm:"column:subject_id" json:"subject_id"`
	StudentID    string `gorm:"column:student_id" json:"student_id"`
	IsAttendance bool   `gorm:"column:is_attendance" json:"is_attendance"`
}

type ClassAttendanceQueryParameters struct {
	ClassID    string
	SubjectIDS []string
	Duration   TimeRange
}
