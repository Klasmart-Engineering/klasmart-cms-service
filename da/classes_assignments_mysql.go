package da

import (
	"context"
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ClassesAssignmentsSQLDA struct {
	BaseDA
}

func (c ClassesAssignmentsSQLDA) BatchInsertTx(ctx context.Context, tx *dbo.DBContext, records []*entity.ClassesAssignmentsRecords) error {
	var models []entity.BatchInsertModeler
	for _, record := range records {
		models = append(models, record)
	}
	err := c.BaseDA.BatchInsertTx(ctx, tx, models...)
	if err != nil {
		log.Error(ctx, "BatchInsertTx: classes_assignments failed",
			log.Any("record", records))
		return err
	}
	return nil
}

func (c ClassesAssignmentsSQLDA) BatchUpdateFinishAndEnd(ctx context.Context, tx *dbo.DBContext, scheduleID string, attendedIDs []string, endTime int64) error {
	if len(attendedIDs) > 0 {
		sql := fmt.Sprintf("update %s set finish_counts=finish_counts+1, last_end_at=? where schedule_id=? and attendance_id in (?)", entity.ClassesAssignmentsRecords{}.TableName())
		err := tx.Exec(sql, endTime, scheduleID, attendedIDs).Error
		if err != nil {
			log.Error(ctx, "BatchUpdateFinish: update failed",
				log.String("sql", sql),
				log.String("schedule_id", scheduleID),
				log.Strings("attend_ids", attendedIDs))
		}
		return err
	}
	return nil
}

func (c ClassesAssignmentsSQLDA) QueryTx(ctx context.Context, tx *dbo.DBContext, condition *ClassesAssignmentsCondition) ([]*entity.ClassesAssignmentsRecords, error) {
	var records []*entity.ClassesAssignmentsRecords
	err := c.BaseDA.QueryTx(ctx, tx, condition, &records)
	if err != nil {
		log.Error(ctx, "QueryTx: classes_assignments failed",
			log.Any("condition", condition))
		return nil, err
	}
	return records, nil
}

func (ClassesAssignmentsSQLDA) CountAttendedTx(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string, attendanceType string) (map[string]int, error) {
	panic("implement me")
}

type ClassesAssignmentsCondition struct {
	ClassID         sql.NullString     `json:"class_id"`
	ClassIDs        entity.NullStrings `json:"class_ids"`
	ScheduleID      sql.NullString     `json:"schedule_id"`
	ScheduleIDs     entity.NullStrings `json:"schedule_ids"`
	AttendanceID    sql.NullString     `json:"attendance_id"`
	AttendanceIDs   entity.NullStrings `json:"attendance_ids"`
	ScheduleType    sql.NullString     `json:"schedule_type"`
	ScheduleTypes   entity.NullStrings `json:"schedule_types"`
	FinishCountsLTE sql.NullInt64      `json:"FinishCountsLTE"`

	OrderBy string    `json:"order_by"`
	Pager   dbo.Pager `json:"pager"`
}

func (c *ClassesAssignmentsCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	if c.ClassID.Valid {
		wheres = append(wheres, "class_id = ?")
		params = append(params, c.ClassID.String)
	}

	if c.ClassIDs.Valid {
		wheres = append(wheres, "class_id in (?)")
		params = append(params, c.ClassIDs.Strings)
	}

	if c.ScheduleID.Valid {
		wheres = append(wheres, "schedule_id = ?")
		params = append(params, c.ScheduleID.String)
	}

	if c.ScheduleIDs.Valid {
		wheres = append(wheres, "schedule_id in (?)")
		params = append(params, c.ScheduleIDs.Strings)
	}

	if c.FinishCountsLTE.Valid {
		wheres = append(wheres, " finish_counts <= ?")
		params = append(params, c.FinishCountsLTE.Int64)
	}

	return wheres, params
}

func (c *ClassesAssignmentsCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *ClassesAssignmentsCondition) GetOrderBy() string {
	return c.OrderBy
}
