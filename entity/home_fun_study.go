package entity

type HomeFunStudy struct {
	ID               string             `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	ScheduleID       string             `gorm:"column:schedule_id;type:varchar(64)" json:"schedule_id"`
	Title            string             `gorm:"column:title;type:varchar(1024)" json:"title"`
	TeacherIDs       string             `gorm:"column:teacher_ids;type:json" json:"teacher_ids"`
	StudentID        string             `gorm:"column:student_id;type:varchar(64)" json:"student_id"`
	Status           HomeFunStudyStatus `gorm:"column:status;type:varchar(128)" json:"status"`
	DueAt            int64              `gorm:"column:due_at;type:bigint" json:"due_at"`
	CompleteAt       int64              `gorm:"column:complete_at;type:bigint" json:"complete_at"`
	LatestFeedbackID string             `gorm:"column:latest_feedback_id;type:varchar(64)" json:"latest_feedback_id"`
	LatestFeedbackAt int64              `gorm:"column:latest_feedback_at;type:bigint" json:"latest_feedback_at"`
	AssessFeedbackID string             `gorm:"column:assess_feedback_id;type:varchar(64)" json:"assess_feedback_id"`
	AssessScore      int                `gorm:"column:assess_score;type:int" json:"assess_score"`
	AssessComment    string             `gorm:"column:assess_comment;type:text" json:"assess_comment"`
	CreateAt         int64              `gorm:"column:create_at;type:bigint;not null" json:"create_at"`
	UpdateAt         int64              `gorm:"column:update_at;type:bigint;not null" json:"update_at"`
	DeleteAt         int64              `gorm:"column:delete_at;type:bigint;not null" json:"delete_at"`
}

func (HomeFunStudy) TableName() string {
	return "home_fun_studies"
}

type HomeFunStudyStatus string

const (
	HomeFunStudyStatusInProgress HomeFunStudyStatus = "in_progress"
	HomeFunStudyStatusComplete   HomeFunStudyStatus = "complete"
)

func (s HomeFunStudyStatus) Valid() bool {
	switch s {
	case HomeFunStudyStatusInProgress, HomeFunStudyStatusComplete:
		return true
	default:
		return false
	}
}

type HomeFunStudyAssessScore int

const (
	HomeFunStudyAssessScorePoor HomeFunStudyAssessScore = iota + 1
	HomeFunStudyAssessScoreFair
	HomeFunStudyAssessScoreAverage
	HomeFunStudyAssessScoreGood
	HomeFunStudyAssessScoreExcellent
)

func (s HomeFunStudyAssessScore) Valid() bool {
	switch s {
	case HomeFunStudyAssessScorePoor,
		HomeFunStudyAssessScoreFair,
		HomeFunStudyAssessScoreAverage,
		HomeFunStudyAssessScoreGood,
		HomeFunStudyAssessScoreExcellent:
		return true
	}
	return false
}

type SaveHomeFunStudyArgs struct {
	ScheduleID     string
	ClassID        string
	LessonName     string
	TeacherIDs     []string
	StudentID      string
	DueAt          int64
	LatestSubmitID string
	LatestSubmitAt int64
}

type ListHomeFunStudiesOrderBy string

const (
	ListHomeFunStudiesOrderByLatestFeedbackAt     ListHomeFunStudiesOrderBy = "latest_feedback_at"
	ListHomeFunStudiesOrderByLatestFeedbackAtDesc ListHomeFunStudiesOrderBy = "-latest_feedback_at"
	ListHomeFunStudiesOrderByCompleteAt           ListHomeFunStudiesOrderBy = "complete_at"
	ListHomeFunStudiesOrderByCompleteAtDesc       ListHomeFunStudiesOrderBy = "-complete_at"
)

func (ob ListHomeFunStudiesOrderBy) Valid() bool {
	switch ob {
	case ListHomeFunStudiesOrderByLatestFeedbackAt,
		ListHomeFunStudiesOrderByLatestFeedbackAtDesc,
		ListHomeFunStudiesOrderByCompleteAt,
		ListHomeFunStudiesOrderByCompleteAtDesc:
		return true
	}
	return false
}

type ListHomeFunStudiesArgs struct {
	Query    string                     `json:"query"`
	Status   *HomeFunStudyStatus        `json:"status"`
	OrderBy  *ListHomeFunStudiesOrderBy `json:"order_by"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
}

type ListHomeFunStudiesResult struct {
	Items []*ListHomeFunStudiesResultItem `json:"items"`
	Total int                             `json:"total"`
}

type ListHomeFunStudiesResultItem struct {
	ID               string             `json:"id"`
	TeacherNames     []string           `json:"teacher_names"`
	StudentName      string             `json:"student_name"`
	Status           HomeFunStudyStatus `json:"status"`
	DueAt            int64              `json:"due_at"`
	LatestFeedbackAt int64              `json:"latest_feedback_at"`
	AssessScore      int                `json:"assess_score"`
	CompleteAt       int64              `json:"complete_at"`
	// debug
	ScheduleID string `json:"schedule_id"`
}

type GetHomeFunStudyResult struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	TeacherNames     []string `json:"teacher_names"`
	StudentName      string   `json:"student_name"`
	DueAt            int64    `json:"due_at"`
	CompleteAt       int64    `json:"complete_at"`
	AssessFeedbackID string   `json:"assess_feedback_id"`
	AssessScore      int      `json:"assess_score"`
	AssessComment    string   `json:"assess_comment"`
}

type AssessHomeFunStudyArgs struct {
	Action        UpdateHomeFunStudyAction `json:"action" enums:"save,complete"`
	ID            string                   `json:"id"`
	AssessScore   HomeFunStudyAssessScore  `json:"score" enums:"1,2,3,4,5"`
	AssessComment string                   `json:"comment"`
}

type UpdateHomeFunStudyAction string

const (
	UpdateHomeFunStudyActionSave     UpdateHomeFunStudyAction = "save"
	UpdateHomeFunStudyActionComplete UpdateHomeFunStudyAction = "complete"
)

func (a UpdateHomeFunStudyAction) Valid() bool {
	switch a {
	case UpdateHomeFunStudyActionSave, UpdateHomeFunStudyActionComplete:
		return true
	}
	return false
}
