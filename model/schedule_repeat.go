package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sort"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type DynamicIntervalFunc func(baseTime int64, isFirst bool) (int, error)

var (
	ErrOverLimit = errors.New("over the limit")
)

type RepeatConfig struct {
	*entity.RepeatOptions
	Location      *time.Location
	RepeatEndYear int
	MaxTime       time.Time
	MinTime       time.Time
}

func NewRepeatConfig(options *entity.RepeatOptions, loc *time.Location) *RepeatConfig {
	cfg := new(RepeatConfig)
	cfg.RepeatOptions = options
	cfg.Location = loc
	cfg.RepeatEndYear = config.Get().Schedule.MaxRepeatYear
	cfg.MaxTime = time.Now().AddDate(cfg.RepeatEndYear, 0, 0).In(loc)
	cfg.MinTime = time.Now().In(loc)
	return cfg
}

type RepeatBaseTimeStamp struct {
	Start int64
	End   int64
}

type RepeatCyclePlan struct {
	ctx context.Context

	repeatCfg       *RepeatConfig
	BaseTimeStamp   *RepeatBaseTimeStamp
	Diff            []*RepeatBaseTimeStamp
	Interval        DynamicIntervalFunc
	sourceTimeStamp *RepeatBaseTimeStamp
}

func NewRepeatCyclePlan(ctx context.Context, baseStart int64, baseEnd int64, repeatCfg *RepeatConfig) (*RepeatCyclePlan, error) {
	general := &RepeatCyclePlan{
		ctx:       ctx,
		repeatCfg: repeatCfg,
		BaseTimeStamp: &RepeatBaseTimeStamp{
			Start: baseStart,
			End:   baseEnd,
		},
		Diff: []*RepeatBaseTimeStamp{
			&RepeatBaseTimeStamp{
				Start: 0,
				End:   0,
			},
		},
		Interval: nil,
		sourceTimeStamp: &RepeatBaseTimeStamp{
			Start: baseStart,
			End:   baseEnd,
		},
	}

	switch repeatCfg.Type {
	case entity.RepeatTypeDaily:
		general.Interval = general.DynamicDayInterval
		return general, nil

	case entity.RepeatTypeWeekly:
		if len(repeatCfg.Weekly.On) <= 0 {
			log.Info(ctx, "NewRepeatCyclePlan:Weekly On invalid", log.Any("Weekly", repeatCfg.Weekly))
			return nil, constant.ErrInvalidArgs
		}
		weeklyOnDiff := make(utils.Int64, len(repeatCfg.Weekly.On))
		for i, item := range repeatCfg.Weekly.On {
			if !item.Valid() {
				log.Info(ctx, "repeatCfg.Weekly.On rule invalid", log.Any("repeatCfg", repeatCfg))
				return nil, constant.ErrInvalidArgs
			}
			selectWeekDayTime := utils.GetTimeByWeekday(baseStart, item.TimeWeekday(), repeatCfg.Location)
			weeklyOnDiff[i] = selectWeekDayTime.Unix()
		}
		sort.Sort(weeklyOnDiff)
		plan := &RepeatCyclePlan{
			ctx:           ctx,
			Interval:      general.DynamicWeekInterval,
			repeatCfg:     repeatCfg,
			BaseTimeStamp: &RepeatBaseTimeStamp{},
			Diff:          make([]*RepeatBaseTimeStamp, 0),
			sourceTimeStamp: &RepeatBaseTimeStamp{
				Start: baseStart,
				End:   baseEnd,
			},
		}
		baseStartEndDiff := baseEnd - baseStart
		for i, wd := range weeklyOnDiff {
			if i == 0 {
				plan.BaseTimeStamp.Start = wd
				plan.BaseTimeStamp.End = wd + baseStartEndDiff
			}
			startDiff := wd - plan.BaseTimeStamp.Start
			endDiff := wd + baseStartEndDiff - plan.BaseTimeStamp.End
			plan.Diff = append(plan.Diff, &RepeatBaseTimeStamp{
				Start: startDiff,
				End:   endDiff,
			})
		}

		return plan, nil

	case entity.RepeatTypeMonthly:
		general.Interval = general.DynamicMonthInterval
		return general, nil

	case entity.RepeatTypeYearly:
		general.Interval = general.DynamicYearlyInterval
		return general, nil

	default:
		return nil, constant.ErrInvalidArgs
	}
}

func (r *RepeatCyclePlan) GenerateTimeByEndRule(endRule *RepeatCycleEndRule) ([]*RepeatBaseTimeStamp, error) {
	if endRule.CycleRuleType == entity.RepeatEndNever {
		endRule.CycleRuleType = entity.RepeatEndAfterTime
		endRule.AfterTime = r.repeatCfg.MaxTime.Unix()
	}
	result := make([]*RepeatBaseTimeStamp, 0)
	if !endRule.CycleRuleType.Valid() {
		log.Info(r.ctx, "GenerateTimeByEndRule:endRule CycleRuleType invalid", log.Any("endRule", endRule))
		return nil, constant.ErrInvalidArgs
	}
	baseStart := time.Unix(r.BaseTimeStamp.Start, 0).In(r.repeatCfg.Location)
	baseEnd := time.Unix(r.BaseTimeStamp.End, 0).In(r.repeatCfg.Location)
	sourceStartTime := time.Unix(r.sourceTimeStamp.Start, 0).In(r.repeatCfg.Location)

	switch endRule.CycleRuleType {
	case entity.RepeatEndAfterCount:
		var count = 0
		var isFirst = true
		for count < endRule.AfterCount && baseEnd.Before(r.repeatCfg.MaxTime) {
			day, err := r.Interval(baseStart.Unix(), isFirst)
			if err == ErrOverLimit {
				continue
			}
			if err != nil {
				log.Error(r.ctx, "GenerateTimeByEndRule:Interval error",
					log.Err(err), log.Any("RepeatCyclePlan", r),
					log.Any("RepeatCyclePlan", r),
					log.Any("endRule", endRule),
				)
				return nil, err
			}
			isFirst = false
			baseStart = baseStart.AddDate(0, 0, day)
			baseEnd = baseEnd.AddDate(0, 0, day)

			for _, d := range r.Diff {
				nextStart := utils.ConvertTime(baseStart.Unix()+d.Start, r.repeatCfg.Location)
				nextEnd := utils.ConvertTime(baseEnd.Unix()+d.End, r.repeatCfg.Location)

				if (nextStart.After(sourceStartTime) || nextStart.Equal(sourceStartTime)) &&
					nextStart.After(r.repeatCfg.MinTime) &&
					nextEnd.Before(r.repeatCfg.MaxTime) &&
					count < endRule.AfterCount {
					result = append(result, &RepeatBaseTimeStamp{
						Start: nextStart.Unix(),
						End:   nextEnd.Unix(),
					})
					count++
				}
			}
		}

	case entity.RepeatEndAfterTime:
		afterTime := time.Unix(endRule.AfterTime, 0).In(r.repeatCfg.Location)
		var isFirst = true
		for baseEnd.Before(afterTime) && baseEnd.Before(r.repeatCfg.MaxTime) {
			day, err := r.Interval(baseStart.Unix(), isFirst)
			if err == ErrOverLimit {
				continue
			}
			if err != nil {
				log.Error(r.ctx, "GenerateTimeByEndRule:Interval error",
					log.Err(err), log.Any("RepeatCyclePlan", r),
					log.Any("RepeatCyclePlan", r),
					log.Any("endRule", endRule),
				)
				return nil, err
			}
			isFirst = false
			baseStart = baseStart.AddDate(0, 0, day)
			baseEnd = baseEnd.AddDate(0, 0, day)

			for _, d := range r.Diff {
				nextStart := utils.ConvertTime(baseStart.Unix()+d.Start, r.repeatCfg.Location)
				nextEnd := utils.ConvertTime(baseEnd.Unix()+d.End, r.repeatCfg.Location)
				if (nextStart.After(sourceStartTime) || nextStart.Equal(sourceStartTime)) &&
					nextStart.After(r.repeatCfg.MinTime) &&
					nextEnd.Before(afterTime) {
					result = append(result, &RepeatBaseTimeStamp{
						Start: nextStart.Unix(),
						End:   nextEnd.Unix(),
					})
				}
			}
		}
	}
	return result, nil
}

func (r *RepeatCyclePlan) DynamicDayInterval(baseTime int64, isFirst bool) (int, error) {
	cfg := r.repeatCfg
	if cfg.Daily.Interval == 0 {
		log.Info(r.ctx, "DynamicDayInterval:Daily Interval invalid", log.Any("Daily", cfg.Daily))
		return 0, constant.ErrInvalidArgs
	}
	if isFirst {
		return 0, nil
	}
	return cfg.Daily.Interval, nil
}
func (r *RepeatCyclePlan) DynamicWeekInterval(baseTime int64, isFirst bool) (int, error) {
	cfg := r.repeatCfg
	if cfg.Weekly.Interval == 0 {
		log.Info(r.ctx, "DynamicWeekInterval:Weekly Interval invalid", log.Any("Daily", cfg.Daily))
		return 0, constant.ErrInvalidArgs
	}
	if isFirst {
		return 0, nil
	}
	return cfg.Weekly.Interval * 7, nil
}
func (r *RepeatCyclePlan) validateMonthlyData(ctx context.Context, monthlyCfg entity.RepeatMonthly) error {
	if !monthlyCfg.OnType.Valid() {
		log.Info(ctx, "DynamicMonthInterval:Monthly OnType invalid", log.Any("Monthly", monthlyCfg))
		return constant.ErrInvalidArgs
	}
	if monthlyCfg.Interval == 0 {
		log.Info(ctx, "DynamicMonthInterval:Monthly Interval invalid", log.Any("Monthly", monthlyCfg))
		return constant.ErrInvalidArgs
	}
	return nil
}
func (r *RepeatCyclePlan) DynamicMonthInterval(baseTime int64, isFirst bool) (int, error) {
	ctx := r.ctx
	cfg := r.repeatCfg
	if err := r.validateMonthlyData(ctx, cfg.Monthly); err != nil {
		return 0, err
	}
	switch cfg.Monthly.OnType {
	case entity.RepeatMonthlyOnDate:
		base := utils.ConvertTime(baseTime, cfg.Location)
		monthStart := utils.StartOfMonth(base.Year(), base.Month(), cfg.Location)
		currentMonthTime := monthStart.AddDate(0, 0, cfg.Monthly.OnDateDay-1).In(cfg.Location)
		isSameMonth := utils.IsSameMonthByTime(monthStart, currentMonthTime)
		if !isSameMonth {
			return 0, ErrOverLimit
		}

		var afterMonthTime time.Time
		if isFirst {
			afterMonthTime = currentMonthTime
		} else {
			afterMonthTime = currentMonthTime.AddDate(0, cfg.Monthly.Interval, 0).In(cfg.Location)
		}

		day := utils.GetTimeDiffToDayByTime(base, afterMonthTime, cfg.Location)
		return int(day), nil
	case entity.RepeatMonthlyOnWeek:
		if !cfg.Monthly.OnWeekSeq.Valid() {
			log.Info(ctx, "DynamicMonthInterval:Monthly OnWeekSeq invalid", log.Any("Monthly", cfg.Monthly))
			return 0, constant.ErrInvalidArgs
		}
		if !cfg.Monthly.OnWeek.Valid() {
			log.Info(ctx, "DynamicMonthInterval:Monthly OnWeek invalid", log.Any("Monthly", cfg.Monthly))
			return 0, constant.ErrInvalidArgs
		}
		base := utils.ConvertTime(baseTime, cfg.Location)
		var afterMonthTime time.Time
		if isFirst {
			afterMonthTime = base
		} else {
			afterMonthTime = base.AddDate(0, cfg.Monthly.Interval, 0).In(cfg.Location)
		}
		date := dateOfWeekday(afterMonthTime.Year(), afterMonthTime.Month(), cfg.Monthly.OnWeek, cfg.Monthly.OnWeekSeq, cfg.Location)
		day := utils.GetTimeDiffToDayByTime(base, date, cfg.Location)
		return int(day), nil
	}
	return 0, constant.ErrInvalidArgs
}
func (r *RepeatCyclePlan) validateYearlyData(ctx context.Context, yearlyCfg entity.RepeatYearly) error {
	if !yearlyCfg.OnType.Valid() {
		log.Info(ctx, "DynamicYearlyInterval:yearly OnType invalid", log.Any("Yearly", yearlyCfg))
		return constant.ErrInvalidArgs
	}
	if yearlyCfg.Interval == 0 {
		log.Info(ctx, "DynamicYearlyInterval:yearly Interval invalid", log.Any("Yearly", yearlyCfg))
		return constant.ErrInvalidArgs
	}
	return nil
}
func (r *RepeatCyclePlan) DynamicYearlyInterval(baseTime int64, isFirst bool) (int, error) {
	ctx := r.ctx
	cfg := r.repeatCfg

	if err := r.validateYearlyData(ctx, cfg.Yearly); err != nil {
		return 0, err
	}
	switch cfg.Yearly.OnType {
	case entity.RepeatYearlyOnDate:
		var interval = cfg.Yearly.Interval
		if isFirst {
			interval = 0
		}
		selectedStartYear := utils.StartOfYearByTimeStamp(baseTime, cfg.Location)
		selectedTime := selectedStartYear.AddDate(interval, cfg.Yearly.OnDateMonth-1, cfg.Yearly.OnDateDay-1)
		if int(selectedTime.Month()) != cfg.Yearly.OnDateMonth {
			log.Error(ctx, "DynamicYearlyInterval:over month limit",
				log.Any("Yearly", cfg.Yearly),
				log.Any("selectedTime", selectedTime),
			)
			return 0, ErrOverLimit
		}
		base := utils.ConvertTime(baseTime, cfg.Location)
		day := utils.GetTimeDiffToDayByTime(base, selectedTime, cfg.Location)

		return int(day), nil
	case entity.RepeatYearlyOnWeek:
		base := utils.ConvertTime(baseTime, cfg.Location)
		var afterYearTime time.Time
		if isFirst {
			afterYearTime = base
		} else {
			afterYearTime = base.AddDate(cfg.Yearly.Interval, 0, 0).In(cfg.Location)
		}
		if !cfg.Yearly.OnWeekSeq.Valid() {
			log.Info(ctx, "DynamicYearlyInterval:Yearly OnWeekSeq invalid", log.Any("Yearly", cfg.Yearly))
			return 0, constant.ErrInvalidArgs
		}
		if !cfg.Yearly.OnWeek.Valid() {
			log.Info(ctx, "DynamicYearlyInterval:Yearly OnWeek invalid", log.Any("Yearly", cfg.Yearly))
			return 0, constant.ErrInvalidArgs
		}
		date := dateOfWeekday(afterYearTime.Year(), time.Month(cfg.Yearly.OnWeekMonth), cfg.Yearly.OnWeek, cfg.Yearly.OnWeekSeq, cfg.Location)
		day := utils.GetTimeDiffToDayByTime(base, date, cfg.Location)

		return int(day), nil
	}
	return 0, constant.ErrInvalidArgs
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
func dateOfWeekday(year int, month time.Month, w entity.RepeatWeekday, seq entity.RepeatWeekSeq, location *time.Location) time.Time {
	switch seq {
	case entity.RepeatWeekSeqFirst, entity.RepeatWeekSeqSecond, entity.RepeatWeekSeqThird, entity.RepeatWeekSeqFourth:
		start := utils.StartOfMonth(year, month, location)
		offset := int(w.TimeWeekday()-start.Weekday()+7)%7 + 7*(seq.Offset()-1)
		result := start.AddDate(0, 0, offset)
		return result
	case entity.RepeatWeekSeqLast:
		end := utils.EndOfMonth(year, month, location)
		offset := int(end.Weekday()-w.TimeWeekday()+7) % 7
		result := end.AddDate(0, 0, -offset)
		return result
	default:
		return time.Now()
	}
}
