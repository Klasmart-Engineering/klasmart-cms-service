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

	tests := []testData{
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
				temp.Daily.End.AfterTime = testTime.AddDate(0, 0, 10).Unix()
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
				temp.Daily.End.AfterTime = testTime.AddDate(0, 0, 20).Unix()
				return &temp
			}(),
			baseTime: testTime.AddDate(0, 0, 2).Add(1 * time.Hour).In(loc),
			enable:   false,
		},
	}
	execute(tests, t, loc)
}

func TestDynamicWeekInterval(t *testing.T) {
	options := entity.RepeatOptions{
		Type: entity.RepeatTypeWeekly,
		Weekly: entity.RepeatWeekly{
			Interval: 1,
			On: []entity.RepeatWeekday{
				entity.RepeatWeekdayWednesday,
				entity.RepeatWeekdayThursday,
			},
			End: entity.RepeatEnd{
				Type:       entity.RepeatEndAfterCount,
				AfterCount: 6,
				AfterTime:  0,
			},
		},
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	//testTime := time.Now() //time.Date(2020, 12, 8, 12, 30, 0, 0, loc)
	testTime := time.Date(2020, 12, 10, 15, 30, 0, 0, loc)
	tests := []testData{
		// RepeatEndAfterCount
		{
			options: func() *entity.RepeatOptions {
				temp := options
				return &temp
			}(),
			baseTime: testTime.In(loc),
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
			baseTime: testTime.In(loc),
			enable:   false,
		},
		// RepeatEndAfterTime
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Weekly.Interval = 1
				temp.Weekly.End.AfterCount = 0
				temp.Weekly.End.Type = entity.RepeatEndAfterTime
				temp.Weekly.End.AfterTime = testTime.AddDate(0, 2, 0).Unix()
				return &temp
			}(),
			baseTime: testTime.AddDate(0, 0, -2).Add(-1 * time.Hour).In(loc),
			enable:   false,
		},
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Weekly.Interval = 2
				temp.Weekly.End.AfterCount = 0
				temp.Weekly.End.Type = entity.RepeatEndAfterTime
				temp.Weekly.End.AfterTime = testTime.AddDate(0, 3, 0).Unix()
				return &temp
			}(),
			baseTime: testTime.AddDate(0, 0, 2).Add(1 * time.Hour).In(loc),
			enable:   false,
		},
	}

	execute(tests, t, loc)
}

func TestDynamicMonthInterval(t *testing.T) {
	options := entity.RepeatOptions{
		Type: entity.RepeatTypeMonthly,
		Monthly: entity.RepeatMonthly{
			Interval:  1,
			OnType:    entity.RepeatMonthlyOnDate,
			OnDateDay: 10,
			End: entity.RepeatEnd{
				Type:       entity.RepeatEndAfterCount,
				AfterCount: 5,
				AfterTime:  0,
			},
		},
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	testTime := time.Date(2020, 12, 10, 9, 30, 0, 0, loc)
	tests := []testData{
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
				temp.Monthly.Interval = 2
				temp.Monthly.OnDateDay = 8
				temp.Monthly.End.AfterCount = 14
				return &temp
			}(),
			baseTime: testTime.Add(1 * time.Hour).In(loc),
			enable:   false,
		},
		// RepeatEndAfterTime
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Monthly.OnType = entity.RepeatMonthlyOnDate
				temp.Monthly.End.Type = entity.RepeatEndAfterTime
				temp.Monthly.End.AfterTime = testTime.AddDate(0, 5, 0).Unix()
				return &temp
			}(),
			baseTime: testTime.Add(1 * time.Hour).In(loc),
			enable:   false,
		},
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Monthly.Interval = 2
				temp.Monthly.OnDateDay = 11
				temp.Monthly.OnType = entity.RepeatMonthlyOnDate
				temp.Monthly.End.Type = entity.RepeatEndAfterTime
				temp.Monthly.End.AfterTime = testTime.AddDate(0, 7, 0).Unix()
				return &temp
			}(),
			baseTime: testTime.Add(-1 * time.Hour).In(loc),
			enable:   false,
		},
		// RepeatEndAfterTime  RepeatMonthlyOnWeek
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Monthly.OnType = entity.RepeatMonthlyOnWeek
				temp.Monthly.OnWeekSeq = entity.RepeatWeekSeqFirst
				temp.Monthly.OnWeek = entity.RepeatWeekdaySunday
				temp.Monthly.End.Type = entity.RepeatEndAfterTime
				temp.Monthly.End.AfterTime = testTime.AddDate(0, 5, 0).Unix()
				return &temp
			}(),
			baseTime: testTime.Add(1 * time.Hour).In(loc),
			enable:   false,
		},
		{
			options: func() *entity.RepeatOptions {
				temp := options
				temp.Monthly.Interval = 2
				temp.Monthly.OnWeekSeq = entity.RepeatWeekSeqLast
				temp.Monthly.OnWeek = entity.RepeatWeekdaySaturday
				temp.Monthly.OnType = entity.RepeatMonthlyOnWeek
				temp.Monthly.End.Type = entity.RepeatEndAfterTime
				temp.Monthly.End.AfterTime = testTime.AddDate(0, 7, 0).Unix()
				return &temp
			}(),
			baseTime: testTime.Add(-1 * time.Hour).In(loc),
			enable:   false,
		},
	}
	execute(tests, t, loc)
}

func execute(tests []testData, t *testing.T, loc *time.Location) {
	for _, item := range tests {
		if !item.enable {
			continue
		}

		t.Log("start time:", item.baseTime.In(time.Now().Location()))
		conf := NewRepeatConfig(item.options, loc)
		plan, _ := NewRepeatCyclePlan(context.Background(), item.baseTime.Unix(), item.baseTime.AddDate(0, 0, 1).Unix(), conf)
		rule, err := NewEndRepeatCycleRule(item.options)
		if err != nil {
			t.Fatal(err)
			return
		}

		result, _ := plan.GenerateTimeByEndRule(rule)
		for _, item := range result {
			temp := utils.ConvertTime(item.Start, loc)
			temp2 := utils.ConvertTime(item.End, loc)
			t.Log(temp, temp2)
		}
	}
}

type testData struct {
	baseTime time.Time
	options  *entity.RepeatOptions
	enable   bool
}

func TestDynamicYearlyInterval(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	testTime := time.Now() //time.Date(2020, 12, 8, 12, 30, 0, 0, loc)
	tests := []testData{
		{
			baseTime: testTime.Add(-1 * time.Hour),
			options: &entity.RepeatOptions{
				Type: entity.RepeatTypeYearly,
				Yearly: entity.RepeatYearly{
					Interval:    1,
					OnType:      entity.RepeatYearlyOnDate,
					OnDateMonth: 12,
					OnDateDay:   20,
					OnWeekMonth: 0,
					OnWeekSeq:   "",
					OnWeek:      "",
					End: entity.RepeatEnd{
						Type:       entity.RepeatEndAfterCount,
						AfterCount: 3,
					},
				},
			},
			enable: true,
		},
		{
			baseTime: testTime.Add(1 * time.Hour),
			options: &entity.RepeatOptions{
				Type: entity.RepeatTypeYearly,
				Yearly: entity.RepeatYearly{
					Interval:    1,
					OnType:      entity.RepeatYearlyOnWeek,
					OnDateMonth: 0,
					OnDateDay:   0,
					OnWeekMonth: 12,
					OnWeekSeq:   entity.RepeatWeekSeqFirst,
					OnWeek:      entity.RepeatWeekdaySunday,
					End: entity.RepeatEnd{
						Type:       entity.RepeatEndAfterCount,
						AfterCount: 3,
					},
				},
			},
			enable: true,
		},
	}

	execute(tests, t, loc)
}

func TestDateOfWeekday(t *testing.T) {
	tt := time.Now()
	r := dateOfWeekday(tt.Year(), tt.Month()+1, entity.RepeatWeekdaySunday, entity.RepeatWeekSeqFirst, tt.Location())
	t.Log(r)
}
func TestDateOfWeekday2(t *testing.T) {
	tt := time.Date(2020, 12, 8, 22, 38, 0, 0, time.Now().Location())
	tt2 := time.Date(2020, 12, 8, 0, 0, 0, 0, time.Now().Location())
	r := utils.GetTimeDiffToDayByTime(tt, tt2, time.Now().Location())
	t.Log(r)
	t.Log(tt.Day() - tt2.Day())
}
func TestStartScheduleRepeat(t *testing.T) {
	loc := time.Now().Location()
	options := &entity.RepeatOptions{
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
	sm := scheduleModel{}
	result, _ := sm.StartScheduleRepeat(context.Background(), &entity.Schedule{
		StartAt: time.Now().In(loc).Unix(),
		EndAt:   time.Now().Add(1 * time.Hour).In(loc).Unix(),
	}, options, loc)
	for _, item := range result {
		t1 := utils.ConvertTime(item.StartAt, loc)
		t2 := utils.ConvertTime(item.EndAt, loc)
		t.Log(t1, t2)
	}
}
