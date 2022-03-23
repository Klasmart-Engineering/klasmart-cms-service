package da

import "gitlab.badanamu.com.cn/calmisland/ro"

const (
	RedisKeyPrefixContentId     = "content:id"
	RedisKeyPrefixContentLock   = "content:lock"
	RedisKeyPrefixContentReview = "content:review"
	RedisKeyPrefixContentAuth   = "content:auth"

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
)

const (
	RedisKeyPrefixAssessmentItem = "assessment:item"

	RedisKeyPrefixAssessmentQueryLearningSummaryTimeFilter = "assessment:query_learning_summary_time_filter"
)

var (
	RedisKeyLazyRefreshCache            = ro.NewStringParameterKey("lazy:refresh:cache:%s")
	RedisKeyLazyRefreshCacheLocker      = ro.NewStringParameterKey("lazy:refresh:locker:%s")
	RedisKeyLazyRefreshCacheDataVersion = ro.NewStringParameterKey("lazy:refresh:data:version:%s")
)
