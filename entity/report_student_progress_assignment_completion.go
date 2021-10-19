package entity

type AssignmentRequest struct {
	ClassID                 string      `json:"class_id" binding:"required"`
	StudentID               string      `json:"student_id" binding:"required"`
	SelectedSubjectIDList   []string    `json:"selected_subject_id_list" binding:"gt=0"`
	UnSelectedSubjectIDList []string    `json:"un_selected_subject_id_list"`
	Durations               []TimeRange `json:"durations" binding:"gt=0"`
}

type AssignmentCompletionRate struct {
	StudentDesignatedSubject    float64   `json:"student_designated_subject"`
	ClassDesignatedSubject      float64   `json:"class_designated_subject"`
	StudentNonDesignatedSubject float64   `json:"student_non_designated_subject"`
	Duration                    TimeRange `json:"duration"`
	StudentCompleteAssignment   int64     `json:"student_complete_assignment"`
	StudentTotalAssignment      int64     `json:"student_total_assignment"`
}

type AssignmentResponse []*AssignmentCompletionRate

type StudentAssignmentStatus struct {
	ScheduleID string `json:"schedule_id" gorm:"column:id"`
	ClassID    string `json:"class_id" gorm:"column:class_id"`
	StudentID  string `json:"student_id" gorm:"student_id"`
	SubjectID  string `json:"subject_id" gorm:"subject_id"`
	Attended   bool   `json:"attended" gorm:"finish_counts"`
	CreateAt   int64  `json:"create_at" gorm:"created_at"`
}
