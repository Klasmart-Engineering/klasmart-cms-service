package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type RepeatCycleRule struct {
	Location    *time.Location
	BaseTime    int64
	IntervalDay int
}

func NewRepeatCycleRule(ctx context.Context, baseTime int64, loc *time.Location, options entity.RepeatOptions) ([]*RepeatCycleRule, error) {
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

type EndRepeatCycleRule struct {
	AfterCount int
	AfterTime  int64
}

//func (r *Repeat) GetDays() []time.Time {
//	var result = make([]time.Time, 0)
//	switch r.Rules.Type {
//	case entity.RepeatTypeDaily:
//
//	}
//}
