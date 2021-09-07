package entity

type ClassesAssignmentOverViewRequest struct {
	ClassIDs  []string    `json:"class_ids" form:"class_ids"`
	Durations []TimeRange `json:"durations" form:"durations"`
}

type ClassesAssignmentOverView struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type ClassesAssignmentsViewRequest struct {
	ClassIDs  []string    `json:"class_ids" form:"class_ids"`
	Durations []TimeRange `json:"durations" form:"durations"`
}

type ClassesAssignmentsDurationRatio struct {
	Key   string  `json:"key"`
	Ratio float32 `json:"ratio"`
}

type ClassesAssignmentsView struct {
	ClassID        string                            `json:"class_id"`
	Total          int                               `json:"total"`
	DurationsRatio []ClassesAssignmentsDurationRatio `json:"durations_ratio"`
}

type ClassesAssignmentsUnattendedStudentsViewRequest struct {
	ClassID   string
	Page      int         `json:"page" form:"page"`
	PageSize  int         `json:"page_size" form:"page_size"`
	Durations []TimeRange `json:"durations" form:"durations"`
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

type ClassesAssignmentsRecords struct {
	ID              string `json:"id"`
	ClassID         string `json:"class_id"`
	ScheduleID      string `json:"schedule_id"`
	AttendanceID    string `json:"attendance_id"`
	ScheduleType    string `json:"schedule_type"`
	ScheduleStartAt int64  `json:"schedule_start_at"`
}

func (ClassesAssignmentsRecords) TableName() string {
	return "classes_assignments_records"
}
