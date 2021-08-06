package entity

type QueryLearningSummaryTimeFilterArgs struct {
	SummaryType LearningSummaryType `json:"summary_type" enums:"live_class,assignment"`
	OrgID       string              `json:"org_id"`
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
	SummaryType LearningSummaryType `json:"summary_type" enums:"live_class,assignment"`
	OrgID       string              `json:"org_id"`
	WeekStart   int64               `json:"week_start"`
	WeekEnd     int64               `json:"week_end"`
}

type LearningSummaryFilterSchool struct {
	ID      string                        `json:"id"`
	Name    string                        `json:"name"`
	Classes []*LearningSummaryFilterClass `json:"classes"`
}

type LearningSummaryFilterClass struct {
	ID       string                          `json:"id"`
	Name     string                          `json:"name"`
	Teachers []*LearningSummaryFilterTeacher `json:"teachers"`
}

type LearningSummaryFilterTeacher struct {
	ID       string                          `json:"id"`
	Name     string                          `json:"name"`
	Students []*LearningSummaryFilterStudent `json:"students"`
}

type LearningSummaryFilterStudent struct {
	ID       string                          `json:"id"`
	Name     string                          `json:"name"`
	Subjects []*LearningSummaryFilterSubject `json:"subjects"`
}

type LearningSummaryFilterSubject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LearningSummaryFilter struct {
	Year      int    `json:"year"`
	WeekStart int64  `json:"week_start"`
	WeekEnd   int64  `json:"week_end"`
	SchoolID  string `json:"school_id"`
	ClassID   string `json:"class_id"`
	TeacherID string `json:"teacher_id"`
	StudentID string `json:"student_id"`
	SubjectID string `json:"subject_id"`
}

type LearningSummaryFilterType string

const (
	LearningSummaryFilterTypeYear    LearningSummaryFilterType = "year"
	LearningSummaryFilterTypeWeek    LearningSummaryFilterType = "week"
	LearningSummaryFilterTypeSchool  LearningSummaryFilterType = "school"
	LearningSummaryFilterTypeClass   LearningSummaryFilterType = "class"
	LearningSummaryFilterTypeTeacher LearningSummaryFilterType = "teacher"
	LearningSummaryFilterTypeStudent LearningSummaryFilterType = "student"
	LearningSummaryFilterTypeSubject LearningSummaryFilterType = "subject"
)

func (l LearningSummaryFilterType) Valid() bool {
	switch l {
	case LearningSummaryFilterTypeYear,
		LearningSummaryFilterTypeWeek,
		LearningSummaryFilterTypeSchool,
		LearningSummaryFilterTypeClass,
		LearningSummaryFilterTypeTeacher,
		LearningSummaryFilterTypeStudent,
		LearningSummaryFilterTypeSubject:
		return true
	}
	return false
}

//type QueryLearningSummaryFilterResultItem struct {
//	Year        int    `json:"year,omitempty"`
//	WeekStart   int64  `json:"week_start,omitempty"`
//	WeekEnd     int64  `json:"week_end,omitempty"`
//	SchoolID    string `json:"school_id,omitempty"`
//	SchoolName  string `json:"school_name,omitempty"`
//	ClassID     string `json:"class_id,omitempty"`
//	ClassName   string `json:"class_name,omitempty"`
//	TeacherID   string `json:"teacher_id,omitempty"`
//	TeacherName string `json:"teacher_name,omitempty"`
//	StudentID   string `json:"student_id,omitempty"`
//	StudentName string `json:"student_name,omitempty"`
//	SubjectID   string `json:"subject_id,omitempty"`
//	SubjectName string `json:"subject_name,omitempty"`
//}

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
	ID   string `json:"id"`
	Name string `json:"name"`
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
