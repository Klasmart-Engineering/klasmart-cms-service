package entity

type StudentsReport struct {
	Items         []*StudentReportItem `json:"items"`
	AssessmentIDs []string             `json:"assessment_ids"`
}

type StudentReportItem struct {
	StudentID         string `json:"student_id"`
	StudentName       string `json:"student_name"`
	Attend            bool   `json:"attend"`
	AchievedCount     int    `json:"achieved_count"`
	NotAchievedCount  int    `json:"not_achieved_count"`
	NotAttemptedCount int    `json:"not_attempted_count"`
}

type StudentDetailReport struct {
	StudentName   string                   `json:"student_name"`
	Attend        bool                     `json:"attend"`
	Categories    []*StudentReportCategory `json:"categories"`
	AssessmentIDs []string                 `json:"assessment_ids"`
}

type StudentReportCategory struct {
	Name              ReportCategory `json:"name"`
	AchievedItems     []string       `json:"achieved_items"`
	NotAchievedItems  []string       `json:"not_achieved_items"`
	NotAttemptedItems []string       `json:"not_attempted_items"`
}

type ReportCategory string

const (
	ReportCategorySpeechLanguagesSkills     ReportCategory = "Speech & Language Skills"
	ReportCategoryFineMotorSkills           ReportCategory = "Fine Motor Skills"
	ReportCategoryGrossMotorSkills          ReportCategory = "Gross Motor Skills"
	ReportCategoryCognitiveSkills           ReportCategory = "Cognitive Skills"
	ReportCategoryPersonalDevelopment       ReportCategory = "Personal Development"
	ReportCategoryLanguageAndNumeracySkills ReportCategory = "Language and Numeracy Skills"
	ReportCategorySocialAndEmotional        ReportCategory = "Social and Emotional"
	ReportCategoryOral                      ReportCategory = "Oral"
	ReportCategoryLiteracy                  ReportCategory = "Literacy"
	ReportCategoryWholeChild                ReportCategory = "Whole-Child"
	ReportCategoryKnowledge                 ReportCategory = "Knowledge"
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
	ReportOutcomeStatusOptionAll          ReportOutcomeStatusOption = "all"
	ReportOutcomeStatusOptionAchieved     ReportOutcomeStatusOption = "achieved"
	ReportOutcomeStatusOptionNotAchieved  ReportOutcomeStatusOption = "not_achieved"
	ReportOutcomeStatusOptionNotAttempted ReportOutcomeStatusOption = "not_attempted"
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
	ReportSortByDesc ReportSortBy = "desc"
	ReportSortByAsc  ReportSortBy = "asc"
)

func (r ReportSortBy) Valid() bool {
	switch r {
	case ReportSortByDesc, ReportSortByAsc:
		return true
	default:
		return false
	}
}

type ReportOutcomeStatus string

const (
	ReportOutcomeStatusAchieved     ReportOutcomeStatus = "achieved"
	ReportOutcomeStatusNotAchieved  ReportOutcomeStatus = "not_achieved"
	ReportOutcomeStatusNotAttempted ReportOutcomeStatus = "not_attempted"
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
	Items  []*StudentReportItem
	Status ReportOutcomeStatusOption
}

func NewSortingStudentReportItems(items []*StudentReportItem, status ReportOutcomeStatusOption) *SortingStudentReportItems {
	return &SortingStudentReportItems{Items: items, Status: status}
}

func (s SortingStudentReportItems) Len() int {
	return len(s.Items)
}

func (s SortingStudentReportItems) Less(i, j int) bool {
	switch s.Status {
	default:
		fallthrough
	case ReportOutcomeStatusOptionAll:
		fallthrough
	case ReportOutcomeStatusOptionAchieved:
		if s.Items[i].AchievedCount != s.Items[j].AchievedCount {
			return s.Items[i].AchievedCount < s.Items[j].AchievedCount
		}
		fallthrough
	case ReportOutcomeStatusOptionNotAchieved:
		if s.Items[i].NotAchievedCount != s.Items[j].NotAchievedCount {
			return s.Items[i].NotAchievedCount < s.Items[j].NotAchievedCount
		}
		fallthrough
	case ReportOutcomeStatusOptionNotAttempted:
		return s.Items[i].NotAttemptedCount < s.Items[j].NotAttemptedCount
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
