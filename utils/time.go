package utils

import (
	"errors"
	"strings"
	"time"
)

type TimeLayout string

const (
	TimeLayoutYear   TimeLayout = "2006"
	TimeLayoutMonth  TimeLayout = "2006-01"
	TimeLayoutDay    TimeLayout = "2006-01-02"
	TimeLayoutHour   TimeLayout = "2006-01-02 15"
	TimeLayoutMinute TimeLayout = "2006-01-02 15:04"
	TimeLayoutSecond TimeLayout = "2006-01-02 15:04:05"
)

func (ts TimeLayout) String() string {
	return string(ts)
}

func FindWeekTimeRange(ts int64, loc *time.Location) (startDay, endDay int64) {
	tt := time.Unix(ts, 0).In(loc)
	offset := int(time.Sunday - tt.Weekday())
	lastOffset := int(time.Saturday - tt.Weekday())

	firstOfWeek := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, offset)
	lastOfWeeK := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, lastOffset)
	startDay = BeginOfDayByTime(firstOfWeek, loc).Unix()
	endDay = EndOfDayByTime(lastOfWeeK, loc).Unix()
	return
}
func FindWorkWeekTimeRange(ts int64, loc *time.Location) (startDay, endDay int64) {
	tt := time.Unix(ts, 0).In(loc)
	offset := int(time.Monday - tt.Weekday())
	lastOffset := int(time.Friday - tt.Weekday())

	firstOfWeek := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	lastOfWeeK := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, lastOffset)
	startDay = BeginOfDayByTime(firstOfWeek, loc).Unix()
	endDay = EndOfDayByTime(lastOfWeeK, loc).Unix()
	return startDay, endDay
}

func FindMonthRange(ts int64, loc *time.Location) (startDay, endDay int64) {
	tt := time.Unix(ts, 0).In(loc)
	currentYear, currentMonth, _ := tt.Date()
	currentLocation := tt.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	startDay = BeginOfDayByTime(firstOfMonth, loc).Unix()
	endDay = EndOfDayByTime(lastOfMonth, loc).Unix()
	return
}

func BeginOfDayByTime(t time.Time, la *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, la)
}

func EndOfDayByTime(t time.Time, la *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, la)
}

func TodayZero(t time.Time) time.Time {
	_, offset := t.Zone()
	duration := time.Second * time.Duration(offset)
	return t.Add(duration).Truncate(time.Hour * 24).Add(-duration)
}
func TodayEnd(t time.Time) time.Time {
	return TodayZero(t).AddDate(0, 0, 1).Add(-1 * time.Second)
}

func TodayZeroByTimeStamp(ts int64, loc *time.Location) time.Time {
	t := time.Unix(ts, 0).In(loc)
	return TodayZero(t)
}
func TodayEndByTimeStamp(ts int64, loc *time.Location) time.Time {
	t := time.Unix(ts, 0).In(loc)
	return TodayEnd(t)
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

func GetTimeDiffToDayByTime(start, end time.Time, loc *time.Location) int64 {
	t1 := BeginOfDayByTime(start, loc).Unix()
	t2 := BeginOfDayByTime(end, loc).Unix()
	return (t2 - t1) / 86400
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

func TimeStampDiff(t1 int64, t2 int64) time.Duration {
	return time.Second * time.Duration(t1-t2)
}

func EndOfYear(year int, loc *time.Location) time.Time {
	return StartOfYear(year+1, loc).Add(-time.Millisecond)
}
func EndOfYearByTimeStamp(ts int64, loc *time.Location) time.Time {
	t := ConvertTime(ts, loc)
	return EndOfYear(t.Year(), loc)
}

func TimeStampString(ts int64, loc *time.Location, layout TimeLayout) string {
	return time.Unix(ts, 0).In(loc).Format(layout.String())
}

func DateBetweenTimeAndFormat(start int64, end int64, loc *time.Location) []string {
	if start >= end {
		return []string{}
	}
	startStr := TimeStampString(start, loc, TimeLayoutDay)
	endStr := TimeStampString(end, loc, TimeLayoutDay)
	startTime, _ := time.Parse(TimeLayoutDay.String(), startStr)
	endTime, _ := time.Parse(TimeLayoutDay.String(), endStr)

	if startTime.Equal(endTime) {
		return []string{startStr}
	}
	result := make([]string, 0)

	for temp := startTime; temp.Before(endTime) || temp.Equal(endTime); {
		result = append(result, temp.Format(TimeLayoutDay.String()))
		temp = temp.AddDate(0, 0, 1)
	}
	return result
}

func StartOfDayByTimeStamp(ts int64, loc *time.Location) int64 {
	t2 := time.Unix(ts, 0).In(loc)
	return time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, loc).Unix()
}
func EndOfDayByTimeStamp(ts int64, loc *time.Location) int64 {
	t := time.Unix(ts, 0).In(loc)
	return EndOfDayByTime(t, loc).Unix()
}
func GetTimeDiffToDayByTimeStamp(start, end int64, loc *time.Location) int64 {
	t1 := time.Unix(start, 0).In(loc)
	t2 := time.Unix(end, 0).In(loc)

	return GetTimeDiffToDayByTime(t1, t2, loc)
}
