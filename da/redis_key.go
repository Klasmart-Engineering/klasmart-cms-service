package da

const (
	RedisKeyPrefixContentCondition = "content:condition"
	RedisKeyPrefixContentId        = "content:id"
	RedisKeyPrefixContentLock      = "content:lock"
	RedisKeyPrefixContentReview    = "content:review"
	RedisKeyPrefixContentAuth      = "content:auth"

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

	RedisKeyPrefixMilestoneMute = "milestone:lock"
)

const (
	RedisKeyPrefixAssessmentItem = "assessment:item"
	RedisKeyPrefixAssessmentLock = "assessment:add"
)
