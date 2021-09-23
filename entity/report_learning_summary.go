package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

type QueryLearningSummaryTimeFilterArgs struct {
	TimeOffset  int                 `json:"time_offset"`
	SummaryType LearningSummaryType `json:"summary_type" enums:"live_class,assignment"`
	OrgID       string              `json:"org_id"`
	SchoolIDs   []string            `json:"school_ids"`
	TeacherID   string              `json:"teacher_id"`
	StudentID   string              `json:"student_id"`
}

type LearningSummaryFilterYear struct {
	Year  int                         `json:"year"`
	Weeks []LearningSummaryFilterWeek `json:"weeks"`
}

type LearningSummaryFilterWeek struct {
	WeekStart int64 `json:"week_start"`
	WeekEnd   int64 `json:"week_end"`
}

type QueryLearningSummaryRemainingFilterArgs struct {
	SummaryType LearningSummaryType                `json:"summary_type" enums:"live_class,assignment"`
	FilterType  LearningSummaryRemainingFilterType `json:"filter_type" enums:"school,class,teacher,student,subject"`
	LearningSummaryFilter
}

type LearningSummaryFilter struct {
	Year      int      `json:"year"`
	WeekStart int64    `json:"week_start"`
	WeekEnd   int64    `json:"week_end"`
	SchoolIDs []string `json:"school_ids"`
	ClassID   string   `json:"class_id"`
	TeacherID string   `json:"teacher_id"`
	StudentID string   `json:"student_id"`
	SubjectID string   `json:"subject_id"`
}

type LearningSummaryRemainingFilterType string

const (
	LearningSummaryFilterTypeSchool  LearningSummaryRemainingFilterType = "school"
	LearningSummaryFilterTypeClass   LearningSummaryRemainingFilterType = "class"
	LearningSummaryFilterTypeTeacher LearningSummaryRemainingFilterType = "teacher"
	LearningSummaryFilterTypeStudent LearningSummaryRemainingFilterType = "student"
	LearningSummaryFilterTypeSubject LearningSummaryRemainingFilterType = "subject"
)

func (l LearningSummaryRemainingFilterType) Valid() bool {
	switch l {
	case LearningSummaryFilterTypeSchool,
		LearningSummaryFilterTypeClass,
		LearningSummaryFilterTypeTeacher,
		LearningSummaryFilterTypeStudent,
		LearningSummaryFilterTypeSubject:
		return true
	}
	return false
}

type QueryLearningSummaryRemainingFilterResultItem struct {
	Year        int    `json:"year,omitempty"`
	WeekStart   int64  `json:"week_start,omitempty"`
	WeekEnd     int64  `json:"week_end,omitempty"`
	SchoolID    string `json:"school_id,omitempty"`
	SchoolName  string `json:"school_name,omitempty"`
	ClassID     string `json:"class_id,omitempty"`
	ClassName   string `json:"class_name,omitempty"`
	TeacherID   string `json:"teacher_id,omitempty"`
	TeacherName string `json:"teacher_name,omitempty"`
	StudentID   string `json:"student_id,omitempty"`
	StudentName string `json:"student_name,omitempty"`
	SubjectID   string `json:"subject_id,omitempty"`
	SubjectName string `json:"subject_name,omitempty"`
}

type QueryLiveClassesSummaryResult struct {
	Attend float64                 `json:"attend"`
	Items  []*LiveClassSummaryItem `json:"items"`
}

type LiveClassSummaryItem struct {
	Status          AssessmentStatus          `json:"status" enums:"in_progress,complete"`
	Absent          bool                      `json:"absent"`
	ClassStartTime  int64                     `json:"class_start_time"`
	ScheduleTitle   string                    `json:"schedule_title"`
	LessonPlanName  string                    `json:"lesson_plan_name"`
	Outcomes        []*LearningSummaryOutcome `json:"outcomes"`
	TeacherFeedback string                    `json:"teacher_feedback"`
	// for debug
	ScheduleID   string `json:"schedule_id"`
	AssessmentID string `json:"assessment_id"`
	// for sorting
	CompleteAt int64 `json:"complete_at"`
	CreateAt   int64 `json:"create_at"`
}

type LearningSummaryOutcome struct {
	ID     string                  `json:"id"`
	Name   string                  `json:"name"`
	Status AssessmentOutcomeStatus `json:"status" enums:"achieved,not_achieved,partially"`
}

type QueryAssignmentsSummaryResult struct {
	StudyCount        int                       `json:"study_count"`
	HomeFunStudyCount int                       `json:"home_fun_study_count"`
	Items             []*AssignmentsSummaryItem `json:"items"`
}

type AssignmentsSummaryItem struct {
	Type            AssessmentType            `json:"assessment_type" enums:"class,live,study,home_fun_study"`
	Status          AssessmentStatus          `json:"status" enums:"in_progress,complete"`
	AssessmentTitle string                    `json:"assessment_title"`
	LessonPlanName  string                    `json:"lesson_plan_name"`
	TeacherFeedback string                    `json:"teacher_feedback"`
	Outcomes        []*LearningSummaryOutcome `json:"outcomes"`
	// for debug
	ScheduleID   string `json:"schedule_id"`
	AssessmentID string `json:"assessment_id"`
	// for sorting
	CompleteAt int64 `json:"complete_at"`
	CreateAt   int64 `json:"create_at"`
}

type LearningSummaryType string

const (
	LearningSummaryTypeInvalid    = "invalid"
	LearningSummaryTypeLiveClass  = "live_class"
	LearningSummaryTypeAssignment = "assignment"
)

func (t LearningSummaryType) Valid() bool {
	switch t {
	case LearningSummaryTypeLiveClass, LearningSummaryTypeAssignment:
		return true
	default:
		return false
	}
}

type AssessmentOutcomeStatus string

const (
	AssessmentOutcomeStatusAchieved    AssessmentOutcomeStatus = "achieved"
	AssessmentOutcomeStatusNotAchieved AssessmentOutcomeStatus = "not_achieved"
	AssessmentOutcomeStatusPartially   AssessmentOutcomeStatus = "partially"
)

type StudentUsageReport struct {
	utils.Pager
}
