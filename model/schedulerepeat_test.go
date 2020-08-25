package model

import (
	"context"
	"testing"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func getScheduleModelForRepeatTest() *scheduleModel {
	return &scheduleModel{
		testScheduleRepeatFlag: true,
	}
}

func Test_repeatScheduleDaily(t *testing.T) {
	const (
		top        = 5
		tail       = 5
		timeLayout = "2006-01-02T15:04:05Z07:00 Monday"
	)

	type args struct {
		ctx      context.Context
		template entity.Schedule
		options  entity.RepeatDaily
	}
	tests := []struct {
		name    string
		args    args
		want    []entity.Schedule
		wantErr bool
	}{
		{
			name: "daily end never",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatDaily{
					Interval: 1,
					End: entity.RepeatEnd{
						Type: entity.RepeatEndNever,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "daily end after count",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatDaily{
					Interval: 1,
					End: entity.RepeatEnd{
						Type:       entity.RepeatEndAfterCount,
						AfterCount: 3,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "daily end after time",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatDaily{
					Interval: 1,
					End: entity.RepeatEnd{
						Type:      entity.RepeatEndAfterTime,
						AfterTime: time.Date(2020, 10, 1, 0, 0, 0, 0, time.Local).Unix(),
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getScheduleModelForRepeatTest().repeatScheduleDaily(tt.args.ctx, &tt.args.template, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("repeatScheduleDaily() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			count := 0
			for index, item := range got {
				count++
				if index >= top && index < len(got)-tail {
					continue
				}
				t.Logf("%3d: %-35s -> %-35s",
					count,
					time.Unix(item.StartAt, 0).Format(timeLayout),
					time.Unix(item.EndAt, 0).Format(timeLayout),
				)
				if index == top-1 && len(got) > top-1 {
					t.Logf("...")
				}
			}
		})
	}
}

func Test_repeatScheduleWeekly(t *testing.T) {
	const (
		top        = 5
		tail       = 5
		timeLayout = "2006-01-02T15:04:05Z07:00 Monday"
	)

	type args struct {
		ctx      context.Context
		template entity.Schedule
		options  entity.RepeatWeekly
	}
	tests := []struct {
		name    string
		args    args
		want    []entity.Schedule
		wantErr bool
	}{
		{
			name: "weekly end never",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatWeekly{
					Interval: 1,
					On:       []entity.RepeatWeekday{entity.RepeatWeekdayMon, entity.RepeatWeekdayTue, entity.RepeatWeekdaySat},
					End: entity.RepeatEnd{
						Type: entity.RepeatEndNever,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "weekly bigger interval and end never",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatWeekly{
					Interval: 3,
					On:       []entity.RepeatWeekday{entity.RepeatWeekdayMon},
					End: entity.RepeatEnd{
						Type: entity.RepeatEndNever,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "weekly end after count",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatWeekly{
					Interval: 1,
					On:       []entity.RepeatWeekday{entity.RepeatWeekdayMon, entity.RepeatWeekdayFri, entity.RepeatWeekdaySun},
					End: entity.RepeatEnd{
						Type:       entity.RepeatEndAfterCount,
						AfterCount: 3,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "weekly end after time",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatWeekly{
					Interval: 1,
					On:       []entity.RepeatWeekday{entity.RepeatWeekdayMon},
					End: entity.RepeatEnd{
						Type:      entity.RepeatEndAfterTime,
						AfterTime: time.Date(2020, 10, 1, 0, 0, 0, 0, time.Local).Unix(),
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getScheduleModelForRepeatTest().repeatScheduleWeekly(tt.args.ctx, &tt.args.template, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("repeatScheduleWeekly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			count := 0
			for index, item := range got {
				count++
				if index >= top && index < len(got)-tail {
					continue
				}
				t.Logf("%3d: %-35s -> %-35s",
					count,
					time.Unix(item.StartAt, 0).Format(timeLayout),
					time.Unix(item.EndAt, 0).Format(timeLayout),
				)
				if index == top-1 && len(got) > top-1 {
					t.Logf("...")
				}
			}
		})
	}
}

func Test_repeatScheduleMonthly(t *testing.T) {
	const (
		top        = 5
		tail       = 5
		timeLayout = "2006-01-02T15:04:05Z07:00 Monday"
	)

	type args struct {
		ctx      context.Context
		template entity.Schedule
		options  entity.RepeatMonthly
	}
	tests := []struct {
		name    string
		args    args
		want    []entity.Schedule
		wantErr bool
	}{
		{
			name: "monthly end never on date",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatMonthly{
					Interval:  1,
					OnType:    entity.RepeatMonthlyOnDate,
					OnDateDay: 1,
					End: entity.RepeatEnd{
						Type: entity.RepeatEndNever,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "monthly end never on week",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatMonthly{
					Interval:  1,
					OnType:    entity.RepeatMonthlyOnWeek,
					OnWeek:    entity.RepeatWeekdayMon,
					OnWeekSeq: entity.RepeatWeekSeqFirst,
					End: entity.RepeatEnd{
						Type: entity.RepeatEndNever,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "monthly end after count on date",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatMonthly{
					Interval:  1,
					OnType:    entity.RepeatMonthlyOnDate,
					OnDateDay: 1,
					End: entity.RepeatEnd{
						Type:       entity.RepeatEndAfterCount,
						AfterCount: 10,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "monthly end after count on week",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatMonthly{
					Interval:  1,
					OnType:    entity.RepeatMonthlyOnWeek,
					OnWeek:    entity.RepeatWeekdayMon,
					OnWeekSeq: entity.RepeatWeekSeqFourth,
					End: entity.RepeatEnd{
						Type:       entity.RepeatEndAfterCount,
						AfterCount: 10,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "monthly end after time on date",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatMonthly{
					Interval:  1,
					OnType:    entity.RepeatMonthlyOnDate,
					OnDateDay: 1,
					End: entity.RepeatEnd{
						Type:      entity.RepeatEndAfterTime,
						AfterTime: time.Date(2020, 12, 1, 0, 0, 0, 0, time.Local).Unix(),
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "monthly end after time on week",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatMonthly{
					Interval:  1,
					OnType:    entity.RepeatMonthlyOnWeek,
					OnWeek:    entity.RepeatWeekdayMon,
					OnWeekSeq: entity.RepeatWeekSeqLast,
					End: entity.RepeatEnd{
						Type:      entity.RepeatEndAfterTime,
						AfterTime: time.Date(2020, 12, 1, 0, 0, 0, 0, time.Local).Unix(),
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getScheduleModelForRepeatTest().repeatScheduleMonthly(tt.args.ctx, &tt.args.template, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("repeatScheduleMonthly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			count := 0
			for index, item := range got {
				count++
				if index >= top && index < len(got)-tail {
					continue
				}
				t.Logf("%3d: %-35s -> %-35s",
					count,
					time.Unix(item.StartAt, 0).Format(timeLayout),
					time.Unix(item.EndAt, 0).Format(timeLayout),
				)
				if index == top-1 && len(got) > top-1 {
					t.Logf("...")
				}
			}
		})
	}
}

func Test_repeatScheduleYearly(t *testing.T) {
	const (
		top        = 5
		tail       = 5
		timeLayout = "2006-01-02T15:04:05Z07:00 Monday"
	)

	type args struct {
		ctx      context.Context
		template entity.Schedule
		options  entity.RepeatYearly
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "yearly end never on date",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 2, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatYearly{
					Interval:    1,
					OnType:      entity.RepeatYearlyOnDate,
					OnDateMonth: 11,
					OnDateDay:   1,
					End: entity.RepeatEnd{
						Type: entity.RepeatEndNever,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "yearly end never on week",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 2, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatYearly{
					Interval:    1,
					OnType:      entity.RepeatYearlyOnWeek,
					OnWeekMonth: 11,
					OnWeekSeq:   entity.RepeatWeekSeqFirst,
					OnWeek:      entity.RepeatWeekdayMon,
					End: entity.RepeatEnd{
						Type: entity.RepeatEndNever,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "yearly end after count on date",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatYearly{
					Interval:    1,
					OnType:      entity.RepeatYearlyOnDate,
					OnDateMonth: 11,
					OnDateDay:   1,
					End: entity.RepeatEnd{
						Type:       entity.RepeatEndAfterCount,
						AfterCount: 10,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "yearly end after count on week",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatYearly{
					Interval:    1,
					OnType:      entity.RepeatYearlyOnWeek,
					OnWeekMonth: 11,
					OnWeekSeq:   entity.RepeatWeekSeqFirst,
					OnWeek:      entity.RepeatWeekdayMon,
					End: entity.RepeatEnd{
						Type:       entity.RepeatEndAfterCount,
						AfterCount: 10,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "yearly end never after time on date",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatYearly{
					Interval:    1,
					OnType:      entity.RepeatYearlyOnDate,
					OnDateMonth: 11,
					OnDateDay:   1,
					End: entity.RepeatEnd{
						Type:      entity.RepeatEndAfterTime,
						AfterTime: time.Date(2021, 12, 1, 0, 0, 0, 0, time.Local).Unix(),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "yearly end never after time on week",
			args: args{
				ctx: context.Background(),
				template: entity.Schedule{
					StartAt: time.Date(2020, 9, 1, 9, 0, 0, 0, time.Local).Unix(),
					EndAt:   time.Date(2020, 9, 1, 10, 0, 0, 0, time.Local).Unix(),
				},
				options: entity.RepeatYearly{
					Interval:    1,
					OnType:      entity.RepeatYearlyOnWeek,
					OnWeekMonth: 11,
					OnWeekSeq:   entity.RepeatWeekSeqFirst,
					OnWeek:      entity.RepeatWeekdayMon,
					End: entity.RepeatEnd{
						Type:      entity.RepeatEndAfterTime,
						AfterTime: time.Date(2021, 12, 1, 0, 0, 0, 0, time.Local).Unix(),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getScheduleModelForRepeatTest().repeatScheduleYearly(tt.args.ctx, &tt.args.template, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("repeatScheduleYearly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			count := 0
			for index, item := range got {
				count++
				if index >= top && index < len(got)-tail {
					continue
				}
				t.Logf("%3d: %-35s -> %-35s",
					count,
					time.Unix(item.StartAt, 0).Format(timeLayout),
					time.Unix(item.EndAt, 0).Format(timeLayout),
				)
				if index == top-1 && len(got) > top-1 {
					t.Logf("...")
				}
			}
		})

	}
}
