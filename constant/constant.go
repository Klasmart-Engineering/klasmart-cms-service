package constant

import "errors"

const (
	TableNameTag             = "tags"
	TableNameSchedule        = "schedules"
	TableNameScheduleTeacher = "schedules_teachers"
)

type GSIName string

func (g GSIName) String() string {
	return string(g)
}

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrDuplicateRecord = errors.New("duplicate record")
	ErrExceededLimit   = errors.New("exceeded limit")
	ErrUnAuthorized    = errors.New("unauthorized")
	ErrFileNotFound    = errors.New("file not found")
	//ErrUnknown = errors.New("unknown error")
	ErrInvalidArgs = errors.New("invalid args")
	ErrConflict    = errors.New("conflict")
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
