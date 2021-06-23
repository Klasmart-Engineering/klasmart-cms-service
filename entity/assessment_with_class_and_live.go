package entity

import (
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/dbo"
)

type AddClassAndLiveAssessmentArgs struct {
	ScheduleID    string   `json:"schedule_id"`
	AttendanceIDs []string `json:"attendance_ids"`
	ClassLength   int      `json:"class_length"`
	ClassEndTime  int64    `json:"class_end_time"`
}

// Valid implement jwt Claims interface
func (a *AddClassAndLiveAssessmentArgs) Valid() error {
	return nil
}

type AddAssessmentResult struct {
	ID string `json:"id"`
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
	AssessmentOrderByClassEndTime     AssessmentOrderBy = "class_end_time"
	AssessmentOrderByClassEndTimeDesc AssessmentOrderBy = "-class_end_time"
	AssessmentOrderByCompleteTime     AssessmentOrderBy = "complete_time"
	AssessmentOrderByCompleteTimeDesc AssessmentOrderBy = "-complete_time"
	AssessmentOrderByCreateAt         AssessmentOrderBy = "create_at"
	AssessmentOrderByCreateAtDesc     AssessmentOrderBy = "-create_at"
)

func (ob AssessmentOrderBy) Valid() bool {
	switch ob {
	case AssessmentOrderByClassEndTime,
		AssessmentOrderByClassEndTimeDesc,
		AssessmentOrderByCompleteTime,
		AssessmentOrderByCompleteTimeDesc,
		AssessmentOrderByCreateAt,
		AssessmentOrderByCreateAtDesc:
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
	ID   string `json:"id"`
	Name string `json:"name"`
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
	ID         string                              `json:"id"`
	Name       string                              `json:"name"`
	OutcomeIDs []string                            `json:"outcome_ids"`
	Materials  []*AssessmentExternalLessonMaterial `json:"materials"`
}

type AssessmentExternalLessonMaterial struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Source     string   `json:"source"`
	OutcomeIDs []string `json:"outcome_ids"`
}

type AssessmentItem struct {
	ID           string                `json:"id"`
	Title        string                `json:"title"`
	Program      AssessmentProgram     `json:"program"`
	Subjects     []*AssessmentSubject  `json:"subjects"`
	Teachers     []*AssessmentTeacher  `json:"teachers"`
	ClassEndTime int64                 `json:"class_end_time"`
	CompleteTime int64                 `json:"complete_time"`
	Status       AssessmentStatus      `json:"status"`
	LessonPlan   *AssessmentLessonPlan `json:"lesson_plan"`
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
	ID               string                          `json:"id"`
	Title            string                          `json:"title"`
	Status           AssessmentStatus                `json:"status"`
	Schedule         *Schedule                       `json:"schedule"`
	RoomID           string                          `json:"room_id"`
	Class            AssessmentClass                 `json:"class"`
	Teachers         []*AssessmentTeacher            `json:"teachers"`
	Students         []*AssessmentStudent            `json:"students"`
	Program          AssessmentProgram               `json:"program"`
	Subjects         []*AssessmentSubject            `json:"subjects"`
	ClassEndTime     int64                           `json:"class_end_time"`
	ClassLength      int                             `json:"class_length"`
	RemainingTime    int64                           `json:"remaining_time"`
	CompleteTime     int64                           `json:"complete_time"`
	LessonPlan       AssessmentDetailContent         `json:"lesson_plan"`
	LessonMaterials  []*AssessmentDetailContent      `json:"lesson_materials"`
	Outcomes         []*AssessmentDetailOutcome      `json:"outcomes"`
	StudentViewItems []*AssessmentStudentViewH5PItem `json:"student_view_items"`
}

type AssessmentDetailOutcome struct {
	OutcomeID     string   `json:"outcome_id"`
	OutcomeName   string   `json:"outcome_name"`
	Assumed       bool     `json:"assumed"`
	Skip          bool     `json:"skip"`
	NoneAchieved  bool     `json:"none_achieved"`
	AttendanceIDs []string `json:"attendance_ids"`
	Checked       bool     `json:"checked"`
}

type AssessmentDetailContent struct {
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

type ListAssessmentsResult struct {
	Total int               `json:"total"`
	Items []*AssessmentItem `json:"items"`
}

type QueryAssessmentsSummaryArgs struct {
	Status      NullAssessmentStatus `json:"status"`
	TeacherName NullString           `json:"teacher_name"`
}

type AssessmentsSummary struct {
	Complete   int `json:"complete"`
	InProgress int `json:"in_progress"`
}

type UpdateAssessmentArgs struct {
	ID               string                         `json:"id"`
	Action           UpdateAssessmentAction         `json:"action" enums:"save,complete"`
	StudentIDs       []string                       `json:"attendance_ids"`
	LessonMaterials  []*UpdateAssessmentContentArgs `json:"lesson_materials"`
	Outcomes         []*UpdateAssessmentOutcomeArgs `json:"outcomes"`
	StudentViewItems []*UpdateAssessmentH5PStudent  `json:"student_view_items"`
}

type UpdateAssessmentOutcomeArgs struct {
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
