package entity

import "gitlab.badanamu.com.cn/calmisland/dbo"

type ListContentAssessmentsArgs struct {
	Type      ContentAssessmentType             `json:"type"`
	Query     string                            `json:"query"`
	QueryType ListContentAssessmentsQueryType   `json:"query_type"`
	Status    NullAssessmentStatus              `json:"status"`
	OrderBy   NullListContentAssessmentsOrderBy `json:"order_by"`
	Pager     dbo.Pager                         `json:"pager"`
}

type ContentAssessmentType string

const (
	ContentAssessmentTypeClassAndLive ContentAssessmentType = "class_and_live"
	ContentAssessmentTypeStudy        ContentAssessmentType = "study"
)

type ListContentAssessmentsQueryType string

const (
	ListContentAssessmentsQueryTypeTeacherName = "teacher_name"
	ListContentAssessmentsQueryTypeClassName   = "class_name"
)

type ListContentAssessmentsResult struct {
	Items []*ListContentAssessmentsResultItem `json:"items"`
	Total int                                 `json:"total"`
}

type ListContentAssessmentsResultItem struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	TeacherNames  []string `json:"teacher_names"`
	ClassName     string   `json:"class_name"`
	DueAt         int64    `json:"due_at"`
	CompleteRate  float64  `json:"complete_rate"`
	RemainingTime int64    `json:"remaining_time"`
	CompleteAt    int64    `json:"complete_at"`
	Comment       string   `json:"comment"`
	// debug
	ScheduleID string `json:"schedule_id"`
}

type GetContentAssessmentDetailResult struct {
	ID               string                              `json:"id"`
	Title            string                              `json:"title"`
	ClassName        string                              `json:"class_name"`
	TeacherNames     []string                            `json:"teacher_names"`
	Students         []*AssessmentStudent                `json:"students"`
	DueAt            int64                               `json:"due_at"`
	LessonPlan       AssessmentContent                   `json:"lesson_plan"`
	LessonMaterials  []*AssessmentContent                `json:"lesson_materials"`
	CompleteRate     float64                             `json:"complete_rate"`
	CompleteAt       int64                               `json:"complete_at"`
	RemainingTime    int64                               `json:"remaining_time"`
	StudentViewItems []*ContentAssessmentStudentViewItem `json:"student_view_items"`
	// debug
	ScheduleID string           `json:"schedule_id"`
	Status     AssessmentStatus `json:"status"`
}

type ContentAssessmentStudentViewItem struct {
	StudentID       string                                     `json:"student_id"`
	StudentName     string                                     `json:"student_name"`
	LessonMaterials []*ContentAssessmentStudentViewContentItem `json:"lesson_materials"`
	Checked         bool                                       `json:"checked"`
}

type ContentAssessmentStudentViewContentItem struct {
	LessonMaterialID   string  `json:"lesson_material_id"`
	LessonMaterialName string  `json:"lesson_material_name"`
	LessonMaterialType string  `json:"lesson_material_type"`
	Answer             string  `json:"answer"`
	MaxScore           float64 `json:"max_score"`
	AchievedScore      float64 `json:"achieved_score"`
	Checked            bool    `json:"checked"`
}

type ListContentAssessmentsOrderBy string

const (
	ListContentAssessmentsOrderByCompleteAt     ListContentAssessmentsOrderBy = "complete_at"
	ListContentAssessmentsOrderByCompleteAtDesc ListContentAssessmentsOrderBy = "-complete_at"
	ListContentAssessmentsOrderByCreateAt       ListContentAssessmentsOrderBy = "create_at"
	ListContentAssessmentsOrderByCreateAtDesc   ListContentAssessmentsOrderBy = "-create_at"
)

func (ob ListContentAssessmentsOrderBy) Valid() bool {
	switch ob {
	case ListContentAssessmentsOrderByCompleteAt,
		ListContentAssessmentsOrderByCompleteAtDesc,
		ListContentAssessmentsOrderByCreateAt,
		ListContentAssessmentsOrderByCreateAtDesc:
		return true
	}
	return false
}

type NullListContentAssessmentsOrderBy struct {
	Value ListContentAssessmentsOrderBy
	Valid bool
}

type UpdateContentAssessmentArgs struct {
	ID               string                                   `json:"id"`
	Action           UpdateAssessmentAction                   `json:"action" enums:"save,complete"`
	StudentIDs       []string                                 `json:"student_ids"`
	LessonMaterials  []UpdateAssessmentContentArgs            `json:"lesson_materials"`
	StudentViewItems []UpdateContentAssessmentStudentViewItem `json:"student_view_items"`
}

type UpdateContentAssessmentStudentViewItem struct {
	StudentID       string                                            `json:"student_id"`
	Comment         string                                            `json:"comment"`
	LessonMaterials []*UpdateContentAssessmentStudentViewMaterialItem `json:"lesson_materials"`
}

type UpdateContentAssessmentStudentViewMaterialItem struct {
	LessonMaterialID string  `json:"lesson_material_id"`
	AchievedScore    float64 `json:"achieved_score"`
}
