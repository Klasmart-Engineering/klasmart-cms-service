package da

const (
	RedisKeyPrefixContentId          = "content:id"
	RedisKeyPrefixContentLock        = "content:lock"
	RedisKeyPrefixContentReview      = "content:review"
	RedisKeyPrefixContentShared      = "content:shared"
	RedisKeyPrefixContentSharedV2    = "content:shared2"
	RedisKeyPrefixContentFolderQuery = "content:folder:query"

	RedisKeyPrefixScheduleID        = "schedule:id"
	RedisKeyPrefixScheduleCondition = "schedule:condition"

	RedisKeyPrefixOutcomeLock      = "outcome:lock"
	RedisKeyPrefixOutcomeReview    = "outcome:review"
	RedisKeyPrefixOutcomeShortcode = "outcome:shortcode"

	RedisKeyPrefixOutcomeCondition = "outcome:condition"
	RedisKeyPrefixOutcomeId        = "outcome:id"

	RedisKeyPrefixOutcomeSetLock = "outcome_set:lock"

	RedisKeyPrefixVerifyCodeLock = "verify_code:lock"

	RedisKeyPrefixFolderName  = "folder:name"
	RedisKeyPrefixFolderShare = "folder:share"

	RedisKeyPrefixShortcodeMute = "shortcode:lock"
	RedisKeyPrefixShortcode     = "shortcode"

	RedisKeyPrefixMilestoneMute        = "milestone:lock"
	RedisKeyPrefixGeneralMilestoneMute = "milestone:general:lock"

	RedisKeyPrefixUser     = "user"
	RedisKeyPrefixUserMute = "user:lock"

	RedisKeyPrefixReportLearnerReportOverview   = "report:learner:overview"
	RedisKeyPrefixReportLearningOutcomeOverview = "report:learning:outcome:overview"
	RedisKeyPrefixReportTeacherUsageOverview    = "report:teacher:usage:overview"
	RedisKeyPrefixReportSkillCoverage           = "report:skill:coverage"
)

const (
	RedisKeyPrefixAssessmentItem = "assessment:item"

	RedisKeyPrefixAssessmentQueryLearningSummaryTimeFilter = "assessment:query_learning_summary_time_filter"
)
