package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var testFlag = true

type RepeatConfig struct {
	*entity.RepeatOptions
	Location *time.Location
}

func NewRepeatConfig(options *entity.RepeatOptions, loc *time.Location) *RepeatConfig {
	return &RepeatConfig{options, loc}
}

type IntervalType string

const (
	IntervalTypeDay   IntervalType = "day"
	IntervalTypeWeek  IntervalType = "week"
	IntervalTypeMonth IntervalType = "month"
	IntervalTypeYear  IntervalType = "year"
)

type RepeatCyclePlan struct {
	repeatCfg *RepeatConfig

	BaseTime int64
	Interval DynamicIntervalFunc
}

func NewRepeatCyclePlan2(ctx context.Context, baseTime int64, loc *time.Location, repeatCfg *RepeatConfig) *RepeatCyclePlan {
	result := &RepeatCyclePlan{
		repeatCfg: repeatCfg,

		BaseTime: baseTime,
		Interval: DefaultDynamicInterval,
	}
	return result
}

func (r *RepeatCyclePlan) GenerateTimeByRule(endRule *RepeatCycleEndRule) ([]int64, error) {
	result := make([]int64, 0)
	if !endRule.CycleRuleType.Valid() {
		return nil, constant.ErrInvalidArgs
	}
	baseTime := time.Unix(r.BaseTime, 0).In(r.repeatCfg.Location)
	maxTime := time.Now().AddDate(r.getMaxRepeatYear(), 0, 0).In(r.repeatCfg.Location)
	minTime := time.Now().In(r.repeatCfg.Location)

	switch endRule.CycleRuleType {
	case entity.RepeatEndAfterCount:
		var count = 0
		var isFirst = true
		for count < endRule.AfterCount && baseTime.Before(maxTime) {
			day, err := r.Interval(baseTime.Unix(), r.repeatCfg, isFirst)
			isFirst = false
			if err != nil {
				return nil, err
			}
			baseTime = baseTime.AddDate(0, 0, day)
			if baseTime.After(minTime) {
				result = append(result, r.BaseTime)
				count++
			}
		}

	case entity.RepeatEndAfterTime:
		afterTime := time.Unix(endRule.AfterTime, 0).In(r.repeatCfg.Location)
		var isFirst = true
		for baseTime.Before(afterTime) && baseTime.Before(maxTime) {
			day, err := r.Interval(baseTime.Unix(), r.repeatCfg, isFirst)
			isFirst = false
			if err != nil {
				return nil, err
			}
			baseTime = baseTime.AddDate(0, 0, day)
			if baseTime.After(minTime) {
				result = append(result, r.BaseTime)
			}
		}
	}
	return result, nil
}

type DynamicIntervalFunc func(baseTime int64, cfg *RepeatConfig, isFirst bool) (int, error)

func DefaultDynamicInterval(baseTime int64, cfg *RepeatConfig, isFirst bool) (int, error) {
	return 0, nil
}
func DynamicDayInterval(baseTime int64, cfg *RepeatConfig, isFirst bool) (int, error) {
	if isFirst {
		return 0, nil
	}
	return cfg.Daily.Interval, nil
}
func DynamicWeekInterval(baseTime int64, cfg *RepeatConfig, isFirst bool) (int, error) {
	if isFirst {
		return 0, nil
	}
	return cfg.Monthly.Interval * 7, nil
}

var (
	ErrOverLimit = errors.New("Over the limit")
)

func DynamicMonthInterval(baseTime int64, cfg *RepeatConfig, isFirst bool) (int, error) {
	if !cfg.Monthly.OnType.Valid() {
		return 0, constant.ErrInvalidArgs
	}
	tu := utils.NewTimeUtil(baseTime, cfg.Location)
	switch cfg.Monthly.OnType {
	case entity.RepeatMonthlyOnDate:
		monthStart := tu.StartOfMonth()
		currentMonthTime := monthStart.AddDate(0, 0, cfg.Monthly.OnDateDay).In(cfg.Location)
		isSameMonth := utils.IsSameMonthByTime(monthStart, currentMonthTime)
		if !isSameMonth {
			return 0, ErrOverLimit
		}

		var afterMonthTime time.Time
		if isFirst {
			isSameMonth2 := utils.IsSameMonthByTime(time.Now().In(cfg.Location), monthStart)
			if isSameMonth2 {
				return currentMonthTime.Day() - tu.ToTime().Day(), nil
			}
			afterMonthTime = currentMonthTime
		} else {
			afterMonthTime = monthStart.AddDate(0, cfg.Monthly.Interval, cfg.Monthly.OnDateDay).In(cfg.Location)
		}
		day := utils.GetTimeDiffToDay(baseTime, afterMonthTime.Unix())
		return int(day), nil
	case entity.RepeatMonthlyOnWeek:
		date, err := dateOfWeekday(baseTime, cfg.Monthly.OnWeek, cfg.Monthly.OnWeekSeq, cfg.Location)
		if err != nil {
			return 0, err
		}
		var afterMonthTime time.Time
		if isFirst {
			isSameMonth2 := utils.IsSameMonthByTime(time.Now().In(cfg.Location), date)
			if isSameMonth2 {
				return date.Day() - tu.ToTime().Day(), nil
			}
			afterMonthTime = date
		} else {
			afterMonthTime = date.AddDate(0, cfg.Monthly.Interval, 0).In(cfg.Location)
		}
		day := utils.GetTimeDiffToDay(baseTime, afterMonthTime.Unix())
		return int(day), nil
	}
	return 0, constant.ErrInvalidArgs
}

//func NewRepeatCycleRule(ctx context.Context, baseTime int64, loc *time.Location, repeatCfg *RepeatConfig) ([]*RepeatCyclePlan, error) {
//	if !repeatCfg.Type.Valid() {
//		return nil, constant.ErrInvalidArgs
//	}
//	result := make([]*RepeatCyclePlan, 0)
//	timeTool := utils.NewTimeUtil(baseTime, loc)
//	switch repeatCfg.Type {
//	case entity.RepeatTypeDaily:
//		result = append(result, &RepeatCyclePlan{
//			Interval: repeatCfg.Daily.Interval,
//			BaseTime: baseTime,
//			Location: loc,
//		})
//
//	case entity.RepeatTypeWeekly:
//		if !repeatCfg.Weekly.Valid() {
//			log.Info(ctx, "repeatCfg.Weekly rule invalid", log.Any("repeatCfg", repeatCfg))
//			return nil, constant.ErrInvalidArgs
//		}
//		for _, item := range repeatCfg.Weekly.On {
//			if !item.Valid() {
//				log.Info(ctx, "repeatCfg.Weekly.On rule invalid", log.Any("repeatCfg", repeatCfg))
//				return nil, constant.ErrInvalidArgs
//			}
//			selectWeekDayTime := timeTool.GetTimeByWeekday(item.TimeWeekday())
//			result = append(result, &RepeatCyclePlan{
//				Interval: repeatCfg.Weekly.Interval * 7,
//				BaseTime: selectWeekDayTime.Unix(),
//				Location: loc,
//			})
//		}
//	case entity.RepeatTypeMonthly:
//		switch repeatCfg.Monthly.OnType {
//		case entity.RepeatMonthlyOnDate:
//			result = append(result, &RepeatCyclePlan{
//				Location:      loc,
//				BaseTime:      baseTime,
//				IntervalDay:   repeatCfg.Monthly.OnDateDay,
//				IntervalMonth: repeatCfg.Monthly.Interval,
//			})
//		case entity.RepeatMonthlyOnWeek:
//			 := dateOfWeekday(baseTime, repeatCfg.Monthly.OnWeek, repeatCfg.Monthly.OnWeekSeq, loc)
//			result = append(result, &RepeatCyclePlan{
//				Location:      loc,
//				BaseTime:      baseTime,
//				IntervalDay:   repeatCfg.Monthly.OnWeekSeq,
//				IntervalMonth: repeatCfg.Monthly.Interval,
//			})
//		}
//	case entity.RepeatTypeYearly:
//	}
//	return nil, nil
//}

func dateOfWeekday(baseTime int64, w entity.RepeatWeekday, seq entity.RepeatWeekSeq, location *time.Location) (time.Time, error) {
	timeTool := utils.NewTimeUtil(baseTime, location)
	switch seq {
	case entity.RepeatWeekSeqFirst, entity.RepeatWeekSeqSecond, entity.RepeatWeekSeqThird, entity.RepeatWeekSeqFourth:
		start := timeTool.StartOfMonth()
		offset := int(w.TimeWeekday()-start.Weekday()+7)%7 + 7*(seq.Offset()-1)
		result := start.AddDate(0, 0, offset)
		return result, nil
	case entity.RepeatWeekSeqLast:
		end := timeTool.EndOfMonth()
		offset := int(end.Weekday()-w.TimeWeekday()+7) % 7
		result := end.AddDate(0, 0, -offset)
		return result, nil
	default:
		return time.Now(), constant.ErrInvalidArgs
	}
}

func (r *RepeatCyclePlan) getMaxRepeatYear() int {
	if testFlag {
		return 2
	}
	return config.Get().Schedule.MaxRepeatYear
}

type RepeatCycleEndRule struct {
	CycleRuleType entity.RepeatEndType
	AfterCount    int
	AfterTime     int64
}

func NewEndRepeatCycleRule(options *entity.RepeatOptions) (*RepeatCycleEndRule, error) {
	if !options.Type.Valid() {
		return nil, constant.ErrInvalidArgs
	}
	result := new(RepeatCycleEndRule)
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
