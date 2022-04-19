package entity

type LearnOutcomeAchievementRequest struct {
	ClassID                 string      `json:"class_id" binding:"required"`
	StudentID               string      `json:"student_id" binding:"required"`
	SelectedSubjectIDList   []string    `json:"selected_subject_id_list" binding:"gt=0"`
	UnSelectedSubjectIDList []string    `json:"un_selected_subject_id_list"`
	Durations               []TimeRange `json:"durations" binding:"gt=0"`
}

type LearnOutcomeAchievementResponse struct {
	Request                               *LearnOutcomeAchievementRequest `json:"request"`
	FirstAchievedCount                    int64                           `json:"first_achieved_count"`
	ReAchievedCount                       int64                           `json:"re_achieved_count"`
	ClassAverageAchievedCount             float64                         `json:"class_average_achieved_count"`
	UnSelectedSubjectsAverageAchieveCount float64                         `json:"un_selected_subjects_average_achieve_count"`

	Items                                 LearnOutcomeAchievementResponseItemSlice `json:"items"`
	StudentAchievedCounts                 map[string]Float64Slice                  `json:"-"`
	UnselectSubjectsStudentAchievedCounts Float64Slice                             `json:"-"`
	LabelID                               string                                   `json:"label_id"`
	LabelParams                           LearningOutcomeAchivementLabelParams     `json:"label_params"`
}

type LearningOutcomeAchivementLabelParams struct {
	AchievedLoCount      float64 `json:"achieved_lo_count"`
	LearntLoCount        float64 `json:"learnt_lo_count"`
	LOCompareClass3week  float64 `json:"lo_compare_class_3_week"`
	LOCompareLastWeek    float64 `json:"lo_compare_last_week"`
	LOReviewCompareClass float64 `json:"lo_review_compare_class"`
	LOCompareLast3Week   float64 `json:"lo_compare_last_3_week"`
	LOCompareClass       float64 `json:"lo_compare_class"`
}

type LearnOutcomeAchievementResponseItemSlice []*LearnOutcomeAchievementResponseItem

func (s *LearnOutcomeAchievementResponse) GetItem(tr TimeRange) (item *LearnOutcomeAchievementResponseItem) {
	for _, responseItem := range s.Items {
		if responseItem.Duration == tr {
			item = responseItem
			return
		}
	}
	return
}

type LearnOutcomeAchievementResponseItem struct {
	Duration TimeRange `json:"duration"`

	FirstAchievedPercentage                     float64 `json:"first_achieved_percentage"`
	ReAchievedPercentage                        float64 `json:"re_achieved_percentage"`
	ClassAverageAchievedPercentage              float64 `json:"class_average_achieved_percentage"`
	UnSelectedSubjectsAverageAchievedPercentage float64 `json:"un_selected_subjects_average_achieved_percentage"`

	FirstAchievedCount int64 `json:"first_achieved_count"`
	ReAchievedCount    int64 `json:"re_achieved_count"`
	UnAchievedCount    int64 `json:"un_achieved_count"`

	FirstAchievedPercentages              Float64Slice            `json:"-"`
	ReAchievedPercentages                 Float64Slice            `json:"-"`
	StudentAchievedPercentages            map[string]Float64Slice `json:"-"`
	UnSelectedSubjectsAchievedPercentages Float64Slice            `json:"-"`
}

type StudentProgressLearnOutcomeCount struct {
	StudentID          string    `json:"student_id" gorm:"column:student_id" `
	SubjectID          string    `json:"subject_id" gorm:"column:subject_id" `
	CompletedCount     int64     `json:"completed_count" gorm:"column:completed_count" `
	AchievedCount      int64     `json:"achieved_count" gorm:"column:achieved_count" `
	FirstAchievedCount int64     `json:"first_achieved_count" gorm:"column:first_achieved_count" `
	Duration           TimeRange `json:"duration" gorm:"column:duration" `
}
