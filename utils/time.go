package utils

import "time"

const (
	Year   string = "2006"
	Month  string = "2006-01"
	Day    string = "2006-01-02"
	Hour   string = "2006-01-02 15"
	Minute string = "2006-01-02 15:04"
	Second string = "2006-01-02 15:04:05"
)

type TimeUtil struct {
	TimeStamp int64
}

func (t TimeUtil) FindWeekTimeRange() (startDay, endDay int64) {
	tt := time.Unix(t.TimeStamp, 0)
	offset := int(time.Sunday - tt.Weekday())
	lastoffset := int(time.Saturday - tt.Weekday())

	firstOfWeek := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	lastOfWeeK := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, lastoffset)
	startDay = BeginOfDayByTime(firstOfWeek).Unix()
	endDay = EndOfDayByTime(lastOfWeeK).Unix()
	return
}
func (t TimeUtil) FindWeekTimeRangeFormat(layout string) (start, end string) {
	s, e := t.FindWorkWeekTimeRange()
	start = time.Unix(s, 0).Format(layout)
	end = time.Unix(e, 0).Format(layout)
	return
}

func (t TimeUtil) FindWorkWeekTimeRange() (startDay, endDay int64) {
	tt := time.Unix(t.TimeStamp, 0)
	offset := int(time.Monday - tt.Weekday())
	lastoffset := int(time.Friday - tt.Weekday())

	firstOfWeek := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	lastOfWeeK := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, lastoffset)
	startDay = BeginOfDayByTime(firstOfWeek).Unix()
	endDay = EndOfDayByTime(lastOfWeeK).Unix()
	return startDay, endDay
}

func (t TimeUtil) FindWorkWeekTimeRangeFormat(layout string) (start, end string) {
	s, e := t.FindWorkWeekTimeRange()
	start = time.Unix(s, 0).Format(layout)
	end = time.Unix(e, 0).Format(layout)
	return
}

func (t TimeUtil) FindMonthRange() (startDay, endDay int64) {
	tt := time.Unix(t.TimeStamp, 0)
	currentYear, currentMonth, _ := tt.Date()
	currentLocation := tt.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	startDay = BeginOfDayByTime(firstOfMonth).Unix()
	endDay = EndOfDayByTime(lastOfMonth).Unix()
	return
}
func (t TimeUtil) FindMonthRangeFormat(layout string) (start, end string) {
	s, e := t.FindMonthRange()
	start = time.Unix(s, 0).Format(layout)
	end = time.Unix(e, 0).Format(layout)
	return
}

func (t TimeUtil) String(layout string) string {
	return time.Unix(t.TimeStamp, 0).Format(layout)
}

func BeginOfDayByTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
func BeginOfDayByTimeStamp(timeStamp int64) time.Time {
	t := time.Unix(timeStamp, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfDayByTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}
func EndOfDayByTimeStamp(timeStamp int64) time.Time {
	t := time.Unix(timeStamp, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}
