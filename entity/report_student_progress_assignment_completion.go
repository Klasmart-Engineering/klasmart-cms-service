package entity

type AssignmentRequest struct {
	ClassID                 string      `json:"class_id" binding:"required"`
	StudentID               string      `json:"student_id" binding:"required"`
	SelectedSubjectIDList   []string    `json:"selected_subject_id_list" binding:"gt=0"`
	UnSelectedSubjectIDList []string    `json:"un_selected_subject_id_list"`
	Durations               []TimeRange `json:"durations" binding:"gt=0"`
}

type AssignmentCompletionRate struct {
	StudentDesignatedSubject    float64 `json:"student_designated_subject"`
	ClassDesignatedSubject      float64 `json:"class_designated_subject"`
	StudentNonDesignatedSubject float64 `json:"student_non_designated_subject"`
}

type AssignmentResponse []AssignmentCompletionRate
