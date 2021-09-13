package entity

type ClassesAssignmentOverViewRequest struct {
	ClassIDs  []string    `json:"class_ids" form:"class_ids"`
	Durations []TimeRange `json:"durations" form:"durations"`
}

type ClassesAssignmentOverView struct {
	Type  string  `json:"type"`
	Count int     `json:"count"`
	Ratio float32 `json:"ratio"`
}

type ClassesAssignmentsViewRequest struct {
	ClassIDs  []string    `json:"class_ids" form:"class_ids"`
	Durations []TimeRange `json:"durations" form:"durations"`
	Type      string      `json:"type"`
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

type ClassesAssignmentsUnattendedViewRequest struct {
	ClassID   string
	Durations []TimeRange `json:"durations" form:"durations"`
	Type      string      `json:"type"`
}

type ScheduleView struct {
	ScheduleID   string `json:"schedule_id"`
	ScheduleName string `json:"schedule_name"`
	Type         string `json:"type"`
}
type ClassesAssignmentsUnattendedStudentsView struct {
	StudentID   string       `json:"student_id"`
	StudentName string       `json:"student_name"` // fill by frontend
	Schedule    ScheduleView `json:"schedule"`
	Time        int64        `json:"time"`
}

type ScheduleInReportType string

const (
	UnknownType ScheduleInReportType = "unknown"
	LiveType    ScheduleInReportType = "live"
	StudyType   ScheduleInReportType = "study"
	HomeFunType ScheduleInReportType = "home_fun"
)

func NewScheduleInReportType(classType ScheduleClassType, isFun bool) ScheduleInReportType {
	if classType == ScheduleClassTypeOnlineClass {
		return LiveType
	}

	if classType == ScheduleClassTypeHomework && !isFun {
		return StudyType
	}

	if classType == ScheduleClassTypeHomework && isFun {
		return HomeFunType
	}
	return UnknownType
}

type ClassesAssignmentsRecords struct {
	ID              string               `gorm:"column:id;primary_key" json:"id"`
	ClassID         string               `gorm:"column:class_id" json:"class_id"`
	ScheduleID      string               `gorm:"column:schedule_id" json:"schedule_id"`
	AttendanceID    string               `gorm:"column:attendance_id" json:"attendance_id"`
	ScheduleType    ScheduleInReportType `gorm:"column:schedule_type" json:"schedule_type"`
	ScheduleStartAt int64                `gorm:"column:schedule_start_at" json:"schedule_start_at"`
	FinishCount     int64                `gorm:"column:finish_counts" json:"finish_counts"`
	LastEndAt       int64                `gorm:"column:last_end_at" json:"last_end_at"`
	CreateAt        int64                `gorm:"column:create_at" json:"create_at"`
}

func (ClassesAssignmentsRecords) TableName() string {
	return "classes_assignments_records"
}

func (c ClassesAssignmentsRecords) GetBatchInsertColsAndValues() (cols []string, values []interface{}) {
	cols = append(cols, "id")
	values = append(values, c.ID)

	cols = append(cols, "class_id")
	values = append(values, c.ClassID)

	cols = append(cols, "schedule_id")
	values = append(values, c.ScheduleID)

	cols = append(cols, "attendance_id")
	values = append(values, c.AttendanceID)

	cols = append(cols, "finish_counts")
	values = append(values, c.FinishCount)

	cols = append(cols, "schedule_type")
	values = append(values, c.ScheduleType)

	cols = append(cols, "schedule_start_at")
	values = append(values, c.ScheduleStartAt)

	cols = append(cols, "last_end_at")
	values = append(values, c.LastEndAt)

	cols = append(cols, "create_at")
	values = append(values, c.CreateAt)

	return
}
