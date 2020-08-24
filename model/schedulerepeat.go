package model

import (
	"context"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"time"
)

func (s *scheduleModel) getMaxRepeatYear() int {
	if s.testScheduleRepeatFlag {
		return 2
	}
	return config.Get().Schedule.MaxRepeatYear
}

func (s *scheduleModel) RepeatSchedule(ctx context.Context, template *entity.Schedule) ([]*entity.Schedule, error) {
	if template == nil {
		err := fmt.Errorf("repeat schedule(include template): require not nil template")
		log.Error(ctx, err.Error())
		return nil, entity.ErrInvalidArgs(err)
	}
	switch template.ModeType {
	case entity.ModeTypeRepeat:
		result := []*entity.Schedule{template}
		items, err := s.repeatSchedule(ctx, template, template.Repeat)
		if err != nil {
			log.Error(ctx, "repeat schedule(include template): call repeat schedule failed",
				log.Err(err),
				log.Any("template", template),
			)
			return nil, err
		}
		result = append(result, items...)
		return result, nil
	default:
		return []*entity.Schedule{template}, nil
	}
}

func (s *scheduleModel) repeatSchedule(ctx context.Context, template *entity.Schedule, options entity.RepeatOptions) ([]*entity.Schedule, error) {
	if template == nil {
		err := fmt.Errorf("repeat schedule: require not nil template")
		log.Error(ctx, err.Error())
		return nil, entity.ErrInvalidArgs(err)
	}
	var result []*entity.Schedule
	switch options.Type {
	case entity.RepeatTypeDaily:
		items, err := s.repeatScheduleDaily(ctx, template, options.Daily)
		if err != nil {
			log.Error(ctx, "repeat schedule: repeat schedule daily failed",
				log.Err(err),
				log.Any("template", template),
				log.Any("daily_options", options.Daily),
			)
			return nil, err
		}
		result = append(result, items...)
	case entity.RepeatTypeWeekly:
		items, err := s.repeatScheduleWeekly(ctx, template, options.Weekly)
		if err != nil {
			log.Error(ctx, "repeat schedule: repeat schedule weekly failed",
				log.Err(err),
				log.Any("template", template),
				log.Any("weekly_options", options.Weekly),
			)
			return nil, err
		}
		result = append(result, items...)
	case entity.RepeatTypeMonthly:
		items, err := s.repeatScheduleMonthly(ctx, template, options.Monthly)
		if err != nil {
			log.Error(ctx, "repeat schedule: repeat schedule monthly failed",
				log.Err(err),
				log.Any("template", template),
				log.Any("monthly_options", options.Monthly),
			)
			return nil, err
		}
		result = append(result, items...)
	case entity.RepeatTypeYearly:
		items, err := s.repeatScheduleYearly(ctx, template, options.Yearly)
		if err != nil {
			log.Error(ctx, "repeat schedule: repeat schedule yearly failed",
				log.Err(err),
				log.Any("template", template),
				log.Any("yearly_options", options.Yearly),
			)
			return nil, err
		}
		result = append(result, items...)
	default:
		err := errors.New("repeat schedule: invalid repeat type")
		log.Error(ctx, err.Error(), log.String("repeat_type", string(options.Type)))
		return nil, entity.ErrInvalidArgs(err)
	}
	return result, nil
}

func (s *scheduleModel) repeatScheduleDaily(ctx context.Context, template *entity.Schedule, options entity.RepeatDaily) ([]*entity.Schedule, error) {
	if template == nil {
		err := fmt.Errorf("repeat schedule daily: require not nil template")
		log.Error(ctx, err.Error())
		return nil, entity.ErrInvalidArgs(err)
	}
	if options.Interval <= 0 {
		log.Debug(ctx, "repeat schedule daily: options interval less than 0", log.Int("interval", options.Interval))
		return nil, nil
	}
	var (
		result      []*entity.Schedule
		maxEndTime  = time.Now().AddDate(s.getMaxRepeatYear(), 0, 0)
		originStart = time.Unix(template.StartAt, 0)
		originEnd   = time.Unix(template.EndAt, 0)
	)
	switch options.End.Type {
	case entity.RepeatEndNever:
		start, end := originStart, originEnd
		for end.Before(maxEndTime) {
			if start.After(originStart) {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, &item)
			}
			start = start.AddDate(0, 0, options.Interval)
			end = end.AddDate(0, 0, options.Interval)
		}
	case entity.RepeatEndAfterCount:
		var (
			start, end = originStart, originEnd
			count      = 0
		)
		for count < options.End.AfterCount && end.Before(maxEndTime) {
			if start.After(originStart) {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, &item)
				count++
			}
			start = start.AddDate(0, 0, options.Interval)
			end = end.AddDate(0, 0, options.Interval)
		}
	case entity.RepeatEndAfterTime:
		var (
			start, end = originStart, originEnd
			afterTime  = time.Unix(options.End.AfterTime, 0)
		)
		for end.Before(afterTime) && end.Before(maxEndTime) {
			if start.After(originStart) {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, &item)
			}
			start = start.AddDate(0, 0, options.Interval)
			end = end.AddDate(0, 0, options.Interval)
		}
	default:
		err := fmt.Errorf("repeat schedule: invalid daily end type")
		log.Error(ctx, err.Error(), log.String("end_type", string(options.End.Type)))
		return nil, entity.ErrInvalidArgs(err)
	}
	return result, nil
}

func (s *scheduleModel) repeatScheduleWeekly(ctx context.Context, template *entity.Schedule, options entity.RepeatWeekly) ([]*entity.Schedule, error) {
	if template == nil {
		err := fmt.Errorf("repeat schedule weekly: require not nil template")
		log.Error(ctx, err.Error())
		return nil, entity.ErrInvalidArgs(err)
	}
	if options.Interval <= 0 {
		log.Debug(ctx, "repeat schedule weekly: options interval less than 0", log.Int("interval", options.Interval))
		return nil, nil
	}
	var (
		result      []*entity.Schedule
		maxEndTime  = time.Now().AddDate(s.getMaxRepeatYear(), 0, 0)
		originStart = time.Unix(template.StartAt, 0)
		originEnd   = time.Unix(template.EndAt, 0)
	)
	switch options.End.Type {
	case entity.RepeatEndNever:
		for _, onWeekday := range options.On {
			var (
				start, end = originStart, originEnd
				first      = true
			)
			for end.Before(maxEndTime) {
				if start.After(originStart) && start.Weekday() == onWeekday.TimeWeekday() {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				var offset int
				if first {
					offset = int(onWeekday.TimeWeekday()-start.Weekday()+7) % 7
					first = false
				} else {
					offset = int(onWeekday.TimeWeekday()-start.Weekday()+7)%7 + 7*options.Interval
				}
				start = start.AddDate(0, 0, offset)
				end = end.AddDate(0, 0, offset)
			}
		}
	case entity.RepeatEndAfterCount:
		for _, onWeekday := range options.On {
			var (
				start, end = originStart, originEnd
				count      = 0
				first      = true
			)
			for count < options.End.AfterCount && end.Before(maxEndTime) {
				if start.After(originStart) && start.Weekday() == onWeekday.TimeWeekday() {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
					count++
				}
				var offset int
				if first {
					offset = int(onWeekday.TimeWeekday()-start.Weekday()+7) % 7
					first = false
				} else {
					offset = int(onWeekday.TimeWeekday()-start.Weekday()+7)%7 + 7*options.Interval
				}
				start = start.AddDate(0, 0, offset)
				end = end.AddDate(0, 0, offset)
			}
		}
	case entity.RepeatEndAfterTime:
		for _, onWeekday := range options.On {
			var (
				start, end = originStart, originEnd
				afterTime  = time.Unix(options.End.AfterTime, 0)
				first      = true
			)
			for end.Before(afterTime) && end.Before(maxEndTime) {
				if start.After(originStart) && start.Weekday() == onWeekday.TimeWeekday() {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				var offset int
				if first {
					offset = int(onWeekday.TimeWeekday()-start.Weekday()+7) % 7
					first = false
				} else {
					offset = int(onWeekday.TimeWeekday()-start.Weekday()+7)%7 + 7*options.Interval
				}
				start = start.AddDate(0, 0, offset)
				end = end.AddDate(0, 0, offset)
			}
		}
	default:
		err := fmt.Errorf("repeat schedule: invalid weekly end type")
		log.Error(ctx, err.Error(), log.String("end_type", string(options.End.Type)))
		return nil, entity.ErrInvalidArgs(err)
	}
	return result, nil
}

func (s *scheduleModel) repeatScheduleMonthly(ctx context.Context, template *entity.Schedule, options entity.RepeatMonthly) ([]*entity.Schedule, error) {
	if template == nil {
		err := fmt.Errorf("repeat schedule monthly: require not nil template")
		log.Error(ctx, err.Error())
		return nil, err
	}
	if options.Interval <= 0 {
		log.Debug(ctx, "repeat schedule monthly: options interval less than 0", log.Int("interval", options.Interval))
		return nil, nil
	}
	var (
		result      []*entity.Schedule
		maxEndTime  = time.Now().AddDate(s.getMaxRepeatYear(), 0, 0)
		originStart = time.Unix(template.StartAt, 0)
		originEnd   = time.Unix(template.EndAt, 0)
	)
	switch options.End.Type {
	case entity.RepeatEndNever:
		switch options.OnType {
		case entity.RepeatMonthlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfMonth(start.Year(), start.Month())
			)
			for {
				start = s.setTimeDatePart(start, timer.Year(), timer.Month(), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(0, options.Interval, 0)
			}
		case entity.RepeatMonthlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfMonth(start.Year(), start.Month())
			)
			for {
				year, month, day := s.dateOfWeekday(timer.Year(), timer.Month(), options.OnWeek, options.OnWeekSeq)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Weekday() == options.OnWeek.TimeWeekday() {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(0, options.Interval, 0)
			}
		}
	case entity.RepeatEndAfterCount:
		switch options.OnType {
		case entity.RepeatMonthlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfMonth(start.Year(), start.Month())
				count      = 0
			)
			for count < options.End.AfterCount {
				start = s.setTimeDatePart(start, timer.Year(), timer.Month(), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
					count++
				}
				timer = timer.AddDate(0, options.Interval, 0)
			}
		case entity.RepeatMonthlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfMonth(start.Year(), start.Month())
				count      = 0
			)
			for count < options.End.AfterCount {
				year, month, day := s.dateOfWeekday(timer.Year(), timer.Month(), options.OnWeek, options.OnWeekSeq)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Weekday() == options.OnWeek.TimeWeekday() {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
					count++
				}
				timer = timer.AddDate(0, options.Interval, 0)
			}
		}
	case entity.RepeatEndAfterTime:
		switch options.OnType {
		case entity.RepeatMonthlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfMonth(start.Year(), start.Month())
				afterTime  = time.Unix(options.End.AfterTime, 0)
			)
			for {
				start = s.setTimeDatePart(start, timer.Year(), timer.Month(), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(afterTime) || end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(0, options.Interval, 0)
			}
		case entity.RepeatMonthlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfMonth(start.Year(), start.Month())
				afterTime  = time.Unix(options.End.AfterTime, 0)
			)
			for {
				year, month, day := s.dateOfWeekday(timer.Year(), timer.Month(), options.OnWeek, options.OnWeekSeq)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(afterTime) || end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Weekday() == options.OnWeek.TimeWeekday() {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(0, options.Interval, 0)
			}
		}
	default:
		err := fmt.Errorf("repeat schedule: invalid monthly end type")
		log.Error(ctx, err.Error(), log.String("end_type", string(options.End.Type)))
		return nil, entity.ErrInvalidArgs(err)
	}
	return result, nil
}

func (s *scheduleModel) repeatScheduleYearly(ctx context.Context, template *entity.Schedule, options entity.RepeatYearly) ([]*entity.Schedule, error) {
	if template == nil {
		err := fmt.Errorf("repeat schedule yearly: require not nil template")
		log.Error(ctx, err.Error())
		return nil, entity.ErrInvalidArgs(err)
	}
	if options.Interval <= 0 {
		log.Debug(ctx, "repeat schedule yearly: options interval less than 0", log.Int("interval", options.Interval))
		return nil, nil
	}
	var (
		result      []*entity.Schedule
		maxEndTime  = time.Now().AddDate(s.getMaxRepeatYear(), 0, 0)
		originStart = time.Unix(template.StartAt, 0)
		originEnd   = time.Unix(template.EndAt, 0)
	)
	switch options.End.Type {
	case entity.RepeatEndNever:
		switch options.OnType {
		case entity.RepeatYearlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfYear(start.Year())
			)
			for {
				start = s.setTimeDatePart(start, timer.Year(), time.Month(options.OnDateMonth), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(options.Interval, 0, 0)
			}
		case entity.RepeatYearlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfYear(start.Year())
			)
			for {
				year, month, day := s.dateOfWeekday(timer.Year(), time.Month(options.OnWeekMonth), options.OnWeek, options.OnWeekSeq)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnWeekMonth) &&
					start.Weekday() == options.OnWeek.TimeWeekday() {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(options.Interval, 0, 0)
			}
		}
	case entity.RepeatEndAfterCount:
		switch options.OnType {
		case entity.RepeatYearlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfYear(start.Year())
				count      = 0
			)
			for count < options.End.AfterCount {
				start = s.setTimeDatePart(start, timer.Year(), time.Month(options.OnDateMonth), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
					count++
				}
				timer = timer.AddDate(options.Interval, 0, 0)
			}
		case entity.RepeatYearlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfYear(start.Year())
				count      = 0
			)
			for count < options.End.AfterCount {
				year, month, day := s.dateOfWeekday(timer.Year(), time.Month(options.OnWeekMonth), options.OnWeek, options.OnWeekSeq)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnWeekMonth) &&
					start.Weekday() == options.OnWeek.TimeWeekday() {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
					count++
				}
				timer = timer.AddDate(options.Interval, 0, 0)
			}
		}
	case entity.RepeatEndAfterTime:
		switch options.OnType {
		case entity.RepeatYearlyOnDate:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfYear(start.Year())
				afterTime  = time.Unix(options.End.AfterTime, 0)
			)
			for {
				start = s.setTimeDatePart(start, timer.Year(), time.Month(options.OnDateMonth), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(afterTime) || end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(options.Interval, 0, 0)
			}
		case entity.RepeatYearlyOnWeek:
			var (
				start, end = originStart, originEnd
				timer      = s.startOfYear(start.Year())
				afterTime  = time.Unix(options.End.AfterTime, 0)
			)
			for {
				year, month, day := s.dateOfWeekday(timer.Year(), time.Month(options.OnWeekMonth), options.OnWeek, options.OnWeekSeq)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(afterTime) || end.After(maxEndTime) {
					break
				}
				if start.After(originStart) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnWeekMonth) &&
					start.Weekday() == options.OnWeek.TimeWeekday() {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(options.Interval, 0, 0)
			}
		}
	default:
		err := fmt.Errorf("repeat schedule: invalid daily end type")
		log.Error(ctx, err.Error(), log.String("end_type", string(options.End.Type)))
		return nil, entity.ErrInvalidArgs(err)
	}
	return result, nil
}

// setTimeDatePart set time date part, include year, month and day
func (s *scheduleModel) setTimeDatePart(src time.Time, year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, src.Hour(), src.Minute(), src.Second(), src.Nanosecond(), src.Location())
}

func (s *scheduleModel) dateOfWeekday(year int, month time.Month, w entity.RepeatWeekday, seq entity.RepeatWeekSeq) (int, time.Month, int) {
	switch seq {
	case entity.RepeatWeekSeqFirst, entity.RepeatWeekSeqSecond, entity.RepeatWeekSeqThird, entity.RepeatWeekSeqFourth:
		start := s.startOfMonth(year, month)
		offset := int(w.TimeWeekday()-start.Weekday()+7)%7 + 7*(seq.Offset()-1)
		result := start.AddDate(0, 0, offset)
		return result.Year(), result.Month(), result.Day()
	case entity.RepeatWeekSeqLast:
		end := s.endOfMonth(year, month)
		offset := int(end.Weekday()-w.TimeWeekday()+7) % 7
		result := end.AddDate(0, 0, -offset)
		return result.Year(), result.Month(), result.Day()
	}
	return 0, 0, 0
}

func (s *scheduleModel) startOfYear(year int) time.Time {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
}

func (s *scheduleModel) startOfMonth(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
}

func (s *scheduleModel) endOfMonth(year int, month time.Month) time.Time {
	return time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.Local).Add(-time.Millisecond)
}

func (s *scheduleModel) startOfDay(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local).Add(-time.Millisecond)
}

func (s *scheduleModel) endOfDay(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 23, 59, 59, 999, time.Local).Add(-time.Millisecond)
}
