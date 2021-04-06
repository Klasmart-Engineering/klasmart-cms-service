package entity

import "strings"

type StudentsAchievementReportResponse struct {
	Items         []*StudentAchievementReportItem `json:"items"`
	AssessmentIDs []string                        `json:"assessment_ids"`
}

type StudentAchievementReportItem struct {
	StudentID         string `json:"student_id"`
	StudentName       string `json:"student_name"`
	Attend            bool   `json:"attend"`
	AchievedCount     int    `json:"achieved_count"`
	NotAchievedCount  int    `json:"not_achieved_count"`
	NotAttemptedCount int    `json:"not_attempted_count"`
}

type StudentAchievementReportResponse struct {
	StudentName   string                                  `json:"student_name"`
	Attend        bool                                    `json:"attend"`
	Categories    []*StudentAchievementReportCategoryItem `json:"categories"`
	AssessmentIDs []string                                `json:"assessment_ids"`
}

type StudentAchievementReportCategoryItem struct {
	Name              string   `json:"name"`
	AchievedItems     []string `json:"achieved_items"`
	NotAchievedItems  []string `json:"not_achieved_items"`
	NotAttemptedItems []string `json:"not_attempted_items"`
}

type ReportLessonPlanInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ListStudentsAchievementReportRequest struct {
	TeacherID    string                    `json:"teacher_id"`
	ClassID      string                    `json:"class_id"`
	LessonPlanID string                    `json:"lesson_plan_id"`
	Status       ReportOutcomeStatusOption `json:"status"`
	SortBy       ReportSortBy              `json:"sort_by"`
	Operator     *Operator                 `json:"-"`
}

type GetStudentAchievementReportRequest struct {
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
	Items  []*StudentAchievementReportItem
	Status ReportOutcomeStatusOption
}

func NewSortingStudentReportItems(items []*StudentAchievementReportItem, status ReportOutcomeStatusOption) *SortingStudentReportItems {
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

type TeacherReport struct {
	Categories []*TeacherReportCategory `json:"categories"`
}

type TeacherReportCategory struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

type TeacherReportSortByCount TeacherReport

func (t *TeacherReportSortByCount) Len() int {
	return len(t.Categories)
}

func (t *TeacherReportSortByCount) Less(i, j int) bool {
	return len(t.Categories[i].Items) < len(t.Categories[j].Items)
}

func (t *TeacherReportSortByCount) Swap(i, j int) {
	t.Categories[i], t.Categories[j] = t.Categories[j], t.Categories[i]
}

type ListStudentsPerformanceReportRequest struct {
	TeacherID    string    `json:"teacher_id"`
	ClassID      string    `json:"class_id"`
	LessonPlanID string    `json:"lesson_plan_id"`
	Operator     *Operator `json:"-"`
}

type ListStudentsPerformanceReportResponse struct {
	Items         []*StudentsPerformanceReportItem `json:"items"`
	AssessmentIDs []string                         `json:"assessment_ids"`
}

type StudentsPerformanceReportItem struct {
	StudentID         string   `json:"student_id"`
	StudentName       string   `json:"student_name"`
	Attend            bool     `json:"attend"`
	AchievedNames     []string `json:"achieved_names"`
	NotAchievedNames  []string `json:"not_achieved_names"`
	NotAttemptedNames []string `json:"not_attempted_names"`
}

type GetStudentPerformanceReportRequest struct {
	StudentID    string    `json:"student_id"`
	TeacherID    string    `json:"teacher_id"`
	ClassID      string    `json:"class_id"`
	LessonPlanID string    `json:"lesson_plan_id"`
	Operator     *Operator `json:"-"`
}

type GetStudentPerformanceReportResponse struct {
	Items         []*StudentPerformanceReportItem `json:"items"`
	AssessmentIDs []string                        `json:"assessment_ids"`
}

type StudentPerformanceReportItem struct {
	StudentID         string   `json:"student_id"`
	StudentName       string   `json:"student_name"`
	Attend            bool     `json:"attend"`
	AchievedNames     []string `json:"achieved_names"`
	NotAchievedNames  []string `json:"not_achieved_names"`
	NotAttemptedNames []string `json:"not_attempted_names"`
	ScheduleID        string   `json:"schedule_id"`
	ScheduleStartTime int64    `json:"schedule_start_time"`
}

type ListStudentsPerformanceH5PReportRequest struct {
	StudentID    string    `json:"student_id"`
	TeacherID    string    `json:"teacher_id"`
	ClassID      string    `json:"class_id"`
	LessonPlanID string    `json:"lesson_plan_id"`
	Operator     *Operator `json:"-"`
}

type ListStudentsPerformanceH5PReportResponse struct {
	Items []*StudentsPerformanceH5PReportItem `json:"items"`
}

type StudentsPerformanceH5PReportItem struct {
	StudentID   string `json:"student_id"`
	StudentName string `json:"student_name"`
	Attend      bool   `json:"attend"`
	SpentTime   int64  `json:"spent_time"`
}

type GetStudentPerformanceH5PReportRequest struct {
	StudentID    string    `json:"student_id"`
	TeacherID    string    `json:"teacher_id"`
	ClassID      string    `json:"class_id"`
	LessonPlanID string    `json:"lesson_plan_id"`
	Operator     *Operator `json:"-"`
}

type GetStudentPerformanceH5PReportResponse struct {
	Items []*StudentPerformanceH5PReportItem `json:"items"`
}

type StudentPerformanceH5PReportItem struct {
	MaterialID              string                   `json:"material_id"`
	MaterialName            string                   `json:"material_name"`
	TotalSpentTime          int64                    `json:"total_spent_time"`
	AvgSpentTime            int64                    `json:"avg_spent_time"`
	ActivityType            ActivityType             `json:"activity_type" enums:"H5P.ImageSequencing,H5P.MemoryGame,H5P.ImagePair,H5P.Flashcards"`
	ActivityImageSequencing *ActivityImageSequencing `json:"activity_image_sequencing,omitempty"`
	ActivityMemoryGame      *ActivityMemoryGame      `json:"activity_memory_game,omitempty"`
	ActivityImagePair       *ActivityImagePair       `json:"activity_image_pair,omitempty"`
	ActivityFlashCards      *ActivityFlashCards      `json:"activity_flash_cards,omitempty"`
}

// region activities

type ActivityType string

const (
	ActivityTypeImageSequencing ActivityType = "H5P.ImageSequencing"
	ActivityTypeMemoryGame      ActivityType = "H5P.MemoryGame"
	ActivityTypeImagePair       ActivityType = "H5P.ImagePair"
	ActivityTypeFlashCards      ActivityType = "H5P.Flashcards"
)

func ParseActivityType(s string) ActivityType {
	return ActivityType(strings.Split(s, " ")[0])
}

func (t ActivityType) Valid() bool {
	switch t {
	case ActivityTypeImageSequencing,
		ActivityTypeMemoryGame,
		ActivityTypeImagePair,
		ActivityTypeFlashCards:
		return true
	default:
		return false
	}
}

type ActivityImageSequencing struct {
	CardsNumber int                                  `json:"cards_number"`
	PlayRecords []*ActivityImageSequencingPlayRecord `json:"play_records"`
	PlayTimes   int                                  `json:"play_times"`
}

type ActivityImageSequencingPlayRecord struct {
	StartTime         int64 `json:"start_time"`
	EndTime           int64 `json:"end_time"`
	Duration          int64 `json:"duration"`
	CorrectCardsCount int   `json:"correct_cards_count"`
}

type ActivityMemoryGame struct {
	PairsNumber int                             `json:"pairs_number"`
	PlayRecords []*ActivityMemoryGamePlayRecord `json:"play_records"`
	PlayTimes   int                             `json:"play_times"`
}

type ActivityMemoryGamePlayRecord struct {
	StartTime   int64 `json:"start_time"`
	EndTime     int64 `json:"end_time"`
	Duration    int64 `json:"duration"`
	ClicksCount int   `json:"clicks_count"`
}

type ActivityImagePair struct {
	ParisNumber int                            `json:"paris_number"`
	PlayRecords []*ActivityImagePairPlayRecord `json:"play_records"`
	PlayTimes   int                            `json:"play_times"`
}

type ActivityImagePairPlayRecord struct {
	StartTime         int64 `json:"start_time"`
	EndTime           int64 `json:"end_time"`
	Duration          int64 `json:"duration"`
	CorrectPairsCount int   `json:"correct_pairs_count"`
}

type ActivityFlashCards struct {
	CardsNumber int                             `json:"cards_number"`
	PlayRecords []*ActivityFlashCardsPlayRecord `json:"play_records"`
}

type ActivityFlashCardsPlayRecord struct {
	StartTime         int64 `json:"start_time"`
	EndTime           int64 `json:"end_time"`
	Duration          int64 `json:"duration"`
	CorrectCardsCount int   `json:"correct_cards_count"`
}

type GetTeachingHoursReportArgs struct {
	OrgID      *string  `json:"org_id`
	SchoolIDs  []string `json:"school_ids"`
	TeacherIDs []string `json:"teacher_ids"`
	Days       int      `json:"days"`
}

type GetTeachingHoursReportResultItem struct {
	TeacherID   string  `json:"teacher_id"`
	TeacherName string  `json:"teacher_name"`
	Durations   []int64 `json:"durations"`
}

// endregion activities
