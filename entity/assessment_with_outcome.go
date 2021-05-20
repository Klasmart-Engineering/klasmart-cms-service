package entity

import "gitlab.badanamu.com.cn/calmisland/dbo"

type AssessmentItem struct {
	ID           string               `json:"id"`
	Title        string               `json:"title"`
	Program      AssessmentProgram    `json:"program"`
	Subjects     []*AssessmentSubject `json:"subjects"`
	Teachers     []*AssessmentTeacher `json:"teachers"`
	ClassEndTime int64                `json:"class_end_time"`
	CompleteTime int64                `json:"complete_time"`
	Status       AssessmentStatus     `json:"status"`
}

type AssessmentStatus string

const (
	AssessmentStatusInProgress AssessmentStatus = "in_progress"
	AssessmentStatusComplete   AssessmentStatus = "complete"
)

func (s AssessmentStatus) Valid() bool {
	switch s {
	case AssessmentStatusInProgress, AssessmentStatusComplete:
		return true
	default:
		return false
	}
}

type NullAssessmentStatus struct {
	Value AssessmentStatus
	Valid bool
}

type AssessmentDetail struct {
	ID           string               `json:"id"`
	Title        string               `json:"title"`
	Class        AssessmentClass      `json:"class"`
	Status       AssessmentStatus     `json:"status"`
	CompleteTime int64                `json:"complete_time"`
	Teachers     []*AssessmentTeacher `json:"teachers"`
	Students     []*AssessmentStudent `json:"students"`
	Program      AssessmentProgram    `json:"program"`
	Subjects     []*AssessmentSubject `json:"subjects"`
	ClassEndTime int64                `json:"class_end_time"`
	ClassLength  int                  `json:"class_length"`
	RoomID       string               `json:"room_id"`

	Plan               AssessmentContentView    `json:"plan"`
	Materials          []*AssessmentContentView `json:"materials"`
	OutcomeAttendances []*OutcomeAttendances    `json:"outcome_attendances"`
}

type OutcomeAttendances struct {
	OutcomeID     string   `json:"outcome_id"`
	OutcomeName   string   `json:"outcome_name"`
	Assumed       bool     `json:"assumed"`
	Skip          bool     `json:"skip"`
	NoneAchieved  bool     `json:"none_achieved"`
	AttendanceIDs []string `json:"attendance_ids"`
	Checked       bool     `json:"checked"`
}

type AssessmentContentView struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Comment    string   `json:"comment"`
	Checked    bool     `json:"checked"`
	OutcomeIDs []string `json:"outcome_ids"`
}

type QueryAssessmentsArgs struct {
	Status      NullAssessmentStatus   `json:"status"`
	TeacherName NullString             `json:"teacher_name"`
	OrderBy     NullAssessmentsOrderBy `json:"order_by"`
	ClassType   NullScheduleClassType  `json:"class_type"`
	Pager       dbo.Pager              `json:"pager"`
}

type AssessmentAllowTeacherIDAndStatusPair struct {
	TeacherID string           `json:"teacher_id"`
	Status    AssessmentStatus `json:"status"`
}

type NullAssessmentAllowTeacherIDAndStatusPairs struct {
	Values []*AssessmentAllowTeacherIDAndStatusPair
	Valid  bool
}

type ListAssessmentsStatus string

const (
	ListAssessmentsStatusAll        ListAssessmentsStatus = "all"
	ListAssessmentsStatusInProgress ListAssessmentsStatus = "in_progress"
	ListAssessmentsStatusComplete   ListAssessmentsStatus = "complete"
)

func (s ListAssessmentsStatus) Valid() bool {
	switch s {
	case ListAssessmentsStatusAll, ListAssessmentsStatusInProgress, ListAssessmentsStatusComplete:
		return true
	default:
		return false
	}
}

func (s ListAssessmentsStatus) AssessmentStatus() AssessmentStatus {
	return AssessmentStatus(s)
}

type ListAssessmentsResult struct {
	Total int               `json:"total"`
	Items []*AssessmentItem `json:"items"`
}

type QueryAssessmentsSummaryArgs struct {
	Status      NullAssessmentStatus  `json:"status"`
	TeacherName NullString            `json:"teacher_name"`
	ClassType   NullScheduleClassType `json:"class_type"`
}

type AssessmentsSummary struct {
	Complete   int `json:"complete"`
	InProgress int `json:"in_progress"`
}

type UpdateAssessmentArgs struct {
	ID                 string                          `json:"id"`
	Action             UpdateAssessmentAction          `json:"action" enums:"save,complete"`
	StudentIDs         *[]string                       `json:"attendance_ids"`
	OutcomeAttendances *[]UpdateOutcomeAttendancesArgs `json:"outcome_attendances"`
	Materials          []UpdateAssessmentMaterialArgs  `json:"materials"`
}

type UpdateOutcomeAttendancesArgs struct {
	OutcomeID     string   `json:"outcome_id"`
	Skip          bool     `json:"skip"`
	NoneAchieved  bool     `json:"none_achieved"`
	AttendanceIDs []string `json:"attendance_ids"`
}

type UpdateAssessmentMaterialArgs struct {
	ID      string `json:"id"`
	Checked bool   `json:"checked"`
	Comment string `json:"comment"`
}
