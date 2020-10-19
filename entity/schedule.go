package entity

import (
	"context"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
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
)

type Schedule struct {
	ID              string            `gorm:"column:id;PRIMARY_KEY"`
	Title           string            `gorm:"column:title;type:varchar(100)"`
	ClassID         string            `gorm:"column:class_id;type:varchar(100)"`
	LessonPlanID    string            `gorm:"column:lesson_plan_id;type:varchar(100)"`
	OrgID           string            `gorm:"column:org_id;type:varchar(100)"`
	StartAt         int64             `gorm:"column:start_at;type:bigint"`
	EndAt           int64             `gorm:"column:end_at;type:bigint"`
	Status          ScheduleStatus    `gorm:"column:status;type:varchar(100)"`
	IsAllDay        bool              `gorm:"column:is_all_day;default:false"`
	SubjectID       string            `gorm:"column:subject_id;type:varchar(100)"`
	ProgramID       string            `gorm:"column:program_id;type:varchar(100)"`
	ClassType       ScheduleClassType `gorm:"column:class_type;type:varchar(100)"`
	DueAt           int64             `gorm:"column:due_at;type:bigint"`
	Description     string            `gorm:"column:description;type:varchar(500)"`
	Attachment      string            `gorm:"column:attachment;type:text;"`
	ScheduleVersion int64             `gorm:"column:version;type:bigint"`
	RepeatID        string            `gorm:"column:repeat_id;type:varchar(100)"`
	RepeatJson      string            `gorm:"column:repeat;type:json;"`
	CreatedID       string            `gorm:"column:created_id;type:varchar(100)"`
	UpdatedID       string            `gorm:"column:updated_id;type:varchar(100)"`
	DeletedID       string            `gorm:"column:deleted_id;type:varchar(100)"`
	CreatedAt       int64             `gorm:"column:created_at;type:bigint"`
	UpdatedAt       int64             `gorm:"column:updated_at;type:bigint"`
	DeleteAt        int64             `gorm:"column:delete_at;type:bigint"`
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
	Title          string            `json:"title" binding:"required"`
	ClassID        string            `json:"class_id" binding:"required"`
	LessonPlanID   string            `json:"lesson_plan_id" binding:"required"`
	TeacherIDs     []string          `json:"teacher_ids" binding:"required,min=1"`
	OrgID          string            `json:"org_id"`
	StartAt        int64             `json:"start_at" binding:"required"`
	EndAt          int64             `json:"end_at" binding:"required"`
	SubjectID      string            `json:"subject_id" binding:"required"`
	ProgramID      string            `json:"program_id" binding:"required"`
	ClassType      ScheduleClassType `json:"class_type" enums:"OnlineClass,OfflineClass,Homework,Task"`
	DueAt          int64             `json:"due_at"`
	Description    string            `json:"description"`
	Attachment     ScheduleShortInfo `json:"attachment"`
	Version        int64             `json:"version"`
	RepeatID       string            `json:"-"`
	Repeat         RepeatOptions     `json:"repeat"`
	IsAllDay       bool              `json:"is_all_day"`
	IsRepeat       bool              `json:"is_repeat"`
	IsForce        bool              `json:"is_force"`
	TimeZoneOffset int               `json:"time_zone_offset"`
	Location       *time.Location    `json:"-"`
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
		SubjectID:       s.SubjectID,
		ProgramID:       s.ProgramID,
		ClassType:       s.ClassType,
		DueAt:           s.DueAt,
		Status:          ScheduleStatusNotStart,
		Description:     s.Description,
		ScheduleVersion: 0,
		CreatedAt:       time.Now().Unix(),
		UpdatedAt:       time.Now().Unix(),
		IsAllDay:        s.IsAllDay,
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
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	StartAt      int64             `json:"start_at"`
	EndAt        int64             `json:"end_at"`
	IsRepeat     bool              `json:"is_repeat"`
	LessonPlanID string            `json:"lesson_plan_id"`
	ClassType    ScheduleClassType `json:"class_type" enums:"OnlineClass,OfflineClass,Homework,Task"`
	Status       ScheduleStatus    `json:"status" enums:"NotStart,Started,Closed"`
}

type ScheduleDetailsView struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Attachment  ScheduleShortInfo `json:"attachment"`
	OrgID       string            `json:"org_id"`
	StartAt     int64             `json:"start_at"`
	EndAt       int64             `json:"end_at"`
	ClassType   ScheduleClassType `json:"class_type" enums:"OnlineClass,OfflineClass,Homework,Task"`
	DueAt       int64             `json:"due_at"`
	Description string            `json:"description"`
	Version     int64             `json:"version"`
	IsAllDay    bool              `json:"is_all_day"`
	IsRepeat    bool              `json:"is_repeat"`
	Repeat      RepeatOptions     `json:"repeat"`
	Status      ScheduleStatus    `json:"status" enums:"NotStart,Started,Closed"`
	ScheduleBasic
}

type ScheduleSearchView struct {
	ID      string `json:"id"`
	StartAt int64  `json:"start_at"`
	EndAt   int64  `json:"end_at"`
	Title   string `json:"title"`
	ScheduleBasic
}
type ScheduleShortInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type SchedulePlain struct {
	Schedule
	TeacherIDs []string
}

type ScheduleBasic struct {
	Class      ScheduleShortInfo   `json:"class"`
	Subject    ScheduleShortInfo   `json:"subject"`
	Program    ScheduleShortInfo   `json:"program"`
	Teachers   []ScheduleShortInfo `json:"teachers"`
	LessonPlan ScheduleShortInfo   `json:"lesson_plan"`
}
type ScheduleVerify struct {
	ClassID      string
	SubjectID    string
	ProgramID    string
	TeacherIDs   []string
	LessonPlanID string
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
