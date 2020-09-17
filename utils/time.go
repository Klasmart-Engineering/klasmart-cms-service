package utils

import (
	"errors"
	"strings"
	"time"
)

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
