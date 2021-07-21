package entity

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type HomeFunStudy struct {
	ID               string                   `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	ScheduleID       string                   `gorm:"column:schedule_id;type:varchar(64)" json:"schedule_id"`
	Title            string                   `gorm:"column:title;type:varchar(1024)" json:"title"`
	TeacherIDs       utils.SQLJSONStringArray `gorm:"column:teacher_ids;type:json" json:"teacher_ids"`
	StudentID        string                   `gorm:"column:student_id;type:varchar(64)" json:"student_id"`
	Status           AssessmentStatus         `gorm:"column:status;type:varchar(128)" json:"status"`
	DueAt            int64                    `gorm:"column:due_at;type:bigint" json:"due_at"`
	CompleteAt       int64                    `gorm:"column:complete_at;type:bigint" json:"complete_at"`
	LatestFeedbackID string                   `gorm:"column:latest_feedback_id;type:varchar(64)" json:"latest_feedback_id"`
	LatestFeedbackAt int64                    `gorm:"column:latest_feedback_at;type:bigint" json:"latest_feedback_at"`
	AssessFeedbackID string                   `gorm:"column:assess_feedback_id;type:varchar(64)" json:"assess_feedback_id"`
	AssessScore      HomeFunStudyAssessScore  `gorm:"column:assess_score;type:int" json:"assess_score"`
	AssessComment    string                   `gorm:"column:assess_comment;type:text" json:"assess_comment"`
	CreateAt         int64                    `gorm:"column:create_at;type:bigint;not null" json:"create_at"`
	UpdateAt         int64                    `gorm:"column:update_at;type:bigint;not null" json:"update_at"`
	DeleteAt         int64                    `gorm:"column:delete_at;type:bigint;not null" json:"delete_at"`
}

func (HomeFunStudy) TableName() string {
	return "home_fun_studies"
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
	ScheduleID       string
	ClassID          string
	LessonName       string
	StudentID        string
	TeacherIDs       []string
	DueAt            int64
	LatestFeedbackID string
	LatestFeedbackAt int64
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

type NullListHomeFunStudiesOrderBy struct {
	Value ListHomeFunStudiesOrderBy
	Valid bool
}

type ListHomeFunStudiesArgs struct {
	Query   string                        `json:"query"`
	Status  NullAssessmentStatus          `json:"status"`
	OrderBy NullListHomeFunStudiesOrderBy `json:"order_by"`
	Pager   dbo.Pager                     `json:"pager"`
}

type ListHomeFunStudiesResult struct {
	Items []*ListHomeFunStudiesResultItem `json:"items"`
	Total int                             `json:"total"`
}

type ListHomeFunStudiesResultItem struct {
	ID               string                  `json:"id"`
	Title            string                  `json:"title"`
	TeacherNames     []string                `json:"teacher_names"`
	StudentName      string                  `json:"student_name"`
	Status           AssessmentStatus        `json:"status"`
	DueAt            int64                   `json:"due_at"`
	LatestFeedbackAt int64                   `json:"latest_feedback_at"`
	AssessScore      HomeFunStudyAssessScore `json:"assess_score"`
	CompleteAt       int64                   `json:"complete_at"`
	LessonPlan       *AssessmentLessonPlan   `json:"lesson_plan"`
	// debug
	ScheduleID string `json:"schedule_id"`
}

type GetHomeFunStudyResult struct {
	ID               string                  `json:"id"`
	ScheduleID       string                  `json:"schedule_id"`
	Title            string                  `json:"title"`
	TeacherIDs       []string                `json:"teacher_ids"`
	TeacherNames     []string                `json:"teacher_names"`
	StudentID        string                  `json:"student_id"`
	StudentName      string                  `json:"student_name"`
	Status           AssessmentStatus        `json:"status"`
	DueAt            int64                   `json:"due_at"`
	CompleteAt       int64                   `json:"complete_at"`
	AssessFeedbackID string                  `json:"assess_feedback_id"`
	AssessScore      HomeFunStudyAssessScore `json:"assess_score" enums:"1,2,3,4,5"`
	AssessComment    string                  `json:"assess_comment"`
	Outcomes         []*HomeFunStudyOutcome  `json:"outcomes"`
}

type HomeFunStudyOutcome struct {
	OutcomeID   string                    `json:"outcome_id"`
	OutcomeName string                    `json:"outcome_name"`
	Assumed     bool                      `json:"assumed"`
	Status      HomeFunStudyOutcomeStatus `json:"status" enums:"achieved,not_achieved,not_attempted"`
}

type HomeFunStudyOutcomeStatus string

const (
	HomeFunStudyOutcomeStatusAchieved     HomeFunStudyOutcomeStatus = "achieved"
	HomeFunStudyOutcomeStatusNotAchieved  HomeFunStudyOutcomeStatus = "not_achieved"
	HomeFunStudyOutcomeStatusNotAttempted HomeFunStudyOutcomeStatus = "not_attempted"
)

type AssessHomeFunStudyArgs struct {
	ID               string                           `json:"id"`
	AssessFeedbackID string                           `json:"assess_feedback_id"`
	AssessScore      HomeFunStudyAssessScore          `json:"assess_score" enums:"1,2,3,4,5"`
	AssessComment    string                           `json:"assess_comment"`
	Action           UpdateHomeFunStudyAction         `json:"action" enums:"save,complete"`
	Outcomes         []*UpdateHomeFunStudyOutcomeArgs `json:"outcomes"`
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

type UpdateHomeFunStudyOutcomeArgs struct {
	OutcomeID string                    `json:"outcome_id"`
	Status    HomeFunStudyOutcomeStatus `json:"status"`
}
