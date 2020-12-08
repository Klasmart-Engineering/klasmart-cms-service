package entity

import (
	"encoding/json"
)

type Assessment struct {
	ID           string           `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	ScheduleID   string           `gorm:"column:schedule_id;type:varchar(64);not null" json:"schedule_id"`
	Title        string           `gorm:"column:title;type:varchar(1024);not null" json:"title"`
	ProgramID    string           `gorm:"column:program_id;type:varchar(64);not null" json:"program_id"`
	SubjectID    string           `gorm:"column:subject_id;type:varchar(64);not null" json:"subject"`
	TeacherIDs   string           `gorm:"column:teacher_ids;type:json;not null" json:"teacher_ids"`
	ClassLength  int              `gorm:"column:class_length;type:int;not null" json:"class_length"`
	ClassEndTime int64            `gorm:"column:class_end_time;type:bigint;not null" json:"class_end_time"`
	CompleteTime int64            `gorm:"column:complete_time;type:bigint;not null" json:"complete_time"`
	Status       AssessmentStatus `gorm:"column:status;type:varchar(128);not null" json:"status"`

	CreateAt int64 `gorm:"column:create_at;type:bigint;not null" json:"create_at"`
	UpdateAt int64 `gorm:"column:update_at;type:bigint;not null" json:"update_at"`
	DeleteAt int64 `gorm:"column:delete_at;type:bigint;not null" json:"delete_at"`
}

func (Assessment) TableName() string {
	return "assessments"
}

func (a Assessment) DecodeTeacherIDs() ([]string, error) {
	if a.TeacherIDs == "[]" {
		return nil, nil
	}
	var result []string
	if err := json.Unmarshal([]byte(a.TeacherIDs), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Assessment) EncodeAndSetTeacherIDs(teacherIDs []string) error {
	if len(teacherIDs) == 0 {
		a.TeacherIDs = "[]"
		return nil
	}
	bs, err := json.Marshal(teacherIDs)
	if err != nil {
		return err
	}
	a.TeacherIDs = string(bs)
	return nil
}

type AssessmentListView struct {
	ID           string              `json:"id"`
	Title        string              `json:"title"`
	Subject      AssessmentSubject   `json:"subject"`
	Program      AssessmentProgram   `json:"program"`
	Teachers     []AssessmentTeacher `json:"teachers"`
	ClassEndTime int64               `json:"class_end_time"`
	CompleteTime int64               `json:"complete_time"`
	Status       AssessmentStatus    `json:"status"`
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

type AssessmentDetailView struct {
	ID                    string                      `json:"id"`
	Title                 string                      `json:"title"`
	Attendances           []*AssessmentAttendanceView `json:"attendances"`
	Subject               AssessmentSubject           `json:"subject"`
	Teachers              []*AssessmentTeacher        `json:"teachers"`
	ClassEndTime          int64                       `json:"class_end_time"`
	ClassLength           int                         `json:"class_length"`
	NumberOfActivities    int                         `json:"number_of_activities"`
	NumberOfOutcomes      int                         `json:"number_of_outcomes"`
	CompleteTime          int64                       `json:"complete_time"`
	Status                AssessmentStatus            `json:"status"`
	OutcomeAttendanceMaps []OutcomeAttendanceMapView  `json:"outcome_attendance_maps"`
}

type OutcomeAttendanceMapView struct {
	OutcomeID     string   `json:"outcome_id"`
	OutcomeName   string   `json:"outcome_name"`
	Assumed       bool     `json:"assumed"`
	AttendanceIDs []string `json:"attendance_ids"`
	Skip          bool     `json:"skip"`
	NoneAchieved  bool     `json:"none_achieved"`
}

type AssessmentAttendanceView struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Checked bool   `json:"checked"`
}

type AssessmentSubject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AssessmentProgram struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AssessmentTeacher struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ListAssessmentsQuery struct {
	Status                         *AssessmentStatus                `json:"status"`
	TeacherIDs                     []string                         `json:"teacher_ids"`
	TeacherName                    *string                          `json:"teacher_name"`
	OrderBy                        *ListAssessmentsOrderBy          `json:"order_by"`
	Page                           int                              `json:"page"`
	PageSize                       int                              `json:"page_size"`
	TeacherAssessmentStatusFilters []*TeacherAssessmentStatusFilter `json:"teacher_assessment_status_filters"`
}

type TeacherAssessmentStatusFilter struct {
	TeacherID string           `json:"teacher_id"`
	Status    AssessmentStatus `json:"status"`
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

type ListAssessmentsOrderBy string

const (
	ListAssessmentsOrderByClassEndTime     ListAssessmentsOrderBy = "class_end_time"
	ListAssessmentsOrderByClassEndTimeDesc ListAssessmentsOrderBy = "-class_end_time"
	ListAssessmentsOrderByCompleteTime     ListAssessmentsOrderBy = "complete_time"
	ListAssessmentsOrderByCompleteTimeDesc ListAssessmentsOrderBy = "-complete_time"
)

func (ob ListAssessmentsOrderBy) Valid() bool {
	switch ob {
	case ListAssessmentsOrderByClassEndTime,
		ListAssessmentsOrderByClassEndTimeDesc,
		ListAssessmentsOrderByCompleteTime,
		ListAssessmentsOrderByCompleteTimeDesc:
		return true
	default:
		return false
	}
}

type ListAssessmentsResult struct {
	Total int                   `json:"total"`
	Items []*AssessmentListView `json:"items"`
}

type AddAssessmentCommand struct {
	Type          LiveTokenType `json:"type" enums:"preview,live"`
	ScheduleID    string        `json:"schedule_id"`
	AttendanceIDs []string      `json:"attendance_ids"`
	ClassLength   int           `json:"class_length"`
	ClassEndTime  int64         `json:"class_end_time"`
}

func (a *AddAssessmentCommand) Valid() error {
	return nil
}

type AddAssessmentResult struct {
	ID string `json:"id"`
}

type UpdateAssessmentCommand struct {
	ID                    string                  `json:"id"`
	Action                UpdateAssessmentAction  `json:"action" enums:"save,complete"`
	AttendanceIDs         *[]string               `json:"attendance_ids"`
	OutcomeAttendanceMaps *[]OutcomeAttendanceMap `json:"outcome_attendance_maps"`
}

type UpdateAssessmentAction string

const (
	UpdateAssessmentActionSave     UpdateAssessmentAction = "save"
	UpdateAssessmentActionComplete UpdateAssessmentAction = "complete"
)

func (a UpdateAssessmentAction) Valid() bool {
	switch a {
	case UpdateAssessmentActionSave, UpdateAssessmentActionComplete:
		return true
	default:
		return false
	}
}

type OutcomeAttendanceMap struct {
	OutcomeID     string   `json:"outcome_id"`
	AttendanceIDs []string `json:"attendance_ids"`
	Skip          bool     `json:"skip"`
	NoneAchieved  bool     `json:"none_achieved"`
}
