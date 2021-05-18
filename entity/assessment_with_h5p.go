package entity

import "gitlab.badanamu.com.cn/calmisland/dbo"

type ListH5PAssessmentsArgs struct {
	Type      AssessmentType              `json:"type"`
	Query     string                      `json:"query"`
	QueryType ListH5PAssessmentsQueryType `json:"query_type"`
	Status    NullAssessmentStatus        `json:"status"`
	OrderBy   NullAssessmentsOrderBy      `json:"order_by"`
	Pager     dbo.Pager                   `json:"pager"`
}

type ListH5PAssessmentsQueryType string

const (
	ListH5PAssessmentsQueryTypeTeacherName = "teacher_name"
	ListH5PAssessmentsQueryTypeClassName   = "class_name"
)

type NullListH5PAssessmentsQueryType struct {
	Value ListH5PAssessmentsQueryType
	Valid bool
}

type ListH5PAssessmentsResult struct {
	Items []*ListH5PAssessmentsResultItem `json:"items"`
	Total int                             `json:"total"`
}

type ListH5PAssessmentsResultItem struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	TeacherNames  []string `json:"teacher_names"`
	ClassName     string   `json:"class_name"`
	DueAt         int64    `json:"due_at"`
	CompleteRate  float64  `json:"complete_rate"`
	RemainingTime int64    `json:"remaining_time"`
	CompleteAt    int64    `json:"complete_at"`
	// debug
	ScheduleID string `json:"schedule_id"`
}

type GetH5PAssessmentDetailResult struct {
	ID               string                          `json:"id"`
	Title            string                          `json:"title"`
	ClassName        string                          `json:"class_name"`
	Teachers         []*AssessmentTeacher            `json:"teachers"`
	Students         []*AssessmentStudent            `json:"students"`
	DueAt            int64                           `json:"due_at"`
	LessonPlan       H5PAssessmentLessonPlan         `json:"lesson_plan"`
	LessonMaterials  []*H5PAssessmentLessonMaterial  `json:"lesson_materials"`
	CompleteRate     float64                         `json:"complete_rate"`
	CompleteAt       int64                           `json:"complete_at"`
	RemainingTime    int64                           `json:"remaining_time"`
	StudentViewItems []*H5PAssessmentStudentViewItem `json:"student_view_items"`
	// debug
	ScheduleID string           `json:"schedule_id"`
	Status     AssessmentStatus `json:"status"`
}

type H5PAssessmentLessonPlan struct {
	ID      string
	Name    string
	Comment string
}

type H5PAssessmentLessonMaterial struct {
	ID      string
	Name    string
	Comment string
	Checked bool
}

type H5PAssessmentStudentViewItem struct {
	StudentID       string                                    `json:"student_id"`
	StudentName     string                                    `json:"student_name"`
	Comment         string                                    `json:"comment"`
	LessonMaterials []*H5PAssessmentStudentViewLessonMaterial `json:"lesson_materials"`
}

type H5PAssessmentStudentViewLessonMaterial struct {
	LessonMaterialID   string  `json:"lesson_material_id"`
	LessonMaterialName string  `json:"lesson_material_name"`
	LessonMaterialType string  `json:"lesson_material_type"`
	Answer             string  `json:"answer"`
	MaxScore           float64 `json:"max_score"`
	AchievedScore      float64 `json:"achieved_score"`
	Attempted          bool    `json:"attempted"`
}

type UpdateH5PAssessmentArgs struct {
	ID               string                               `json:"id"`
	Action           UpdateAssessmentAction               `json:"action" enums:"save,complete"`
	StudentIDs       []string                             `json:"student_ids"`
	LessonMaterials  []UpdateAssessmentContentArgs        `json:"lesson_materials"`
	StudentViewItems []UpdateH5PAssessmentStudentViewItem `json:"student_view_items"`
}

type UpdateH5PAssessmentStudentViewItem struct {
	StudentID       string                                        `json:"student_id"`
	Comment         string                                        `json:"comment"`
	LessonMaterials []*UpdateH5PAssessmentStudentViewMaterialItem `json:"lesson_materials"`
}

type UpdateH5PAssessmentStudentViewMaterialItem struct {
	LessonMaterialID string  `json:"lesson_material_id"`
	AchievedScore    float64 `json:"achieved_score"`
}

type AddH5PAssessmentStudyInput struct {
	ScheduleID string `json:"schedule_id"`
}

type AssessmentH5PRoom struct {
	CompleteRate    float64
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
	ContentID        string
	ContentName      string
	ContentType      string
	Answer           string
	Answers          []string
	MaxPossibleScore float64
	AchievedScore    float64
	Scores           []float64
}
