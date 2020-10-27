package entity

type StudentsReport struct {
	Items []StudentReportItem `json:"items"`
}

type StudentReportItem struct {
	StudentName       string `json:"student_name"`
	AllAchievedCount  int    `json:"all_achieved_count"`
	NotAchievedCount  int    `json:"not_achieved_count"`
	NotAttemptedCount int    `json:"not_attempted_count"`
}

type StudentDetailReport struct {
	Categories []StudentReportCategory `json:"categories"`
}

type StudentReportCategory struct {
	Name              ReportCategory `json:"name"`
	AllAchievedItems  []string       `json:"all_achieved_items"`
	NotAchievedItems  []string       `json:"not_achieved_items"`
	NotAttemptedItems []string       `json:"not_attempted_items"`
}

type ReportCategory string

const (
	ReportCategorySpeechLanguagesSkills     = "Speech & Language Skills"
	ReportCategoryFineMotorSkills           = "Fine Motor Skills"
	ReportCategoryGrossMotorSkills          = "Gross Motor Skills"
	ReportCategoryCognitiveSkills           = "Cognitive Skills"
	ReportCategoryPersonalDevelopment       = "Personal Development"
	ReportCategoryLanguageAndNumeracySkills = "Language and Numeracy Skills"
	ReportCategorySocialAndEmotional        = "Social and Emotional"
	ReportCategoryOral                      = "Oral"
	ReportCategoryLiteracy                  = "Literacy"
	ReportCategoryWholeChild                = "Whole-Child"
	ReportCategoryKnowledge                 = "Knowledge"
)

var reportCategoryOrderMap = map[ReportCategory]int{
	ReportCategorySpeechLanguagesSkills:     0,
	ReportCategoryFineMotorSkills:           1,
	ReportCategoryGrossMotorSkills:          2,
	ReportCategoryCognitiveSkills:           3,
	ReportCategoryPersonalDevelopment:       4,
	ReportCategoryLanguageAndNumeracySkills: 5,
	ReportCategorySocialAndEmotional:        6,
	ReportCategoryOral:                      7,
	ReportCategoryLiteracy:                  8,
	ReportCategoryWholeChild:                9,
	ReportCategoryKnowledge:                 10,
}

func ReportCategoryOrder(category ReportCategory) int {
	return reportCategoryOrderMap[category]
}

type ReportLessonPlanInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ListStudentsReportCommand struct {
	TeacherID    string    `json:"teacher_id"`
	ClassID      string    `json:"class_id"`
	LessonPlanID string    `json:"lesson_plan_id"`
	Status       string    `json:"status"`
	SortBy       string    `json:"sort_by"`
	Operator     *Operator `json:"-"`
}

type GetStudentDetailReportCommand struct {
	StudentID    string    `json:"student_id"`
	TeacherID    string    `json:"teacher_id"`
	ClassID      string    `json:"class_id"`
	LessonPlanID string    `json:"lesson_plan_id"`
	Operator     *Operator `json:"-"`
}

type ReportOutcomeStatusOptions string

const (
	ReportOutcomeStatusOptionsAll          = "all"
	ReportOutcomeStatusOptionsAllAchieved  = "all_achieved"
	ReportOutcomeStatusOptionsNotAchieved  = "not_achieved"
	ReportOutcomeStatusOptionsNotAttempted = "not_attempted"
)

func (o ReportOutcomeStatusOptions) Valid() bool {
	switch o {
	case ReportOutcomeStatusOptionsAll,
		ReportOutcomeStatusOptionsAllAchieved,
		ReportOutcomeStatusOptionsNotAchieved,
		ReportOutcomeStatusOptionsNotAttempted:
		return true
	default:
		return false
	}
	return false
}

type ReportSortBy string

const (
	ReportSortByDescending = "descending"
	ReportSortByAscending  = "ascending"
)

func (r ReportSortBy) Valid() bool {
	switch r {
	case ReportSortByDescending, ReportSortByAscending:
		return true
	default:
		return false
	}
	return false
}

type ReportOutcomeStatus string

const (
	ReportOutcomeStatusAllAchieved  = "all_achieved"
	ReportOutcomeStatusNotAchieved  = "not_achieved"
	ReportOutcomeStatusNotAttempted = "not_attempted"
)

var reportOutcomeStatusPriorityMap = map[ReportOutcomeStatus]int{
	ReportOutcomeStatusNotAttempted: 1,
	ReportOutcomeStatusNotAchieved:  2,
	ReportOutcomeStatusAllAchieved:  3,
}

func ReportOutcomeStatusPriority(status ReportOutcomeStatus) int {
	return reportOutcomeStatusPriorityMap[status]
}

type AssessmentOutcomeKey struct {
	AssessmentID string `json:"assessment_id"`
	OutcomeID    string `json:"outcome_id"`
}
