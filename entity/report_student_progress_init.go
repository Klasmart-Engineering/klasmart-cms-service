package entity

type ProgressInitRequest struct {
	ClassID                 string      `json:"class_id" binding:"required"`
	StudentID               string      `json:"student_id" binding:"required"`
	SelectedSubjectIDList   []string    `json:"selected_subject_id_list" binding:"gt=0"`
	UnSelectedSubjectIDList []string    `json:"un_selected_subject_id_list"`
	Durations               []TimeRange `json:"durations" binding:"gt=0"`
}

type ProgressInitResponse struct {
	LabelID     string `json:"labelID"`
	LabelParams string `json:"labelParams"`
}
