package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestDemo(t *testing.T) {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	time2 := BeginOfDayByTime(time.Now(), loc)
	fmt.Println(time2.Unix())
	fmt.Println(time.Unix(time2.Unix(), 0).In(loc))
	loc = GetTimeLocationByOffset(8 * 60 * 60)
	fmt.Println(time.Now().In(loc))
	fmt.Println(time.Now().In(loc).Zone())
}

func TestTimeUtil_FindWeekTimeRange(t *testing.T) {
	// Sunday
	tm, _ := time.Parse(Day, "2020-08-16")
	// Saturday
	tm2, _ := time.Parse(Day, "2020-08-15")
	// Friday
	tm3, _ := time.Parse(Day, "2020-08-14")
	testData := []struct {
		targetTime int64
	}{
		{targetTime: time.Now().Unix()},
		{targetTime: tm.Unix()},
		{targetTime: tm2.Unix()},
		{targetTime: tm3.Unix()},
	}
	for _, item := range testData {
		timeUtil := NewTimeUtil(item.targetTime, time.Local)
		start, end := timeUtil.FindWeekTimeRangeFormat(Second)
		fmt.Println("start:", start, "end:", end)
	}
}
func TestTimeUtil_FindWorkWeekTimeRange(t *testing.T) {
	// Sunday
	tm, _ := time.Parse(Day, "2020-08-16")
	// Saturday
	tm2, _ := time.Parse(Day, "2020-08-15")
	// Friday
	tm3, _ := time.Parse(Day, "2020-08-14")
	testData := []struct {
		targetTime int64
	}{
		{targetTime: time.Now().Unix()},
		{targetTime: tm.Unix()},
		{targetTime: tm2.Unix()},
		{targetTime: tm3.Unix()},
	}
	for _, item := range testData {
		timeUtil := NewTimeUtil(item.targetTime, time.Local)
		start, end := timeUtil.FindWorkWeekTimeRangeFormat(Second)
		fmt.Println("start:", start, "end:", end)
	}
}

func TestTimeUtil_FindMonthRange(t *testing.T) {
	// Sunday
	tm, _ := time.Parse(Day, "2020-07-16")
	// Saturday
	tm2, _ := time.Parse(Day, "2020-08-15")
	// Friday
	tm3, _ := time.Parse(Day, "2020-09-14")
	testData := []struct {
		targetTime int64
	}{
		{targetTime: time.Now().Unix()},
		{targetTime: tm.Unix()},
		{targetTime: tm2.Unix()},
		{targetTime: tm3.Unix()},
	}
	for _, item := range testData {
		timeUtil := NewTimeUtil(item.targetTime, time.Local)
		start, end := timeUtil.FindMonthRangeFormat(Second)
		fmt.Println("start:", start, "end:", end)
	}
}

func TestTodayZero(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t2 := TodayZero(time.Now().In(loc))
	fmt.Println(t2)
}
