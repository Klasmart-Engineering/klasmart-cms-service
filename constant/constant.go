package constant

import (
	"errors"
	"time"
)

const (
	TableNameSchedule        = "schedules"
	TableNameScheduleTeacher = "schedules_teachers"

	TableNameAge               = "ages"
	TableNameClassType         = "class_types"
	TableNameDevelopmental     = "developmentals"
	TableNameGrade             = "grades"
	TableNameLessonType        = "lesson_types"
	TableNameProgram           = "programs"
	TableNameSkill             = "skills"
	TableNameSubject           = "subjects"
	TableNameVisibilitySetting = "visibility_settings"

	TableNameProgramAge         = "programs_ages"
	TableNameProgramDevelopment = "programs_developments"
	TableNameProgramGrade       = "programs_grades"
	TableNameProgramSubject     = "programs_subjects"
	TableNameDevelopmentalSkill = "developments_skills"
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
	ErrInvalidArgs       = errors.New("invalid args")
	ErrConflict          = errors.New("conflict")
	ErrOperateNotAllowed = errors.New("operation not allowed")
	ErrInternalServer    = errors.New("internal server error")
	ErrForbidden         = errors.New("forbidden")
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

const (
	LockedByNoBody = "-"

	NoVisibilitySetting = "-"
)

const (
	ShortcodeBaseCustom = 36
	ShortcodeShowLength = 5
)

const (
	URLOrganizationIDParameter = "org_id"
	DefaultRole                = "admin"
	DefaultSalt                = "Kidsloop2@GetHashK3y"
)

const (
	NoSearchItem = "{nothing}"
	Self         = "{self}"
)

const (
	LoginByPassword = ""
	LoginByCode     = "temp_code"

	AccountPhone = "" // as default
	AccountEmail = "email"
)

const (
	DefaultPeriod uint8 = 120
	DefaultWindow uint8 = 5
)
