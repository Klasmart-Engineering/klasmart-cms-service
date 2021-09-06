package entity

type ClassesAssignmentCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}
type ClassesAssignmentsOverView []ClassesAssignmentCount

type ClassesAssignmentsDurationRatio struct {
	Key   string  `json:"key"`
	Ratio float32 `json:"ratio"`
}
type ClassesAssignmentDetailView struct {
	ClassID        string                            `json:"class_id"`
	Total          int                               `json:"total"`
	DurationsRatio []ClassesAssignmentsDurationRatio `json:"durations_ratio"`
}

type ClassesAssignmentsUnattendedStudentsView struct {
	StudentID string `json:"student_id"`
	Schedule  struct {
		ScheduleID   string `json:"schedule_id"`
		ScheduleName string `json:"schedule_name"`
		Type         string `json:"type"`
	} `json:"schedule"`
	Time int64 `json:"time"`
}
