package model

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"golang.org/x/net/context"
	"testing"
	"time"
)

func TestDynamicDayInterval_RepeatEndAfterCount(t *testing.T) {
	options := entity.RepeatOptions{
		Type: entity.RepeatTypeDaily,
		Daily: entity.RepeatDaily{
			Interval: 1,
			End: entity.RepeatEnd{
				Type:       entity.RepeatEndAfterCount,
				AfterCount: 3,
				AfterTime:  0,
			},
		},
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	testTime := time.Now() //time.Date(2020, 12, 8, 12, 30, 0, 0, loc)

	tests := []struct {
		baseTime time.Time
		options  *entity.RepeatOptions
		enable   bool
	}{
		// RepeatEndAfterCount
		{
			options: func() *entity.RepeatOptions {
				temp := options
				return &temp
			}(),
			baseTime: testTime.Add(1 * time.Hour).In(loc),
			enable:   false,
		},
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Daily.Interval = 3
				temp.Daily.End.AfterCount = 5
				return &temp
			}(),
			baseTime: testTime.Add(-1 * time.Hour).In(loc),
			enable:   false,
		},
		// RepeatEndAfterTime
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Daily.Interval = 3
				temp.Daily.End.AfterCount = 0
				temp.Daily.End.Type = entity.RepeatEndAfterTime
				temp.Daily.End.AfterTime = testTime.AddDate(0, 1, 0).Unix()
				return &temp
			}(),
			baseTime: testTime.AddDate(0, 0, -2).Add(-1 * time.Hour).In(loc),
			enable:   true,
		},
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Daily.Interval = 5
				temp.Daily.End.AfterCount = 0
				temp.Daily.End.Type = entity.RepeatEndAfterTime
				temp.Daily.End.AfterTime = testTime.AddDate(0, 2, 0).Unix()
				return &temp
			}(),
			baseTime: testTime.AddDate(0, 0, 2).Add(1 * time.Hour).In(loc),
			enable:   true,
		},
	}
	for _, item := range tests {
		if !item.enable {
			continue
		}
		after := utils.ConvertTime(item.options.Daily.End.AfterTime, loc)
		t.Log("base time:", item.baseTime, "-----end after time:", after)
		conf := NewRepeatConfig(item.options, loc)
		plan, _ := NewRepeatCyclePlan(context.Background(), item.baseTime.Unix(), conf)
		rule, err := NewEndRepeatCycleRule(item.options)
		if err != nil {
			t.Fatal(err)
			return
		}

		result, _ := plan[0].GenerateTimeByRule(rule)
		for _, item := range result {
			temp := utils.ConvertTime(item, loc)
			t.Log(temp)
		}
	}
}

func TestDynamicWeekInterval(t *testing.T) {
	options := entity.RepeatOptions{
		Type: entity.RepeatTypeWeekly,
		Weekly: entity.RepeatWeekly{
			Interval: 1,
			On: []entity.RepeatWeekday{
				entity.RepeatWeekdayMonday,
				entity.RepeatWeekdayTuesday,
			},
			End: entity.RepeatEnd{
				Type:       entity.RepeatEndAfterCount,
				AfterCount: 3,
				AfterTime:  0,
			},
		},
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	testTime := time.Now() //time.Date(2020, 12, 8, 12, 30, 0, 0, loc)
	tests := []struct {
		baseTime time.Time
		options  *entity.RepeatOptions
		enable   bool
	}{
		// RepeatEndAfterCount
		{
			options: func() *entity.RepeatOptions {
				temp := options
				return &temp
			}(),
			baseTime: testTime.Add(1 * time.Hour).In(loc),
			enable:   true,
		},
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Weekly.Interval = 2
				temp.Weekly.On = []entity.RepeatWeekday{
					entity.RepeatWeekdaySaturday,
					entity.RepeatWeekdayTuesday,
					entity.RepeatWeekdaySunday,
				}
				temp.Weekly.End.AfterCount = 5
				return &temp
			}(),
			baseTime: testTime.Add(-1 * time.Hour).In(loc),
			enable:   true,
		},
		// RepeatEndAfterTime
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Daily.Interval = 3
				temp.Daily.End.AfterCount = 0
				temp.Daily.End.Type = entity.RepeatEndAfterTime
				temp.Daily.End.AfterTime = testTime.AddDate(0, 1, 0).Unix()
				return &temp
			}(),
			baseTime: testTime.AddDate(0, 0, -2).Add(-1 * time.Hour).In(loc),
			enable:   false,
		},
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Daily.Interval = 5
				temp.Daily.End.AfterCount = 0
				temp.Daily.End.Type = entity.RepeatEndAfterTime
				temp.Daily.End.AfterTime = testTime.AddDate(0, 2, 0).Unix()
				return &temp
			}(),
			baseTime: testTime.AddDate(0, 0, 2).Add(1 * time.Hour).In(loc),
			enable:   false,
		},
	}

	for _, item := range tests {
		if !item.enable {
			continue
		}
		after := utils.ConvertTime(item.options.Daily.End.AfterTime, loc)
		t.Log("base time:", item.baseTime, "-----end after time:", after)
		conf := NewRepeatConfig(item.options, loc)
		plan, _ := NewRepeatCyclePlan(context.Background(), item.baseTime.Unix(), conf)
		rule, err := NewEndRepeatCycleRule(item.options)
		if err != nil {
			t.Fatal(err)
			return
		}
		//result, _ := GenerateTimeByRule(plan, rule)
		//for _, item := range result {
		//	temp := utils.ConvertTime(item, loc)
		//	t.Log(temp)
		//}
		for _, item := range plan {
			result, _ := item.GenerateTimeByRule(rule)
			for _, item := range result {
				temp := utils.ConvertTime(item, loc)
				t.Log(temp)
			}
		}
	}
}
