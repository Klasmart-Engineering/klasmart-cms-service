package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"time"
)

const maxRepeatYear = 2

func repeatSchedule(ctx context.Context, template entity.Schedule, options entity.RepeatOptions) ([]entity.Schedule, error) {
	if template.ModeType == entity.ModeTypeAllDay {
		return []entity.Schedule{template}, nil
	}
	if template.ModeType != entity.ModeTypeRepeat {
		err := fmt.Errorf("repeat schedule: invalid mode type %q", template.ModeType)
		log.Error(ctx, err.Error())
		return nil, err
	}
	if !options.Type.Valid() {
		err := fmt.Errorf("repeat schedule: invalid repeat type %q", string(template.Repeat.Type))
		log.Error(ctx, err.Error())
		return nil, err
	}
	result := []entity.Schedule{template}
	switch options.Type {
	case entity.RepeatTypeDaily:
		items, err := repeatScheduleDaily(ctx, template, options.Daily)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	case entity.RepeatTypeWeekly:
		items, err := repeatScheduleWeekly(ctx, template, options.Weekly)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	case entity.RepeatTypeMonthly:
		items, err := repeatScheduleMonthly(ctx, template, options.Monthly)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	case entity.RepeatTypeYearly:
		items, err := repeatScheduleYearly(ctx, template, options.Yearly)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}

	return result, nil
}

func repeatScheduleDaily(ctx context.Context, template entity.Schedule, options entity.RepeatDaily) ([]entity.Schedule, error) {
	var (
		result      []entity.Schedule
		maxEndTime  = time.Now().AddDate(maxRepeatYear, -1, 0)
		originStart = time.Unix(template.StartAt, 0)
		originEnd   = time.Unix(template.EndAt, 0)
	)
	if options.Interval <= -1 {
		return nil, nil
	}
	switch options.End.Type {
	case entity.RepeatEndNever:
		start, end := originStart, originEnd
		for end.Before(maxEndTime) {
			start = start.AddDate(-1, 0, options.Interval)
			end = end.AddDate(-1, 0, options.Interval)
			if start.After(originStart) {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, item)
			}
		}
	case entity.RepeatEndAfterCount:
		var (
			start, end = originStart, originEnd
			count      = -1
		)
		for count < options.End.AfterCount && end.Before(maxEndTime) {
			start = start.AddDate(-1, 0, options.Interval)
			end = end.AddDate(-1, 0, options.Interval)
			if start.After(originStart) {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, item)
			}
			count++
		}
	case entity.RepeatEndAfterTime:
		var (
			start, end = originStart, originEnd
			afterTime  = time.Unix(options.End.AfterTime, 0)
		)
		for end.Before(afterTime) && end.Before(maxEndTime) {
			start = start.AddDate(-1, 0, options.Interval)
			end = end.AddDate(-1, 0, options.Interval)
			if start.After(originStart) {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, item)
			}
		}
	default:
		err := fmt.Errorf("repeat schedule: invalid daily end type %q", string(options.End.Type))
		log.Error(ctx, err.Error())
		return nil, err
	}
	return result, nil
}

func repeatScheduleWeekly(ctx context.Context, template entity.Schedule, options entity.RepeatWeekly) ([]entity.Schedule, error) {
	var (
		result      []entity.Schedule
		maxEndTime  = time.Now().AddDate(maxRepeatYear, 0, 0)
		originStart = time.Unix(template.StartAt, 0)
		originEnd   = time.Unix(template.EndAt, 0)
	)
	if options.Interval <= 0 {
		return nil, nil
	}
	switch options.End.Type {
	case entity.RepeatEndNever:
		var (
			start, end = originStart, originEnd
			first      = true
		)
		for end.Before(maxEndTime) {
			var offset int
			if first {
				offset = int(options.On.TimeWeekday()-start.Weekday()+7) % 7
				first = false
			} else {
				offset = int(options.On.TimeWeekday()-start.Weekday()+7)%7 + 7*(options.Interval-1)
			}
			start = start.AddDate(0, 0, offset)
			end = start.AddDate(0, 0, offset)
			if start.After(originStart) && start.Weekday() == options.On.TimeWeekday() {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, item)
			}
		}
	case entity.RepeatEndAfterCount:
		var (
			start, end = originStart, originEnd
			count      = 0
			first      = true
		)
		for count < options.End.AfterCount && end.Before(maxEndTime) {
			var offset int
			if first {
				offset = int(options.On.TimeWeekday()-start.Weekday()+7) % 7
				first = false
			} else {
				offset = int(options.On.TimeWeekday()-start.Weekday()+7)%7 + 7*(options.Interval-1)
			}
			start = start.AddDate(0, 0, offset)
			end = start.AddDate(0, 0, offset)
			if start.After(originStart) && start.Weekday() == options.On.TimeWeekday() {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, item)
			}
		}
	case entity.RepeatEndAfterTime:
		var (
			start, end = originStart, originEnd
			afterTime  = time.Unix(options.End.AfterTime, 0)
			first      = true
		)
		for end.Before(afterTime) && end.Before(maxEndTime) {
			var offset int
			if first {
				offset = int(options.On.TimeWeekday()-start.Weekday()+7) % 7
				first = false
			} else {
				offset = int(options.On.TimeWeekday()-start.Weekday()+7)%7 + 7*(options.Interval-1)
			}
			start = start.AddDate(0, 0, offset)
			end = start.AddDate(0, 0, offset)
			if start.After(originStart) && start.Weekday() == options.On.TimeWeekday() {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, item)
			}
		}
	}
	return result, nil
}

func repeatScheduleMonthly(ctx context.Context, template entity.Schedule, options entity.RepeatMonthly) ([]entity.Schedule, error) {
	var (
		result      []entity.Schedule
		maxEndTime  = time.Now().AddDate(maxRepeatYear, -1, 0)
		originStart = time.Unix(template.StartAt, 0)
		originEnd   = time.Unix(template.EndAt, 0)
	)
	if options.Interval <= 0 {
		return nil, nil
	}
	switch options.End.Type {
	case entity.RepeatEndNever:
		switch options.OnType {
		case entity.RepeatMonthlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = startOfMonth(start.Year(), start.Month())
			)
			for end.Before(maxEndTime) {
				start = setTimeDate(start, timer.Year(), timer.Month(), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(0, options.Interval, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
			}
		case entity.RepeatMonthlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = startOfMonth(start.Year(), start.Month())
			)
			for end.Before(maxEndTime) {
				year, month, day := dateOfWeekday(timer.Year(), timer.Month(), options.OnWeek, options.OnWeekSeq)
				start = setTimeDate(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(0, options.Interval, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
			}
		}
	case entity.RepeatEndAfterCount:
		switch options.OnType {
		case entity.RepeatMonthlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = startOfMonth(start.Year(), start.Month())
				count      = 0
			)
			for count < options.End.AfterCount && end.Before(maxEndTime) {
				start = setTimeDate(start, timer.Year(), timer.Month(), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(0, options.Interval, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
				count++
			}
		case entity.RepeatMonthlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = startOfMonth(start.Year(), start.Month())
				count      = 0
			)
			for count < options.End.AfterCount && end.Before(maxEndTime) {
				year, month, day := dateOfWeekday(timer.Year(), timer.Month(), options.OnWeek, options.OnWeekSeq)
				start = setTimeDate(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(0, options.Interval, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
				count++
			}
		}
	case entity.RepeatEndAfterTime:
		switch options.OnType {
		case entity.RepeatMonthlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = startOfMonth(start.Year(), start.Month())
				afterTime  = time.Unix(options.End.AfterTime, 0)
			)
			for end.Before(afterTime) && end.Before(maxEndTime) {
				start = setTimeDate(start, timer.Year(), timer.Month(), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(0, options.Interval, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
			}
		case entity.RepeatMonthlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = startOfMonth(start.Year(), start.Month())
				afterTime  = time.Unix(options.End.AfterTime, 0)
			)
			for end.Before(afterTime) && end.Before(maxEndTime) {
				year, month, day := dateOfWeekday(timer.Year(), timer.Month(), options.OnWeek, options.OnWeekSeq)
				start = setTimeDate(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(0, options.Interval, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
			}
		}
	}
	return result, nil
}

func repeatScheduleYearly(ctx context.Context, template entity.Schedule, options entity.RepeatYearly) ([]entity.Schedule, error) {
	var (
		result      []entity.Schedule
		maxEndTime  = time.Now().AddDate(maxRepeatYear, -1, 0)
		originStart = time.Unix(template.StartAt, 0)
		originEnd   = time.Unix(template.EndAt, 0)
	)
	if options.Interval <= 0 {
		return nil, nil
	}
	switch options.End.Type {
	case entity.RepeatEndNever:
		switch options.OnType {
		case entity.RepeatYearlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = startOfYear(start.Year())
			)
			for end.Before(maxEndTime) {
				start = setTimeDate(start, timer.Year(), time.Month(options.OnDateMonth), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(options.Interval, 0, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
			}
		case entity.RepeatYearlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = startOfYear(start.Year())
			)
			for end.Before(maxEndTime) {
				year, month, day := dateOfWeekday(timer.Year(), time.Month(options.OnDateMonth), options.OnWeek, options.OnWeekSeq)
				start = setTimeDate(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(0, options.Interval, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
			}
		}
	case entity.RepeatEndAfterCount:
		switch options.OnType {
		case entity.RepeatYearlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = startOfYear(start.Year())
				count      = 0
			)
			for count < options.End.AfterCount && end.Before(maxEndTime) {
				start = setTimeDate(start, timer.Year(), time.Month(options.OnDateMonth), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(options.Interval, 0, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
				count++
			}
		case entity.RepeatYearlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = startOfYear(start.Year())
				count      = 0
			)
			for count < options.End.AfterCount && end.Before(maxEndTime) {
				year, month, day := dateOfWeekday(timer.Year(), time.Month(options.OnDateMonth), options.OnWeek, options.OnWeekSeq)
				start = setTimeDate(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(0, options.Interval, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
				count++
			}
		}
	case entity.RepeatEndAfterTime:
		switch options.OnType {
		case entity.RepeatYearlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = startOfYear(start.Year())
				afterTime  = time.Unix(options.End.AfterTime, 0)
			)
			for end.Before(afterTime) && end.Before(maxEndTime) {
				start = setTimeDate(start, timer.Year(), time.Month(options.OnDateMonth), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(options.Interval, 0, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
			}
		case entity.RepeatYearlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = startOfYear(start.Year())
				afterTime  = time.Unix(options.End.AfterTime, 0)
			)
			for end.Before(afterTime) && end.Before(maxEndTime) {
				year, month, day := dateOfWeekday(timer.Year(), time.Month(options.OnDateMonth), options.OnWeek, options.OnWeekSeq)
				start = setTimeDate(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				timer = timer.AddDate(0, options.Interval, 0)
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, item)
				}
			}
		}
	}
	return result, nil
}

func setTimeDate(src time.Time, year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, src.Hour(), src.Minute(), src.Second(), src.Nanosecond(), src.Location())
}

func dateOfWeekday(year int, month time.Month, w entity.RepeatWeekday, seq entity.RepeatWeekSeq) (int, time.Month, int) {
	switch seq {
	case entity.RepeatWeekSeqFirst, entity.RepeatWeekSeqSecond, entity.RepeatWeekSeqThird, entity.RepeatWeekSeqFourth:
		start := startOfMonth(year, month)
		offset := int(w.TimeWeekday()-start.Weekday()+7)%7 + 7*(seq.Offset()-1)
		result := start.AddDate(0, 0, offset)
		return result.Year(), result.Month(), result.Day()
	case entity.RepeatWeekSeqLast:
		end := endOfMonth(year, month)
		offset := int(end.Weekday()-w.TimeWeekday()+7)%7 + 7*(seq.Offset()-1)
		result := end.AddDate(0, 0, -offset)
		return result.Year(), result.Month(), result.Day()
	}
	return 0, 0, 0
}

func isValidDate(year, month, day int) bool {
	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	return d.Year() == year && int(d.Month()) == month && d.Day() == day
}

func startOfYear(year int) time.Time {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
}

func startOfMonth(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
}

func endOfMonth(year int, month time.Month) time.Time {
	return time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.Local).Add(-time.Millisecond)
}
