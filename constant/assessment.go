package constant

import "time"

const (
	AssessmentNoClass = "NoClass"

	ListOptionAll                  = "all"
	AssessmentDefaultRemainingTime = 7 * 24 * time.Hour

	AssessmentQueryLearningSummaryTimeFilterCacheExpiration = 24 * time.Hour

	LearningSummaryFilterOptionNoneID   = "none"
	LearningSummaryFilterOptionNoneName = "None"

	AssessmentBatchPageSize = 500

	AssessmentInitializedKey = "Initialized"
	AssessmentHistoryFlag    = 1
	AssessmentCurrentFlag    = 2
)

const (
	TableNameAssessmentV2                 = "assessments_v2"
	TableNameAssessmentsUsersV2           = "assessments_users_v2"
	TableNameAssessmentsContentsV2        = "assessments_contents_v2"
	TableNameAssessmentReviewerFeedbackV2 = "assessments_reviewer_feedback_v2"
	TableNameAssessmentsUsersOutcomesV2   = "assessments_users_outcomes_v2"
)
