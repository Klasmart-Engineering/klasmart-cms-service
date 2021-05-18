package entity

import "database/sql"

type Assessment struct {
	ID           string           `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	ScheduleID   string           `gorm:"column:schedule_id;type:varchar(64);not null" json:"schedule_id"`
	Type         AssessmentType   `gorm:"column:type;type:varchar(1024);not null" json:"type"` // add: 2021-05-15
	Title        string           `gorm:"column:title;type:varchar(1024);not null" json:"title"`
	CompleteTime int64            `gorm:"column:complete_time;type:bigint;not null" json:"complete_time"`
	Status       AssessmentStatus `gorm:"column:status;type:varchar(128);not null" json:"status"`

	CreateAt int64 `gorm:"column:create_at;type:bigint;not null" json:"create_at"`
	UpdateAt int64 `gorm:"column:update_at;type:bigint;not null" json:"update_at"`
	DeleteAt int64 `gorm:"column:delete_at;type:bigint;not null" json:"delete_at"`

	// Union Fields
	ClassLength  int   `gorm:"column:class_length;type:int;not null" json:"class_length"`
	ClassEndTime int64 `gorm:"column:class_end_time;type:bigint;not null" json:"class_end_time"`
}

type AssessmentType string

const (
	AssessmentTypeClassAndLiveOutcome = "class_and_live_outcome"
	AssessmentTypeClassAndLiveH5P     = "class_and_live_h5p"
	AssessmentTypeStudyH5P            = "study_h5p"
)

func (t AssessmentType) Valid() bool {
	switch t {
	case AssessmentTypeClassAndLiveOutcome, AssessmentTypeClassAndLiveH5P, AssessmentTypeStudyH5P:
		return true
	default:
		return false
	}
}

type NullAssessmentType struct {
	Value AssessmentType
	Valid bool
}

func (Assessment) TableName() string {
	return "assessments"
}

type AddAssessmentArgs struct {
	Type          AssessmentType `json:"type"`
	ScheduleID    string         `json:"schedule_id"`
	AttendanceIDs []string       `json:"attendance_ids"`
	ClassLength   int            `json:"class_length"`
	ClassEndTime  int64          `json:"class_end_time"`
}

func (a *AddAssessmentArgs) Valid() error {
	return nil
}

type AddAssessmentResult struct {
	IDs []string `json:"id"`
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

type AssessmentOrderBy string

const (
	ListAssessmentsOrderByClassEndTime     AssessmentOrderBy = "class_end_time"
	ListAssessmentsOrderByClassEndTimeDesc AssessmentOrderBy = "-class_end_time"
	ListAssessmentsOrderByCompleteTime     AssessmentOrderBy = "complete_time"
	ListAssessmentsOrderByCompleteTimeDesc AssessmentOrderBy = "-complete_time"
	ListAssessmentsOrderByCreateAt         AssessmentOrderBy = "create_at"
	ListAssessmentsOrderByCreateAtDesc     AssessmentOrderBy = "-create_at"
)

func (ob AssessmentOrderBy) Valid() bool {
	switch ob {
	case ListAssessmentsOrderByClassEndTime,
		ListAssessmentsOrderByClassEndTimeDesc,
		ListAssessmentsOrderByCompleteTime,
		ListAssessmentsOrderByCompleteTimeDesc,
		ListAssessmentsOrderByCreateAt,
		ListAssessmentsOrderByCreateAtDesc:
		return true
	default:
		return false
	}
}

type NullAssessmentsOrderBy struct {
	Value AssessmentOrderBy
	Valid bool
}

type AssessmentView struct {
	*Assessment
	Schedule        *Schedule                   `json:"schedule"`
	RoomID          string                      `json:"room_id"`
	Program         AssessmentProgram           `json:"program"`
	Subjects        []*AssessmentSubject        `json:"subjects"`
	Teachers        []*AssessmentTeacher        `json:"teachers"`
	Students        []*AssessmentStudent        `json:"students"`
	Class           AssessmentClass             `json:"class"`
	LessonPlan      *AssessmentLessonPlan       `json:"lesson_plan"`
	LessonMaterials []*AssessmentLessonMaterial `json:"lesson_materials"`
}

type AssessmentProgram struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AssessmentTeacher struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AssessmentStudent struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Checked bool   `json:"checked"`
}

type AssessmentClass struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AssessmentSubject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ConvertToViewsOptions struct {
	CheckedStudents  sql.NullBool
	EnableProgram    bool
	EnableSubjects   bool
	EnableTeachers   bool
	EnableStudents   bool
	EnableClass      bool
	EnableLessonPlan bool
}

type AssessmentLessonPlan struct {
	ID        string                      `json:"id"`
	Name      string                      `json:"name"`
	Materials []*AssessmentLessonMaterial `json:"materials"`
}

type AssessmentLessonMaterial struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	FileType FileType `json:"file_type"`
	Comment  string   `json:"comment"`
	Source   string   `json:"source"`
	Checked  bool     `json:"checked"`
}

type AssessmentExternalLessonPlan struct {
	ID        string                              `json:"id"`
	Name      string                              `json:"name"`
	Materials []*AssessmentExternalLessonMaterial `json:"materials"`
}

type AssessmentExternalLessonMaterial struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Source string `json:"source"`
}
