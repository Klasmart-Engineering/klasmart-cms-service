package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"time"
)

func (s *scheduleModel) RepeatSchedule(ctx context.Context, template *entity.Schedule, options *entity.RepeatOptions, location *time.Location) ([]*entity.Schedule, error) {
	if options == nil || !options.Type.Valid() {
		return []*entity.Schedule{template}, nil
	}
	result := []*entity.Schedule{template}
	items, err := s.repeatSchedule(ctx, template, options, location)
	if err != nil {
		log.Error(ctx, "repeat schedule(include template): call repeat schedule failed",
			log.Err(err),
			log.Any("template", template),
		)
		return nil, err
	}
	result = append(result, items...)
	return result, nil
}

func (s *scheduleModel) getMaxRepeatYear() int {
	if s.testScheduleRepeatFlag {
		return 2
	}
	return config.Get().Schedule.MaxRepeatYear
}

func (s *scheduleModel) repeatSchedule(ctx context.Context, template *entity.Schedule, options *entity.RepeatOptions, location *time.Location) ([]*entity.Schedule, error) {
	var (
		result []*entity.Schedule
		items  []*entity.Schedule
		err    error
	)
	switch options.Type {
	case entity.RepeatTypeDaily:
		items, err = s.repeatScheduleDaily(ctx, template, &options.Daily, location)
	case entity.RepeatTypeWeekly:
		items, err = s.repeatScheduleWeekly(ctx, template, &options.Weekly, location)
	case entity.RepeatTypeMonthly:
		items, err = s.repeatScheduleMonthly(ctx, template, &options.Monthly, location)
	case entity.RepeatTypeYearly:
		items, err = s.repeatScheduleYearly(ctx, template, &options.Yearly, location)
	}
	if err != nil {
		log.Error(ctx, "repeat schedule failed",
			log.Err(err),
			log.Any("template", template),
			log.Any("options", options),
		)
		return nil, err
	}
	result = append(result, items...)
	return result, nil
}

func (s *scheduleModel) repeatScheduleDaily(ctx context.Context, template *entity.Schedule, options *entity.RepeatDaily, location *time.Location) ([]*entity.Schedule, error) {
	var (
		result       []*entity.Schedule
		originStart  = time.Unix(template.StartAt, 0).In(location)
		originEnd    = time.Unix(template.EndAt, 0).In(location)
		minStartTime = s.nextWeekStart(originStart)
		maxEndTime   = time.Now().AddDate(s.getMaxRepeatYear(), 0, 0).In(location)
	)
	switch options.End.Type {
	case entity.RepeatEndNever:
		start, end := originStart, originEnd
		for end.Before(maxEndTime) {
			if start.After(minStartTime) {
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
			if start.After(minStartTime) {
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
			afterTime  = time.Unix(options.End.AfterTime, 0).In(location)
		)
		for end.Before(afterTime) && end.Before(maxEndTime) {
			if start.After(minStartTime) {
				item := template.Clone()
				item.StartAt, item.EndAt = start.Unix(), end.Unix()
				result = append(result, &item)
			}
			start = start.AddDate(0, 0, options.Interval)
			end = end.AddDate(0, 0, options.Interval)
		}
	default:
		log.Info(ctx, "repeat schedule: invalid daily end type", log.String("end_type", string(options.End.Type)))
		return nil, constant.ErrInvalidArgs
	}
	return result, nil
}

func (s *scheduleModel) repeatScheduleWeekly(ctx context.Context, template *entity.Schedule, options *entity.RepeatWeekly, location *time.Location) ([]*entity.Schedule, error) {
	if options.Interval <= 0 {
		log.Info(ctx, "repeat schedule weekly: options interval less than 0", log.Int("interval", options.Interval))
		return nil, nil
	}
	var (
		result       []*entity.Schedule
		originStart  = time.Unix(template.StartAt, 0).In(location)
		originEnd    = time.Unix(template.EndAt, 0).In(location)
		minStartTime = s.nextWeekStart(originStart)
		maxEndTime   = time.Now().AddDate(s.getMaxRepeatYear(), 0, 0).In(location)
	)
	switch options.End.Type {
	case entity.RepeatEndNever:
		for _, onWeekday := range options.On {
			var (
				start, end = originStart, originEnd
				first      = true
			)
			for end.Before(maxEndTime) {
				if start.After(minStartTime) && start.Weekday() == onWeekday.TimeWeekday() {
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
				first      = true
				count      = 0
			)
			for count < options.End.AfterCount && end.Before(maxEndTime) {
				if start.After(minStartTime) && start.Weekday() == onWeekday.TimeWeekday() {
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
				afterTime  = time.Unix(options.End.AfterTime, 0).In(location)
				first      = true
			)
			for end.Before(afterTime) && end.Before(maxEndTime) {
				if start.After(minStartTime) && start.Weekday() == onWeekday.TimeWeekday() {
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
		log.Info(ctx, "repeat schedule: invalid weekly end type", log.String("end_type", string(options.End.Type)))
		return nil, constant.ErrInvalidArgs
	}
	return result, nil
}

func (s *scheduleModel) repeatScheduleMonthly(ctx context.Context, template *entity.Schedule, options *entity.RepeatMonthly, location *time.Location) ([]*entity.Schedule, error) {
	if options.Interval <= 0 {
		log.Info(ctx, "repeat schedule monthly: options interval less than 0", log.Int("interval", options.Interval))
		return nil, nil
	}
	var (
		result       []*entity.Schedule
		originStart  = time.Unix(template.StartAt, 0).In(location)
		originEnd    = time.Unix(template.EndAt, 0).In(location)
		minStartTime = s.nextMonthStart(originStart)
		maxEndTime   = time.Now().AddDate(s.getMaxRepeatYear(), 0, 0).In(location)
	)
	switch options.End.Type {
	case entity.RepeatEndNever:
		var (
			start, end = originStart, originEnd
			timer      = s.startOfMonth(start.Year(), start.Month(), location)
		)
		switch options.OnType {
		case entity.RepeatMonthlyOnDate:
			for {
				start = s.setTimeDatePart(start, timer.Year(), timer.Month(), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(0, options.Interval, 0)
			}
		case entity.RepeatMonthlyOnWeek:
			for {
				year, month, day := s.dateOfWeekday(timer.Year(), timer.Month(), options.OnWeek, options.OnWeekSeq, location)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
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
		var (
			start, end = originStart, originEnd
			timer      = s.startOfMonth(start.Year(), start.Month(), location)
		)
		switch options.OnType {
		case entity.RepeatMonthlyOnDate:
			for len(result) < options.End.AfterCount {
				start = s.setTimeDatePart(start, timer.Year(), timer.Month(), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(0, options.Interval, 0)
			}
		case entity.RepeatMonthlyOnWeek:
			for len(result) < options.End.AfterCount {
				year, month, day := s.dateOfWeekday(timer.Year(), timer.Month(), options.OnWeek, options.OnWeekSeq, location)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Weekday() == options.OnWeek.TimeWeekday() {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(0, options.Interval, 0)
			}
		}
	case entity.RepeatEndAfterTime:
		var (
			start, end = originStart, originEnd
			timer      = s.startOfMonth(start.Year(), start.Month(), location)
			afterTime  = time.Unix(options.End.AfterTime, 0).In(location)
		)
		switch options.OnType {
		case entity.RepeatMonthlyOnDate:
			for {
				start = s.setTimeDatePart(start, timer.Year(), timer.Month(), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(afterTime) || end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
					start.Year() == timer.Year() && start.Month() == timer.Month() &&
					start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(0, options.Interval, 0)
			}
		case entity.RepeatMonthlyOnWeek:
			for {
				year, month, day := s.dateOfWeekday(timer.Year(), timer.Month(), options.OnWeek, options.OnWeekSeq, location)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(afterTime) || end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
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
		log.Info(ctx, "repeat schedule: invalid monthly end type", log.String("end_type", string(options.End.Type)))
		return nil, constant.ErrInvalidArgs
	}
	return result, nil
}

func (s *scheduleModel) repeatScheduleYearly(ctx context.Context, template *entity.Schedule, options *entity.RepeatYearly, location *time.Location) ([]*entity.Schedule, error) {
	if options.Interval <= 0 {
		log.Info(ctx, "repeat schedule yearly: options interval less than 0", log.Int("interval", options.Interval))
		return nil, nil
	}
	var (
		result       []*entity.Schedule
		originStart  = time.Unix(template.StartAt, 0).In(location)
		originEnd    = time.Unix(template.EndAt, 0).In(location)
		minStartTime = s.nextYearStart(originStart)
		maxEndTime   = time.Now().AddDate(s.getMaxRepeatYear(), 0, 0).In(location)
	)
	switch options.End.Type {
	case entity.RepeatEndNever:
		var (
			start, end = originStart, originEnd
			timer      = s.startOfYear(start.Year(), location)
		)
		switch options.OnType {
		case entity.RepeatYearlyOnDate:
			for {
				start = s.setTimeDatePart(start, timer.Year(), time.Month(options.OnDateMonth), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(options.Interval, 0, 0)
			}
		case entity.RepeatYearlyOnWeek:
			for {
				year, month, day := s.dateOfWeekday(timer.Year(), time.Month(options.OnWeekMonth), options.OnWeek, options.OnWeekSeq, location)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
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
		var (
			start, end = originStart, originEnd
			timer      = s.startOfYear(start.Year(), location)
		)
		switch options.OnType {
		case entity.RepeatYearlyOnDate:
			for len(result) < options.End.AfterCount {
				start = s.setTimeDatePart(start, timer.Year(), time.Month(options.OnDateMonth), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(options.Interval, 0, 0)
			}
		case entity.RepeatYearlyOnWeek:
			for len(result) < options.End.AfterCount {
				year, month, day := s.dateOfWeekday(timer.Year(), time.Month(options.OnWeekMonth), options.OnWeek, options.OnWeekSeq, location)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
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
	case entity.RepeatEndAfterTime:
		var (
			start, end = originStart, originEnd
			timer      = s.startOfYear(start.Year(), location)
			afterTime  = time.Unix(options.End.AfterTime, 0).In(location)
		)
		switch options.OnType {
		case entity.RepeatYearlyOnDate:
			for {
				start = s.setTimeDatePart(start, timer.Year(), time.Month(options.OnDateMonth), options.OnDateDay)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(afterTime) || end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
					start.Year() == timer.Year() &&
					start.Month() == time.Month(options.OnDateMonth) && start.Day() == options.OnDateDay {
					item := template.Clone()
					item.StartAt, item.EndAt = start.Unix(), end.Unix()
					result = append(result, &item)
				}
				timer = timer.AddDate(options.Interval, 0, 0)
			}
		case entity.RepeatYearlyOnWeek:
			for {
				year, month, day := s.dateOfWeekday(timer.Year(), time.Month(options.OnWeekMonth), options.OnWeek, options.OnWeekSeq, location)
				start = s.setTimeDatePart(start, year, month, day)
				end = originEnd.Add(start.Sub(originStart))
				if end.After(afterTime) || end.After(maxEndTime) {
					break
				}
				if start.After(minStartTime) &&
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
		log.Info(ctx, "repeat schedule: invalid daily end type", log.String("end_type", string(options.End.Type)))
		return nil, constant.ErrInvalidArgs
	}
	return result, nil
}

// setTimeDatePart set time date part, include year, month and day
func (s *scheduleModel) setTimeDatePart(src time.Time, year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, src.Hour(), src.Minute(), src.Second(), src.Nanosecond(), src.Location())
}

func (s *scheduleModel) dateOfWeekday(year int, month time.Month, w entity.RepeatWeekday, seq entity.RepeatWeekSeq, location *time.Location) (int, time.Month, int) {
	switch seq {
	case entity.RepeatWeekSeqFirst, entity.RepeatWeekSeqSecond, entity.RepeatWeekSeqThird, entity.RepeatWeekSeqFourth:
		start := s.startOfMonth(year, month, location)
		offset := int(w.TimeWeekday()-start.Weekday()+7)%7 + 7*(seq.Offset()-1)
		result := start.AddDate(0, 0, offset)
		return result.Year(), result.Month(), result.Day()
	case entity.RepeatWeekSeqLast:
		end := s.endOfMonth(year, month, location)
		offset := int(end.Weekday()-w.TimeWeekday()+7) % 7
		result := end.AddDate(0, 0, -offset)
		return result.Year(), result.Month(), result.Day()
	}
	return 0, 0, 0
}

func (s *scheduleModel) startOfYear(year int, location *time.Location) time.Time {
	return time.Date(year, 1, 1, 0, 0, 0, 0, location)
}

func (s *scheduleModel) startOfMonth(year int, month time.Month, location *time.Location) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, location)
}

func (s *scheduleModel) endOfMonth(year int, month time.Month, location *time.Location) time.Time {
	return time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, location).Add(-time.Millisecond)
}

func (s *scheduleModel) startOfDay(year int, month time.Month, day int, location *time.Location) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func (s *scheduleModel) endOfDay(year int, month time.Month, day int, location *time.Location) time.Time {
	return time.Date(year, month, day, 23, 59, 59, 999999999, location)
}

func (s *scheduleModel) nextDayStart(t time.Time) time.Time {
	newTime := t.AddDate(0, 0, 1)
	return time.Date(newTime.Year(), newTime.Month(), newTime.Day(), 0, 0, 0, 0, t.Location())
}

func (s *scheduleModel) nextWeekStart(t time.Time) time.Time {
	var newTime time.Time
	if t.Weekday() == time.Monday {
		newTime = t.AddDate(0, 0, 7)
	} else {
		newTime = t.AddDate(0, 0, int(time.Monday-t.Weekday()+7)%7)
	}
	return time.Date(newTime.Year(), newTime.Month(), newTime.Day(), 0, 0, 0, 0, t.Location())
}

func (s *scheduleModel) nextMonthStart(t time.Time) time.Time {
	newTime := t.AddDate(0, 1, 0)
	return time.Date(newTime.Year(), newTime.Month(), newTime.Day(), 0, 0, 0, 0, t.Location())
}

func (s *scheduleModel) nextYearStart(t time.Time) time.Time {
	newTime := t.AddDate(1, 0, 0)
	return time.Date(newTime.Year(), newTime.Month(), newTime.Day(), 0, 0, 0, 0, t.Location())
}
