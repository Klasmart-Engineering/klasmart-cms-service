package constant

import (
	"errors"
	"time"
)

const (
	TableNameTag             = "tags"
	TableNameSchedule        = "schedules"
	TableNameScheduleTeacher = "schedules_teachers"
)

const (
	ScheduleDefaultCacheExpiration = 3 * time.Minute
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

const (
	PresignDurationMinutes       = 60 * 24 * time.Minute
	PresignUploadDurationMinutes = 60 * time.Minute
)

const (
	LiveTokenExpiresAt = 24 * 30 * time.Hour
	LiveTokenIssuedAt  = 30 * time.Second
)
