package entity

type LearnOutcomeAchievementRequest struct {
	ClassID                 string      `json:"class_id" binding:"required"`
	StudentID               string      `json:"student_id" binding:"required"`
	SelectedSubjectIDList   []string    `json:"selected_subject_id_list" binding:"gt=0"`
	UnSelectedSubjectIDList []string    `json:"un_selected_subject_id_list"`
	Durations               []TimeRange `json:"durations" binding:"gt=0"`
}

type LearnOutcomeAchievementResponse struct {
	FirstAchievedCount                    int64 `json:"first_achieved_count"`
	ReAchievedCount                       int64 `json:"re_achieved_count"`
	ClassAverageAchievedCount             int64 `json:"class_average_achieved_count"`
	UnSelectedSubjectsAverageAchieveCount int64 `json:"un_selected_subjects_average_achieve_count"`

	Items []*LearnOutcomeAchievementResponseItem `json:"items"`
}

type LearnOutcomeAchievementResponseItem struct {
	Duration TimeRange `json:"duration"`

	FirstAchievedPercentage                    float64 `json:"first_achieved_percentage"`
	ReAchievedPercentage                       float64 `json:"re_achieved_percentage"`
	ClassAverageAchievePercent                 float64 `json:"class_average_achieve_percent"`
	UnSelectedSubjectsAverageAchievePercentage float64 `json:"un_selected_subjects_average_achieve_percentage"`
}
