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

var (
	weekDay = map[string]int{
		"Sunday":    0,
		"Monday":    1,
		"Tuesday":   2,
		"Wednesday": 3,
		"Thursday":  4,
		"Friday":    5,
		"Saturday":  6,
	}
)

type TimeUtil struct {
	TimeStamp int64
}

func (t TimeUtil) FindWeekTimeRange() (startDay, endDay int64) {
	targetTime := time.Unix(t.TimeStamp, 0)
	targetBegin := BeginOfDay(targetTime)

	end := (time.Saturday - targetTime.Weekday()).String()
	plus := int64(86400 * weekDay[end])
	owe := int64(86400 * (7 - 1 - weekDay[end]))
	endDay = targetBegin.Unix() + plus + 86399
	startDay = targetBegin.Unix() - owe

	return startDay, endDay
}

func (t TimeUtil) FindWeekTimeRangeFormat(layout string) (start, end string) {
	s, e := t.FindWeekTimeRange()
	start = time.Unix(s, 0).Format(layout)
	end = time.Unix(e, 0).Format(layout)
	return
}

func (t TimeUtil) String(layout string) string {
	return time.Unix(t.TimeStamp, 0).Format(layout)
}

func BeginOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}
