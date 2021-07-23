package entity

type QueryLearningSummaryFilterItemsArgs struct {
	Type LearningSummaryFilterType `json:"type" enums:"year,week,school,class,teacher,student,subject"`
	*LearningSummaryFilter
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

type QueryLearningSummaryFilterResultItem struct {
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
	Absent          bool                      `json:"absent"`
	ClassStartTime  int64                     `json:"class_start_time"`
	ScheduleTitle   string                    `json:"schedule_title"`
	LessonPlanName  string                    `json:"lesson_plan_name"`
	Outcomes        []*LearningSummaryOutcome `json:"outcomes"`
	TeacherFeedback string                    `json:"teacher_feedback"`
	// for debug
	ScheduleID   string `json:"schedule_id"`
	AssessmentID string `json:"assessment_id"`
}

type LearningSummaryOutcome struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type QueryAssignmentsSummaryResult struct {
	Completed         float64                               `json:"completed"`
	StudyItems        []*AssignmentsSummaryStudyItem        `json:"study_items"`
	HomeFunStudyItems []*AssignmentsSummaryHomeFunStudyItem `json:"home_fun_study_items"`
}

type AssignmentsSummaryStudyItem struct {
	AssessmentTitle string                    `json:"assessment_title"`
	LessonPlanName  string                    `json:"lesson_plan_name"`
	Outcomes        []*LearningSummaryOutcome `json:"outcomes"`
	TeacherFeedback string                    `json:"teacher_feedback"`
	// for debug
	ScheduleID   string `json:"schedule_id"`
	AssessmentID string `json:"assessment_id"`
}

type AssignmentsSummaryHomeFunStudyItem struct {
	AssessmentTitle string                    `json:"assessment_title"`
	Outcomes        []*LearningSummaryOutcome `json:"outcomes"`
	TeacherFeedback string                    `json:"teacher_feedback"`
	// for debug
	ScheduleID   string `json:"schedule_id"`
	AssessmentID string `json:"assessment_id"`
}
