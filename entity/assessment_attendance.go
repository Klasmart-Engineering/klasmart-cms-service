package entity

type AssessmentAttendance struct {
	ID           string                     `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	AssessmentID string                     `gorm:"column:assessment_id;type:varchar(64);not null" json:"assessment_id"`
	AttendanceID string                     `gorm:"column:attendance_id;type:varchar(64);not null" json:"attendance_id"`
	Checked      bool                       `gorm:"column:checked;type:boolean;not null" json:"checked"`
	Origin       AssessmentAttendanceOrigin `gorm:"column:origin;type:varchar(128);not null" json:"origin"`
	Role         AssessmentAttendanceRole   `gorm:"column:role;type:varchar(128);not null" json:"role"`
}

func (AssessmentAttendance) TableName() string {
	return "assessments_attendances"
}

type AssessmentAttendanceOrigin string

var (
	AssessmentAttendanceOriginClassRoaster AssessmentAttendanceOrigin = "class_roaster"
	AssessmentAttendanceOriginParticipants AssessmentAttendanceOrigin = "participants"
)

func (a AssessmentAttendanceOrigin) Valid() bool {
	switch a {
	case AssessmentAttendanceOriginClassRoaster, AssessmentAttendanceOriginParticipants:
		return true
	default:
		return false
	}
}

type AssessmentAttendanceRole string

var (
	AssessmentAttendanceRoleTeacher AssessmentAttendanceRole = "teacher"
	AssessmentAttendanceRoleStudent AssessmentAttendanceRole = "student"
)

func (a AssessmentAttendanceRole) Valid() bool {
	switch a {
	case AssessmentAttendanceRoleTeacher, AssessmentAttendanceRoleStudent:
		return true
	default:
		return false
	}
}

type NullAssessmentAttendanceRole struct {
	Value AssessmentAttendanceRole
	Valid bool
}

type AddAttendancesInput struct {
	AssessmentID      string
	ScheduleID        string
	AttendanceIDs     []string
	ScheduleRelations []*ScheduleRelation
}

type BatchAddAttendancesInput struct {
	Items []*BatchAddAttendancesInputItem
}

type BatchAddAttendancesInputItem struct {
	AssessmentID  string
	ScheduleID    string
	AttendanceIDs []string
}
