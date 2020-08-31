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

func (t TimeUtil) FindWeekTimeRange(la *time.Location) (startDay, endDay int64) {
	tt := time.Unix(t.TimeStamp, 0).In(la)
	offset := int(time.Sunday - tt.Weekday())
	lastoffset := int(time.Saturday - tt.Weekday())

	firstOfWeek := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, la).AddDate(0, 0, offset)
	lastOfWeeK := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, la).AddDate(0, 0, lastoffset)
	startDay = BeginOfDayByTime(firstOfWeek, la).Unix()
	endDay = EndOfDayByTime(lastOfWeeK, la).Unix()
	return
}
func (t TimeUtil) FindWeekTimeRangeFormat(la *time.Location, layout string) (start, end string) {
	s, e := t.FindWorkWeekTimeRange(la)
	start = time.Unix(s, 0).Format(layout)
	end = time.Unix(e, 0).Format(layout)
	return
}

func (t TimeUtil) FindWorkWeekTimeRange(la *time.Location) (startDay, endDay int64) {
	tt := time.Unix(t.TimeStamp, 0).In(la)
	offset := int(time.Monday - tt.Weekday())
	lastoffset := int(time.Friday - tt.Weekday())

	firstOfWeek := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	lastOfWeeK := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, lastoffset)
	startDay = BeginOfDayByTime(firstOfWeek, la).Unix()
	endDay = EndOfDayByTime(lastOfWeeK, la).Unix()
	return startDay, endDay
}

func (t TimeUtil) FindWorkWeekTimeRangeFormat(la *time.Location, layout string) (start, end string) {
	s, e := t.FindWorkWeekTimeRange(la)
	start = time.Unix(s, 0).Format(layout)
	end = time.Unix(e, 0).Format(layout)
	return
}

func (t TimeUtil) FindMonthRange(la *time.Location) (startDay, endDay int64) {
	tt := time.Unix(t.TimeStamp, 0).In(la)
	currentYear, currentMonth, _ := tt.Date()
	currentLocation := tt.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	startDay = BeginOfDayByTime(firstOfMonth, la).Unix()
	endDay = EndOfDayByTime(lastOfMonth, la).Unix()
	return
}
func (t TimeUtil) FindMonthRangeFormat(la *time.Location, layout string) (start, end string) {
	s, e := t.FindMonthRange(la)
	start = time.Unix(s, 0).Format(layout)
	end = time.Unix(e, 0).Format(layout)
	return
}

func (t TimeUtil) String(layout string, la *time.Location) string {
	return time.Unix(t.TimeStamp, 0).In(la).Format(layout)
}

func BeginOfDayByTime(t time.Time, la *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, la)
}
func BeginOfDayByTimeStamp(timeStamp int64, la *time.Location) time.Time {
	t := time.Unix(timeStamp, 0).In(la)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, la)
}

func EndOfDayByTime(t time.Time, la *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, la)
}
func EndOfDayByTimeStamp(timeStamp int64, la *time.Location) time.Time {
	t := time.Unix(timeStamp, 0).In(la)
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, la)
}

func TodayZero(now time.Time) time.Time {
	_, offset := now.Zone()
	duration := time.Second * time.Duration(offset)
	return now.Add(duration).Truncate(time.Hour * 24).Add(-duration)
}
