package entity

type Assessment struct {
	ID           string           `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	ScheduleID   string           `gorm:"column:schedule_id;type:varchar(64);not null" json:"schedule_id"`
	Title        string           `gorm:"column:title;type:varchar(1024);not null" json:"title"`
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

type AssessmentView struct {
	*Assessment
	Program  AssessmentProgram    `json:"program"`
	Subjects []*AssessmentSubject `json:"subjects"`
	Teachers []*AssessmentTeacher `json:"teachers"`
	Students []*AssessmentStudent `json:"students"`
}

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

type AssessmentStudent struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Checked bool   `json:"checked"`
}

type AssessmentClass struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AssessmentOutcomeView struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Checked bool   `json:"checked"`
}

type AssessmentContentView struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Comment    string   `json:"comment"`
	Checked    bool     `json:"checked"`
	OutcomeIDs []string `json:"outcome_ids"`
}

type QueryAssessmentsArgs struct {
	Status      *AssessmentStatus       `json:"status"`
	TeacherName *string                 `json:"teacher_name"`
	OrderBy     *ListAssessmentsOrderBy `json:"order_by"`
	ClassType   *ScheduleClassType      `json:"class_type"`
	Page        int                     `json:"page"`
	PageSize    int                     `json:"page_size"`
}

type AssessmentTeacherIDAndStatusPair struct {
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
	Total int               `json:"total"`
	Items []*AssessmentItem `json:"items"`
}

type QueryAssessmentsSummaryArgs struct {
	Status      *AssessmentStatus  `json:"status"`
	TeacherName *string            `json:"teacher_name"`
	ClassType   *ScheduleClassType `json:"class_type"`
}

type AssessmentsSummary struct {
	Complete   int `json:"complete"`
	InProgress int `json:"in_progress"`
}

type AddAssessmentArgs struct {
	ScheduleID    string   `json:"schedule_id"`
	AttendanceIDs []string `json:"attendance_ids"`
	ClassLength   int      `json:"class_length"`
	ClassEndTime  int64    `json:"class_end_time"`
}

func (a *AddAssessmentArgs) Valid() error {
	return nil
}

type AddAssessmentResult struct {
	ID string `json:"id"`
}

type UpdateAssessmentArgs struct {
	ID                 string                          `json:"id"`
	Action             UpdateAssessmentAction          `json:"action" enums:"save,complete"`
	StudentIDs         *[]string                       `json:"attendance_ids"`
	OutcomeAttendances *[]UpdateOutcomeAttendancesArgs `json:"outcome_attendances"`
	Materials          []UpdateAssessmentMaterialArgs  `json:"materials"`
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
