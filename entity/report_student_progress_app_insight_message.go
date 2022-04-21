package entity

type AppInsightMessageRequest struct {
	ClassID   string `json:"class_id" binding:"required"`
	StudentID string `json:"student_id" binding:"required"`
	OrgID     string `json:"org_id" binding:"required"`
	EndTime   int    `json:"end_time"`
}

type AppInsightMessageResponse struct {
	LearningOutcomeAchivementLabelID     string                               `json:"learning_outcome_achivement_label_id"`
	LearningOutcomeAchivementLabelParams LearningOutcomeAchivementLabelParams `json:"learning_outcome_achivement_label_params"`
	AttedanceLabelID                     string                               `json:"attedance_label_id"`
	AttedanceLabelParams                 AttedanceLabelParams                 `json:"attedance_label_params"`
	AssignmentLabelID                    string                               `json:"assignment_label_id"`
	AssignmentLabelParams                AssignmentLabelParams                `json:"assignment_label_params"`
}
