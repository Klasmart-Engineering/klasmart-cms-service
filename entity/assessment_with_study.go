package entity

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
)

type ListStudyAssessmentsArgs struct {
	ClassTypes []ScheduleClassType           `json:"class_types"`
	Query      string                        `json:"query"`
	QueryType  ListStudyAssessmentsQueryType `json:"query_type"`
	Status     NullAssessmentStatus          `json:"status"`
	OrderBy    NullAssessmentsOrderBy        `json:"order_by"`
	Pager      dbo.Pager                     `json:"pager"`
}

type ListStudyAssessmentsQueryType string

const (
	ListStudyAssessmentsQueryTypeTeacherName ListStudyAssessmentsQueryType = "teacher_name"
	ListStudyAssessmentsQueryTypeClassName   ListStudyAssessmentsQueryType = "class_name"
)

type NullListStudyAssessmentsQueryType struct {
	Value ListStudyAssessmentsQueryType
	Valid bool
}

type ListStudyAssessmentsResult struct {
	Items []*ListStudyAssessmentsResultItem `json:"items"`
	Total int                               `json:"total"`
}

type ListStudyAssessmentsResultItem struct {
	ID            string                `json:"id"`
	Title         string                `json:"title"`
	TeacherNames  []string              `json:"teacher_names"`
	ClassName     string                `json:"class_name"`
	DueAt         int64                 `json:"due_at"`
	CompleteRate  float64               `json:"complete_rate"`
	RemainingTime int64                 `json:"remaining_time"`
	CompleteAt    int64                 `json:"complete_at"`
	LessonPlan    *AssessmentLessonPlan `json:"lesson_plan"`
	// debug
	ScheduleID string `json:"schedule_id"`
	CreateAt   int64  `json:"create_at"`
}

type AddH5PAssessmentStudyInput struct {
	ScheduleID string `json:"schedule_id"`
}

type AssessmentH5PRoom struct {
	AnyoneAttempted bool
	Users           []*AssessmentH5PUser
	UserMap         map[string]*AssessmentH5PUser
}

type AssessmentH5PUser struct {
	UserID     string
	Comment    string
	Contents   []*AssessmentH5PContentScore
	ContentMap map[string]*AssessmentH5PContentScore
}

type AssessmentH5PContentScore struct {
	H5PID            string
	ContentID        string
	ContentName      string
	ContentType      string
	Answer           string
	Answers          []string
	MaxPossibleScore float64
	AchievedScore    float64
	Scores           []float64
}
