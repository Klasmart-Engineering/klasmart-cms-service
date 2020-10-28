package entity

type StudentsReport struct {
	Items []StudentReportItem `json:"items"`
}

type StudentReportItem struct {
	StudentID         string `json:"student_id"`
	StudentName       string `json:"student_name"`
	AchievedCount     int    `json:"achieved_count"`
	NotAchievedCount  int    `json:"not_achieved_count"`
	NotAttemptedCount int    `json:"not_attempted_count"`
}

type StudentDetailReport struct {
	StudentName string                  `json:"student_name"`
	Categories  []StudentReportCategory `json:"categories"`
}

type StudentReportCategory struct {
	Name              ReportCategory `json:"name"`
	AchievedItems     []string       `json:"achieved_items"`
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
	TeacherID    string                    `json:"teacher_id"`
	ClassID      string                    `json:"class_id"`
	LessonPlanID string                    `json:"lesson_plan_id"`
	Status       ReportOutcomeStatusOption `json:"status"`
	SortBy       ReportSortBy              `json:"sort_by"`
	Operator     *Operator                 `json:"-"`
}

type GetStudentDetailReportCommand struct {
	StudentID    string    `json:"student_id"`
	TeacherID    string    `json:"teacher_id"`
	ClassID      string    `json:"class_id"`
	LessonPlanID string    `json:"lesson_plan_id"`
	Operator     *Operator `json:"-"`
}

type ReportOutcomeStatusOption string

const (
	ReportOutcomeStatusOptionAll          = "all"
	ReportOutcomeStatusOptionAchieved     = "achieved"
	ReportOutcomeStatusOptionNotAchieved  = "not_achieved"
	ReportOutcomeStatusOptionNotAttempted = "not_attempted"
)

func (o ReportOutcomeStatusOption) Valid() bool {
	switch o {
	case ReportOutcomeStatusOptionAll,
		ReportOutcomeStatusOptionAchieved,
		ReportOutcomeStatusOptionNotAchieved,
		ReportOutcomeStatusOptionNotAttempted:
		return true
	default:
		return false
	}
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
}

type ReportOutcomeStatus string

const (
	ReportOutcomeStatusAchieved     = "achieved"
	ReportOutcomeStatusNotAchieved  = "not_achieved"
	ReportOutcomeStatusNotAttempted = "not_attempted"
)

var reportOutcomeStatusPriorityMap = map[ReportOutcomeStatus]int{
	ReportOutcomeStatusNotAttempted: 1,
	ReportOutcomeStatusNotAchieved:  2,
	ReportOutcomeStatusAchieved:     3,
}

func ReportOutcomeStatusPriority(status ReportOutcomeStatus) int {
	return reportOutcomeStatusPriorityMap[status]
}

type AssessmentOutcomeKey struct {
	AssessmentID string `json:"assessment_id"`
	OutcomeID    string `json:"outcome_id"`
}

type SortingStudentReportItems struct {
	Items  []StudentReportItem
	Status ReportOutcomeStatusOption
}

func NewSortingStudentReportItems(items []StudentReportItem, status ReportOutcomeStatusOption) *SortingStudentReportItems {
	return &SortingStudentReportItems{Items: items, Status: status}
}

func (s SortingStudentReportItems) Len() int {
	return len(s.Items)
}

func (s SortingStudentReportItems) Less(i, j int) bool {
	switch s.Status {
	case ReportOutcomeStatusOptionNotAchieved:
		return s.Items[i].NotAchievedCount < s.Items[j].NotAchievedCount
	case ReportOutcomeStatusOptionNotAttempted:
		return s.Items[i].NotAttemptedCount < s.Items[j].NotAttemptedCount
	case ReportOutcomeStatusOptionAll:
		fallthrough
	case ReportOutcomeStatusOptionAchieved:
		fallthrough
	default:
		return s.Items[i].AchievedCount < s.Items[j].AchievedCount
	}
}

func (s SortingStudentReportItems) Swap(i, j int) {
	s.Items[i], s.Items[j] = s.Items[j], s.Items[i]
}

type ReportAttendanceOutcomeData struct {
	AllOutcomeIDs                         []string
	AllAttendanceIDExistsMap              map[string]bool
	AllAttendanceID2OutcomeIDsMap         map[string][]string
	SkipAttendanceID2OutcomeIDsMap        map[string][]string
	AchievedAttendanceID2OutcomeIDsMap    map[string][]string
	NotAchievedAttendanceID2OutcomeIDsMap map[string][]string
}
