package entity

type LearnerUsageRequest struct {
	Durations       []TimeRange `json:"durations" form:"durations" binding:"gt=0"`
	ContentTypeList []string    `json:"content_type_list" form:"content_type_list" binding:"gt=0"`
}

type LearnerUsageResponse struct {
}
