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
	CreateAt   int64  `json:"create_at"`
}

type GetStudyAssessmentDetailResult struct {
	ID               string                           `json:"id"`
	Title            string                           `json:"title"`
	ClassName        string                           `json:"class_name"`
	Teachers         []*AssessmentTeacher             `json:"teachers"`
	Students         []*AssessmentStudent             `json:"students"`
	DueAt            int64                            `json:"due_at"`
	LessonPlan       StudyAssessmentLessonPlan        `json:"lesson_plan"`
	LessonMaterials  []*StudyAssessmentLessonMaterial `json:"lesson_materials"`
	CompleteAt       int64                            `json:"complete_at"`
	RemainingTime    int64                            `json:"remaining_time"`
	StudentViewItems []*AssessmentStudentViewH5PItem  `json:"student_view_items"`
	// debug
	ScheduleID string           `json:"schedule_id"`
	Status     AssessmentStatus `json:"status"`
}

type StudyAssessmentLessonPlan struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

type StudyAssessmentLessonMaterial struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Checked bool   `json:"checked"`
}

type UpdateStudyAssessmentArgs struct {
	ID               string                                  `json:"id"`
	Action           UpdateAssessmentAction                  `json:"action" enums:"save,complete"`
	StudentIDs       []string                                `json:"student_ids"`
	LessonMaterials  []*UpdateAssessmentContentArgs          `json:"lesson_materials"`
	StudentViewItems []*UpdateStudyAssessmentStudentViewItem `json:"student_view_items"`
}

type UpdateStudyAssessmentStudentViewItem struct {
	StudentID       string                                          `json:"student_id"`
	Comment         string                                          `json:"comment"`
	LessonMaterials []*UpdateStudyAssessmentStudentViewMaterialItem `json:"lesson_materials"`
}

type UpdateStudyAssessmentStudentViewMaterialItem struct {
	LessonMaterialID string  `json:"lesson_material_id"`
	AchievedScore    float64 `json:"achieved_score"`
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
	ContentID        string
	ContentName      string
	ContentType      string
	Answer           string
	Answers          []string
	MaxPossibleScore float64
	AchievedScore    float64
	Scores           []float64
}

type AddStudyInput struct {
	ScheduleID    string
	ScheduleTitle string
	ClassID       string
	LessonPlanID  string
	Attendances   []*ScheduleRelation
}
