package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IOutcomeAttendanceDA interface {
	dbo.DataAccesser
	BatchGetByAssessmentIDAndOutcomeIDs(ctx context.Context, tx *dbo.DBContext, assessmentID string, outcomeIDs []string) ([]*entity.OutcomeAttendance, error)
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.OutcomeAttendance) error
	BatchDeleteByAssessmentIDAndOutcomeIDs(ctx context.Context, tx *dbo.DBContext, assessmentID string, outcomeIDs []string) error
	BatchGetByAssessmentIDs(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) ([]*entity.OutcomeAttendance, error)
	BatchGetByAssessmentIDsAndAttendanceID(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string, attendanceID string) ([]*entity.OutcomeAttendance, error)
}

var (
	outcomeAttendanceDAInstance     IOutcomeAttendanceDA
	outcomeAttendanceDAInstanceOnce = sync.Once{}
)

func GetOutcomeAttendanceDA() IOutcomeAttendanceDA {
	outcomeAttendanceDAInstanceOnce.Do(func() {
		outcomeAttendanceDAInstance = &outcomeAttendanceDA{}
	})
	return outcomeAttendanceDAInstance
}

type outcomeAttendanceDA struct {
	dbo.BaseDA
}

func (d *outcomeAttendanceDA) BatchGetByAssessmentIDAndOutcomeIDs(ctx context.Context, tx *dbo.DBContext, assessmentID string, outcomeIDs []string) ([]*entity.OutcomeAttendance, error) {
	var items []*entity.OutcomeAttendance
	if err := tx.
		Where("assessment_id = ?", assessmentID).
		Where("outcome_id in (?)", outcomeIDs).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (*outcomeAttendanceDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.OutcomeAttendance) error {
	if len(items) == 0 {
		return nil
	}
	columns := []string{"id", "assessment_id", "outcome_id", "attendance_id"}
	var matrix [][]interface{}
	for _, item := range items {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		matrix = append(matrix, []interface{}{item.ID, item.AssessmentID, item.OutcomeID, item.AttendanceID})
	}
	format, values := SQLBatchInsert(entity.OutcomeAttendance{}.TableName(), columns, matrix)
	if err := tx.Exec(format, values...).Error; err != nil {
		log.Error(ctx, "batch insert outcomes_attendances: batch insert failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

func (*outcomeAttendanceDA) BatchDeleteByAssessmentIDAndOutcomeIDs(ctx context.Context, tx *dbo.DBContext, assessmentID string, outcomeIDs []string) error {
	if len(outcomeIDs) == 0 {
		return nil
	}
	if err := tx.
		Where("assessment_id = ?", assessmentID).
		Where("outcome_id in (?)", outcomeIDs).
		Delete(entity.OutcomeAttendance{}).Error; err != nil {
		log.Error(ctx, "batch delete attendances by outcome ids: batch delete failed",
			log.Err(err),
			log.Strings("outcome_ids", outcomeIDs),
		)
		return err
	}
	return nil
}

func (d *outcomeAttendanceDA) BatchGetByAssessmentIDs(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) ([]*entity.OutcomeAttendance, error) {
	if len(assessmentIDs) == 0 {
		return nil, nil
	}
	var items []*entity.OutcomeAttendance
	if err := tx.
		Where("assessment_id in (?)", assessmentIDs).
		Find(&items).Error; err != nil {
		log.Error(ctx, "batch get outcome attendance by assessment ids: find failed",
			log.Err(err),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}
	return items, nil
}

func (d *outcomeAttendanceDA) BatchGetByAssessmentIDsAndAttendanceID(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string, attendanceID string) ([]*entity.OutcomeAttendance, error) {
	if len(assessmentIDs) == 0 || attendanceID == "" {
		return nil, nil
	}
	var items []*entity.OutcomeAttendance
	if err := tx.
		Where("assessment_id in (?) and attendance_id = ?", assessmentIDs, attendanceID).
		Find(&items).Error; err != nil {
		log.Error(ctx, "BatchGetByAssessmentIDsAndAttendanceID: call Find failed",
			log.Err(err),
			log.Strings("assessment_ids", assessmentIDs),
			log.String("attendance_id", attendanceID),
		)
		return nil, err
	}
	return items, nil
}

type QueryAssessmentOutcomeAttendanceCondition struct {
	AttendanceIDs                 entity.NullStrings
	AssessmentIDAndOutcomeIDPairs entity.NullAssessmentOutcomeKeys
}

func (c *QueryAssessmentOutcomeAttendanceCondition) GetConditions() ([]string, []interface{}) {
	t := NewSQLTemplate("")
	if c.AttendanceIDs.Valid {
		t.Appendf("attendance_id in (?)", c.AttendanceIDs.Strings)
	}
	if c.AssessmentIDAndOutcomeIDPairs.Valid {
		temp := NewSQLTemplate("")
		for _, pair := range c.AssessmentIDAndOutcomeIDPairs.Value {
			temp.Appendf("assessment_id = ? and outcome_id = ?", pair.AssessmentID, pair.OutcomeID)
		}
		t.AppendResult(temp.Or())
	}
	return t.DBOConditions()
}

func (c *QueryAssessmentOutcomeAttendanceCondition) GetPager() *dbo.Pager { return nil }

func (c *QueryAssessmentOutcomeAttendanceCondition) GetOrderBy() string { return "" }
