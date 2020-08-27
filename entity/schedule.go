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

type ScheduleClassType string

const (
	ScheduleClassTypeOnlineClass  ScheduleClassType = "OnlineClass"
	ScheduleClassTypeOfflineClass ScheduleClassType = "OfflineClass"
	ScheduleClassTypeHomework     ScheduleClassType = "Homework"
	ScheduleClassTypeTask         ScheduleClassType = "Task"
)

type Schedule struct {
	ID           string `gorm:"column:id;PRIMARY_KEY"`
	Title        string `gorm:"column:title;type:varchar(100)"`
	ClassID      string `gorm:"column:class_id;type:varchar(100)"`
	LessonPlanID string `gorm:"column:lesson_plan_id;type:varchar(100)"`
	OrgID        string `gorm:"column:org_id;type:varchar(100)"`
	StartAt      int64  `gorm:"column:start_at;type:bigint"`
	EndAt        int64  `gorm:"column:end_at;type:bigint"`
	Status       string `gorm:"column:status;type:varchar(100)"`
	IsAllDay     bool   `gorm:"column:is_all_day;default:false"`
	SubjectID    string `gorm:"column:subject_id;type:varchar(100)"`
	ProgramID    string `gorm:"column:program_id;type:varchar(100)"`
	ClassType    string `gorm:"column:class_type;type:varchar(100)"`
	DueAt        int64  `gorm:"column:due_at;type:bigint"`
	Description  string `gorm:"column:description;type:varchar(500)"`
	Attachment   string `gorm:"column:attachment_url;type:varchar(500)"`
	Version      int64  `gorm:"column:'version';type:bigint"`
	RepeatID     string `gorm:"column:repeat_id;type:varchar(100)"`
	RepeatJson   string `gorm:"column:repeat;type:varchar(500)"`
	CreatedID    string `gorm:"column:created_id;type:varchar(100)"`
	UpdatedID    string `gorm:"column:updated_id;type:varchar(100)"`
	DeletedID    string `gorm:"column:deleted_id;type:varchar(100)"`
	CreatedAt    int64  `gorm:"column:created_at;type:bigint"`
	UpdatedAt    int64  `gorm:"column:updated_at;type:bigint"`
	DeleteAt     int64  `gorm:"column:delete_at;type:bigint"`
}

func (Schedule) TableName() string {
	return constant.TableNameSchedule
}

func (Schedule) IndexNameRepeatIDAndStartAt() string {
	return "repeat_id_and_start_at"
}

func (s Schedule) Clone() Schedule {
	newItem := s
	return newItem
}

type ScheduleAddView struct {
	Title        string        `json:"title" binding:"required"`
	ClassID      string        `json:"class_id" binding:"required"`
	LessonPlanID string        `json:"lesson_plan_id" binding:"required"`
	TeacherIDs   []string      `json:"teacher_ids" binding:"required"`
	OrgID        string        `json:"org_id"`
	StartAt      int64         `json:"start_at" binding:"required"`
	EndAt        int64         `json:"end_at" binding:"required"`
	SubjectID    string        `json:"subject_id" binding:"required"`
	ProgramID    string        `json:"program_id" binding:"required"`
	ClassType    string        `json:"class_type"`
	DueAt        int64         `json:"due_at"`
	Description  string        `json:"description"`
	Attachment   string        `json:"attachment_path"`
	Version      int64         `json:"version"`
	RepeatID     string        `json:"-"`
	Repeat       RepeatOptions `json:"repeat"`
	IsAllDay     bool          `json:"is_all_day"`
	IsRepeat     bool          `json:"is_repeat"`
	IsForce      bool          `json:"is_force"`
}

func (s *ScheduleAddView) Convert() (*Schedule, error) {
	schedule := &Schedule{
		Title:        s.Title,
		ClassID:      s.ClassID,
		LessonPlanID: s.LessonPlanID,
		OrgID:        s.OrgID,
		StartAt:      s.StartAt,
		EndAt:        s.EndAt,
		SubjectID:    s.SubjectID,
		ProgramID:    s.ProgramID,
		ClassType:    s.ClassType,
		DueAt:        s.DueAt,
		Description:  s.Description,
		Attachment:   s.Attachment,
		Version:      0,
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    0,
		RepeatID:     s.RepeatID,
		IsAllDay:     s.IsAllDay,
	}
	if s.IsRepeat {
		b, err := json.Marshal(s.Repeat)
		if err != nil {
			return nil, err
		}
		schedule.RepeatJson = string(b)

		if schedule.RepeatID == "" {
			schedule.RepeatID = utils.NewID()
		}
	}

	return schedule, nil
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
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Attachment  string        `json:"attachment"`
	OrgID       string        `json:"org_id"`
	StartAt     int64         `json:"start_at"`
	EndAt       int64         `json:"end_at"`
	ClassType   string        `json:"class_type"`
	DueAt       int64         `json:"due_at"`
	Description string        `json:"description"`
	Version     int64         `json:"version"`
	IsAllDay    bool          `json:"is_all_day"`
	RepeatID    string        `json:"repeat_id"`
	Repeat      RepeatOptions `json:"repeat"`
	ScheduleBasic
}

type ScheduleSeachView struct {
	ID      string `json:"id"`
	StartAt int64  `json:"start_at"`
	EndAt   int64  `json:"end_at"`
	ScheduleBasic
}
type ScheduleShortInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ScheduleBasic struct {
	Class      ScheduleShortInfo   `json:"class"`
	Subject    ScheduleShortInfo   `json:"subject"`
	Program    ScheduleShortInfo   `json:"program"`
	Teachers   []ScheduleShortInfo `json:"teachers"`
	LessonPlan ScheduleShortInfo   `json:"lesson_plan"`
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
