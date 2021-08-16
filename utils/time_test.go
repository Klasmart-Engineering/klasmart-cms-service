package utils

import (
	"fmt"
	"github.com/go-playground/assert/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"reflect"
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

func TestStartOfDayByTimeStamp(t *testing.T) {
	loc := time.Now().Location()
	start := StartOfDayByTimeStamp(time.Now().Unix(), loc)
	t.Log(ConvertTime(start, loc))
}
func TestEndOfDayByTimeStamp(t *testing.T) {
	loc := time.Now().Location()
	end := EndOfDayByTimeStamp(time.Now().Unix(), loc)
	t.Log(ConvertTime(end, loc))
}
func TestGetTimeDiffToDayByTimeStamp(t *testing.T) {
	loc := time.Now().Location()
	start := time.Now()
	end := time.Now().AddDate(0, 0, -1)
	r := GetTimeDiffToDayByTimeStamp(start.Unix(), end.Unix(), loc)
	t.Log(start)
	t.Log(end)
	t.Log(r)
}

func TestTemp(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now().Add(7 * time.Hour)
	t2 := now.In(loc)
	t.Log(t2)

	end := TodayEndByTimeStamp(t2.Unix(), time.Local)
	t.Log(end)
	end2 := TodayEndByTimeStamp(now.Unix(), time.Local)
	t.Log(end2)
	t.Log(loc)
}

func TestEndOfDayByTime(t *testing.T) {
	type args struct {
		t  time.Time
		la *time.Location
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "t1",
			args: args{
				t:  time.Now(),
				la: time.FixedZone("report_teaching_load", 8),
			},
			want: time.Time{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EndOfDayByTime(tt.args.t, tt.args.la).AddDate(0, 0, 1); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EndOfDayByTime() = %v, want %v", got, tt.want)
				t.Log(got.Unix())
			} else {
				t.Log(got.Unix())
			}
		})
	}
}

func TestFindWeekTimeRangeFromMonday(t *testing.T) {
	type args struct {
		ts  int64
		loc *time.Location
	}
	tests := []struct {
		name         string
		args         args
		wantStartDay int64
		wantEndDay   int64
	}{
		{
			name: "t1",
			args: args{
				ts:  1628841045,
				loc: time.FixedZone("test", 28800),
			},
			wantStartDay: 0,
			wantEndDay:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStartDay, gotEndDay := FindWeekTimeRangeFromMonday(tt.args.ts, tt.args.loc)
			if gotStartDay != tt.wantStartDay {
				t.Errorf("FindWeekTimeRangeFromMonday() gotStartDay = %v, want %v", time.Unix(gotStartDay, 0), time.Unix(tt.wantStartDay, 0))
			}
			if gotEndDay != tt.wantEndDay {
				t.Errorf("FindWeekTimeRangeFromMonday() gotEndDay = %v, want %v", time.Unix(gotEndDay, 0), time.Unix(tt.wantEndDay, 0))
			}
		})
	}
}
