package da

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"strings"
)

type IReportDA interface {
	GetAchievedAttendanceID2OutcomeCountMap(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) (map[string]int, error)
	GetAchievedAttendanceID2OutcomeIDsMap(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) (map[string][]string, error)
}

func GetReportDA() IReportDA {
	return &reportDA{}
}

type reportDA struct{}

func (r *reportDA) GetAchievedAttendanceID2OutcomeCountMap(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) (map[string]int, error) {
	values := make([]string, len(scheduleIDs))
	for i, scheduleID := range scheduleIDs {
		values[i] = fmt.Sprintf("%q", scheduleID)
	}
	rows, err := tx.Raw(`select attendance_id, count(outcome_id) count from outcomes_attendances 
		where assessment_id in (
			select id from assessments where schedule_id in (?)
		) group by attendance_id`, strings.Join(values, ",")).Rows()
	if err != nil {
		log.Error(ctx, "get achieved count: call gorm raw failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
		)
		return nil, err
	}
	defer rows.Close()
	type Item struct {
		AttendanceID string `gorm:"column:attendance_id"`
		Count        int    `gorm:"column:count"`
	}
	var items []*Item
	if err = rows.Scan(&items); err != nil {
		log.Error(ctx, "get achieved count: rows scan failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
		)
		return nil, err
	}
	result := map[string]int{}
	for _, item := range items {
		result[item.AttendanceID] = item.Count
	}
	return result, nil
}

func (r *reportDA) GetAchievedAttendanceID2OutcomeIDsMap(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) (map[string][]string, error) {
	values := make([]string, len(scheduleIDs))
	for i, scheduleID := range scheduleIDs {
		values[i] = fmt.Sprintf("%q", scheduleID)
	}
	rows, err := tx.Raw(`select attendance_id, count(outcome_id) count from outcomes_attendances 
		where assessment_id in (
			select id from assessments where schedule_id in (?)
		) group by attendance_id`, strings.Join(values, ",")).Rows()
	if err != nil {
		log.Error(ctx, "get achieved count: call gorm raw failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
		)
		return nil, err
	}
	defer rows.Close()
	type Item struct {
		AttendanceID string `gorm:"column:attendance_id"`
		OutcomeID    string `gorm:"column:outcome_id"`
	}
	var items []*Item
	if err = rows.Scan(&items); err != nil {
		log.Error(ctx, "get achieved count: rows scan failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
		)
		return nil, err
	}
	result := map[string][]string{}
	for _, item := range items {
		result[item.AttendanceID] = append(result[item.AttendanceID], item.AttendanceID)
	}
	return result, nil
}
