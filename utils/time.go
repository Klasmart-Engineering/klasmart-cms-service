package utils

import (
	"errors"
	"strings"
	"time"
)

type TimeLayout string

const (
	Year   TimeLayout = "2006"
	Month  TimeLayout = "2006-01"
	Day    TimeLayout = "2006-01-02"
	Hour   TimeLayout = "2006-01-02 15"
	Minute TimeLayout = "2006-01-02 15:04"
	Second TimeLayout = "2006-01-02 15:04:05"
)

type TimeUtil struct {
	TimeStamp int64
	Location  *time.Location
}

func NewTimeUtil(timeStamp int64, la *time.Location) *TimeUtil {
	return &TimeUtil{
		TimeStamp: timeStamp,
		Location:  la,
	}
}

func (t *TimeUtil) FindWeekTimeRange() (startDay, endDay int64) {
	tt := time.Unix(t.TimeStamp, 0).In(t.Location)
	offset := int(time.Sunday - tt.Weekday())
	lastoffset := int(time.Saturday - tt.Weekday())

	firstOfWeek := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, t.Location).AddDate(0, 0, offset)
	lastOfWeeK := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, t.Location).AddDate(0, 0, lastoffset)
	startDay = BeginOfDayByTime(firstOfWeek, t.Location).Unix()
	endDay = EndOfDayByTime(lastOfWeeK, t.Location).Unix()
	return
}
func (t *TimeUtil) FindWeekTimeRangeFormat(layout string) (start, end string) {
	s, e := t.FindWeekTimeRange()
	start = time.Unix(s, 0).Format(layout)
	end = time.Unix(e, 0).Format(layout)
	return
}

func (t *TimeUtil) FindWorkWeekTimeRange() (startDay, endDay int64) {
	tt := time.Unix(t.TimeStamp, 0).In(t.Location)
	offset := int(time.Monday - tt.Weekday())
	lastoffset := int(time.Friday - tt.Weekday())

	firstOfWeek := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	lastOfWeeK := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, lastoffset)
	startDay = BeginOfDayByTime(firstOfWeek, t.Location).Unix()
	endDay = EndOfDayByTime(lastOfWeeK, t.Location).Unix()
	return startDay, endDay
}

func (t *TimeUtil) FindWorkWeekTimeRangeFormat(layout string) (start, end string) {
	s, e := t.FindWorkWeekTimeRange()
	start = time.Unix(s, 0).Format(layout)
	end = time.Unix(e, 0).Format(layout)
	return
}

func (t *TimeUtil) FindMonthRange() (startDay, endDay int64) {
	tt := time.Unix(t.TimeStamp, 0).In(t.Location)
	currentYear, currentMonth, _ := tt.Date()
	currentLocation := tt.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	startDay = BeginOfDayByTime(firstOfMonth, t.Location).Unix()
	endDay = EndOfDayByTime(lastOfMonth, t.Location).Unix()
	return
}
func (t *TimeUtil) FindMonthRangeFormat(layout string) (start, end string) {
	s, e := t.FindMonthRange()
	start = time.Unix(s, 0).Format(layout)
	end = time.Unix(e, 0).Format(layout)
	return
}

func (t *TimeUtil) String(layout string) string {
	return time.Unix(t.TimeStamp, 0).In(t.Location).Format(layout)
}

func BeginOfDayByTime(t time.Time, la *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, la)
}
func (t *TimeUtil) BeginOfDayByTimeStamp() time.Time {
	t2 := time.Unix(t.TimeStamp, 0).In(t.Location)
	return time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t.Location)
}

func EndOfDayByTime(t time.Time, la *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, la)
}
func (t *TimeUtil) EndOfDayByTimeStamp() time.Time {
	t2 := time.Unix(t.TimeStamp, 0).In(t.Location)
	return time.Date(t2.Year(), t2.Month(), t2.Day(), 23, 59, 59, 999999999, t.Location)
}

func TodayZero(now time.Time) time.Time {
	_, offset := now.Zone()
	duration := time.Second * time.Duration(offset)
	return now.Add(duration).Truncate(time.Hour * 24).Add(-duration)
}

// offset:Second
func GetTimeLocationByOffset(offset int) *time.Location {
	return time.FixedZone("UTC", offset)
}

func GetTimeLocationByName(tz string) (*time.Location, error) {
	if strings.TrimSpace(tz) == "" {
		return nil, errors.New("time_zone is require")
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, err
	}
	return loc, nil
}

// get time by weekday
func (tu *TimeUtil) GetTimeByWeekday(weekday time.Weekday) time.Time {
	t := time.Unix(tu.TimeStamp, 0).In(tu.Location)
	offset := int(weekday - t.Weekday())
	if weekday == time.Sunday {
		offset = offset + 7
	}
	newTime := t.AddDate(0, 0, offset)
	return newTime
}

func (tu *TimeUtil) NextDayStart(t time.Time) time.Time {
	newTime := t.AddDate(0, 0, 1)
	return time.Date(newTime.Year(), newTime.Month(), newTime.Day(), 0, 0, 0, 0, tu.Location)
}

func IsSameDay(t1, t2 int64, loc *time.Location) bool {
	time1 := time.Unix(t1, 0).In(loc)
	time2 := time.Unix(t2, 0).In(loc)

	return IsSameDayByTime(time1, time2)
}

func IsSameDayByTime(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()

	return y1 == y2 && m1 == m2 && d1 == d2
}

func IsSameMonth(t1, t2 int64, loc *time.Location) bool {
	time1 := time.Unix(t1, 0).In(loc)
	time2 := time.Unix(t2, 0).In(loc)

	return IsSameMonthByTime(time1, time2)
}

func IsSameMonthByTime(t1, t2 time.Time) bool {
	y1, m1, _ := t1.Date()
	y2, m2, _ := t2.Date()

	return y1 == y2 && m1 == m2
}

//func GetTimeDiffToDay(start, end int64) int64 {
//	return (end - start) / 86400
//}
func GetTimeDiffToDayByTime(start, end time.Time, loc *time.Location) int64 {
	t1 := BeginOfDayByTime(start, loc).Unix()
	t2 := BeginOfDayByTime(end, loc).Unix()
	return ((t2 - t1) / 86400) - 1
}

func SetTimeDatePart(src time.Time, year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, src.Hour(), src.Minute(), src.Second(), src.Nanosecond(), src.Location())
}

func ConvertTime(ts int64, loc *time.Location) time.Time {
	return time.Unix(ts, 0).In(loc)
}

func StartOfYear(year int, loc *time.Location) time.Time {
	return time.Date(year, 1, 1, 0, 0, 0, 0, loc)
}
func StartOfYearByTimeStamp(ts int64, loc *time.Location) time.Time {
	t := ConvertTime(ts, loc)
	return StartOfYear(t.Year(), loc)
}

func StartOfMonth(year int, month time.Month, loc *time.Location) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, loc)
}

func EndOfMonth(year int, month time.Month, loc *time.Location) time.Time {
	return time.Date(year, month+1, 1, 0, 0, 0, 0, loc).Add(-time.Millisecond)
}

func GetTimeByWeekday(ts int64, weekday time.Weekday, loc *time.Location) time.Time {
	t := time.Unix(ts, 0).In(loc)
	offset := int(weekday - t.Weekday())
	if weekday == time.Sunday {
		offset = offset + 7
	}
	newTime := t.AddDate(0, 0, offset)
	return newTime
}
