package entity

import (
	"fmt"
	"time"
)

type Assessment struct {
	ID           string           `json:"id"`
	ScheduleID   string           `json:"schedule_id"`
	Title        string           `json:"title"`
	ProgramID    string           `json:"program_id"`
	SubjectID    string           `json:"subject"`
	TeacherID    string           `json:"teacher_id"`
	ClassLength  int              `json:"class_length"`
	ClassEndTime int64            `json:"class_end_time"`
	CompleteTime int64            `json:"complete_time"`
	Status       AssessmentStatus `json:"status"`

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
	DeletedAt int64 `json:"deleted_at"`
}

func (Assessment) TableName() string {
	return "assessments"
}

type AssessmentListView struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Subject      AssessmentSubject `json:"subject"`
	Program      AssessmentProgram `json:"program"`
	Teacher      AssessmentTeacher `json:"teacher"`
	ClassEndTime int64             `json:"class_end_time"`
	CompleteTime int64             `json:"complete_time"`
	Status       AssessmentStatus  `json:"status"`
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
	ID                    string                     `json:"id"`
	Title                 string                     `json:"title"`
	Attendances           []AssessmentAttendanceView `json:"attendances"`
	Subject               AssessmentSubject          `json:"subject"`
	Teacher               AssessmentTeacher          `json:"teacher"`
	ClassEndTime          int64                      `json:"class_end_time"`
	ClassLength           int                        `json:"class_length"`
	NumberOfActivities    int                        `json:"number_of_activities"`
	NumberOfOutcomes      int                        `json:"number_of_outcomes"`
	CompleteTime          int64                      `json:"complete_time"`
	OutcomeAttendanceMaps []OutcomeAttendanceMapView `json:"outcome_attendance_maps"`
}

type OutcomeAttendanceMapView struct {
	OutcomeID     string   `json:"outcome_id"`
	OutcomeName   string   `json:"outcome_name"`
	Assumed       bool     `json:"assumed"`
	AttendanceIDs []string `json:"attendance_ids"`
}

type AssessmentAttendanceView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

type ListAssessmentsCommand struct {
	Status      *ListAssessmentsStatus  `json:"status"`
	TeacherName *string                 `json:"teacher_name"`
	TeacherIDs  *[]string               `json:"teacher_ids"`
	OrderBy     *ListAssessmentsOrderBy `json:"order_by"`
	Page        *int                    `json:"page"`
	PageSize    *int                    `json:"page_size"`
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
	ScheduleID    string           `json:"schedule_id"`
	ClassID       string           `json:"class_id"`
	ClassName     string           `json:"class_name"`
	LessonName    string           `json:"lesson_name"`
	AttendanceIDs []string         `json:"attendance_ids"`
	ProgramID     string           `json:"program_id"`
	SubjectID     string           `json:"subject_id"`
	TeacherID     string           `json:"teacher_id"`
	ClassLength   int              `json:"class_length"`
	ClassEndTime  int64            `json:"class_end_time"`
	CompleteTime  int64            `json:"-"`
	Status        AssessmentStatus `json:"-"`
}

func (cmd AddAssessmentCommand) Title() string {
	return fmt.Sprintf("%s-%s-%s",
		time.Unix(cmd.ClassEndTime, 0).Format("20060102"),
		cmd.ClassName,
		cmd.LessonName,
	)
}

type UpdateAssessmentCommand struct {
	ID                    string                  `json:"id"`
	Action                UpdateAssessmentAction  `json:"action"`
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
	AttendanceIDs []string `json:"attendance_i_ds"`
}
