package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

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

type RepeatWeek string

const (
	RepeatWeekSun RepeatWeek = "sun"
	RepeatWeekMon RepeatWeek = "mon"
	RepeatWeekTue RepeatWeek = "tue"
	RepeatWeekWed RepeatWeek = "wed"
	RepeatWeekThu RepeatWeek = "thu"
	RepeatWeekFri RepeatWeek = "fri"
	RepeatWeekSat RepeatWeek = "sat"
)

func (w RepeatWeek) Valid() bool {
	switch w {
	case RepeatWeekSun, RepeatWeekMon, RepeatWeekTue, RepeatWeekWed, RepeatWeekThu, RepeatWeekFri, RepeatWeekSat:
		return true
	default:
		return false
	}
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

type RepeatOptions struct {
	ID      string        `json:"id"`
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
	Interval int        `json:"interval"`
	On       RepeatWeek `json:"on"`
	End      RepeatEnd  `json:"end"`
}

type RepeatMonthly struct {
	Interval  int                 `json:"interval"`
	OnType    RepeatMonthlyOnType `json:"on_type"`
	OnDateDay int                 `json:"on_date_day"`
	OnWeekSeq RepeatWeekSeq       `json:"on_week_seq"`
	OnWeek    RepeatWeek          `json:"on_week"`
	End       RepeatEndType       `json:"end"`
}

type RepeatYearly struct {
	Interval    int                `json:"interval"`
	OnType      RepeatYearlyOnType `json:"on_type"`
	OnDateMonth int                `json:"on_date_month"`
	OnDateDay   int                `json:"on_date_day"`
	OnWeekMonth int                `json:"on_week_month"`
	OnWeekSeq   RepeatWeekSeq      `json:"on_week_seq"`
	OnWeek      RepeatWeek         `json:"on_week"`
	End         RepeatEndType      `json:"end"`
}

type RepeatEnd struct {
	Type       RepeatEndType `json:"type"`
	Never      bool          `json:"never"`
	AfterCount int           `json:"after_count"`
	AfterTime  int64         `json:"after_time"`
}

type Schedule struct {
	ID           string   `dynamodbav:"id"`
	Title        string   `dynamodbav:"title"`
	ClassID      string   `dynamodbav:"class_id"`
	LessonPlanID string   `dynamodbav:"lesson_plan_id"`
	TeacherIDs   []string `dynamodbav:"teacher_ids"`
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
}

func (Schedule) TableName() string {
	return constant.TableNameSchedule
}
