package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestDemo(t *testing.T) {

	time2 := BeginOfDay(time.Now())
	fmt.Println(time2.Unix())

	fmt.Println(time.Unix(time2.Unix(), 0).Format(Second))
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
		timeUtil := TimeUtil{TimeStamp: item.targetTime}
		start, end := timeUtil.FindWeekTimeRangeFormat(Second)
		fmt.Println("start:", start, "end:", end)
	}
}
