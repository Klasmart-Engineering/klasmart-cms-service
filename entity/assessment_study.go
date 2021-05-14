package entity

import "gitlab.badanamu.com.cn/calmisland/dbo"

type ListStudiesArgs struct {
	Query   string                        `json:"query"`
	Status  NullAssessmentStatus          `json:"status"`
	OrderBy NullListHomeFunStudiesOrderBy `json:"order_by"`
	Pager   dbo.Pager                     `json:"pager"`
}

type ListStudiesResult struct {
	Items []*ListStudiesResultItem `json:"items"`
	Total int                      `json:"total"`
}

type ListStudiesResultItem struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	TeacherNames  []string `json:"teacher_names"`
	ClassName     string   `json:"class_name"`
	DueAt         int64    `json:"due_at"`
	CompleteRate  float64  `json:"complete_rate"`
	RemainingTime float64  `json:"remaining_time"`
	CompleteAt    int64    `json:"complete_at"`
	Comment       string   `json:"comment"`
	// debug
	ScheduleID string `json:"schedule_id"`
}

type GetStudyDetailResult struct {
	ID               string                            `json:"id"`
	Title            string                            `json:"title"`
	ClassName        string                            `json:"class_name"`
	TeacherNames     []string                          `json:"teacher_names"`
	Students         []*AssessmentStudent              `json:"students"`
	DueAt            int64                             `json:"due_at"`
	LessonPlan       AssessmentContent                 `json:"lesson_plan"`
	LessonMaterials  []*AssessmentContent              `json:"lesson_materials"`
	CompleteRate     float64                           `json:"complete_rate"`
	CompleteAt       int64                             `json:"complete_at"`
	RemainingTime    int64                             `json:"remaining_time"`
	StudentViewItems []*AssessmentStudyStudentViewItem `json:"student_view_items"`
	// debug
	ScheduleID string           `json:"schedule_id"`
	Status     AssessmentStatus `json:"status"`
}

type AssessmentStudyStudentViewItem struct {
	StudentID   string                                   `json:"student_id"`
	StudentName string                                   `json:"student_name"`
	Contents    []*AssessmentStudyStudentViewContentItem `json:"contents"`
}

type AssessmentStudyStudentViewContentItem struct {
	ContentID             string  `json:"content_id"`
	ContentName           string  `json:"content_name"`
	ContentType           string  `json:"content_type"`
	Answer                string  `json:"answer"`
	AnswerType            string  `json:"answer_type"`
	MaxScore              float64 `json:"max_score"`
	AchievedScore         float64 `json:"achieved_score"`
	AchievedScoreEditable bool    `json:"achieved_score_editable"`
	AchievedPercentage    float64 `json:"achieved_percentage"`
}

type ListStudiesOrderBy string

const (
	ListStudiesOrderByCompleteAt     ListStudiesOrderBy = "complete_at"
	ListStudiesOrderByCompleteAtDesc ListStudiesOrderBy = "-complete_at"
)

func (ob ListStudiesOrderBy) Valid() bool {
	switch ob {
		ListStudiesOrderByCompleteAt,
		ListStudiesOrderByCompleteAtDesc:
		return true
	}
	return false
}

type NullListStudiesOrderBy struct {
	Value ListStudiesOrderBy
	Valid bool
}
