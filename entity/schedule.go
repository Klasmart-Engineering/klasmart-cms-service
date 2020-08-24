package entity

import (
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"time"
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
	RepeatWeekdaySun RepeatWeekday = "sun"
	RepeatWeekdayMon RepeatWeekday = "mon"
	RepeatWeekdayTue RepeatWeekday = "tue"
	RepeatWeekdayWed RepeatWeekday = "wed"
	RepeatWeekdayThu RepeatWeekday = "thu"
	RepeatWeekdayFri RepeatWeekday = "fri"
	RepeatWeekdaySat RepeatWeekday = "sat"
)

func (w RepeatWeekday) Valid() bool {
	switch w {
	case RepeatWeekdaySun, RepeatWeekdayMon, RepeatWeekdayTue, RepeatWeekdayWed, RepeatWeekdayThu, RepeatWeekdayFri, RepeatWeekdaySat:
		return true
	default:
		return false
	}
}

func (w RepeatWeekday) TimeWeekday() time.Weekday {
	switch w {
	case RepeatWeekdaySun:
		return time.Sunday
	case RepeatWeekdayMon:
		return time.Monday
	case RepeatWeekdayTue:
		return time.Tuesday
	case RepeatWeekdayWed:
		return time.Wednesday
	case RepeatWeekdayThu:
		return time.Thursday
	case RepeatWeekdayFri:
		return time.Friday
	case RepeatWeekdaySat:
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
	Type    RepeatType    `json:"type"`
	Daily   RepeatDaily   `json:"daily"`
	Weekly  RepeatWeekly  `json:"weekly"`
	Monthly RepeatMonthly `json:"monthly"`
	Yearly  RepeatYearly  `json:"yearly"`
}

type RepeatDaily struct {
	Interval int       `json:"interval"`
	End      RepeatEnd `json:"end"`
}

type RepeatWeekly struct {
	Interval int             `json:"interval"`
	On       []RepeatWeekday `json:"on"`
	End      RepeatEnd       `json:"end"`
}

type RepeatMonthly struct {
	Interval  int                 `json:"interval"`
	OnType    RepeatMonthlyOnType `json:"on_type"`
	OnDateDay int                 `json:"on_date_day"`
	OnWeekSeq RepeatWeekSeq       `json:"on_week_seq"`
	OnWeek    RepeatWeekday       `json:"on_week"`
	End       RepeatEnd           `json:"end"`
}

type RepeatYearly struct {
	Interval    int                `json:"interval"`
	OnType      RepeatYearlyOnType `json:"on_type"`
	OnDateMonth int                `json:"on_date_month"`
	OnDateDay   int                `json:"on_date_day"`
	OnWeekMonth int                `json:"on_week_month"`
	OnWeekSeq   RepeatWeekSeq      `json:"on_week_seq"`
	OnWeek      RepeatWeekday      `json:"on_week"`
	End         RepeatEnd          `json:"end"`
}

type RepeatEnd struct {
	Type       RepeatEndType `json:"type"`
	AfterCount int           `json:"after_count"`
	AfterTime  int64         `json:"after_time"`
}

type Schedule struct {
	ID           string        `gorm:"column:id;PRIMARY_KEY" dynamodbav:"id"`
	Title        string        `gorm:"column:title;type:varchar(100)" dynamodbav:"title"`
	ClassID      string        `gorm:"column:class_id;type:varchar(100)" dynamodbav:"class_id"`
	LessonPlanID string        `gorm:"column:lesson_plan_id;type:varchar(100)" dynamodbav:"lesson_plan_id"`
	TeacherIDs   []string      `dynamodbav:"teacher_ids"`
	OrgID        string        `gorm:"column:org_id;type:varchar(100)" dynamodbav:"org_id"`
	StartAt      int64         `gorm:"column:start_at;type:bigint" dynamodbav:"start_at"`
	EndAt        int64         `gorm:"column:end_at;type:bigint" dynamodbav:"end_at"`
	ModeType     string        `gorm:"column:mode_type;type:varchar(100)" dynamodbav:"mode_type"`
	SubjectID    string        `gorm:"column:subject_id;type:varchar(100)" dynamodbav:"subject_id"`
	ProgramID    string        `gorm:"column:program_id;type:varchar(100)" dynamodbav:"program_id"`
	ClassType    string        `gorm:"column:class_type;type:varchar(100)" dynamodbav:"class_type"`
	DueAt        int64         `gorm:"column:due_at;type:bigint" dynamodbav:"due_at"`
	Description  string        `gorm:"column:description;type:varchar(500)" dynamodbav:"description"`
	Attachment   string        `gorm:"column:attachment_id;type:varchar(500)" dynamodbav:"attachment_id"`
	Version      int64         `gorm:"column:version;type:bigint" dynamodbav:"version"`
	RepeatID     string        `gorm:"column:repeat_id;type:varchar(100)" dynamodbav:"repeat_id"`
	Repeat       RepeatOptions `dynamodbav:"repeat"`
	RepeatJson   string        `gorm:"column:repeat;type:varchar(500)" dynamodbav:"repeat"`
	CreatedID    string        `gorm:"column:created_id;type:varchar(100)" dynamodbav:"created_id"`
	UpdatedID    string        `gorm:"column:updated_id;type:varchar(100)" dynamodbav:"updated_id"`
	DeletedID    string        `gorm:"column:position;type:varchar(100)"`
	CreatedAt    int64         `gorm:"column:created_at;type:bigint" dynamodbav:"created_at"`
	UpdatedAt    int64         `gorm:"column:updated_at;type:bigint" dynamodbav:"updated_at"`
	DeletedAt    int64         `gorm:"column:deleted_at;type:bigint"`
}

func (Schedule) TableName() string {
	return constant.TableNameSchedule
}

func (Schedule) IndexNameRepeatIDAndStartAt() string {
	return "repeat_id_and_start_at"
}

func (s Schedule) Clone() Schedule {
	newItem := s
	newItem.TeacherIDs = append([]string{}, s.TeacherIDs...)
	return newItem
}

const (
	ModeTypeAllDay = "AllDay"
	ModeTypeRepeat = "Repeat"
	ModeTypeNone   = "None"
)

type ScheduleAddView struct {
	Title        string        `json:"title" validate:"required"`
	ClassID      string        `json:"class_id"`
	LessonPlanID string        `json:"lesson_plan_id"`
	TeacherIDs   []string      `json:"teacher_ids"`
	OrgID        string        `json:"org_id" validate:"required"`
	StartAt      int64         `json:"start_at" validate:"required"`
	EndAt        int64         `json:"end_at" validate:"required"`
	ModeType     string        `json:"mode_type"`
	SubjectID    string        `json:"subject_id"`
	ProgramID    string        `json:"program_id"`
	ClassType    string        `json:"class_type"`
	DueAt        int64         `json:"due_at"`
	Description  string        `json:"description"`
	AttachmentID string        `json:"attachment_id"`
	Version      int64         `json:"version"`
	RepeatID     string        `json:"repeat_id"`
	Repeat       RepeatOptions `json:"repeat"`
}

func (s *ScheduleAddView) Convert() *Schedule {
	repeatJson, _ := json.Marshal(s.Repeat)

	schedule := &Schedule{
		Title:        s.Title,
		ClassID:      s.ClassID,
		LessonPlanID: s.LessonPlanID,
		TeacherIDs:   s.TeacherIDs,
		OrgID:        s.OrgID,
		StartAt:      s.StartAt,
		EndAt:        s.EndAt,
		ModeType:     s.ModeType,
		SubjectID:    s.SubjectID,
		ProgramID:    s.ProgramID,
		ClassType:    s.ClassType,
		DueAt:        s.DueAt,
		Description:  s.Description,
		Attachment:   s.AttachmentID,
		Version:      0,
		Repeat:       s.Repeat,
		RepeatJson:   string(repeatJson),
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    0,
		RepeatID:     s.RepeatID,
	}
	if schedule.RepeatID == "" {
		schedule.RepeatID = utils.NewID()
	}
	return schedule
}

type ScheduleUpdateView struct {
	ID       string           `json:"id"`
	EditType ScheduleEditType `json:"edit_type"`

	ScheduleAddView
}

type ScheduleListView struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	StartAt int64  `json:"start_at"`
	EndAt   int64  `json:"end_at"`
}

type ScheduleDetailsView struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	//Class       ShortInfo     `json:"class"`
	//LessonPlan  ShortInfo     `json:"lesson_plan"`
	//Teachers    []ShortInfo   `json:"teachers"`
	//Subject     ShortInfo     `json:"subject"`
	//Program     ShortInfo     `json:"program"`
	//Attachment  ShortInfo     `json:"attachment"`
	OrgID       string        `json:"org_id"`
	StartAt     int64         `json:"start_at"`
	EndAt       int64         `json:"end_at"`
	ModeType    string        `json:"mode_type"`
	ClassType   string        `json:"class_type"`
	DueAt       int64         `json:"due_at"`
	Description string        `json:"description"`
	Version     int64         `json:"version"`
	RepeatID    string        `json:"repeat_id"`
	Repeat      RepeatOptions `json:"repeat"`
	ScheduleBasic
}

type ScheduleSeachView struct {
	ID      string `json:"id"`
	StartAt int64  `json:"start_at"`
	EndAt   int64  `json:"end_at"`
	ScheduleBasic
	//LessonPlan ShortInfo   `json:"lesson_plan"`
	//Class      ShortInfo   `json:"class"`
	//Subject    ShortInfo   `json:"subject"`
	//Program    ShortInfo   `json:"program"`
	//Teachers   []ShortInfo `json:"teachers"`
}
type ShortInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ScheduleBasic struct {
	Class      ShortInfo   `json:"class"`
	Subject    ShortInfo   `json:"subject"`
	Program    ShortInfo   `json:"program"`
	Teachers   []ShortInfo `json:"teachers"`
	LessonPlan ShortInfo   `json:"lesson_plan"`
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
