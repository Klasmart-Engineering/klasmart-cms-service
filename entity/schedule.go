package entity

import (
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
	ID           string   `dynamodbav:"id"`
	Title        string   `dynamodbav:"title"`
	ClassID      string   `dynamodbav:"class_id"`
	LessonPlanID string   `dynamodbav:"lesson_plan_id"`
	TeacherIDs   []string `dynamodbav:"teacher_ids"`
	OrgID        string   `dynamodbav:"org_id"`
	StartAt      int64    `dynamodbav:"start_at"`
	EndAt        int64    `dynamodbav:"end_at"`
	ModeType     string   `dynamodbav:"mode_type"`
	SubjectID    string   `dynamodbav:"subject_id"`
	ProgramID    string   `dynamodbav:"program_id"`
	ClassType    string   `dynamodbav:"class_type"`
	DueAt        int64    `dynamodbav:"due_at"`
	Description  string   `dynamodbav:"description"`
	AttachmentID string   `dynamodbav:"attachment_id"`
	Version      int64    `dynamodbav:"version"`

	RepeatID string        `dynamodbav:"repeat_id"`
	Repeat   RepeatOptions `dynamodbav:"repeat"`

	CreatedID string `dynamodbav:"created_id"`
	UpdatedID string `dynamodbav:"updated_id"`
	//DeletedID string `dynamodbav:"deleted_id"`

	CreatedAt int64 `dynamodbav:"created_at"`
	UpdatedAt int64 `dynamodbav:"updated_at"`
	//DeletedAt int64 `dynamodbav:"deleted_at"`
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
)

type ScheduleAddView struct {
	Title        string        `json:"title"`
	ClassID      string        `json:"class_id"`
	LessonPlanID string        `json:"lesson_plan_id"`
	TeacherIDs   []string      `json:"teacher_ids"`
	OrgID        string        `json:"org_id"`
	StartAt      int64         `json:"start_at"`
	EndAt        int64         `json:"end_at"`
	ModeType     string        `json:"mode_type"`
	SubjectID    string        `json:"subject_id"`
	ProgramID    string        `json:"program_id"`
	ClassType    string        `json:"class_type"`
	DueAt        int64         `json:"due_at"`
	Description  string        `json:"description"`
	AttachmentID string        `json:"attachment_id"`
	Version      int64         `json:"version"`
	Repeat       RepeatOptions `json:"repeat"`
}

func (s *ScheduleAddView) Convert() *Schedule {
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
		AttachmentID: s.AttachmentID,
		Version:      0,
		Repeat:       s.Repeat,
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    0,
		RepeatID:     utils.NewID(),
	}
	return schedule
}

type ScheduleUpdateView struct {
	ID string `json:"id"`
	ScheduleAddView
}

type ScheduleListView struct {
}

type ScheduleDetailsView struct {
}

// ScheduleEditType include delete and edit
type ScheduleEditType string

const (
	ScheduleEditOnlyCurrent   ScheduleEditType = "only_current"
	ScheduleEditWithFollowing ScheduleEditType = "with_following"
)
