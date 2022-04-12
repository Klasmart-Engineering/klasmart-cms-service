package constant

import (
	"errors"
	"time"
)

const (
	KidsloopCN                  = "CN"
	TableNameSchedule           = "schedules"
	TableNameScheduleFeedback   = "schedules_feedbacks"
	TableNameScheduleRelation   = "schedules_relations"
	TableNameScheduleReview     = "schedules_reviews"
	TableNameFeedbackAssignment = "feedbacks_assignments"

	TableNameClassType   = "class_types"
	TableNameLessonType  = "lesson_types"
	TableNameUserSetting = "user_settings"

	TableNameProgramGroup = "programs_groups"

	TableNameOrganizationProperty = "organizations_properties"

	TableNameStudentUsageRecord = "student_usage_records"
)

const (
	ScheduleDefaultCacheExpiration = 3 * time.Minute
)

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrDuplicateRecord = errors.New("duplicate record")
	ErrExceededLimit   = errors.New("exceeded limit")
	ErrUnAuthorized    = errors.New("unauthorized")
	ErrFileNotFound    = errors.New("file not found")
	//ErrUnknown = errors.New("unknown error")
	ErrInvalidArgs        = errors.New("invalid args")
	ErrConflict           = errors.New("conflict")
	ErrOperateNotAllowed  = errors.New("operation not allowed")
	ErrInternalServer     = errors.New("internal server error")
	ErrForbidden          = errors.New("forbidden")
	ErrHasLocked          = errors.New("has locked")
	ErrOverflow           = errors.New("over flow")
	ErrOutOfDate          = errors.New("out of date")
	ErrSqlBuilderFailed   = errors.New("sql builder failed")
	ErrAmsHttpFailed      = errors.New("ams http failed")
	ErrBadUsageOfKl2Cache = errors.New("bad usage of kl2cache see log for detail")
	ErrAssertFailed       = errors.New("assert failed")
	ErrAmsDataFailed      = errors.New("ams data failed")
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
	LiveTokenExpiresAt              = 24 * 30 * time.Hour
	LiveTokenIssuedAt               = 30 * time.Second
	LiveClassTypeLiveTokenExpiresAt = 15 * time.Minute
)

const (
	LockedByNoBody = "-"

	NoVisibilitySetting = "-"

	FolderRootPath      = "/"
	FolderPathSeparator = "/"
)

const (
	ShortcodeBaseCustom    = 36
	ShortcodeShowLength    = 5
	ShortcodeMaxShowLength = 32
	ShortcodeSpace         = ShortcodeBaseCustom * ShortcodeBaseCustom * ShortcodeBaseCustom * ShortcodeBaseCustom * ShortcodeBaseCustom
	ShortcodeFindStep      = 200
)

const (
	URLOrganizationIDParameter = "org_id"
	DefaultRole                = "admin"
	DefaultSalt                = "Kidsloop2@GetHashK3y"
)

const (
	NoSearchItem = "{nothing}"
	Self         = "{self}"

	ShareToAll = "{share_all}"

	TeacherManualSeparator  = "-"
	FolderItemLinkSeparator = "-"

	TeacherManualAssetsKeyword = "Teacher Manual"

	StringArraySeparator = ","
)

const (
	LoginByPassword = ""
	LoginByCode     = "temp_code"

	AccountPhone = "" // as default
	AccountEmail = "email"

	AccountExist      = "exist"
	AccountUnregister = "unregister"
)

const (
	DefaultPeriod     uint8 = 120
	DefaultWindow     uint8 = 5
	ValidDays               = 30
	BounceMax               = 5
	BounceInterval          = 2 //unit, hour
	ShareAllBatchSize       = 1000
)

const (
	ScheduleAllowEditTime   = 5 * time.Minute
	ScheduleAllowGoLiveTime = 15 * time.Minute
)

const (
	// 150 * 3000
	ScheduleRelationBatchInsertCount = 3000
	// 750 * 800
	ScheduleBatchInsertCount = 800
)

const (
	MilestoneNameLength = 200
)

const (
	CookieKey = "Cookie"
)

const (
	VisibilitySettingsGroupOrg    = "org"
	VisibilitySettingsGroupSchool = "school"
)

const (
	ResponseHeaderKeyInternalErrorType = "X-Internal-Error-Type"
)

const (
	UserDefaultCacheExpiration           = 4 * time.Hour
	UserPermissionDefaultCacheExpiration = 15 * time.Minute
)

const (
	DBDefaultConnMaxLifetime = 4 * time.Hour
	DBDefaultSlowThreshold   = 200 * time.Millisecond
)

const (
	ContentFolderQueryCacheRefreshDuration = time.Minute
	ContentFolderQueryCacheExpiration      = 0 // never expire
)

const (
	ReportQueryCacheExpiration      = time.Minute * 5
	ReportQueryCacheRefreshDuration = time.Minute
)
