package entity

type AssignmentRequest struct {
	ClassID                 string      `json:"class_id" binding:"required"`
	StudentID               string      `json:"student_id" binding:"required"`
	SelectedSubjectIDList   []string    `json:"selected_subject_id_list" binding:"gt=0"`
	UnSelectedSubjectIDList []string    `json:"un_selected_subject_id_list"`
	Durations               []TimeRange `json:"durations" binding:"gt=0"`
}

type AssignmentCompletionRate struct {
	StudentID                   string    `json:"student_id"`
	StudentCompleteAssignment   int64     `json:"student_complete_assignment"`
	StudentTotalAssignment      int64     `json:"student_total_assignment"`
	StudentDesignatedSubject    float64   `json:"student_designated_subject"`
	ClassDesignatedSubject      float64   `json:"class_designated_subject"`
	StudentNonDesignatedSubject float64   `json:"student_non_designated_subject"`
	Duration                    TimeRange `json:"duration"`
}

type AssignmentResponse []*AssignmentCompletionRate

type StudentAssignmentStatus struct {
	ClassID   string `json:"class_id" gorm:"column:class_id"`
	StudentID string `json:"student_id" gorm:"column:student_id"`
	SubjectID string `json:"subject_id" gorm:"column:subject_id"`
	Total     int64  `json:"total" gorm:"column:total"`
	Finish    int64  `json:"finish" gorm:"column:finish"`
	Duration  string `json:"duration" gorm:"column:duration"`
}
