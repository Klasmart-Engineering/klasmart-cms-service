package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type RepeatType string

const (
	RepeatTypeDaily   RepeatType = "daily"
	RepeatTypeWeekly  RepeatType = "weekly"
	RepeatTypeMonthly RepeatType = "monthly"
	RepeatTypeYearly  RepeatType = "yearly"
)

func (t RepeatType) Valid() bool {
	switch t {
	case RepeatTypeDaily, RepeatTypeWeekly, RepeatTypeMonthly, RepeatTypeYearly:
		return true
	default:
		return false
	}
}

type RepeatEndType string

const (
	RepeatEndNever      RepeatEndType = "never"
	RepeatEndAfterCount RepeatEndType = "after_count"
	RepeatEndAfterTime  RepeatEndType = "after_time"
)

type RepeatWeekday string

const (
	RepeatWeekdaySunday    RepeatWeekday = "Sunday"
	RepeatWeekdayMonday    RepeatWeekday = "Monday"
	RepeatWeekdayTuesday   RepeatWeekday = "Tuesday"
	RepeatWeekdayWednesday RepeatWeekday = "Wednesday"
	RepeatWeekdayThursday  RepeatWeekday = "Thursday"
	RepeatWeekdayFriday    RepeatWeekday = "Friday"
	RepeatWeekdaySaturday  RepeatWeekday = "Saturday"
)

func (w RepeatWeekday) Valid() bool {
	switch w {
	case RepeatWeekdaySunday, RepeatWeekdayMonday, RepeatWeekdayTuesday, RepeatWeekdayWednesday, RepeatWeekdayThursday, RepeatWeekdayFriday, RepeatWeekdaySaturday:
		return true
	default:
		return false
	}
}

func (w RepeatWeekday) TimeWeekday() time.Weekday {
	switch w {
	case RepeatWeekdaySunday:
		return time.Sunday
	case RepeatWeekdayMonday:
		return time.Monday
	case RepeatWeekdayTuesday:
		return time.Tuesday
	case RepeatWeekdayWednesday:
		return time.Wednesday
	case RepeatWeekdayThursday:
		return time.Thursday
	case RepeatWeekdayFriday:
		return time.Friday
	case RepeatWeekdaySaturday:
		return time.Saturday
	}
	return 0
}

type RepeatMonthlyOnType string

const (
	RepeatMonthlyOnDate RepeatMonthlyOnType = "date"
	RepeatMonthlyOnWeek RepeatMonthlyOnType = "week"
)

func (t RepeatMonthlyOnType) Valid() bool {
	switch t {
	case RepeatMonthlyOnDate, RepeatMonthlyOnWeek:
		return true
	default:
		return false
	}
}

type RepeatYearlyOnType string

const (
	RepeatYearlyOnDate RepeatYearlyOnType = "date"
	RepeatYearlyOnWeek RepeatYearlyOnType = "week"
)

func (t RepeatYearlyOnType) Valid() bool {
	switch t {
	case RepeatYearlyOnDate, RepeatYearlyOnWeek:
		return true
	default:
		return false
	}
}

func (t RepeatEndType) Valid() bool {
	switch t {
	case RepeatEndNever, RepeatEndAfterCount, RepeatEndAfterTime:
		return true
	default:
		return false
	}
}

type RepeatWeekSeq string

const (
	RepeatWeekSeqFirst  RepeatWeekSeq = "first"
	RepeatWeekSeqSecond RepeatWeekSeq = "second"
	RepeatWeekSeqThird  RepeatWeekSeq = "third"
	RepeatWeekSeqFourth RepeatWeekSeq = "fourth"
	RepeatWeekSeqLast   RepeatWeekSeq = "last"
)

func (s RepeatWeekSeq) Valid() bool {
	switch s {
	case RepeatWeekSeqFirst, RepeatWeekSeqSecond, RepeatWeekSeqThird, RepeatWeekSeqFourth, RepeatWeekSeqLast:
		return true
	default:
		return false
	}
}

func (s RepeatWeekSeq) Offset() int {
	switch s {
	case RepeatWeekSeqFirst:
		return 1
	case RepeatWeekSeqSecond:
		return 2
	case RepeatWeekSeqThird:
		return 3
	case RepeatWeekSeqFourth:
		return 4
	case RepeatWeekSeqLast:
		return -1
	}
	return 0
}

type RepeatOptions struct {
	Type    RepeatType    `json:"type,omitempty" enums:"daily,weekly,monthly,yearly"`
	Daily   RepeatDaily   `json:"daily,omitempty"`
	Weekly  RepeatWeekly  `json:"weekly,omitempty"`
	Monthly RepeatMonthly `json:"monthly,omitempty"`
	Yearly  RepeatYearly  `json:"yearly,omitempty"`
}

type RepeatDaily struct {
	Interval int       `json:"interval,omitempty"`
	End      RepeatEnd `json:"end,omitempty"`
}

type RepeatWeekly struct {
	Interval int             `json:"interval,omitempty"`
	On       []RepeatWeekday `json:"on,omitempty" enums:"Sunday,Monday,Tuesday,Wednesday,Thursday,Friday,Saturday"`
	End      RepeatEnd       `json:"end,omitempty"`
}

type RepeatMonthly struct {
	Interval  int                 `json:"interval,omitempty"`
	OnType    RepeatMonthlyOnType `json:"on_type,omitempty" enums:"date,week"`
	OnDateDay int                 `json:"on_date_day,omitempty"`
	OnWeekSeq RepeatWeekSeq       `json:"on_week_seq,omitempty" enums:"first,second,third,fourth,last"`
	OnWeek    RepeatWeekday       `json:"on_week,omitempty" enums:"Sunday,Monday,Tuesday,Wednesday,Thursday,Friday,Saturday"`
	End       RepeatEnd           `json:"end,omitempty"`
}

type RepeatYearly struct {
	Interval    int                `json:"interval,omitempty"`
	OnType      RepeatYearlyOnType `json:"on_type,omitempty" enums:"date,week"`
	OnDateMonth int                `json:"on_date_month,omitempty"`
	OnDateDay   int                `json:"on_date_day,omitempty"`
	OnWeekMonth int                `json:"on_week_month,omitempty"`
	OnWeekSeq   RepeatWeekSeq      `json:"on_week_seq,omitempty" enums:"first,second,third,fourth,last"`
	OnWeek      RepeatWeekday      `json:"on_week,omitempty" enums:"Sunday,Monday,Tuesday,Wednesday,Thursday,Friday,Saturday"`
	End         RepeatEnd          `json:"end,omitempty"`
}

type RepeatEnd struct {
	Type       RepeatEndType `json:"type,omitempty"  enums:"never,after_count,after_time"`
	AfterCount int           `json:"after_count,omitempty"`
	AfterTime  int64         `json:"after_time,omitempty"`
}

type ScheduleClassType string

const (
	ScheduleClassTypeOnlineClass  ScheduleClassType = "OnlineClass"
	ScheduleClassTypeOfflineClass ScheduleClassType = "OfflineClass"
	ScheduleClassTypeHomework     ScheduleClassType = "Homework"
	ScheduleClassTypeTask         ScheduleClassType = "Task"
	ScheduleClassTypeSubject      ScheduleClassType = "Subject"
)

func (s ScheduleClassType) Valid() bool {
	switch s {
	case ScheduleClassTypeOnlineClass, ScheduleClassTypeOfflineClass, ScheduleClassTypeHomework, ScheduleClassTypeTask:
		return true
	default:
		return false
	}
}

func (s ScheduleClassType) ToLabel() ScheduleClassTypeLabel {
	switch s {
	case ScheduleClassTypeOnlineClass:
		return ScheduleClassTypeLabelOnlineClass
	case ScheduleClassTypeOfflineClass:
		return ScheduleClassTypeLabelOfflineClass
	case ScheduleClassTypeHomework:
		return ScheduleClassTypeLabelHomework
	case ScheduleClassTypeTask:
		return ScheduleClassTypeLabelTask
	}
	return ScheduleClassTypeLabelInvalid
}

type NullScheduleClassType struct {
	Value ScheduleClassType
	Valid bool
}

type NullScheduleClassTypes struct {
	Value []ScheduleClassType
	Valid bool
}

type ScheduleClassTypeLabel string

const (
	ScheduleClassTypeLabelInvalid      ScheduleClassTypeLabel = "schedule_detail_invalid"
	ScheduleClassTypeLabelOnlineClass  ScheduleClassTypeLabel = "schedule_detail_online_class"
	ScheduleClassTypeLabelOfflineClass ScheduleClassTypeLabel = "schedule_detail_offline_class"
	ScheduleClassTypeLabelHomework     ScheduleClassTypeLabel = "schedule_detail_homework"
	ScheduleClassTypeLabelTask         ScheduleClassTypeLabel = "schedule_detail_task"
)

func (s ScheduleClassTypeLabel) String() string {
	return string(s)
}

//Live (online class)
//Class (offline class)
//Study (homework)
//Task (task)
func (s ScheduleClassType) ConvertToLiveClassType() LiveClassType {
	switch s {
	case ScheduleClassTypeOnlineClass:
		return LiveClassTypeLive
	case ScheduleClassTypeOfflineClass:
		return LiveClassTypeClass
	case ScheduleClassTypeHomework:
		return LiveClassTypeStudy
	case ScheduleClassTypeTask:
		return LiveClassTypeTask
	default:
		return LiveClassTypeInvalid
	}
}

func (s ScheduleClassType) String() string {
	return string(s)
}

type Schedule struct {
	ID    string `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Title string `gorm:"column:title;type:varchar(100)" json:"title"`
	// disabled
	ClassID      string         `gorm:"column:class_id;type:varchar(100)" json:"class_id"`
	LessonPlanID string         `gorm:"column:lesson_plan_id;type:varchar(100)" json:"lesson_plan_id"`
	OrgID        string         `gorm:"column:org_id;type:varchar(100)" json:"org_id"`
	StartAt      int64          `gorm:"column:start_at;type:bigint" json:"start_at"`
	EndAt        int64          `gorm:"column:end_at;type:bigint" json:"end_at"`
	Status       ScheduleStatus `gorm:"column:status;type:varchar(100)" json:"status"`
	IsAllDay     bool           `gorm:"column:is_all_day;default:false" json:"is_all_day"`
	// disabled
	SubjectID       string            `gorm:"column:subject_id;type:varchar(100)" json:"subject_id"`
	ProgramID       string            `gorm:"column:program_id;type:varchar(100)" json:"program_id"`
	ClassType       ScheduleClassType `gorm:"column:class_type;type:varchar(100)" json:"class_type"`
	DueAt           int64             `gorm:"column:due_at;type:bigint" json:"due_at"`
	Description     string            `gorm:"column:description;type:varchar(500)" json:"description"`
	Attachment      string            `gorm:"column:attachment;type:text;" json:"attachment"`
	ScheduleVersion int64             `gorm:"column:version;type:bigint" json:"schedule_version"`
	RepeatID        string            `gorm:"column:repeat_id;type:varchar(100)" json:"repeat_id"`
	RepeatJson      string            `gorm:"column:repeat;type:json;" json:"repeat_json"`
	IsHidden        bool              `gorm:"column:is_hidden;default:false" json:"is_hidden"`
	IsHomeFun       bool              `gorm:"column:is_home_fun;default:false" json:"is_home_fun"`
	CreatedID       string            `gorm:"column:created_id;type:varchar(100)" json:"created_id"`
	UpdatedID       string            `gorm:"column:updated_id;type:varchar(100)" json:"updated_id"`
	DeletedID       string            `gorm:"column:deleted_id;type:varchar(100)" json:"deleted_id"`
	CreatedAt       int64             `gorm:"column:created_at;type:bigint" json:"created_at"`
	UpdatedAt       int64             `gorm:"column:updated_at;type:bigint" json:"updated_at"`
	DeleteAt        int64             `gorm:"column:delete_at;type:bigint" json:"delete_at"`
}

func (s *Schedule) ToScheduleSimplified() *ScheduleSimplified {
	return &ScheduleSimplified{
		ID:           s.ID,
		LessonPlanID: s.LessonPlanID,
		OrgID:        s.OrgID,
	}
}

type ScheduleStatus string

func (s ScheduleStatus) Valid() bool {
	switch s {
	case ScheduleStatusNotStart, ScheduleStatusStarted, ScheduleStatusClosed:
		return true
	default:
		return false
	}
}

func (s ScheduleStatus) GetScheduleStatus(input ScheduleStatusInput) ScheduleStatus {
	status := s
	switch input.ClassType {
	case ScheduleClassTypeHomework, ScheduleClassTypeTask:
		if input.DueAt > 0 {
			endAt := utils.TodayEndByTimeStamp(input.DueAt, time.Local).Unix()
			if endAt < time.Now().Unix() {
				status = ScheduleStatusClosed
			}
		}
	case ScheduleClassTypeOfflineClass, ScheduleClassTypeOnlineClass:
		if input.EndAt > 0 && input.EndAt < time.Now().Unix() {
			status = ScheduleStatusClosed
		}
	}
	return status
}

type ScheduleStatusInput struct {
	EndAt     int64
	DueAt     int64
	ClassType ScheduleClassType
}

const (
	ScheduleStatusNotStart ScheduleStatus = "NotStart"
	ScheduleStatusStarted  ScheduleStatus = "Started"
	ScheduleStatusClosed   ScheduleStatus = "Closed"
)

func (Schedule) TableName() string {
	return constant.TableNameSchedule
}

func (s Schedule) Clone() Schedule {
	newItem := s
	return newItem
}

type ScheduleAddView struct {
	Title                  string            `json:"title"`
	ClassID                string            `json:"class_id"`
	LessonPlanID           string            `json:"lesson_plan_id"`
	ClassRosterTeacherIDs  []string          `json:"class_roster_teacher_ids"`
	ClassRosterStudentIDs  []string          `json:"class_roster_student_ids"`
	ParticipantsTeacherIDs []string          `json:"participants_teacher_ids"`
	ParticipantsStudentIDs []string          `json:"participants_student_ids"`
	OrgID                  string            `json:"org_id"`
	StartAt                int64             `json:"start_at"`
	EndAt                  int64             `json:"end_at"`
	SubjectIDs             []string          `json:"subject_ids"`
	ProgramID              string            `json:"program_id"`
	ClassType              ScheduleClassType `json:"class_type" enums:"OnlineClass,OfflineClass,Homework,Task"`
	DueAt                  int64             `json:"due_at"`
	Description            string            `json:"description"`
	Attachment             ScheduleShortInfo `json:"attachment"`
	Version                int64             `json:"version"`
	RepeatID               string            `json:"-"`
	Repeat                 RepeatOptions     `json:"repeat"`
	IsAllDay               bool              `json:"is_all_day"`
	IsRepeat               bool              `json:"is_repeat"`
	IsForce                bool              `json:"is_force"`
	TimeZoneOffset         int               `json:"time_zone_offset"`
	Location               *time.Location    `json:"-"`
	IsHomeFun              bool              `json:"is_home_fun"`
	OutcomeIDs             []string          `json:"outcome_ids"`
}

type ScheduleEditValidation struct {
	ClassRosterTeacherIDs  []string
	ClassRosterStudentIDs  []string
	ParticipantsTeacherIDs []string
	ParticipantsStudentIDs []string
	ClassID                string
	ClassType              ScheduleClassType
	Title                  string
	OutcomeIDs             []string
}

func (s *ScheduleAddView) ToSchedule(ctx context.Context) (*Schedule, error) {
	schedule := &Schedule{
		ID:              utils.NewID(),
		Title:           s.Title,
		ClassID:         s.ClassID,
		LessonPlanID:    s.LessonPlanID,
		OrgID:           s.OrgID,
		StartAt:         s.StartAt,
		EndAt:           s.EndAt,
		ProgramID:       s.ProgramID,
		ClassType:       s.ClassType,
		DueAt:           s.DueAt,
		Status:          ScheduleStatusNotStart,
		Description:     s.Description,
		ScheduleVersion: 0,
		CreatedAt:       time.Now().Unix(),
		UpdatedAt:       time.Now().Unix(),
		IsAllDay:        s.IsAllDay,
		IsHomeFun:       s.IsHomeFun,
	}
	if schedule.ClassType != ScheduleClassTypeHomework {
		schedule.IsHomeFun = false
	}
	if s.IsRepeat {
		b, err := json.Marshal(s.Repeat)
		if err != nil {
			log.Warn(ctx, "ToSchedule:marshal schedule repeat error", log.Any("repeat", s.Repeat))
			return nil, err
		}
		schedule.RepeatJson = string(b)
		schedule.RepeatID = utils.NewID()
	} else {
		schedule.RepeatJson = "{}"
	}
	b, err := json.Marshal(s.Attachment)
	if err != nil {
		log.Warn(ctx, "ToSchedule:marshal attachment error", log.Any("attachment", s.Attachment))
		return nil, err
	}
	schedule.Attachment = string(b)
	return schedule, nil
}

type ScheduleUpdateView struct {
	ID       string           `json:"id"`
	EditType ScheduleEditType `json:"repeat_edit_options"  enums:"only_current,with_following"`

	ScheduleAddView
}

type ScheduleListView struct {
	ID                 string            `json:"id"`
	Title              string            `json:"title"`
	StartAt            int64             `json:"start_at"`
	EndAt              int64             `json:"end_at"`
	IsRepeat           bool              `json:"is_repeat"`
	LessonPlanID       string            `json:"lesson_plan_id"`
	ClassType          ScheduleClassType `json:"class_type" enums:"OnlineClass,OfflineClass,Homework,Task"`
	ClassTypeLabel     ScheduleShortInfo `json:"class_type_label"`
	Status             ScheduleStatus    `json:"status" enums:"NotStart,Started,Closed"`
	ClassID            string            `json:"class_id"`
	DueAt              int64             `json:"due_at"`
	IsHidden           bool              `json:"is_hidden"`
	RoleType           ScheduleRoleType  `json:"role_type"`
	ExistFeedback      bool              `json:"exist_feedback"`
	IsHomeFun          bool              `json:"is_home_fun"`
	ExistAssessment    bool              `json:"exist_assessment"`
	CompleteAssessment bool              `json:"complete_assessment"`
}

type ScheduleDateView struct {
	DayDate string
}

type ScheduleDetailsView struct {
	ID                   string                        `json:"id"`
	Title                string                        `json:"title"`
	Attachment           ScheduleShortInfo             `json:"attachment"`
	OrgID                string                        `json:"org_id"`
	StartAt              int64                         `json:"start_at"`
	EndAt                int64                         `json:"end_at"`
	ClassType            ScheduleClassType             `json:"class_type" enums:"OnlineClass,OfflineClass,Homework,Task"`
	ClassTypeLabel       ScheduleShortInfo             `json:"class_type_label"`
	DueAt                int64                         `json:"due_at"`
	Description          string                        `json:"description"`
	Version              int64                         `json:"version"`
	IsAllDay             bool                          `json:"is_all_day"`
	IsRepeat             bool                          `json:"is_repeat"`
	Repeat               RepeatOptions                 `json:"repeat"`
	Status               ScheduleStatus                `json:"status" enums:"NotStart,Started,Closed"`
	RealTimeStatus       ScheduleRealTimeView          `json:"real_time_status"`
	ClassRosterTeachers  []*ScheduleAccessibleUserView `json:"class_roster_teachers"`
	ClassRosterStudents  []*ScheduleAccessibleUserView `json:"class_roster_students"`
	ParticipantsTeachers []*ScheduleAccessibleUserView `json:"participants_teachers"`
	ParticipantsStudents []*ScheduleAccessibleUserView `json:"participants_students"`
	IsHidden             bool                          `json:"is_hidden"`
	IsHomeFun            bool                          `json:"is_home_fun"`
	RoleType             ScheduleRoleType              `json:"role_type" enums:"Student,Teacher,Unknown"`
	ExistFeedback        bool                          `json:"exist_feedback"`
	LessonPlan           *ScheduleLessonPlan           `json:"lesson_plan"`
	Class                *ScheduleAccessibleUserView   `json:"class"`
	Subjects             []*ScheduleShortInfo          `json:"subjects"`
	Program              *ScheduleShortInfo            `json:"program"`
	Teachers             []*ScheduleAccessibleUserView `json:"teachers"`
	ExistAssessment      bool                          `json:"exist_assessment"`
	CompleteAssessment   bool                          `json:"complete_assessment"`
	OutcomeIDs           []string                      `json:"outcome_ids"`
}

type ScheduleRoleType string

const (
	ScheduleRoleTypeStudent ScheduleRoleType = "Student"
	ScheduleRoleTypeTeacher ScheduleRoleType = "Teacher"
	ScheduleRoleTypeUnknown ScheduleRoleType = "Unknown"
)

type ScheduleAccessibleUserView struct {
	ID     string               `json:"id"`
	Name   string               `json:"name"`
	Type   ScheduleRelationType `json:"type"`
	Enable bool                 `json:"enable"`
}

type ScheduleAccessibleUserInput struct {
	ClassRosterClassID     string
	ClassRosterTeacherIDs  []string
	ClassRosterStudentIDs  []string
	ParticipantsTeacherIDs []string
	ParticipantsStudentIDs []string
}

type ScheduleBasicDataInput struct {
	ScheduleID   string
	ClassID      string
	ProgramID    string
	LessonPlanID string
	SubjectIDs   []string
	TeacherIDs   []string
	StudentIDs   []string
}

type ScheduleUserInput struct {
	ID   string               `json:"id"`
	Type ScheduleRelationType `json:"type"`
}

type ScheduleSearchView struct {
	ID      string `json:"id"`
	StartAt int64  `json:"start_at"`
	EndAt   int64  `json:"end_at"`
	DueAt   int64  `json:"due_at"`
	Title   string `json:"title"`
	ScheduleBasic
}

type ScheduleSimplified struct {
	ID           string `json:"id"`
	LessonPlanID string `json:"lesson_plan_id"`
	OrgID        string `json:"org_id"`
}

type ScheduleShortInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SchedulePlain struct {
	*Schedule
}

type ScheduleVariable struct {
	*Schedule
	RoomID   string               `json:"room_id"`
	Subjects []*ScheduleShortInfo `json:"subjects"`
}

type ScheduleInclude struct {
	Subject bool
}

type ScheduleBasic struct {
	Class          *ScheduleAccessibleUserView `json:"class"`
	Subjects       []*ScheduleShortInfo        `json:"subjects"`
	Program        *ScheduleShortInfo          `json:"program"`
	Members        []*ScheduleShortInfo        `json:"teachers"`
	MemberTeachers []*ScheduleShortInfo        `json:"member_teachers"`
	StudentCount   int                         `json:"student_count"`
	LessonPlan     *ScheduleShortInfo          `json:"lesson_plan"`
}

type ScheduleLessonPlan struct {
	ID        string                        `json:"id"`
	Name      string                        `json:"name"`
	IsAuth    bool                          `json:"is_auth"`
	Materials []*ScheduleLessonPlanMaterial `json:"materials"`
}

type ScheduleLessonPlanMaterial struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ScheduleVerify struct {
	ClassID      string
	SubjectIDs   []string
	ProgramID    string
	LessonPlanID string
	ClassType    ScheduleClassType
	IsHomeFun    bool
	OutcomeIDs   []string
}

// ScheduleEditType include delete and edit
type ScheduleEditType string

const (
	ScheduleEditOnlyCurrent   ScheduleEditType = "only_current"
	ScheduleEditWithFollowing ScheduleEditType = "with_following"
)

func (t ScheduleEditType) Valid() bool {
	switch t {
	case ScheduleEditOnlyCurrent, ScheduleEditWithFollowing:
		return true
	default:
		return false
	}
}

type SchedulePageView struct {
	Total int                   `json:"total"`
	Data  []*ScheduleSearchView `json:"data"`
}

type ScheduleSimplifiedPageView struct {
	Total int                   `json:"total"`
	Data  []*ScheduleSimplified `json:"data"`
}

type ScheduleIDsCondition struct {
	ClassID      string
	LessonPlanID string
	StartAt      int64
}

type ScheduleViewType string

const (
	ScheduleViewTypeDay      ScheduleViewType = "day"
	ScheduleViewTypeWorkweek ScheduleViewType = "work_week"
	ScheduleViewTypeWeek     ScheduleViewType = "week"
	ScheduleViewTypeMonth    ScheduleViewType = "month"
	ScheduleViewTypeYear     ScheduleViewType = "year"
	ScheduleViewTypeFullView ScheduleViewType = "full_view"
)

func (s ScheduleViewType) String() string {
	return string(s)
}

type ScheduleRealTimeView struct {
	ID               string `json:"id"`
	LessonPlanIsAuth bool   `json:"lesson_plan_is_auth"`
}

type ScheduleConflictView struct {
	ClassRosterTeachers  []ScheduleConflictUserView `json:"class_roster_teachers"`
	ClassRosterStudents  []ScheduleConflictUserView `json:"class_roster_students"`
	ParticipantsTeachers []ScheduleConflictUserView `json:"participants_teachers"`
	ParticipantsStudents []ScheduleConflictUserView `json:"participants_students"`
}
type ScheduleConflictUserView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ScheduleConflictInput struct {
	ClassRosterTeacherIDs  []string
	ClassRosterStudentIDs  []string
	ParticipantsTeacherIDs []string
	ParticipantsStudentIDs []string
	ClassID                string
	IgnoreScheduleID       string
	StartAt                int64
	EndAt                  int64
	IsRepeat               bool
	RepeatOptions          RepeatOptions
	Location               *time.Location
}

type ScheduleRelationInput struct {
	ScheduleID             string
	ClassRosterClassID     string
	ClassRosterTeacherIDs  []string
	ClassRosterStudentIDs  []string
	ParticipantsTeacherIDs []string
	ParticipantsStudentIDs []string
	SubjectIDs             []string
	LearningOutcomeIDs     []string
}

type ProcessScheduleDueAtInput struct {
	StartAt   int64
	EndAt     int64
	DueAt     int64
	ClassType ScheduleClassType
	Location  *time.Location
}
type ProcessScheduleDueAtView struct {
	StartAt int64
	EndAt   int64
	DueAt   int64
}

const ScheduleFilterInvalidValue = "-1"
const ScheduleFilterUndefinedClass = "Undefined"

type ScheduleFilterClass struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	OperatorRoleType ScheduleRoleType `json:"operator_role_type" enums:"Student,Teacher,Unknown"`
}

type ScheduleFilterSchool struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ScheduleFilterOption string

const (
	ScheduleFilterAnyTime  ScheduleFilterOption = "any_time"
	ScheduleFilterOnlyMine ScheduleFilterOption = "only_mine"
)

type ScheduleShowOption string

const (
	ScheduleShowOptionHidden  ScheduleShowOption = "hidden"
	ScheduleShowOptionVisible ScheduleShowOption = "visible"
)

func (s ScheduleShowOption) IsValid() bool {
	switch s {
	case ScheduleShowOptionHidden, ScheduleShowOptionVisible:
		return true
	default:
		return false
	}
}

type ScheduleViewDetail struct {
	ID             string               `json:"id"`
	IsRepeat       bool                 `json:"is_repeat"`
	LessonPlanID   string               `json:"lesson_plan_id"`
	ClassID        string               `json:"class_id"`
	Title          string               `json:"title"`
	Attachment     ScheduleShortInfo    `json:"attachment"`
	StartAt        int64                `json:"start_at"`
	EndAt          int64                `json:"end_at"`
	ClassType      ScheduleShortInfo    `json:"class_type"`
	ClassTypeLabel ScheduleShortInfo    `json:"class_type_label"`
	DueAt          int64                `json:"due_at"`
	Status         ScheduleStatus       `json:"status" enums:"NotStart,Started,Closed"`
	IsHidden       bool                 `json:"is_hidden"`
	IsHomeFun      bool                 `json:"is_home_fun"`
	RoleType       ScheduleRoleType     `json:"role_type" enums:"Student,Teacher,Unknown"`
	ExistFeedback  bool                 `json:"exist_feedback"`
	LessonPlan     *ScheduleLessonPlan  `json:"lesson_plan"`
	Class          *ScheduleShortInfo   `json:"class"`
	Teachers       []*ScheduleShortInfo `json:"teachers"`
	Students       []*ScheduleShortInfo `json:"students"`
	RoomID         string               `json:"room_id"`
	//LiveToken     string               `json:"live_token"`
	ExistAssessment    bool     `json:"exist_assessment"`
	CompleteAssessment bool     `json:"complete_assessment"`
	Description        string   `json:"description"`
	OutcomeIDs         []string `json:"outcome_ids"`
}

type ScheduleTeachingLoadInput struct {
	OrgID      string
	SchoolIDs  []string
	ClassIDs   []string
	TeacherIDs []string
	TimeRanges []*ScheduleTimeRange
}

type ScheduleTimeRange struct {
	StartAt int64
	EndAt   int64
}

type ScheduleTeachingLoadView struct {
	TeacherID string
	ClassType ScheduleClassType
	Durations []*ScheduleTeachingDuration
}

type ScheduleTeachingDuration struct {
	StartAt  int64
	EndAt    int64
	Duration int64
}

type ScheduleTimeViewQuery struct {
	ViewType       string   `json:"view_type"`
	TimeAt         int64    `json:"time_at"`
	TimeZoneOffset int      `json:"time_zone_offset"`
	SchoolIDs      []string `json:"school_ids"`
	TeacherIDs     []string `json:"teacher_ids"`
	ClassIDs       []string `json:"class_ids"`
	SubjectIDs     []string `json:"subject_ids"`
	ProgramIDs     []string `json:"program_ids"`
	ClassTypes     []string `json:"class_types"`
	StartAtGe      int64    `json:"start_at_ge"`
	EndAtLe        int64    `json:"end_at_le"`
	Anytime        bool     `json:"anytime"`
	OrderBy        string   `json:"order_by"`
}

type ScheduleQueryCondition struct {
	IDs                NullStrings
	OrgID              sql.NullString
	StartAtGe          sql.NullInt64
	StartAtLt          sql.NullInt64
	ClassTypes         NullStrings
	IsHomefun          sql.NullBool
	RelationSchoolIDs  NullStrings
	RelationClassIDs   NullStrings
	RelationTeacherIDs NullStrings
	RelationStudentIDs NullStrings
	RelationSubjectIDs NullStrings
	CreateAtGe         sql.NullInt64
	CreateAtLt         sql.NullInt64
	DeleteAt           sql.NullInt64
}

type ScheduleTimeView struct {
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	StartAt          int64             `json:"start_at"`
	EndAt            int64             `json:"end_at"`
	DueAt            int64             `json:"due_at"`
	ClassType        ScheduleClassType `json:"class_type" enums:"OnlineClass,OfflineClass,Homework,Task"`
	Status           ScheduleStatus    `json:"status" enums:"NotStart,Started,Closed"`
	ClassID          string            `json:"class_id"`
	IsHomeFun        bool              `json:"is_home_fun"`
	IsRepeat         bool              `json:"is_repeat"`
	LessonPlanID     string            `json:"lesson_plan_id"`
	AssessmentStatus AssessmentStatus  `json:"assessment_status"`
}

// request and response
type ScheduleTimeViewListRequest struct {
	ViewType             string   `json:"view_type"`
	TimeAt               int64    `json:"time_at"`
	TimeZoneOffset       int      `json:"time_zone_offset"`
	SchoolIDs            []string `json:"school_ids"`
	TeacherIDs           []string `json:"teacher_ids"`
	ClassIDs             []string `json:"class_ids"`
	SubjectIDs           []string `json:"subject_ids"`
	ProgramIDs           []string `json:"program_ids"`
	ClassTypes           []string `json:"class_types"`
	StartAtGe            int64    `json:"start_at_ge"`
	EndAtLe              int64    `json:"end_at_le"`
	DueAtEq              int64    `json:"due_at_eq"`
	Anytime              bool     `json:"anytime"`
	WithAssessmentStatus bool     `json:"with_assessment_status"`
	OrderBy              string   `json:"order_by"`
	PageSize             int      `json:"page_size" binding:"min=-1,max=1000"`
	Page                 int      `json:"page"`
}

type ScheduleTimeViewListResponse struct {
	Total int                 `json:"total"`
	Data  []*ScheduleTimeView `json:"data"`
}
