package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var testFlag = true

type RepeatCycleRule struct {
	Location    *time.Location
	BaseTime    int64
	IntervalDay int
}

func NewRepeatCycleRule(ctx context.Context, baseTime int64, loc *time.Location, options *entity.RepeatOptions) ([]*RepeatCycleRule, error) {
	if !options.Type.Valid() {
		return nil, constant.ErrInvalidArgs
	}
	result := make([]*RepeatCycleRule, 0)

	switch options.Type {
	case entity.RepeatTypeDaily:
		result = append(result, &RepeatCycleRule{
			IntervalDay: options.Daily.Interval,
			BaseTime:    baseTime,
			Location:    loc,
		})

	case entity.RepeatTypeWeekly:
		if !options.Weekly.Valid() {
			log.Info(ctx, "options.Weekly rule invalid", log.Any("options", options))
			return nil, constant.ErrInvalidArgs
		}
		for _, item := range options.Weekly.On {
			if !item.Valid() {
				log.Info(ctx, "options.Weekly.On rule invalid", log.Any("options", options))
				return nil, constant.ErrInvalidArgs
			}
			tu := utils.NewTimeUtil(baseTime, loc)
			selectWeekDayTime := tu.GetTimeByWeekday(item.TimeWeekday())
			result = append(result, &RepeatCycleRule{
				IntervalDay: options.Weekly.Interval * 7,
				BaseTime:    selectWeekDayTime.Unix(),
				Location:    loc,
			})
		}
	case entity.RepeatTypeMonthly:
	case entity.RepeatTypeYearly:
	}
	return nil, nil
}

func (r *RepeatCycleRule) getMaxRepeatYear() int {
	if testFlag {
		return 2
	}
	return config.Get().Schedule.MaxRepeatYear
}

func (r *RepeatCycleRule) GenerateTimeByRule(endRule *EndRepeatCycleRule) ([]int64, error) {
	result := make([]int64, 0)
	if !endRule.CycleRuleType.Valid() {
		return nil, constant.ErrInvalidArgs
	}
	baseTime := time.Unix(r.BaseTime, 0).In(r.Location)
	maxTime := time.Now().AddDate(r.getMaxRepeatYear(), 0, 0).In(r.Location)
	minTime := time.Now().In(r.Location)

	switch endRule.CycleRuleType {
	case entity.RepeatEndAfterCount:
		var count = 0
		for count < endRule.AfterCount && baseTime.Before(maxTime) {
			if baseTime.After(minTime) {
				result = append(result, r.BaseTime)
				count++
			}
			baseTime = baseTime.AddDate(0, 0, r.IntervalDay)
		}
	case entity.RepeatEndAfterTime:
		afterTime := time.Unix(endRule.AfterTime, 0).In(r.Location)
		for baseTime.Before(afterTime) && baseTime.Before(maxTime) {
			if baseTime.After(minTime) {
				result = append(result, r.BaseTime)
			}
			baseTime = baseTime.AddDate(0, 0, r.IntervalDay)
		}
	}
	return result, nil
}

type EndRepeatCycleRule struct {
	CycleRuleType entity.RepeatEndType
	AfterCount    int
	AfterTime     int64
}

func NewEndRepeatCycleRule(options *entity.RepeatOptions) (*EndRepeatCycleRule, error) {
	if !options.Type.Valid() {
		return nil, constant.ErrInvalidArgs
	}
	result := new(EndRepeatCycleRule)
	switch options.Type {
	case entity.RepeatTypeDaily:
		result.CycleRuleType = options.Daily.End.Type
		result.AfterCount = options.Daily.End.AfterCount
		result.AfterTime = options.Daily.End.AfterTime
	case entity.RepeatTypeWeekly:
		result.CycleRuleType = options.Weekly.End.Type
		result.AfterCount = options.Weekly.End.AfterCount
		result.AfterTime = options.Weekly.End.AfterTime
	case entity.RepeatTypeMonthly:
		result.CycleRuleType = options.Monthly.End.Type
		result.AfterCount = options.Monthly.End.AfterCount
		result.AfterTime = options.Monthly.End.AfterTime
	case entity.RepeatTypeYearly:
		result.CycleRuleType = options.Yearly.End.Type
		result.AfterCount = options.Yearly.End.AfterCount
		result.AfterTime = options.Yearly.End.AfterTime
	}
	return result, nil
}

//func (r *Repeat) GetDays() []time.Time {
//	var result = make([]time.Time, 0)
//	switch r.Rules.Type {
//	case entity.RepeatTypeDaily:
//
//	}
//}
