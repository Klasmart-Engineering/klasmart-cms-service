package utils

import (
	"fmt"
	"github.com/go-playground/assert/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"testing"
	"time"
)

func TestTodayZero(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t2 := TodayZero(time.Now().In(loc))
	fmt.Println(t2)
}

func TestGetTimeDiffToDay(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	base := time.Date(2020, 12, 8, 12, 30, 0, 0, loc)
	t.Log(base)
	tests := []struct {
		t1 time.Time
		t2 time.Time
		r  int
	}{
		{base, base.Add(1 * time.Hour), 0},
		{base, base.AddDate(0, 0, -5), -5},
		{base, base.AddDate(0, 0, 5), 5},
		{base, base.AddDate(0, 1, 5), 36},
	}
	for _, item := range tests {
		result := GetTimeDiffToDayByTime(item.t1, item.t2, loc)
		assert.Equal(t, result, int64(item.r))
	}
}

func TestCheckedDiffToMinuteByTimeStamp(t *testing.T) {
	loc := time.Now().Location()
	s := time.Now().Unix()
	e := time.Now().Add(20 * time.Minute).In(loc).Unix()
	result := TimeStampDiff(e, s)

	t.Log(result > constant.ScheduleAllowEditTime)
}

func TestDateBetweenTimeAndFormat(t *testing.T) {
	loc := time.Now().Location()
	testData := []struct {
		start int64
		end   int64
	}{
		{
			start: time.Now().Unix(),
			end:   time.Now().Add(20 * time.Minute).Unix(),
		},
		{
			start: time.Now().Unix(),
			end:   time.Now().AddDate(0, 0, 1).Add(-20 * time.Minute).Unix(),
		},
		{
			start: time.Now().Unix(),
			end:   time.Now().AddDate(0, 0, 3).Add(-20 * time.Minute).Unix(),
		},
	}
	for _, item := range testData {
		result := DateBetweenTimeAndFormat(item.start, item.end, loc)
		t.Log(result)
	}
}

func TestEndOfYearByTimeStamp(t *testing.T) {
	end := EndOfYearByTimeStamp(time.Now().Unix(), time.Now().Location())
	t.Log(end)
	start := StartOfYearByTimeStamp(time.Now().Unix(), time.Now().Location())
	t.Log(start)
}
