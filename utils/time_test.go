package utils

import (
	"fmt"
	"github.com/go-playground/assert/v2"
	"testing"
	"time"
)

func TestTodayZero(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t2 := TodayZero(time.Now().In(loc))
	fmt.Println(t2)
}

func TestTimeUtil_GetTimeByWeekday(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	tu := NewTimeUtil(time.Now().Unix(), loc)
	tt := tu.GetTimeByWeekday(time.Monday)
	t.Log(time.Now())
	t.Log(tt)
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
