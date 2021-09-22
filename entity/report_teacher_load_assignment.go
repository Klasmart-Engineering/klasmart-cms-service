package entity

type TeacherLoadAssignmentRequest struct {
	TeacherIDList []string `json:"teacher_id_list"`
	ClassIDList   []string `json:"class_id_list"`

	ClassTypeList []ClassType `json:"class_type_list"`
	Duration      TimeRange   `json:"duration"`
}

type TeacherLoadAssignmentResponse struct {
	TeacherID                  string  `json:"teacher_id"`
	TeacherName                string  `json:"teacher_name"`
	CountOfClasses             int64   `json:"count_of_classes"`
	CountOfStudents            int64   `json:"count_of_students"`
	CountOfScheduledAssignment int64   `json:"count_of_scheduled_assignment"`
	CountOfCompletedAssignment int64   `json:"count_of_completed_assignment"`
	FeedbackPercentage         float64 `json:"feedback_percentage"`
	CountOfPendingAssignment   int64   `json:"count_of_pending_assignment"`
	AvgDaysOfPendingAssignment int64   `json:"avg_days_of_pending_assignment"`
}
