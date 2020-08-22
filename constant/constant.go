package constant

import "errors"

const (
	TableNameTag             = "tags"
	TableNameSchedule        = "schedules"
	TableNameTeacherSchedule = "teachers_schedules"
)

type GSIName string

const (
	GSI_TeacherSchedule_TeacherAtStartAt GSIName = "teacher_id_and_start_at"
)

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrDuplicateRecord = errors.New("duplicate record")
	ErrExceededLimit   = errors.New("exceeded limit")
	ErrUnauthorized    = errors.New("unauthorized")
	//ErrUnknown = errors.New("unknown error")
)

// tag States
const (
	Enable   = 1
	Disabled = 2
)

const (
	DefaultPageSize  = 10
	DefaultPageIndex = 1
)
