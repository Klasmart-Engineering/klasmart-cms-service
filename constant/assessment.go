package constant

import "time"

const (
	AssessmentNoClass = "NoClass"

	ListOptionAll                  = "all"
	AssessmentDefaultRemainingTime = 7 * 24 * time.Hour

	AssessmentQueryLearningSummaryTimeFilterCacheExpiration = 24 * time.Hour

	LearningSummaryFilterOptionNoneID   = "none"
	LearningSummaryFilterOptionNoneName = "None"
)
