package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"time"
)

const maxRepeatYear = 2

func RepeatSchedule(ctx context.Context, template entity.Schedule) ([]entity.Schedule, error) {
	options := template.Repeat
	var result []entity.Schedule
	switch options.Type {
	case entity.RepeatTypeDaily:
		if options.Daily.Interval <= 0 {
			err := fmt.Errorf("repeat schedule: invalid daily interval options")
			log.Error(ctx, err.Error())
			return nil, err
		}
		switch options.Daily.End.Type {
		case entity.RepeatEndNever:
			maxEndTime := time.Now().AddDate(maxRepeatYear, 0, 0)
			start, end := time.Unix(template.StartAt, 0), time.Unix(template.EndAt, 0)
			for end.Before(maxEndTime) {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, item)
				start = start.AddDate(0, 0, options.Daily.Interval)
				end = end.AddDate(0, 0, options.Daily.Interval)
			}
		case entity.RepeatEndAfterCount:
			maxEndTime := time.Now().AddDate(maxRepeatYear, 0, 0)
			start, end := time.Unix(template.StartAt, 0), time.Unix(template.EndAt, 0)
			count := 0
			for count < options.Daily.End.AfterCount && end.Before(maxEndTime) {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, item)
				start = start.AddDate(0, 0, options.Daily.Interval)
				end = end.AddDate(0, 0, options.Daily.Interval)
				count++
			}
		case entity.RepeatEndAfterTime:
		default:
			err := fmt.Errorf("repeat schedule: invalid daily end type %q", string(options.Daily.End.Type))
			log.Error(ctx, err.Error())
			return nil, err
		}
	case entity.RepeatTypeWeekly:
	case entity.RepeatTypeMonthly:
	case entity.RepeatTypeYearly:
	default:
		err := fmt.Errorf("repeat schedule: invalid options type %q", string(options.Type))
		log.Error(ctx, err.Error())
		return nil, err
	}
	return result, nil
}
